package main

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	workflow "gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang_workflow"
	"golang.org/x/exp/slog"
)

func main() {

	fmt.Println("Constructing Wiring Spec")

	wiring := blueprint.NewWiringSpec()

	// Create the wiring spec

	workflow.SetWorkflowSpecPath("path/to/workflow/spec")

	workflow.Add(wiring, "b", "LeafService")
	workflow.Add(wiring, "a", "nonLeafService", "b")

	// Do the building and print some stuff

	var b strings.Builder
	b.WriteString("WiringSpec:\n")
	b.WriteString(wiring.String())
	slog.Info(b.String())

	bp := wiring.Build()
	bp.InstantiateAll()

	b.Reset()
	b.WriteString("Blueprint:\n")
	b.WriteString(bp.String())
	slog.Info(b.String())
}
