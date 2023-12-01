// Пакет main - это точка входа для анализатора staticcheck с дополнительной проверкой.
package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// Определение карты подключаемых правил.
	checks := map[string]bool{
		"SA1015": true, // Добавляем анализаторы класса SA.
		"SA1010": true,
		"S1004":  true, // Добавляем еще один анализатор другого класса.
	}
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		// Добавление в массив нужных проверок.
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	mychecks = append(mychecks, myMainCheck)
	multichecker.Main(
		mychecks...,
	)
}

// myMainCheck - это пользовательский анализатор для проверки вызовов os.Exit в функции main.
var myMainCheck = &analysis.Analyzer{
	Name: "maincheck",
	Doc:  "Проверка прямого вызова os.Exit в функции main",
	Run:  runMainCheck,
}

// runMainCheck - функция, выполняющая проверку вызовов os.Exit в функции main.
func runMainCheck(pass *analysis.Pass) (interface{}, error) {
	if pass.Analyzer.Name == "maincheck" {
		if pass.Pkg.Name() == "main" {
			ast.Inspect(pass.Files[0], func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := fun.X.(*ast.Ident); ok && ident.Name == "os" && fun.Sel.Name == "Exit" {
							pass.Reportf(call.Pos(), "не вызывайте os.Exit напрямую в функции main")
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
