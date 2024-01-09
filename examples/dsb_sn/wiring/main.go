// Package main provides an application for compiling different
// wiring specs for DeathStarBench SocialNetwork application.
//
// To display options and usage, invoke:
//
//  go run main.go -h
package main

import (
	"github.com/Blueprint-uServices/blueprint/examples/dsb_sn/wiring/specs"
	"github.com/Blueprint-uServices/blueprint/plugins/wiringcmd"
	"github.com/Blueprint-uServices/blueprint/plugins/workflow"
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
