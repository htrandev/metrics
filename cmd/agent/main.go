package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
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
	url, err := buildURL(addr, metric)
	if err != nil {
		return fmt.Errorf("build url for [%+v]", metric)
	}

	_, err = client.R().Post(url)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	return nil
}

func buildURL(addr string, m models.Metric) (string, error) {
	u := url.URL{
		Scheme: "http",
		Host:   addr,
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
