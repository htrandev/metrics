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
)

func main() {
	if err := run(); err != nil {
		log.Printf("run ends with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	log.Println("init config")
	flags := parseFlags()

	r := chi.NewRouter()

	s := repository.NewMemStorageRepository()
	metricHandler := handler.NewMetricsHandler(s)

	lm, err := middleware.NewLogger("info")
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	r.With(
		middleware.MethodChecker(http.MethodGet),
		lm.Logger(),
	).Get("/", metricHandler.GetAll)

	r.With(
		middleware.MethodChecker(http.MethodGet),
		lm.Logger(),
	).Get("/value/{metricType}/{metricName}", metricHandler.Get)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		lm.Logger(),
	).Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.Update)

	srv := http.Server{
		Addr:    flags.addr,
		Handler: r,
	}

	log.Println("start serving")
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
