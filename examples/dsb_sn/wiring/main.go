// Package main provides an application for compiling different
// wiring specs for DeathStarBench SocialNetwork application.
//
// To display options and usage, invoke:
//
//	go run main.go -h
package main

import (
	_ "github.com/blueprint-uservices/blueprint/examples/dsb_sn/tests"
	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/wiring/specs"
	_ "github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
)

func main() {
	// Configure the location of our workflow spec and tests
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork")
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/dsb_sn/tests")

	// Build a supported wiring spec
	name := "SocialNetwork"
	cmdbuilder.MakeAndExecute(
		name,
		specs.Docker,
	)
}
