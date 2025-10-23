package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/agent"
	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Printf("run agent ends with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	log.Println("init config")
	conf, err := parseFlags()
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	log.Println("init logger")
	zl, err := logger.NewZapLogger(conf.logLvl)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	zl.Info("init tickers")
	poolTicker := time.NewTicker(conf.pollInterval)
	defer poolTicker.Stop()

	reportTicker := time.NewTicker(conf.reportInterval)
	defer reportTicker.Stop()

	zl.Info("init resty client")

	zl.Info("init agent")
	agent := agent.New(conf.addr)

	zl.Info("init collection")
	collection := model.NewCollection()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	for {
		var send bool
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-poolTicker.C:
		case <-reportTicker.C:
			send = true
		}
		zl.Info("collect metrics")

		metrics := collection.Collect()

		if send {
			zl.Info("send metrics")
			if err := agent.SendManyMetrics(ctx, metrics); err != nil {
				zl.Error("can't send many metric", zap.Error(err))
			}
		}
	}
}
