package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type config struct {
	addr           string
	reportInterval time.Duration
	pollInterval   time.Duration
}

func parseFlags() (*config, error) {
	var c config

	var poll, report int

	flag.StringVar(&c.addr, "a", "localhost:8080", "address to run server")
	flag.IntVar(&report, "r", 10, "report interval in seconds")
	flag.IntVar(&poll, "p", 2, "poll interval in seconds")

	flag.Parse()

	c.pollInterval = time.Duration(poll) * time.Second
	c.reportInterval = time.Duration(report) * time.Second

	if addr := os.Getenv("ADDRESS"); addr != "" {
		c.addr = addr
	}
	if r := os.Getenv("REPORT_INTERVAL"); r != "" {
		v, err := strconv.Atoi(r)
		if err != nil {
			return nil, fmt.Errorf("parse report interval: %w", err)
		}
		c.reportInterval = time.Duration(v) * time.Second
	}
	if p := os.Getenv("POLL_INTERVAL"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("parse poll interval: %w", err)
		}
		c.pollInterval = time.Duration(v) * time.Second
	}

	return &c, nil
}
