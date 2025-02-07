package analysers_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/npavlov/go-metrics-service/pkg/analysers"
)

func TestExitCheckAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analysers.ExitCheckAnalyser, "exitcheck")
}
