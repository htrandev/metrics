package local

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/mailru/easyjson"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/internal/repository"
)

// StorageOptions параметры локального хранилища.
type StorageOptions struct {
	FileName string
	Interval time.Duration
	Logger   *zap.Logger
	MaxRetry int
}

// MemStorage реализует in-memory хранилище метрик.
type MemStorage struct {
	metrics map[string]model.Metric

	file    *os.File
	scanner *bufio.Scanner

	opts *StorageOptions

	mu sync.RWMutex
}

// NewRepository создает новое хранилище без восстановления из файла.
func NewRepository(opts *StorageOptions) (*MemStorage, error) {
	flag := os.O_RDWR | os.O_CREATE
	storage, err := new(flag, opts)
	if err != nil {
		return nil, fmt.Errorf("create new repository: %w", err)
	}
	return storage, nil
}

// NewRestore создает хранилище с восстановлением метрик из файла.
func NewRestore(opts *StorageOptions) (*MemStorage, error) {
	flag := os.O_RDWR | os.O_CREATE | os.O_APPEND

	storage, err := new(flag, opts)
	if err != nil {
		return nil, fmt.Errorf("create new restore: %w", err)
	}

	storage.opts.Logger.Info("restore old metrics")
	if err := storage.restore(); err != nil {
		storage.Close()
		return nil, fmt.Errorf("memstorage: restore: %w", err)
	}
	return storage, nil
}

func new(flag int, opts *StorageOptions) (*MemStorage, error) {
	metrics := make(map[string]model.Metric)
	f, err := os.OpenFile(opts.FileName, flag, 0664)
	if err != nil {
		return nil, fmt.Errorf("restore: open file: %w", err)
	}
	if opts.MaxRetry == 0 {
		opts.MaxRetry = 3
	}
	storage := &MemStorage{
		metrics: metrics,
		file:    f,
		scanner: bufio.NewScanner(f),
		opts:    opts,
	}

	storage.opts.Logger.Info("start flusher")
	if opts.Interval > 0 {
		go storage.flush(opts.Interval)
	}
	return storage, nil
}

// Ping проверяет доступность хранилища.
func (m *MemStorage) Ping(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

// Set записывает значение метрики.
// Если метрика уже существует, то ничего не делает.
func (m *MemStorage) Set(Ctx context.Context, request *model.Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.metrics[request.Name]; !ok {
		m.metrics[request.Name] = *request
	}

	return nil
}

// Store записывает новое значение метрики.
// Если метрика существует, то обновляет ее значение.
func (m *MemStorage) Store(ctx context.Context, request *model.Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	metric, ok := m.metrics[request.Name]
	if !ok {
		m.metrics[request.Name] = *request
		return nil
	}

	switch request.Value.Type {
	case model.TypeGauge:
		metric.Value.Gauge = request.Value.Gauge
	case model.TypeCounter:
		metric.Value.Counter += request.Value.Counter
	}

	m.metrics[request.Name] = metric
	return nil
}

// StoreMany записывает новое значение метрик.
// Если метрика существует, то обновляет ее значение.
func (m *MemStorage) StoreMany(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		log.Println("repository/storeMany: request is nil")
		return nil
	}
	errs := make([]error, 0, len(metrics))
	for _, metric := range metrics {
		if err := m.Store(ctx, &metric); err != nil {
			err = fmt.Errorf("repository/storeMany: store [%+v]: %w", metric, err)
			errs = append(errs, err)
			continue
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// StoreManyWithRetry записывает новое значение метрик.
// Если метрика существует, то обновляет ее значение.
// При ошибке пытается записать еще maxRetry раз
func (m *MemStorage) StoreManyWithRetry(ctx context.Context, metrics []model.Metric) error {
	err := m.StoreMany(ctx, metrics)
	if err != nil {
		for i := 0; i < m.opts.MaxRetry; i++ {
			if err := m.StoreMany(ctx, metrics); err != nil {
				continue
			}
			return nil
		}
		return fmt.Errorf("repository/storeManyWithRetry: reach retry limits: %w", err)
	}
	return nil
}

// Get возвращает метрику по имени.
func (m *MemStorage) Get(ctx context.Context, name string) (model.Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metric, ok := m.metrics[name]
	if !ok {
		return model.Metric{}, fmt.Errorf("repository/get: metric with name [%s]: %w", name, repository.ErrNotFound)
	}
	return metric, nil
}

// GetAll возвращает все метрики.
func (m *MemStorage) GetAll(ctx context.Context) ([]model.Metric, error) {
	metrics := make([]model.Metric, 0, len(m.metrics))

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, metric := range m.metrics {
		metrics = append(metrics, metric)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})
	return metrics, nil
}

// flush асинхронно сохраняет все метрики в файл каждые Interval секунд.
func (m *MemStorage) flush(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()

		for _, metric := range m.metrics {
			data, err := easyjson.Marshal(metric)
			if err != nil {
				m.opts.Logger.Error(
					"marshal metric",
					zap.Any("metric", metric),
					zap.Error(err),
					zap.String("scope", "memstorage/flush"),
				)
			}
			if _, err := m.file.Write(data); err != nil {
				m.opts.Logger.Error(
					"write data",
					zap.String("data", string(data)),
					zap.Error(err),
					zap.String("scope", "memstorage/flush"),
				)
			}

			if _, err := m.file.Write([]byte("\n")); err != nil {
				m.opts.Logger.Error(
					"write new line",
					zap.Error(err),
					zap.String("scope", "memstorage/flush"),
				)
			}
		}
		m.mu.Unlock()
	}
}

// restore восстанавливает метрики из JSONL файла при запуске.
func (m *MemStorage) restore() error {
	ctx := context.Background()
	// читаем из файла пока не дойдем до конца
	for m.scanner.Scan() {
		data := m.scanner.Bytes()

		var metric model.Metric
		if err := easyjson.Unmarshal(data, &metric); err != nil {
			m.opts.Logger.Error("can't unmarshal data from file", zap.Error(err), zap.String("scope", "restore"))
			continue
		}

		m.opts.Logger.Debug("set metric", zap.Any("metric", metric))
		if err := m.Set(ctx, &metric); err != nil {
			m.opts.Logger.Error("can't set metric", zap.Error(err), zap.Any("metric", m), zap.String("scope", "restore"))
			continue
		}
	}

	// проверяем наличие ошибки
	if m.scanner.Err() != nil {
		return fmt.Errorf("scan: %w", m.scanner.Err())
	}
	return nil
}

// Close закрывает файл.
func (m *MemStorage) Close() error {
	return m.file.Close()
}

// Up проверяет готовность хранилища.
func (m *MemStorage) Up() error {
	return nil
}
