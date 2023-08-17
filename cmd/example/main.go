package main

import (
	"fmt"
	"os"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/workflow"
	"golang.org/x/exp/slog"
)

func serviceDefaults(wiring blueprint.WiringSpec, serviceName string) {
	procName := fmt.Sprintf("p%s", serviceName)
	// opentelemetry.WrapService(wiring, serviceName)
	golang.CreateProcess(wiring, procName, serviceName)

}

func main() {

	fmt.Println("Constructing Wiring Spec")

	wiring := blueprint.NewWiringSpec("example")

	// Create the wiring spec

	workflow.Init("path/to/workflow/spec")

	workflow.Define(wiring, "b", "LeafService")
	workflow.Define(wiring, "a", "nonLeafService", "b")

	serviceDefaults(wiring, "a")
	serviceDefaults(wiring, "b")

	// Do the building and print some stuff

	var b strings.Builder
	b.WriteString("WiringSpec:\n")
	b.WriteString(wiring.String())
	slog.Info(b.String())

	bp := wiring.GetBlueprint()
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
