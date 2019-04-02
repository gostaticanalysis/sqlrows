// +build !go1.12

package main

import (
	"github.com/gostaticanalysis/sqlrows/passes/sqlrows"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(sqlrows.Analyzer) }
