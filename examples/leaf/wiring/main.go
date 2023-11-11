package main

import (
	"fmt"
	"os"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/clientpool"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/dockerdeployment"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/healthchecker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplenosqldb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workload"
	"golang.org/x/exp/slog"
)

func serviceDefaults(spec wiring.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("proc%s", serviceName)
	ctrName := fmt.Sprintf("ctr%s", serviceName)
	// opentelemetry.Instrument(spec, serviceName)
	clientpool.Create(spec, serviceName, 5)
	healthchecker.AddHealthCheckAPI(spec, serviceName)
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func main() {

	fmt.Println("Constructing spec Spec")

	// Initialize blueprint compiler
	linuxcontainer.RegisterBuilders()
	dockerdeployment.RegisterBuilders()

	// --------------

	spec := wiring.NewWiringSpec("leaf_example")

	// Create the wiring spec
	workflow.Init("../workflow")

	b_database := simplenosqldb.Define(spec, "b_database")
	b_cache := simplecache.Define(spec, "b_cache")
	b := workflow.Define(spec, "b", "LeafServiceImpl", b_cache, b_database)

	a := workflow.Define(spec, "a", "NonLeafService", b)

	pa := serviceDefaults(spec, a)
	pb := serviceDefaults(spec, b)

	client := workload.Generator(spec, a)

	// -----------------

	slog.Info("spec Spec: \n" + spec.String())

	// Build the IR for our specific nodes
	nodesToInstantiate := []string{pa, pb, client}
	application, err := spec.BuildIR(nodesToInstantiate...)
	if err != nil {
		slog.Error("Unable to build blueprint, exiting", "error", err)
		slog.Info("Application: \n" + application.String())
		os.Exit(1)
	}

	slog.Info("Application: \n" + application.String())

	// Below here is a WIP on generating code

	err = application.GenerateArtifacts("tmp")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	fmt.Println("Exiting")
}
