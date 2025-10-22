package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/internal/repository/local"
	"github.com/htrandev/metrics/internal/repository/postgres"
	"github.com/htrandev/metrics/internal/router"
	"github.com/htrandev/metrics/internal/service/metrics"
	"github.com/htrandev/metrics/migrations"
	"github.com/htrandev/metrics/pkg/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	zl.Info("init storage")
	storage, err := newStorage(ctx, flags, zl)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}
	defer storage.Close()

	zl.Info("init metric service")
	metricService := metrics.NewService(&metrics.ServiseOptions{
		Logger:  zl,
		Storage: storage,
	})

	zl.Info("init handler")
	metricHandler := handler.NewMetricsHandler(zl, metricService)

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

func newStorage(ctx context.Context, cfg flags, logger *zap.Logger) (model.Storager, error) {
	var storage model.Storager
	var err error

	switch {
	case cfg.databaseDsn != "":
		db, err := sql.Open("pgx", cfg.databaseDsn)
		if err != nil {
			return nil, fmt.Errorf("open db: %w", err)
		}
		storage = postgres.New(db, "/migrations")

		logger.Info("init provider")
		provider, err := goose.NewProvider(database.DialectPostgres, db, migrations.Embed)
		if err != nil {
			return nil, fmt.Errorf("goose: create new provider: %w", err)
		}

		logger.Info("up migrations")
		if _, err := provider.Up(ctx); err != nil {
			return nil, fmt.Errorf("goose: provider up: %w", err)
		}
	case cfg.restore:
		storage, err = local.NewRestore(&local.StorageOptions{
			FileName: cfg.filePath,
			Interval: cfg.storeInterval,
			Logger:   logger,
		})
		if err != nil {
			return nil, fmt.Errorf("creating restore: %w", err)
		}
	default:
		storage, err = local.NewRepository(&local.StorageOptions{
			FileName: cfg.filePath,
			Interval: cfg.storeInterval,
			Logger:   logger,
		})
		if err != nil {
			return nil, fmt.Errorf("creating default storage: %w", err)
		}
	}

	return storage, nil
}
