package main

import (
	"flag"
	"os"
)

type flags struct {
	addr string
}

func parseFlags() *flags {
	var f flags
	flag.StringVar(&f.addr, "a", "localhost:8080", "address to run server")
	flag.Parse()

	if addr := os.Getenv("ADDRESS"); addr != "" {
		f.addr = addr
	}

	return &f
}
