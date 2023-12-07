// Package main provides the HotelReservation application from the DeathStarBench suite.
//
// Run with go run examples/dsb_hotel/wiring/main.go
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
