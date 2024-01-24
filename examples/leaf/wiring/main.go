// Package main provides the LeafApp application, a simple application designed for demonstrating Blueprint
// usage and not as a realistic executable application.
//
// The wiring specs in the [specs] directory illustrate the usage of various Blueprint plugins.
//
// Leaf is also used by Blueprint developers while developing plugins.
//
// # Usage
//
// To display usage, run
//
//	go run . -h
package main

import (
	"github.com/blueprint-uservices/blueprint/examples/leaf/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
)

func main() {
	// Configure the location of our workflow spec
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/leaf/workflow")

	// Build a supported wiring spec
	name := "LeafApp"
	cmdbuilder.MakeAndExecute(
		name,
		specs.Docker,
		specs.Thrift,
		specs.HTTP,
		specs.TimeoutDemo,
		specs.TimeoutRetriesDemo,
		specs.Xtrace_Logger,
		specs.OT_Logger,
		specs.Govector,
	)
}
