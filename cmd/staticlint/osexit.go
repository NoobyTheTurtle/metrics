package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// osExitAnalyzer проверяет запрет использования прямого вызова os.Exit в функции main пакета main.
var osExitAnalyzer = &analysis.Analyzer{
	Name:     "osexit",
	Doc:      "check for direct os.Exit calls in main function of main package",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		if funcDecl.Name.Name != "main" {
			return
		}

		ast.Inspect(funcDecl, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if isProhibitedCall(call, pass) {
					pass.Reportf(call.Pos(), "direct os.Exit calls are not allowed in main function of main package")
				}
			}
			return true
		})
	})

	return nil, nil
}

func isProhibitedCall(call *ast.CallExpr, pass *analysis.Pass) bool {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		return isProhibitedSelectorCall(fun, pass)
	}
	return false
}

func isProhibitedSelectorCall(sel *ast.SelectorExpr, pass *analysis.Pass) bool {
	if obj := pass.TypesInfo.ObjectOf(sel.Sel); obj != nil {
		pkg := obj.Pkg()
		if pkg == nil {
			return false
		}

		pkgPath := pkg.Path()
		funcName := sel.Sel.Name

		if pkgPath == "os" && funcName == "Exit" {
			return true
		}
	}

	return false
}
