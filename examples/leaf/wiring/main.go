// Package main provides the LeafApp application, primarily used for
// testing compilation during the plugin development process.
//
// Run with go run examples/leaf/wiring/main.go
package main

import (
	"gitlab.mpi-sws.org/cld/blueprint/examples/leaf/wiring/specs"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
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
		specs.Logging,
	)
}
