// Package main provides an application for compiling different wiring specs for TrainTicket application.
//
// To display options and usage, invoke:
//
//   go run main.go -h
package main

import (
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/wiringcmd"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

func main() {
	// Configure the location of our workflow spec
	workflow.Init("../workflow", "../tests")

	// Build a supported wiring spec
	name := "TrainTicket"
	wiringcmd.MakeAndExecute(
		name,
		specs.Docker,
	)
}
