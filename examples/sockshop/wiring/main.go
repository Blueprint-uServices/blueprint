// An application for compiling the SockShop application.
// Provides a number of different wiring specs for compiling
// the application in different configurations.
//
// To display options and usage, invoke:
//
//	go run main.go -h
package main

import (
	"github.com/blueprint-uservices/blueprint/examples/sockshop/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

func main() {
	// Configure the location of our workflow spec
	workflow.Init("../workflow", "../tests", "../workload")

	// Build a supported wiring spec
	name := "SockShop"
	cmdbuilder.MakeAndExecute(
		name,
		specs.Basic,
		specs.GRPC,
		specs.Docker,
		specs.DockerRabbit,
	)
}
