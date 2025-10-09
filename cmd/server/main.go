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

	"github.com/go-chi/chi/v5"
	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/handler/middleware"
	"github.com/htrandev/metrics/internal/repository"
	"github.com/htrandev/metrics/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Printf("run ends with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	flags := parseFlags()

	// zl, err := logger.NewZapLogger(flags.logLvl)
	zl, err := logger.NewZapLogger("error")
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	zl.Info("init config")

	s := repository.NewMemStorageRepository()
	metricHandler := handler.NewMetricsHandler(zl, s)

	// router, err := router.New(metricHandler)
	// if err != nil {
	// 	return fmt.Errorf("can't create new router: %w", err)
	// }

	r := chi.NewRouter()

	r.With(
		middleware.MethodChecker(http.MethodGet),
		middleware.Logger(zl),
	).Get("/", metricHandler.GetAll)

	r.With(
		middleware.MethodChecker(http.MethodGet),
		middleware.Logger(zl),
	).Get("/value/{metricType}/{metricName}", metricHandler.Get)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		middleware.Logger(zl),
	).Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.Update)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		middleware.Logger(zl),
		middleware.ContentType(),
	).Post("/update/", metricHandler.UpdateViaBody)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		middleware.Logger(zl),
		middleware.ContentType(),
	).Post("/value/", metricHandler.GetViaBody)

	srv := http.Server{
		Addr:    flags.addr,
		Handler: r,
	}

	zl.Info("start serving")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("can't start server: %v", err)
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
