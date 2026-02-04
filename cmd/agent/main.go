package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/htrandev/metrics/internal/agent"
	"github.com/htrandev/metrics/internal/info"
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
	info.PrintBuildInfo()

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

	zl.Info("init resty client")
	client := resty.New().
		SetTimeout(30 * time.Second)

	zl.Info("init collection")
	collection := model.NewCollection()

	zl.Info("init agent")
	agent := agent.New(&agent.AgentOptions{
		Addr:           conf.addr,
		Key:            conf.key,
		MaxRetry:       conf.maxRetry,
		Logger:         zl,
		Client:         client,
		Collector:      collection,
		RateLimit:      conf.rateLimit,
		PollInterval:   conf.pollInterval,
		ReportInterval: conf.reportInterval,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	agent.Run(ctx)
	return nil
}
