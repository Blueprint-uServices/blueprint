package main

import (
	"fmt"
	"os"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang_process"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang_workflow"
	"golang.org/x/exp/slog"
)

func main() {

	fmt.Println("Constructing Wiring Spec")

	wiring := blueprint.NewWiringSpec("example")

	// Create the wiring spec

	golang_workflow.Init("path/to/workflow/spec")

	golang_workflow.Define(wiring, "b", "LeafService")
	golang_workflow.Define(wiring, "a", "nonLeafService", "b")

	golang_process.Define(wiring, "pa", "a")
	golang_process.Define(wiring, "pb", "b")

	// Do the building and print some stuff

	var b strings.Builder
	b.WriteString("WiringSpec:\n")
	b.WriteString(wiring.String())
	slog.Info(b.String())

	bp := wiring.Blueprint()
	bp.Instantiate("pa", "pb")

	application, err := bp.Build()
	if err != nil {
		slog.Error("Unable to build blueprint, exiting", "error", err)
		os.Exit(1)
	}

	b.Reset()
	b.WriteString("Application:\n")
	b.WriteString(application.String())
	slog.Info(b.String())
}
