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
	"github.com/htrandev/metrics/internal/config"
	"github.com/htrandev/metrics/internal/info"
	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/pkg/crypto"
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
	conf, err := config.GetAgentConfig()
	if err != nil {
		return fmt.Errorf("get agent config: %w", err)
	}

	log.Println("init logger")
	zl, err := logger.NewZapLogger(conf.LogLvl)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	zl.Info("init resty client")
	client := resty.New().
		SetTimeout(30 * time.Second)

	zl.Info("init collection")
	collection := model.NewCollection()

	zl.Info("init public key")
	publicKey, err := crypto.PublicKey(conf.PublicKeyFile)
	if err != nil {
		return fmt.Errorf("init public key: %w", err)
	}

	zl.Info("init agent")
	agent := agent.New(&agent.AgentOptions{
		Addr:           conf.Addr,
		Signature:      conf.Signature,
		MaxRetry:       conf.MaxRetry,
		Key:            publicKey,
		Logger:         zl,
		Client:         client,
		Collector:      collection,
		RateLimit:      conf.RateLimit,
		PollInterval:   conf.PollInterval,
		ReportInterval: conf.ReportInterval,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	agent.Run(ctx)
	return nil
}
