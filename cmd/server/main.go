package main

import (
	"log"
	"net/http"
	"os"

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
	mux := http.NewServeMux()
	s := repository.NewMemStorageRepository()
	updateHandler := handler.NewUpdateHandler(s)

	mux.Handle("/update/{metricType}/{metricName}/{metricValue}", updateHandler)
	log.Print("Start serving")
	return http.ListenAndServe("localhost:8080", mux)
}
