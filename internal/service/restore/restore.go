package restore

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/htrandev/metrics/internal/model"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

type Storage interface {
	Store(context.Context, *model.Metric) error
}

type RestoreService struct {
	storage Storage
	logger  *zap.Logger

	file    *os.File
	scanner *bufio.Scanner
}

func NewService(fileName string, s Storage, logger *zap.Logger) (*RestoreService, error) {
	f, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("restore: open file: %w", err)
	}

	return &RestoreService{
		storage: s,
		file:    f,
		scanner: bufio.NewScanner(f),
		logger:  logger,
	}, nil
}

func (r *RestoreService) Restore(ctx context.Context) error {
	// читаем из файла пока не дойдем до конца
	for r.scanner.Scan() {
		data := r.scanner.Bytes()

		var m model.Metric
		if err := easyjson.Unmarshal(data, &m); err != nil {
			r.logger.Error("can't unmarshal data from file", zap.Error(err), zap.String("scope", "restore"))
			continue
		}

		r.logger.Debug("store metric", zap.Any("metric", m))
		if err := r.storage.Store(ctx, &m); err != nil {
			r.logger.Error("can't store metric", zap.Error(err), zap.Any("metric", m), zap.String("scope", "restore"))
			continue
		}
	}

	// проверяем наличие ошибки
	if r.scanner.Err() != nil {
		return fmt.Errorf("scan: %w", r.scanner.Err())
	}
	return nil
}

func (r *RestoreService) Close() error {
	return r.file.Close()
}
