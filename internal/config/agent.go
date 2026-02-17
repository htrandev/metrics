package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Agent struct {
	Addr           string        `mapstructure:"ADDRESS"`
	ReportInterval time.Duration `mapstructure:"REPORT_INTERVAL"`
	PollInterval   time.Duration `mapstructure:"POLL_INTERVAL"`
	LogLvl         string        `mapstructure:"LOG_LEVEL"`
	MaxRetry       int           `mapstructure:"MAX_RETRY"`
	Signature      string        `mapstructure:"SIGNATURE"`
	RateLimit      int           `mapstructure:"RATE_LIMIT"`
	PublicKeyFile  string        `mapstructure:"CRYPTO_KEY"`
}

// GetAgentConfig return a server configuration.
func GetAgentConfig() (Agent, error) {
	v := viper.New()

	filepath := getConfigFilePath()
	v.SetConfigFile(filepath)

	if err := v.ReadInConfig(); err != nil {
		return Agent{}, fmt.Errorf("load config file: %w", err)
	}

	flagVals := parseAgentFlags(v)

	v.AutomaticEnv()

	for key := range flagVals {
		if envVal, exists := os.LookupEnv(key); exists {
			v.Set(key, envVal)
		}
	}

	var a Agent
	if err := v.Unmarshal(&a); err != nil {
		return Agent{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return a, nil
}

func parseAgentFlags(v *viper.Viper) map[string]any {
	var (
		addr          = pflag.String("a", "localhost:8080", "address to run server")
		report        = pflag.Int("r", 10, "report interval in seconds")
		poll          = pflag.Int("p", 2, "poll interval in seconds")
		logLvl        = pflag.String("lvl", "debug", "log level")
		maxRetry      = pflag.Int("maxRetry", 3, "max number of retries")
		signature     = pflag.String("k", "", "secret key")
		rateLimit     = pflag.Int("l", 3, "agent rate limit")
		publicKeyFile = pflag.String("crypto-key", "", "path to public key file")
	)

	pflag.Parse()

	flagVals := map[string]any{
		"ADDRESS":         *addr,
		"REPORT_INTERVAL": *report,
		"POLL_INTERVAL":   *poll,
		"LOG_LEVEL":       *logLvl,
		"MAX_RETRY":       *maxRetry,
		"SIGNATURE":       *signature,
		"RATE_LIMIT":      *rateLimit,
		"CRYPTO_KEY":      *publicKeyFile,
	}
	for key, val := range flagVals {
		if val != nil {
			v.Set(key, val)
		}
	}

	return flagVals
}
