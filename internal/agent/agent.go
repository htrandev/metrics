package agent

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/handler/middleware"
	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/pkg/netutil"
	"github.com/htrandev/metrics/pkg/sign"
)

// Collector предоставляет интерфейс взаимодействия со сборщик метрик.
type Collector interface {
	Collect() []model.Metric
	CollectGopsutil() ([]model.Metric, error)
}

// defaultOpts определяет параметры для агента по умолчанию.
func defaultOpts() *AgentOptions {
	return &AgentOptions{
		Addr:           "localhost:8080",
		Signature:      "",
		MaxRetry:       3,
		RateLimit:      3,
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Client:         resty.New(),
		Logger:         zap.NewNop(),
	}
}

// AgentOptions определяет параметры для Агента.
type AgentOptions struct {
	Addr      string
	Signature string
	MaxRetry  int
	RateLimit int
	Ip        string

	PollInterval   time.Duration
	ReportInterval time.Duration

	Key       *rsa.PublicKey
	Client    *resty.Client
	Logger    *zap.Logger
	Collector Collector
}

// validateOptions валидирует параметры агента и подставляет значения по умолчанию,
// если параметр не был предоставлен.
func validateOptions(opts *AgentOptions) *AgentOptions {
	if opts == nil {
		opts = defaultOpts()
	}

	if opts.MaxRetry <= 0 {
		opts.MaxRetry = 3
	}

	if opts.RateLimit <= 0 {
		opts.RateLimit = 3
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

	if opts.Client == nil {
		opts.Client = resty.New()
	}

	ip, err := netutil.GetLocalIp()
	if err != nil {
		opts.Logger.Error("get local ip",
			zap.String("method", "SendManyMetrics"),
			zap.Error(err),
		)
	}
	opts.Ip = ip.String()
	return opts
}

// Agent определяет агента для сбора метрик и отправки их на сервер.
type Agent struct {
	opts *AgentOptions
}

// New возвращает новый экземпляр агента.
func New(opts *AgentOptions) *Agent {
	return &Agent{opts: validateOptions(opts)}
}

// Run собирает метрики и отправляет их на сервер.
func (a *Agent) Run(ctx context.Context) {
	a.opts.Logger.Info("running agent")
	var wg sync.WaitGroup

	collectChan := make(chan []model.Metric)
	a.Collect(ctx, collectChan)

	for i := 0; i < a.opts.RateLimit; i++ {
		wg.Add(1)

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

// Collect собирает метрики и передаёт их в полученный канал(collectChan).
func (a *Agent) Collect(ctx context.Context, collectChan chan<- []model.Metric) {
	go func() {
		poller := a.poller(ctx)

		reportTicker := time.NewTicker(a.opts.ReportInterval)
		defer reportTicker.Stop()
		for {

			select {
			case <-ctx.Done():
				return
			case <-reportTicker.C:
			}

			metrics := <-poller
			collectChan <- metrics
		}
	}()
}

// poller возвращает канал, в который отправляются полученные метрики.
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

// SendSingleMetric отправляет за раз одну метрику на сервер.
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

// SendManyMetrics отправляет за раз несколько метрик на сервер.
func (a *Agent) SendManyMetrics(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	req := buildManyRequest(metrics)
	body, err := a.buildManyBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}

	url := buildManyURL(a.opts.Addr)

	r := a.opts.Client.R().
		SetHeader(middleware.IpHeader, a.opts.Ip).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		SetContext(ctx)

	if a.opts.Signature != "" {
		s := sign.Signature(a.opts.Signature)
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

// SendManyWithRetry отправляет за раз несколько метрик на сервер,
// если произошла ошибка, пытается повторно отправить указанное в MaxRetry количество раз.
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
