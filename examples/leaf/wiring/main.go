package main

import (
	"fmt"
	"os"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplenosqldb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workload"
	"golang.org/x/exp/slog"
)

func serviceDefaults(wiring blueprint.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("p%s", serviceName)
	// opentelemetry.Instrument(wiring, serviceName)
	grpc.Deploy(wiring, serviceName)
	return goproc.CreateProcess(wiring, procName, serviceName)
}

func main() {

	fmt.Println("Constructing Wiring Spec")

	wiring := blueprint.NewWiringSpec("leaf_example")

	// Create the wiring spec
	workflow.Init("../workflow")

	// b_cache := memcached.PrebuiltProcess(wiring, "b_cache")
	b_database := simplenosqldb.Define(wiring, "b_database")
	b_cache := simplecache.Define(wiring, "b_cache")
	b := workflow.Define(wiring, "b", "LeafServiceImpl", b_cache, b_database)

	// b := workflow.Define(wiring, "b", "LeafServiceImpl")
	// a := workflow.Define(wiring, "a", "NonLeafServiceImpl", b) // Will fail, because no constructors returning the impl directly
	a := workflow.Define(wiring, "a", "NonLeafService", b)

	pa := serviceDefaults(wiring, a)
	pb := serviceDefaults(wiring, b)
	// proc := goproc.CreateProcess(wiring, "proc", a, b)

	// client := goproc.CreateClientProcess(wiring, "client", a)
	client := workload.Generator(wiring, a)

	// Let's print out all of the nodes currently defined in the wiring spec
	slog.Info("Wiring Spec: \n" + wiring.String())

	bp, err := wiring.GetBlueprint()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	bp.Instantiate(pa, pb, client)
	// bp.Instantiate(proc)

	application, err := bp.Build()
	if err != nil {
		slog.Error("Unable to build blueprint, exiting", "error", err)
		slog.Info("Application: \n" + application.String())
		os.Exit(1)
	}

	slog.Info("Application: \n" + application.String())

	// Below here is a WIP on generating code
	err = application.Children["pa"].(*goproc.Process).GenerateArtifacts("tmp")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = application.Children["pb"].(*goproc.Process).GenerateArtifacts("tmp")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = application.Children[client].(*goproc.Process).GenerateArtifacts("tmp")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	fmt.Println("Exiting")
}
