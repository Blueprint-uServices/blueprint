// Package main provides an application for compiling different
// wiring specs for DeathStarBench SocialNetwork application.
//
// To display options and usage, invoke:
//
//  go run main.go -h
package main

import (
	"gitlab.mpi-sws.org/cld/blueprint/examples/leaf/wiring/specs"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func main() {
	// Configure the location of our workflow spec
	workflow.Init("../workflow", "../tests")

	// Build a supported wiring spec
	name := "SocialNetwork"
	wiringcmd.MakeAndExecute(
		name,
		specs.Docker,
	)
}
