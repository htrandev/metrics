package config

import (
	"flag"
	"fmt"
	"log"
	"time"
)

// Server represents server configuration
type Server struct {
	Addr           string        `env:"address"`
	LogLvl         string        `env:"log_level"`
	StoreInterval  time.Duration `env:"store_interval"`
	StoreFilePath  string        `env:"store_file"`
	Restore        bool          `env:"restore"`
	DatabaseDsn    string        `env:"database_dsn"`
	MaxRetry       int           `env:"max_retry"`
	Signature      string        `env:"signature"`
	AuditFile      string        `env:"audit_file"`
	AuditURL       string        `env:"audit_url"`
	PprofAddr      string        `env:"pprof_address"`
	PrivateKeyFile string        `env:"crypto_key"`
}

// serverConfigFromFile represents configuration from file
type serverConfigFromFile struct {
	Addr           string        `json:"address"`
	LogLvl         string        `json:"log_level"`
	StoreInterval  time.Duration `json:"store_interval"`
	StoreFilePath  string        `json:"store_file"`
	Restore        bool          `json:"restore"`
	DatabaseDsn    string        `json:"database_dsn"`
	MaxRetry       int           `json:"max_retry"`
	Signature      string        `json:"signature"`
	AuditFile      string        `json:"audit_file"`
	AuditURL       string        `json:"audit_url"`
	PprofAddr      string        `json:"pprof_address"`
	PrivateKeyFile string        `json:"crypto_key"`
}

// GetServerConfig return a server configuration.
func GetServerConfig() (Server, error) {
	var s Server

	if configPath := getConfigFilePath(); configPath != "" {
		var cfg *serverConfigFromFile
		err := loadConfFromFile(configPath, cfg)
		if err != nil {
			log.Printf("can't read env from file [%s]: %v\n", configPath, err)
		} else {
			err := s.applyServerConfig(cfg)
			if err != nil {
				return s, fmt.Errorf("apply server config: %w", err)
			}
		}
	}

	return s, nil
}

// applyServerConfig set server env value
//
// priority order: env -> flag -> config file.
func (s *Server) applyServerConfig(cfg *serverConfigFromFile) error {
	var err error

	addr := flag.String("a", "localhost:8080", "address to run server")
	logLvl := flag.String("lvl", "debug", "log level")
	storeInterval := flag.Int("i", 300, "interval of writeing metrics")
	storeFilePath := flag.String("f", "metrics.log", "path to file to write metrics")
	restore := flag.Bool("r", false, "restore previous metrics")
	databaseDsn := flag.String("d", "", "db dsn")
	maxRetry := flag.Int("maxRetry", 3, "pg max retry")
	signature := flag.String("k", "", "secret key")
	auditFile := flag.String("audit-file", "", "file path to save audit")
	auditURL := flag.String("audit-url", "", "url to send audit")
	pprofAddr := flag.String("pprod-addr", "localhost:6060", "pprof address")
	privateKeyFile := flag.String("crypto-key", "", "path to private key file")

	flag.Parse()

	s.Addr = selectStringValue("ADDRESS", addr, cfg.Addr)
	s.LogLvl = selectStringValue("LOG_LEVEL", logLvl, cfg.LogLvl)
	s.StoreInterval, err = selectTimeDurationVal("STORE_INTERVAL", *storeInterval, cfg.StoreInterval)
	if err != nil {
		return err
	}
	s.StoreFilePath = selectStringValue("FILE_STORAGE_PATH", storeFilePath, cfg.StoreFilePath)
	s.Restore, err = selectBoolVal("RESTORE", restore, cfg.Restore)
	s.DatabaseDsn = selectStringValue("DATABASE_DSN", databaseDsn, cfg.DatabaseDsn)
	s.MaxRetry, err = selectIntVal("MAX_RETRY", *maxRetry, cfg.MaxRetry)
	if err != nil {
		return err
	}
	s.Signature = selectStringValue("SIGNATURE", signature, cfg.Signature)
	s.AuditFile = selectStringValue("AUDIT_FILE", auditFile, cfg.AuditFile)
	s.AuditURL = selectStringValue("AUDIT_URL", auditURL, cfg.AuditURL)
	s.PprofAddr = selectStringValue("PPROF_ADDRESS", pprofAddr, cfg.PprofAddr)
	s.PrivateKeyFile = selectStringValue("CRYPTO_KEY", privateKeyFile, cfg.PrivateKeyFile)

	return nil
}
