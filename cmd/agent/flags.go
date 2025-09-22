package main

import "time"

type config struct {
	addr           string
	reportInterval time.Duration
	pollInterval   time.Duration
}

func parseFlags() *config {
	var c config

	return &c
}
