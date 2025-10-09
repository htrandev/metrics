package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	models "github.com/htrandev/metrics/internal/model"
)

func main() {
	if err := run(); err != nil {
		log.Printf("run ends with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	log.Println("init config")
	conf, err := parseFlags()
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	log.Println("init tickers")
	poolTicker := time.NewTicker(conf.pollInterval)
	defer poolTicker.Stop()

	reportTicker := time.NewTicker(conf.reportInterval)
	defer reportTicker.Stop()

	client := resty.New()
	collection := models.NewCollection()
	url := buildURL(conf.addr)

	for {
		var send bool
		select {
		case <-poolTicker.C:
		case <-reportTicker.C:
			send = true
		}
		log.Println("collect metrics")
		metrics := collection.Collect()

		if send {
			for _, metric := range metrics {
				if err := sendMetric(client, url, metric); err != nil {
					return fmt.Errorf("send metric: %w", err)
				}
			}
		}
	}

}

func sendMetric(client *resty.Client, url string, metric models.Metric) error {
	req := buildRequest(metric)

	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	return nil
}

func buildURL(addr string) string {
	u := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   "/update/",
	}

	return u.String()
}

func buildRequest(metric models.Metric) models.Metrics {
	m := models.Metrics{
		ID:    metric.Name,
		MType: metric.Value.Type.String(),
	}

	switch metric.Value.Type {
	case models.TypeGauge:
		m.Value = &metric.Value.Gauge
	case models.TypeCounter:
		m.Delta = &metric.Value.Counter
	}

	return m
}
