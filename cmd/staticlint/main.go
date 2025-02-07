// Command staticlint runs multichecker to perform static analysis of the project code.
//
// This tool uses the standard Go analyzers, as well as additional
// checks from the staticcheck, simple, stylecheck, and quickfix packages.
//
// Pre-build the binary and run:
// make build-checker
// ./cmd/checker ./...
package main

import (
	"github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/npavlov/go-metrics-service/pkg/analysers"
)

// main creates and runs multichecker with the selected analyzers.
func main() {
	// Checks to include
	checks := map[string]bool{
		"S1000":  true, // Simple code optimizations
		"ST1000": true, // Code style
		"QF1001": true, // Automatic code correction
	}

	// Basic analyzers
	mychecks := []*analysis.Analyzer{
		analysers.ExitCheckAnalyser, // Custom analyzer
		printf.Analyzer,             // Check Printf formats
		shadow.Analyzer,             // Detect hidden variables
		shift.Analyzer,              // Check bit shifts
		structtag.Analyzer,          // Validate structure tags
		analyzer.Analyzer,           // go-critic analyzer
		errcheck.Analyzer,           // Check for unhandled errors
	}

	// Add analyzers from staticcheck, simple, stylecheck and quickfix
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range simple.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range stylecheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range quickfix.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// Launch multichecker
	multichecker.Main(mychecks...)
}
