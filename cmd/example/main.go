package main

import (
	"fmt"
	"os"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/opentelemetry"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/workflow"
	"golang.org/x/exp/slog"
)

func serviceDefaults(wiring blueprint.WiringSpec, serviceName string) {
	procName := fmt.Sprintf("p%s", serviceName)
	opentelemetry.Instrument(wiring, serviceName)
	grpc.Deploy(wiring, serviceName)
	golang.CreateProcess(wiring, procName, serviceName)

}

func main() {

	fmt.Println("Constructing Wiring Spec")

	wiring := blueprint.NewWiringSpec("example")

	// Create the wiring spec

	workflow.Init("path/to/workflow/spec")

	b := workflow.Define(wiring, "b", "LeafService")
	a := workflow.Define(wiring, "a", "nonLeafService", b)

	serviceDefaults(wiring, a)
	serviceDefaults(wiring, b)
	golang.CreateProcess(wiring, "proc", a, b)

	// Let's print out all of the nodes currently defined in the wiring spec
	slog.Info("Wiring Spec: \n" + wiring.String())

	bp := wiring.GetBlueprint()
	bp.Instantiate("pa", "pb")
	// bp.Instantiate("proc")

	application, err := bp.Build()
	if err != nil {
		slog.Error("Unable to build blueprint, exiting", "error", err)
		slog.Info("Application: \n" + application.String())
		os.Exit(1)
	}

	slog.Info("Application: \n" + application.String())
}
