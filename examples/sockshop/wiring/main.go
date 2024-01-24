// An application for compiling the SockShop application.
// Provides a number of different wiring specs for compiling
// the application in different configurations.
//
// To display options and usage, invoke:
//
//	go run main.go -h
package main

import (
	"fmt"

	_ "github.com/blueprint-uservices/blueprint/examples/sockshop/tests"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
)

func main() {
	fmt.Println("hello world")
	// Make sure tests and workflow can be found
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/sockshop/tests")
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/sockshop/workflow")

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
