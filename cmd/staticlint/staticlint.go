package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// Определяем map подключаемых правил
	checks := map[string]bool{
		"SA1015": true, // Добавляем анализаторы класса SA
		"SA1010": true,
		"S1004":  true, // Добавляем еще один анализатор другого класса
	}
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		// Добавляем в массив нужные проверки
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	mychecks = append(mychecks, myMainCheck)
	multichecker.Main(
		mychecks...,
	)
}

var myMainCheck = &analysis.Analyzer{
	Name: "maincheck",
	Doc:  "Check for direct os.Exit call in main function",
	Run:  runMainCheck,
}

func runMainCheck(pass *analysis.Pass) (interface{}, error) {
	if pass.Analyzer.Name == "maincheck" {
		if pass.Pkg.Name() == "main" {
			ast.Inspect(pass.Files[0], func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := fun.X.(*ast.Ident); ok && ident.Name == "os" && fun.Sel.Name == "Exit" {
							pass.Reportf(call.Pos(), "do not call os.Exit directly in main function")
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
