package linter

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/htrandev/metrics/cmd/linter/analyzer"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
