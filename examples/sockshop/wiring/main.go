// Package main provides an application for compiling a number of different
// wiring specs for the SockShop application.
//
// Run with go run examples/sockshop/wiring/main.go
package main

import (
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/wiring/specs"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func main() {
	// Configure the location of our workflow spec
	workflow.Init("../workflow")

	// Build a supported wiring spec
	name := "SockShop"
	wiringcmd.MakeAndExecute(
		name,
		specs.Basic,
		specs.GRPC,
		specs.Docker,
	)
}
