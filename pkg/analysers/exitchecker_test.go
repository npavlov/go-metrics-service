package analysers_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/npavlov/go-metrics-service/pkg/analysers"
)

func TestExitCheckAnalyzer(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analysers.ExitCheckAnalyser, "exitcheck")
}

func TestExitCheckAnalyzerNotMainFunc(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analysers.ExitCheckAnalyser, "notmainfunc")
}

func TestExitCheckAnalyzerNotMainPackage(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analysers.ExitCheckAnalyser, "notmainpackage")
}
