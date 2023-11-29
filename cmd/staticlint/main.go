package main

import (
	mychecker "github.com/wurt83ow/tinyurl/cmd/staticlint/multichecker"

	gocritic "github.com/go-critic/go-critic/checkers/analyzer"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
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

func main() {

	filter := "SA"
	var statch []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {

		// all elements already have a prefix specified by the filter, added in case in the future
		// analyzers with a prefix different from the given one will be added.
		if v.Analyzer.Name[0:2] == filter {
			statch = append(statch, v.Analyzer)
		}
	}

	// определяем map подключаемых правил
	style := map[string]bool{
		"ST1000": true,
		"ST1001": true,
		"ST1003": true,
	}
	var stylch []*analysis.Analyzer
	for _, v := range stylecheck.Analyzers {
		// добавляем в массив нужные проверки
		if style[v.Analyzer.Name] {
			stylch = append(stylch, v.Analyzer)
		}
	}

	// определяем map подключаемых правил
	smpl := map[string]bool{
		"S1000": true,
		"S1001": true,
		"S1002": true,
	}
	var simplch []*analysis.Analyzer
	for _, v := range simple.Analyzers {
		// добавляем в массив нужные проверки
		if smpl[v.Analyzer.Name] {
			simplch = append(simplch, v.Analyzer)
		}
	}

	// определяем map подключаемых правил
	qckfx := map[string]bool{
		"qf1001": true,
		"qf1002": true,
		"qf1003": true,
	}
	var quickch []*analysis.Analyzer
	for _, v := range quickfix.Analyzers {
		// добавляем в массив нужные проверки
		if qckfx[v.Analyzer.Name] {
			quickch = append(quickch, v.Analyzer)
		}
	}

	gocriticAnalyzer := gocritic.Analyzer

	ineffassignAnalyzer := ineffassign.Analyzer
	myDenyOsExitAnalyzer := mychecker.Analyzer

	allcheck := []*analysis.Analyzer{
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpmux.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		slog.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer}

	allcheck = append(allcheck, statch...)
	allcheck = append(allcheck, stylch...)
	allcheck = append(allcheck, simplch...)
	allcheck = append(allcheck, quickch...)
	allcheck = append(allcheck, ineffassignAnalyzer)
	allcheck = append(allcheck, gocriticAnalyzer)
	allcheck = append(allcheck, myDenyOsExitAnalyzer)

	multichecker.Main(allcheck...)
}
