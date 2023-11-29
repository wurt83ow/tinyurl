package multichecker

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "deny_os_exit",
	Doc:  "prevents the use of a direct call to os.Exit in the main function of the main package.",
	Run:  runDenyOsExitAnalyzer,
}

func runDenyOsExitAnalyzer(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, inspectFuncDecl(pass))
		}
	}
	return nil, nil
}

func inspectFuncDecl(pass *analysis.Pass) func(n ast.Node) bool {
	return func(n ast.Node) bool {
		if stmt, ok := n.(*ast.FuncDecl); ok {
			checkMainFunction(pass, stmt)
		}
		return true
	}
}

func checkMainFunction(pass *analysis.Pass, stmt *ast.FuncDecl) {
	if stmt.Name.Name == "main" {
		for _, bodyStmt := range stmt.Body.List {
			if callExpr, ok := bodyStmt.(*ast.ExprStmt); ok {
				checkExitCall(pass, callExpr)
			}
		}
	}
}

func checkExitCall(pass *analysis.Pass, callExpr *ast.ExprStmt) {
	if call, ok := callExpr.X.(*ast.CallExpr); ok {
		if selExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "os" && selExpr.Sel.Name == "Exit" {
				pass.Report(analysis.Diagnostic{
					Pos:     selExpr.Pos(),
					Message: "avoid direct os.Exit calls in the main function",
				})
			}
		}
	}
}
