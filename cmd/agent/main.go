package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	models "github.com/htrandev/metrics/internal/model"
	"github.com/mailru/easyjson"
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
				if err := sendMetric(client, conf.addr, metric); err != nil {
					return fmt.Errorf("send metric: %w", err)
				}
			}
		}
	}

}

func sendMetric(client *resty.Client, addr string, metric models.Metric) error {
	url, err := buildURL(addr)
	if err != nil {
		return fmt.Errorf("build url for [%+v]", metric)
	}

	req := buildRequest(metric)
	log.Printf("send metric: [%+v] type: %s", req, req.MType)

	body, err := easyjson.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(url)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	return nil
}

func buildURL(addr string) (string, error) {
	u := url.URL{
		Scheme: "http",
		Host:   addr,
	}
	var err error
	u.Path, err = url.JoinPath("update/")

	if err != nil {
		return "", fmt.Errorf("join path: %w", err)
	}

	return u.String(), nil
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
