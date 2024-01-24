// Package main provides an application for compiling different wiring specs for TrainTicket application.
//
// To display options and usage, invoke:
//
//	go run main.go -h
package main

import (
	_ "github.com/blueprint-uservices/blueprint/examples/train_ticket/tests"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
)

func main() {
	// Configure the location of our tests
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/train_ticket/tests")

	// Build a supported wiring spec
	name := "TrainTicket"
	cmdbuilder.MakeAndExecute(
		name,
		specs.Docker,
	)
}
