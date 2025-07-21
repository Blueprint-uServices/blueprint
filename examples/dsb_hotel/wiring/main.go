// Package main provides an application for compiling a number of different
// wiring specs for the Hotel Reservation application from the DeathStarBench suite.
//
// To display options and usage, invoke:
//
//	go run main.go -h
package main

import (
	_ "github.com/blueprint-uservices/blueprint/examples/dsb_hotel/tests"
	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
)

func main() {
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/dsb_hotel/tests")

	name := "Hotel"
	cmdbuilder.MakeAndExecute(
		name,
		// Change the spec below to select the default experiment/topology.
		// For example: specs.Original, specs.Chain, specs.Fanin, specs.Fanout
		// You can also use the -w flag to select the spec at runtime.
		specs.Fanout,
	)
}
