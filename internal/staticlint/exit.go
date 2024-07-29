// Package staticlint describes analyzer for checking existance os.Exit function direct call at
// main function of main package.
package staticlint

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check using os.Exit in function main of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue // check only "main" package, skip others
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.FuncDecl:
				if n.Name.Name != "main" {
					return false // check only "main" function, skip other
				}

				ast.Inspect(n.Body, func(internalNode ast.Node) bool {
					switch bodyNode := internalNode.(type) {
					case *ast.CallExpr:
						callexpr(pass, bodyNode)
						return false // skip anonymous functions in args
					case *ast.FuncDecl:
						return false // skip nested function declaration
					default:
						return true
					}
				})

				return false // already checked by nested ast.Insert call
			default:
				return true
			}
		})
	}

	return nil, nil
}

func callexpr(pass *analysis.Pass, call *ast.CallExpr) {
	fun, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	ident, ok := fun.X.(*ast.Ident)
	if !ok {
		return
	}

	if ident.Name == "os" && fun.Sel.Name == "Exit" {
		pass.Reportf(ident.NamePos, "call os.Exit at main func of package main")
	}
}
