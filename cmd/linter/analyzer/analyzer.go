package analyzer

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = (&analysis.Analyzer{
	Name: "badexit",
	Doc:  "check panic and log.Fatal/os.Exit outside main",
	Run:  run,
})

func run(pass *analysis.Pass) (interface{}, error) {
	selExpr := func(x *ast.SelectorExpr) bool {
		pkg, ok := x.X.(*ast.Ident)

		isLogFatal := pkg.Name == "log" && strings.Contains(x.Sel.Name, "Fatal")
		isOsExit := pkg.Name == "os" && x.Sel.Name == "Exit"

		if ok && (isLogFatal || isOsExit) {
			pass.Reportf(x.Pos(), "calling %s.%s outside main", pkg.Name, x.Sel.Name)
		}
		return true
	}

	panicExpr := func(x *ast.CallExpr) bool {
		if fun, ok := x.Fun.(*ast.Ident); ok {
			if fun.Name == "panic" {
				pass.Reportf(x.Pos(), "don't use panic")
			}
		}
		return true
	}

	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.SelectorExpr:
				selExpr(x)
			case *ast.CallExpr:
				panicExpr(x)
			}
			return true
		})
	}

	return nil, nil
}
