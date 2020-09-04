package main

import (
	"github.com/joehattori/checkunderscore"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(checkunderscore.Analyzer) }
