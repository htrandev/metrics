package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/repository"
	"github.com/htrandev/metrics/internal/router"
	"github.com/htrandev/metrics/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Printf("run ends with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	flags := parseFlags()

	zl, err := logger.NewZapLogger(flags.logLvl)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	zl.Info("init config")

	s := repository.NewMemStorageRepository()
	metricHandler := handler.NewMetricsHandler(zl, s)

	router, err := router.New(metricHandler)
	if err != nil {
		return fmt.Errorf("can't create new router: %w", err)
	}

	zl.Info("", zap.String("addr", flags.addr))
	srv := http.Server{
		Addr:    flags.addr,
		Handler: router,
	}

	zl.Info("start serving")
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zl.Error("can't start server", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
