package checkunderscore_test

import (
	"testing"

	"checkunderscore"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, checkunderscore.Analyzer, "a")
}

