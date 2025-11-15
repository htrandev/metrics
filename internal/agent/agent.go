package agent

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/pkg/sign"
)

type Collector interface {
	Collect() []model.Metric
	CollectGopsutil() ([]model.Metric, error)
}

var defaultOpts = &AgentOptions{
	Addr:           "localhost:8080",
	Key:            "",
	MaxRetry:       3,
	RateLimit:      3,
	PollInterval:   2 * time.Second,
	ReportInterval: 10 * time.Second,
	Client:         resty.New(),
	Logger:         zap.NewNop(),
}

type AgentOptions struct {
	Addr      string
	Key       string
	MaxRetry  int
	RateLimit int

	PollInterval   time.Duration
	ReportInterval time.Duration

	Client    *resty.Client
	Logger    *zap.Logger
	Collector Collector
}

func validateOptions(opts *AgentOptions) *AgentOptions {
	if opts == nil {
		opts = defaultOpts
	}

	if opts.MaxRetry <= 0 {
		opts.MaxRetry = 3
	}

	if opts.RateLimit <= 0 {
		opts.MaxRetry = 3
	}

	if opts.Addr == "" {
		opts.Addr = "localhost:8080"
	}

	if opts.ReportInterval == 0 {
		opts.ReportInterval = 10 * time.Second
	}
	if opts.PollInterval == 0 {
		opts.PollInterval = 2 * time.Second
	}

	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	return opts
}

type Agent struct {
	opts *AgentOptions
}

func New(opts *AgentOptions) *Agent {
	return &Agent{opts: validateOptions(opts)}
}

func (a *Agent) Run(ctx context.Context) {
	a.opts.Logger.Info("running agent")
	var wg sync.WaitGroup

	for i := 0; i < a.opts.RateLimit; i++ {
		wg.Add(1)

		collectChan := make(chan []model.Metric)
		a.Collect(ctx, collectChan)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case metrics := <-collectChan:
					a.opts.Logger.Info("send metrics with retry")

					start := time.Now()
					err := a.SendManyWithRetry(ctx, metrics)
					elapsed := time.Since(start)

					if err != nil {
						a.opts.Logger.Error("can't send many metric", zap.Error(err))
						continue
					}

					a.opts.Logger.Debug("successfully send metrics",
						zap.Int("batch size", len(metrics)),
						zap.String("elapsed", elapsed.String()),
					)
				}
			}
		}()
	}

	wg.Wait()
	a.opts.Logger.Info("finish running agent")
}

func (a *Agent) Collect(ctx context.Context, collectChan chan<- []model.Metric) {
	go func() {
		poller := a.poller(ctx)

		reportTicker := time.NewTicker(a.opts.ReportInterval)
		defer reportTicker.Stop()

		select {
		case <-ctx.Done():
			return
		case <-reportTicker.C:
		}

		metrics := <-poller
		collectChan <- metrics
	}()
}

func (a *Agent) poller(ctx context.Context) <-chan []model.Metric {
	pollChan := make(chan []model.Metric)

	go func() {
		defer close(pollChan)

		pollTicker := time.NewTicker(a.opts.PollInterval)
		defer pollTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-pollTicker.C:
			}
			// собираем метрики
			a.opts.Logger.Info("collect metrics")
			metrics := a.opts.Collector.Collect()
			a.opts.Logger.Info("collect gopsutil metrics")
			gopsUtilMetrics, err := a.opts.Collector.CollectGopsutil()
			if err != nil {
				a.opts.Logger.Error("failde to collect gopsutil metrics", zap.Error(err))
			} else {
				metrics = append(metrics, gopsUtilMetrics...)
			}

			// пытаемся отправить в канал,
			// если канал полон, то продолжаем собирать метрики
			select {
			case pollChan <- metrics:
			default:
			}
		}
	}()

	return pollChan
}

func (a *Agent) SendSingleMetric(ctx context.Context, metric model.Metric) error {
	req := buildSingleRequest(metric)
	body, err := buildSingleBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}
	url := buildSingleURL(a.opts.Addr)

	_, err = a.opts.Client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		SetContext(ctx).
		Post(url)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	return nil
}

func (a *Agent) SendManyMetrics(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	req := buildManyRequest(metrics)
	body, err := buildManyBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}

	url := buildManyURL(a.opts.Addr)

	r := a.opts.Client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		SetContext(ctx)

	if a.opts.Key != "" {
		s := sign.Signature(a.opts.Key)
		signature := s.Sign(body)
		hash := base64.RawURLEncoding.EncodeToString(signature)
		r.SetHeader("HashSHA256", hash)
	}

	_, err = r.Post(url)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	return nil
}

func (a *Agent) SendManyWithRetry(ctx context.Context, metrics []model.Metric) error {
	err := a.SendManyMetrics(ctx, metrics)

	if err != nil {
		a.opts.Logger.Error("send many metrics", zap.Error(err), zap.String("scope", "agent/sendWithRetry"))
		a.opts.Logger.Info("try to resend metrics", zap.String("scope", "agent/sendWithRetry"))

		for i := 0; i < a.opts.MaxRetry; i++ {
			a.opts.Logger.Debug("", zap.Int("retry", i+1))
			retryDelay := i*2 + 1
			time.Sleep(time.Second * time.Duration(retryDelay))
			err := a.SendManyMetrics(ctx, metrics)
			if err != nil {
				a.opts.Logger.Error("send many metrics", zap.Int("retry", i+1), zap.Error(err), zap.String("scope", "agent/sendWithRetry"))
				continue
			}
			return nil
		}
		return fmt.Errorf("send many with retries: %w", err)
	}
	return nil
}
