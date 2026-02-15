package config

import (
	"flag"
	"fmt"
	"log"
	"time"
)

type Agent struct {
	Addr           string        `env:"address"`
	ReportInterval time.Duration `env:"report_interval"`
	PollInterval   time.Duration `env:"poll_interval"`
	LogLvl         string        `env:"log_level"`
	MaxRetry       int           `env:"max_retry"`
	Signature      string        `env:"signature"`
	RateLimit      int           `env:"rate_limit"`
	PublicKeyFile  string        `env:"crypto_key"`
}

type agentConfigFromFile struct {
	Addr           string        `json:"address"`
	ReportInterval time.Duration `json:"report_interval"`
	PollInterval   time.Duration `json:"poll_interval"`
	LogLvl         string        `json:"log_level"`
	MaxRetry       int           `json:"max_retry"`
	Signature      string        `json:"signature"`
	RateLimit      int           `json:"rate_limit"`
	PublicKeyFile  string        `json:"crypto_key"`
}

// GetAgentConfig return a server configuration.
func GetAgentConfig() (Agent, error) {
	var a Agent

	if configPath := getConfigFilePath(); configPath != "" {
		var cfg *agentConfigFromFile
		err := loadConfFromFile(configPath, cfg)
		if err != nil {
			log.Printf("can't read env from file [%s]: %v\n", configPath, err)
		} else {
			err := a.applyAgentConfig(cfg)
			if err != nil {
				return a, fmt.Errorf("apply agent config: %w", err)
			}
		}
	}

	return a, nil
}

// applyAgentConfig set server env value
//
// priority order: env -> flag -> config file.
func (a *Agent) applyAgentConfig(cfg *agentConfigFromFile) error {
	var err error

	addr := flag.String("a", "localhost:8080", "address to run server")
	report := flag.Int("r", 10, "report interval in seconds")
	poll := flag.Int("p", 2, "poll interval in seconds")
	logLvl := flag.String("lvl", "debug", "log level")
	maxRetry := flag.Int("maxRetry", 3, "max number of retries")
	signature := flag.String("k", "", "secret key")
	rateLimit := flag.Int("l", 3, "agent rate limit")
	publicKeyFile := flag.String("crypto-key", "", "path to public key file")

	flag.Parse()

	a.Addr = selectStringValue("ADDRESS", addr, cfg.Addr)
	a.ReportInterval, err = selectTimeDurationVal("REPORT_INTERVAL", *report, cfg.ReportInterval)
	if err != nil {
		return err
	}
	a.PollInterval, err = selectTimeDurationVal("POLL_INTERVAL", *poll, cfg.PollInterval)
	if err != nil {
		return err
	}
	a.LogLvl = selectStringValue("LOG_LEVEL", logLvl, cfg.LogLvl)
	a.MaxRetry, err = selectIntVal("MAX_RETRY", *maxRetry, cfg.MaxRetry)
	if err != nil {
		return err
	}
	a.Signature = selectStringValue("SIGNATURE", signature, cfg.Signature)
	a.RateLimit, err = selectIntVal("RATE_LIMIT", *rateLimit, cfg.RateLimit)
	if err != nil {
		return err
	}
	a.PublicKeyFile = selectStringValue("CRYPTO_KEY", publicKeyFile, cfg.PublicKeyFile)

	return nil
}
