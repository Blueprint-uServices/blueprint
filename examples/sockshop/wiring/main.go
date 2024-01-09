// Package main provides an application for compiling a number of different
// wiring specs for the SockShop application.
//
// To display options and usage, invoke:
//
//	go run main.go -h
package main

import (
	"github.com/blueprint-uservices/blueprint/plugins/wiringcmd"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

func main() {
	// Configure the location of our workflow spec
	workflow.Init("../workflow", "../tests")

	// Build a supported wiring spec
	name := "SockShop"
	wiringcmd.MakeAndExecute(
		name,
		Basic,
		GRPC,
		Docker,
		DockerRabbit,
	)
}
