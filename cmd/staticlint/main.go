// Package main provides a tool to run a collection of static analysis tools in Golang.
//
// Usage:
//  1. Import necessary analyzers from relevant packages.
//  2. Create maps to specify included analyzers for different rule sets (e.g., staticcheck, stylecheck).
//  3. Create slices of *analysis.Analyzer for each rule set.
//  4. Combine all analyzers into one slice.
//  5. Use multichecker.Main with the combined analyzers to run static analysis.
//
// Analyzers:
//   - staticcheck: A set of analyzers provided by staticcheck.io.
//   - stylecheck: A set of style-related analyzers.
//   - simple: A set of simple code quality analyzers.
//   - quickfix: Analyzers with quick fixes for issues.
//   - gocritic: Analyzers from the go-critic project.
//   - ineffassign: Analyzer to detect ineffectual assignments.
//   - mychecker: Custom analyzer to deny the use of direct os.Exit calls in the main function of the main package.
//
// Notes:
//   - Adjust the filter variable to choose specific analyzers from staticcheck.
//   - Modify the maps (style, smpl, qckfx) to include or exclude specific analyzers.
package main

import (
	gocritic "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	mychecker "github.com/wurt83ow/tinyurl/cmd/staticlint/multichecker"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// main is the entry point of the program.
func main() {

	// filter defines a prefix for selecting specific analyzers from staticcheck.
	filter := "SA"
	var statch []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		// Select analyzers with the specified prefix.
		if v.Analyzer.Name[0:2] == filter {
			statch = append(statch, v.Analyzer)
		}
	}

	// style is a map of included style-related analyzers.
	style := map[string]bool{
		"ST1000": true,
		"ST1001": true,
		"ST1003": true,
	}
	var stylch []*analysis.Analyzer
	for _, v := range stylecheck.Analyzers {
		// Include analyzers based on the style map.
		if style[v.Analyzer.Name] {
			stylch = append(stylch, v.Analyzer)
		}
	}
	// smpl is a map of included simple code quality analyzers.
	smpl := map[string]bool{
		"S1000": true,
		"S1001": true,
		"S1002": true,
	}
	var simplch []*analysis.Analyzer
	for _, v := range simple.Analyzers {
		// Include analyzers based on the smpl map.
		if smpl[v.Analyzer.Name] {
			simplch = append(simplch, v.Analyzer)
		}
	}

	// qckfx is a map of included analyzers with quick fixes.
	qckfx := map[string]bool{
		"qf1001": true,
		"qf1002": true,
		"qf1003": true,
	}
	var quickch []*analysis.Analyzer
	for _, v := range quickfix.Analyzers {
		// Include analyzers based on the qckfx map.
		if qckfx[v.Analyzer.Name] {
			quickch = append(quickch, v.Analyzer)
		}
	}

	// gocriticAnalyzer is an analyzer from the go-critic project.
	gocriticAnalyzer := gocritic.Analyzer

	// ineffassignAnalyzer is an analyzer to detect ineffectual assignments.
	ineffassignAnalyzer := ineffassign.Analyzer

	// myDenyOsExitAnalyzer is a custom analyzer to deny the use of direct os.Exit calls in the main function.
	myDenyOsExitAnalyzer := mychecker.Analyzer

	allcheck := []*analysis.Analyzer{
		// Common Go Analysis Tools
		appends.Analyzer,             // Detects calls to append where the result is not used.
		asmdecl.Analyzer,             // Reports assembly declarations and calls to assembly functions.
		assign.Analyzer,              // Detects implicit assignments.
		atomic.Analyzer,              // Detects common mistakes using the sync/atomic package.
		atomicalign.Analyzer,         // Detects non-atomic struct field alignment.
		bools.Analyzer,               // Suggests simplified or corrected boolean expressions.
		buildssa.Analyzer,            // Builds a single, unified SSA form for the entire package.
		buildtag.Analyzer,            // Detects and suggests fixes for suspicious build tags.
		cgocall.Analyzer,             // Detects misuse of cgo.
		composite.Analyzer,           // Suggests a range of simplifications for composite literals.
		copylock.Analyzer,            // Detects passing a lock by value.
		ctrlflow.Analyzer,            // Inspects control flow structures.
		deepequalerrors.Analyzer,     // Finds calls to reflect.DeepEqual that can produce errors.
		defers.Analyzer,              // Checks for deferred calls within commonly-mistaken code.
		directive.Analyzer,           // Checks for poorly formatted or unnecessary build constraints.
		errorsas.Analyzer,            // Finds common cases of using errors.As incorrectly.
		fieldalignment.Analyzer,      // Checks for struct fields that could have been wider if aligned differently.
		findcall.Analyzer,            // Looks for the occurrence of a specified function call.
		framepointer.Analyzer,        // Detects incorrect or suspicious uses of the frame pointer.
		httpmux.Analyzer,             // Detects suspicious http.ServeMux usage.
		httpresponse.Analyzer,        // Suggests improvements for HTTP response handling.
		ifaceassert.Analyzer,         // Suggests a type switch instead of a series of type assertions.
		inspect.Analyzer,             // Examines all functions.
		loopclosure.Analyzer,         // Detects references to loop variables from within nested functions.
		lostcancel.Analyzer,          // Finds contexts that are not canceled when their owner function returns.
		nilfunc.Analyzer,             // Checks for self assignments of functions to their own receiver.
		nilness.Analyzer,             // Finds dereferences of possibly nil pointers.
		pkgfact.Analyzer,             // Detects incorrect usage of go:linkname.
		printf.Analyzer,              // Checks printf-like functions for consistency.
		reflectvaluecompare.Analyzer, // Detects equality comparisons against reflect.Value.
		shadow.Analyzer,              // Finds variables that may have been unintentionally shadowed.
		shift.Analyzer,               // Detects shifts that equal their left-hand operand.
		sigchanyzer.Analyzer,         // Checks for common mistakes using signals and channels.
		slog.Analyzer,                // Checks for common mistakes using the log package.
		sortslice.Analyzer,           // Detects slice sorts with ineffective comparison functions.
		stdmethods.Analyzer,          // Checks for calls to methods defined on interfaces in the stdlib.
		stringintconv.Analyzer,       // Suggests simplifications for conversions between strings and integers.
		structtag.Analyzer,           // Suggests simplifications for struct field tags.
		testinggoroutine.Analyzer,    // Checks that tests do not leave goroutines undeleted.
		tests.Analyzer,               // Finds suspicious test code.
		timeformat.Analyzer,          // Suggests changes to time format strings.
		unmarshal.Analyzer,           // Checks for unmarshal calls with invalid struct tags.
		unreachable.Analyzer,         // Detects unreachable code.
		unsafeptr.Analyzer,           // Detects unsafe.Pointer conversions that can be removed.
		unusedresult.Analyzer,        // Detects unused results of function calls.
		unusedwrite.Analyzer,         // Detects writes to variables that are never read.
		usesgenerics.Analyzer,        // Checks for uses of type parameters with constraints.
	}

	allcheck = append(allcheck, statch...)
	allcheck = append(allcheck, stylch...)
	allcheck = append(allcheck, simplch...)
	allcheck = append(allcheck, quickch...)
	allcheck = append(allcheck, ineffassignAnalyzer)
	allcheck = append(allcheck, gocriticAnalyzer)
	allcheck = append(allcheck, myDenyOsExitAnalyzer)

	multichecker.Main(allcheck...)
}
