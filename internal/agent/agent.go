package agent

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
)

type Client interface {
	Send(ctx context.Context, metrics []model.MetricDto) error
}

// Collector предоставляет интерфейс взаимодействия со сборщик метрик.
type Collector interface {
	Collect() []model.MetricDto
	CollectGopsutil() ([]model.MetricDto, error)
}

// defaultOpts определяет параметры для агента по умолчанию.
func defaultOpts() *AgentOptions {
	return &AgentOptions{
		RateLimit:      3,
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Logger:         zap.NewNop(),
	}
}

// AgentOptions определяет параметры для Агента.
type AgentOptions struct {
	RateLimit int

	PollInterval   time.Duration
	ReportInterval time.Duration

	Logger    *zap.Logger
	Client    Client
	Collector Collector
}

// validateOptions валидирует параметры агента и подставляет значения по умолчанию,
// если параметр не был предоставлен.
func validateOptions(opts *AgentOptions) *AgentOptions {
	if opts == nil {
		opts = defaultOpts()
	}

	if opts.RateLimit <= 0 {
		opts.RateLimit = 3
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

	collectChan := make(chan []model.MetricDto)
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
					err := a.opts.Client.Send(ctx, metrics)
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
func (a *Agent) Collect(ctx context.Context, collectChan chan<- []model.MetricDto) {
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
func (a *Agent) poller(ctx context.Context) <-chan []model.MetricDto {
	pollChan := make(chan []model.MetricDto)

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
