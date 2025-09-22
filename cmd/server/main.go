package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/htrandev/metrics/internal/handler"
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

	log.Println("start serving")
	r.Get("/", metricHandler.GetAll)
	r.Get("/value/{metricType}/{metricName}", metricHandler.Get)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.Update)
	return http.ListenAndServe(flags.addr, r)
}
