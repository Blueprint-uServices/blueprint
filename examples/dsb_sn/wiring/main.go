// Package main provides an application for compiling different
// wiring specs for DeathStarBench SocialNetwork application.
//
// To display options and usage, invoke:
//
//  go run main.go -h
package main

import (
	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/wiringcmd"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
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
