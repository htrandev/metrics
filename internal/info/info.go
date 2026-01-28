package info

import "fmt"

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func PrintBuildInfo() {
	fmt.Printf("Build version: %s\n", getValue(buildVersion))
	fmt.Printf("Build date: %s\n", getValue(buildDate))
	fmt.Printf("Build commit: %s\n", getValue(buildCommit))
}

func getValue(s string) string {
	if s != "" {
		return s
	}
	return "N/A"
}
