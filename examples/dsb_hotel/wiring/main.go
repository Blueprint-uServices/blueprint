// Package main provides an application for compiling a number of different
// wiring specs for the Hotel Reservation application from the DeathStarBench suite.
//
// To display options and usage, invoke:
//
//	go run main.go -h
package main

import (
	"gitlab.mpi-sws.org/cld/blueprint/examples/dsb_hotel/wiring/specs"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func main() {
	workflow.Init("../workflow", "../tests")

	name := "Hotel"
	wiringcmd.MakeAndExecute(
		name,
		specs.Original,
	)
}
