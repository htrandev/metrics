package main

import (
	"flag"
	"time"
)

type config struct {
	addr           string
	reportInterval time.Duration
	pollInterval   time.Duration
}

func parseFlags() *config {
	var c config
	var poll, report int

	flag.StringVar(&c.addr, "a", "localhost:8080", "address to run server")
	flag.IntVar(&report, "r", 10, "report interval in seconds")
	flag.IntVar(&poll, "p", 2, "poll interval in seconds")

	flag.Parse()

	c.pollInterval = time.Duration(poll) * time.Second
	c.reportInterval = time.Duration(report) * time.Second

	return &c
}
