package sqlrows_test

import (
	"testing"

	"github.com/gostaticanalysis/sqlrows/passes/sqlrows"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, sqlrows.Analyzer, "a")
}
