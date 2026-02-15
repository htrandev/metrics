package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/audit"
	"github.com/htrandev/metrics/internal/config"
	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/info"
	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/internal/repository/local"
	"github.com/htrandev/metrics/internal/repository/postgres"
	"github.com/htrandev/metrics/internal/router"
	"github.com/htrandev/metrics/internal/service/metrics"
	"github.com/htrandev/metrics/migrations"
	"github.com/htrandev/metrics/pkg/crypto"
	"github.com/htrandev/metrics/pkg/logger"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		log.Printf("run ends with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	info.PrintBuildInfo()

	log.Println("init config")
	cfg, err := config.GetServerConfig()
	if err != nil {
		return fmt.Errorf("get server config: %w", err)
	}

	log.Println("init logger")
	zl, err := logger.NewZapLogger(cfg.LogLvl)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zl.Info("init storage")
	storage, err := newStorage(ctx, cfg, zl)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}
	defer storage.Close()

	zl.Info("init metric service")
	metricService := metrics.NewService(&metrics.ServiсeOptions{
		Logger:  zl,
		Storage: storage,
	})

	zl.Info("init publisher")
	auditor := audit.NewAuditor()

	zl.Info("init subscribers")
	subs := make([]audit.Observer, 0, 2)

	if cfg.AuditFile != "" {
		zl.Info("init file auditor")
		flag := os.O_RDWR | os.O_CREATE | os.O_APPEND
		f, err := os.OpenFile(cfg.AuditFile, flag, 0664)
		if err != nil {
			return fmt.Errorf("open audit file: %w", err)
		}

		fileAudit := audit.NewFile(uuid.New(), f, zl)
		subs = append(subs, fileAudit)
	}

	if cfg.AuditURL != "" {
		zl.Info("init url auditor")
		auditClient := resty.New().
			SetTimeout(30 * time.Second)

		urlAudit := audit.NewURL(uuid.New(), cfg.AuditURL, auditClient, zl)
		subs = append(subs, urlAudit)
	}

	zl.Info("register subscribers")
	registerSubscribers(auditor, subs...)

	zl.Info("init handler")
	metricHandler := handler.NewMetricsHandler(zl, metricService, auditor)

	zl.Info("init private key")
	privateKey, err := crypto.PrivateKey(cfg.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("init private key: %w", err)
	}

	zl.Info("init router")
	router := router.New(cfg.Signature, privateKey, zl, metricHandler)

	wg := sync.WaitGroup{}
	wg.Add(2) // debug and http servers

	pprofSrv := http.Server{Addr: cfg.PprofAddr}
	go func() {
		defer wg.Done()

		zl.Info(fmt.Sprintf("start pprof on http://%s/debug/pprof/", cfg.PprofAddr))
		if err := pprofSrv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	srv := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}
	go func() {
		defer wg.Done()

		zl.Info("start serving", zap.String("addr", cfg.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("can't start server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-stop

	wg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := pprofSrv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown pprof server: %w", err)
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}

func newStorage(ctx context.Context, cfg config.Server, logger *zap.Logger) (model.Storager, error) {
	var storage model.Storager
	var err error

	switch {
	case cfg.DatabaseDsn != "":
		db, err := sql.Open("pgx", cfg.DatabaseDsn)
		if err != nil {
			return nil, fmt.Errorf("open db: %w", err)
		}
		storage = postgres.New(db, cfg.MaxRetry)

		logger.Info("init provider")
		provider, err := goose.NewProvider(database.DialectPostgres, db, migrations.Embed)
		if err != nil {
			return nil, fmt.Errorf("goose: create new provider: %w", err)
		}

		logger.Info("up migrations")
		if _, err := provider.Up(ctx); err != nil {
			return nil, fmt.Errorf("goose: provider up: %w", err)
		}
	case cfg.Restore:
		storage, err = local.NewRestore(&local.StorageOptions{
			FileName: cfg.StoreFilePath,
			Interval: cfg.StoreInterval,
			Logger:   logger,
			MaxRetry: cfg.MaxRetry,
		})
		if err != nil {
			return nil, fmt.Errorf("creating restore: %w", err)
		}
	default:
		storage, err = local.NewRepository(&local.StorageOptions{
			FileName: cfg.StoreFilePath,
			Interval: cfg.StoreInterval,
			Logger:   logger,
		})
		if err != nil {
			return nil, fmt.Errorf("creating default storage: %w", err)
		}
	}

	return storage, nil
}

func registerSubscribers(p *audit.Auditor, subs ...audit.Observer) {
	for _, sub := range subs {
		p.Register(sub)
	}
}
