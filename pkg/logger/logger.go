package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// NewZapLogger возвращает новый экземпляр логгера.
func NewZapLogger(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build config: %w", err)
	}

	return zl, nil
}
