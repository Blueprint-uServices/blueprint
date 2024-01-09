// Package main provides the LeafApp application, primarily used for
// testing compilation during the plugin development process.
//
// Run with go run examples/leaf/wiring/main.go
package main

import (
	"github.com/blueprint-uservices/blueprint/examples/leaf/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/wiringcmd"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

func main() {
	// Configure the location of our workflow spec
	workflow.Init("../workflow")

	// Build a supported wiring spec
	name := "LeafApp"
	wiringcmd.MakeAndExecute(
		name,
		specs.Docker,
		specs.Thrift,
		specs.HTTP,
		specs.TimeoutDemo,
		specs.TimeoutRetriesDemo,
	)
}
