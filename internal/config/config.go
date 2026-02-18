package config

import (
	"os"
	"strings"
)

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
