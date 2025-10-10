package main

import (
	"flag"
	"os"
)

type flags struct {
	addr   string
	logLvl string
}

func parseFlags() *flags {
	var f flags
	flag.StringVar(&f.addr, "a", "localhost:8080", "address to run server")
	flag.StringVar(&f.logLvl, "lvl", "debug", "log level")
	flag.Parse()

	if addr := os.Getenv("ADDRESS"); addr != "" {
		f.addr = addr
	}
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		f.logLvl = lvl
	}

	return &f
}
