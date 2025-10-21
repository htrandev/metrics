package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type flags struct {
	addr          string
	logLvl        string
	storeInterval time.Duration
	filePath      string
	restore       bool
	databaseDsn   string
}

func parseFlags() (flags, error) {
	var f flags
	var storeInterval int

	flag.StringVar(&f.addr, "a", "localhost:8080", "address to run server")
	flag.StringVar(&f.logLvl, "lvl", "debug", "log level")
	flag.IntVar(&storeInterval, "i", 300, "interval of writeing metrics")
	flag.StringVar(&f.filePath, "f", "metrics.log", "path to file to write metrics")
	flag.BoolVar(&f.restore, "r", false, "restore previous metrics")
	flag.StringVar(&f.databaseDsn, "d", "", "db dsn")

	flag.Parse()

	if addr := os.Getenv("ADDRESS"); addr != "" {
		f.addr = addr
	}
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		f.logLvl = lvl
	}

	if si := os.Getenv("STORE_INTERVAL"); si != "" {
		v, err := strconv.Atoi(si)
		if err != nil {
			return f, fmt.Errorf("parse store interval: %w", err)
		}
		f.storeInterval = time.Duration(v) * time.Second
	} else {
		f.storeInterval = time.Duration(storeInterval) * time.Second
	}

	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		f.filePath = filePath
	}

	if r := os.Getenv("RESTORE"); r != "" {
		restore, err := strconv.ParseBool(r)
		if err != nil {
			return f, fmt.Errorf("parse restore: %w", err)
		}
		f.restore = restore
	}

	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		f.databaseDsn = dsn
	}

	return f, nil
}
