package main

import (
	"flag"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/workflow/parser"
)

// For testing parser
func main() {
	srcDir := flag.String("src", "examples/leaf/workflow/leaf", "path to a workflow spec")
	flag.Parse()

	p := parser.NewSpecParser(*srcDir)
	p.ParseSpec()
}
