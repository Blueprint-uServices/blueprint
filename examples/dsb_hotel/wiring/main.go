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
	_ "github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
)

func main() {
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow")
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/dsb_hotel/tests")

	name := "Hotel"
	cmdbuilder.MakeAndExecute(
		name,
		specs.Original,
	)
}
