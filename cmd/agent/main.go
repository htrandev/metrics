package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mailru/easyjson"

	"github.com/htrandev/metrics/internal/model"
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
	collection := model.NewCollection()
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
			log.Println("send metrics")
			for _, metric := range metrics {
				if err := sendMetric(client, url, metric); err != nil {
					log.Printf("can't send metric [%+v]: %v", metric, err)
				}
			}
		}
	}

}

func sendMetric(client *resty.Client, url string, metric model.Metric) error {
	req := buildRequest(metric)
	body, err := buildBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}

	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
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

func buildRequest(metric model.Metric) model.Metrics {
	m := model.Metrics{
		ID:    metric.Name,
		MType: metric.Value.Type.String(),
	}

	switch metric.Value.Type {
	case model.TypeGauge:
		m.Value = &metric.Value.Gauge
	case model.TypeCounter:
		m.Delta = &metric.Value.Counter
	}

	return m
}

func buildBody(m model.Metrics) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	p, err := easyjson.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("can't marshal metrics: %w", err)
	}
	_, err = gz.Write(p)
	if err != nil {
		return nil, fmt.Errorf("can't write: %w", err)
	}
	gz.Close()
	return buf.Bytes(), nil
}
