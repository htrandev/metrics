package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// loadConfFromFile loads configuration from given file to given interface variable.
func loadConfFromFile(configFilePath string, e any) error {
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("flags: read file: %w", err)
	}
	if err := json.Unmarshal(content, e); err != nil {
		return fmt.Errorf("flags: unmarshal file content: %w", err)
	}
	return nil
}

// getConfigFilePath return path to config file from env or comands line argument.
func getConfigFilePath() string {
	if v := os.Getenv("CONFIG"); v != "" {
		return v
	}

	args := os.Args[1:]
	for i := range len(args) {
		a := args[i]

		if a == "-c" || a == "-config" {
			if i+1 < len(args) {
				return args[i+1]
			}
			return ""
		}

		if val, exists := strings.CutPrefix(a, "-c="); exists {
			return val
		}
		if val, exists := strings.CutPrefix(a, "-config="); exists {
			return val
		}
	}

	return ""
}

func selectStringValue(envName string, flagVal *string, cfgFileVal string) string {
	if envVal, exists := os.LookupEnv(envName); exists {
		return envVal
	} else if flagVal != nil {
		return *flagVal
	}
	return cfgFileVal
}

func selectTimeDurationVal(envName string, flagVal int, cfgFileVal time.Duration) (time.Duration, error) {
	if envVal, exists := os.LookupEnv(envName); exists {
		v, err := time.ParseDuration(envVal)
		if err != nil {
			return 0, fmt.Errorf("parse [%s]: %w", envName, err)
		}
		return time.Duration(v) * time.Second, nil
	} else if flagVal != 0 {
		return time.Duration(flagVal) * time.Second, nil
	}
	return cfgFileVal * time.Second, nil
}

func selectIntVal[T ~int | ~int32 | ~int64](envName string, flagVal, cfgFileVal T) (T, error) {
	if val, exists := os.LookupEnv(envName); exists {
		v, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("convert env [%s] value: %w", envName, err)
		}
		return T(v), nil
	}
	if flagVal != 0 {
		return flagVal, nil
	}
	return cfgFileVal, nil
}

func selectBoolVal(envName string, flagVal *bool, cfgFileVal bool) (bool, error) {
	if val, exists := os.LookupEnv(envName); exists {
		v, err := strconv.ParseBool(val)
		if err != nil {
			return false, fmt.Errorf("convert env [%s] value: %w", envName, err)
		}
		return v, nil
	}
	if flagVal != nil {
		return *flagVal, nil
	}
	return cfgFileVal, nil
}
