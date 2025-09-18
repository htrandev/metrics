package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	models "github.com/htrandev/metrics/internal/model"
)

const (
	host = "localhost"
	port = "8080"
)

const (
	poolInterval   = time.Second * 2
	reportInterval = time.Second * 10
)

func main() {
	if err := run(); err != nil {
		log.Printf("run ends with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	log.Println("init tickers")
	poolTicker := time.NewTicker(poolInterval)
	defer poolTicker.Stop()

	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	for {
		var send bool
		select {
		case <-poolTicker.C:
		case <-reportTicker.C:
			send = true
		}
		log.Println("collect metrics")
		metrics := models.Collect()

		if send {
			for _, metric := range metrics {
				if err := sendMetric(metric); err != nil {
					return fmt.Errorf("send metric: %w", err)
				}
			}
		}
	}

}

func sendMetric(metric models.Metric) error {
	url, err := buildURL(metric)
	if err != nil {
		return fmt.Errorf("build url for [%+v]", metric)
	}

	client := resty.New()
	_, err = client.R().Post(url)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	return nil
}

func buildURL(m models.Metric) (string, error) {
	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
	}
	var err error

	switch m.Value.Type {
	case models.TypeGauge:
		u.Path, err = url.JoinPath("update", m.Value.Type.String(), m.Name, strconv.FormatFloat(m.Value.Gauge, 'f', -1, 64))
	case models.TypeCounter:
		u.Path, err = url.JoinPath("update", m.Value.Type.String(), m.Name, strconv.FormatInt(m.Value.Counter, 10))
	}

	if err != nil {
		return "", fmt.Errorf("join path: %w", err)
	}

	log.Printf("send metric: [%+v] type: %s", m, m.Value.Type)
	return u.String(), nil
}
