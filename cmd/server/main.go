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
	"github.com/htrandev/metrics/internal/repository/file"
	"github.com/htrandev/metrics/internal/repository/memstorage"
	"github.com/htrandev/metrics/internal/router"
	"github.com/htrandev/metrics/internal/service/metrics"
	"github.com/htrandev/metrics/internal/service/restore"
	"github.com/htrandev/metrics/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Printf("run ends with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	log.Println("init config")
	flags, err := parseFlags()
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	log.Println("init logger")
	zl, err := logger.NewZapLogger(flags.logLvl)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zl.Info("init mem storage")
	memStorageRepository := memstorage.NewRepository()

	zl.Info("init file storage")
	fileRepository, err := file.NewRepository(flags.filePath)
	if err != nil {
		return fmt.Errorf("init file repository: %w", err)
	}
	defer func() { _ = fileRepository.Close() }()

	zl.Info("init metric service")
	metricService := metrics.NewService(memStorageRepository)

	zl.Info("init handler")
	metricHandler := handler.NewMetricsHandler(
		zl,
		metricService,
		fileRepository,
		flags.storeInterval,
	)

	zl.Info("run flusher")
	go metricHandler.Run(ctx)

	if flags.restore {
		zl.Info("restore previous metrics")
		restoreService, err := restore.NewService(flags.filePath, memStorageRepository, zl)
		if err != nil {
			return fmt.Errorf("init restore service: %w", err)
		}
		defer func() { _ = restoreService.Close() }()

		restoreCtx, restoreCancel := context.WithTimeout(ctx, 1*time.Minute)
		defer restoreCancel()

		if err := restoreService.Restore(restoreCtx); err != nil {
			return fmt.Errorf("can't restore metrics ferom file: %w", err)
		}

	}

	zl.Info("init router")
	router, err := router.New(zl, metricHandler)
	if err != nil {
		return fmt.Errorf("can't create new router: %w", err)
	}

	srv := http.Server{
		Addr:    flags.addr,
		Handler: router,
	}

	zl.Info("start serving")
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("can't start server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	shutDownCtx, shutDownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutDownCancel()

	if err := srv.Shutdown(shutDownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
