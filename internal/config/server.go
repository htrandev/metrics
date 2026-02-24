package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Server represents server configuration.
type Server struct {
	Addr           string        `mapstructure:"ADDRESS"`
	LogLvl         string        `mapstructure:"LOG_LEVEL"`
	StoreInterval  time.Duration `mapstructure:"STORE_INTERVAL"`
	StoreFilePath  string        `mapstructure:"STORE_FILE"`
	Restore        bool          `mapstructure:"RESTORE"`
	DatabaseDsn    string        `mapstructure:"DATABASE_DSN"`
	MaxRetry       int           `mapstructure:"MAX_RETRY"`
	Signature      string        `mapstructure:"SIGNATURE"`
	AuditFile      string        `mapstructure:"AUDIT_FILE"`
	AuditURL       string        `mapstructure:"AUDIT_URL"`
	PprofAddr      string        `mapstructure:"PPROF_ADDRESS"`
	PrivateKeyFile string        `mapstructure:"CRYPTO_KEY"`
	TrustedSubnet  string        `mapstructure:"TRUSTED_SUBNET"`
}

// GetServerConfig return a server configuration.
func GetServerConfig() (Server, error) {
	v := viper.New()

	filepath := getConfigFilePath()
	v.SetConfigFile(filepath)

	if err := v.ReadInConfig(); err != nil {
		return Server{}, fmt.Errorf("load config file: %w", err)
	}

	flagVals := parseServerFlags(v)

	v.AutomaticEnv()

	for key := range flagVals {
		if envVal, exists := os.LookupEnv(key); exists {
			v.Set(key, envVal)
		}
	}

	var s Server
	if err := v.Unmarshal(&s); err != nil {
		return Server{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return s, nil
}

// parseServerFlags parse server flags.
func parseServerFlags(v *viper.Viper) map[string]any {
	var (
		addr           = pflag.String("a", "localhost:8080", "address to run server")
		logLvl         = pflag.String("lvl", "debug", "log level")
		storeInterval  = pflag.Int("i", 300, "interval of writeing metrics")
		storeFilePath  = pflag.String("f", "metrics.log", "path to file to write metrics")
		restore        = pflag.Bool("r", false, "restore previous metrics")
		databaseDsn    = pflag.String("d", "", "db dsn")
		maxRetry       = pflag.Int("maxRetry", 3, "pg max retry")
		signature      = pflag.String("k", "", "secret key")
		auditFile      = pflag.String("audit-file", "", "file path to save audit")
		auditURL       = pflag.String("audit-url", "", "url to send audit")
		pprofAddr      = pflag.String("pprof-addr", "localhost:6060", "pprof address")
		privateKeyFile = pflag.String("crypto-key", "", "path to private key file")
		trustedSubnet  = pflag.String("t", "", "trusted subnet")
	)
	pflag.Parse()

	// store flag values in a map for later merging
	flagVals := map[string]any{
		"ADDRESS":        *addr,
		"LOG_LEVEL":      *logLvl,
		"STORE_INTERVAL": *storeInterval,
		"STORE_FILE":     *storeFilePath,
		"RESTORE":        *restore,
		"DATABASE_DSN":   *databaseDsn,
		"MAX_RETRY":      *maxRetry,
		"SIGNATURE":      *signature,
		"AUDIT_FILE":     *auditFile,
		"AUDIT_URL":      *auditURL,
		"PPROF_ADDRESS":  *pprofAddr,
		"CRYPTO_KEY":     *privateKeyFile,
		"TRUSTED_SUBNET": *trustedSubnet,
	}

	for key, val := range flagVals {
		if val != nil {
			v.Set(key, val)
		}
	}

	return flagVals
}
