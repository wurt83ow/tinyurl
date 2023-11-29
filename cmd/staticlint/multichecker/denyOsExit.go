// Package multichecker provides an analyzer to prevent the use of a direct call to os.Exit
// in the main function of the main package.
package multichecker

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer is the main entry point for the deny_os_exit analyzer.
var Analyzer = &analysis.Analyzer{
	Name: "deny_os_exit",
	Doc:  "prevents the use of a direct call to os.Exit in the main function of the main package.",
	Run:  runDenyOsExitAnalyzer,
}

// runDenyOsExitAnalyzer is the main function that will be executed during the analysis.
// It inspects each file in the main package and checks for the use of os.Exit in the main function.
func runDenyOsExitAnalyzer(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, inspectFuncDecl(pass))
		}
	}
	return nil, nil
}

// inspectFuncDecl is a helper function that returns a function to inspect each function declaration in the file.
func inspectFuncDecl(pass *analysis.Pass) func(n ast.Node) bool {
	return func(n ast.Node) bool {
		if stmt, ok := n.(*ast.FuncDecl); ok {
			checkMainFunction(pass, stmt)
		}
		return true
	}
}

// checkMainFunction checks if the provided function declaration is the main function,
// and then checks each statement in the body for os.Exit calls.
func checkMainFunction(pass *analysis.Pass, stmt *ast.FuncDecl) {
	if stmt.Name.Name == "main" {
		for _, bodyStmt := range stmt.Body.List {
			if callExpr, ok := bodyStmt.(*ast.ExprStmt); ok {
				checkExitCall(pass, callExpr)
			}
		}
	}
}

// checkExitCall checks if the provided expression statement contains a direct call to os.Exit,
// and reports a diagnostic if such a call is found.
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
