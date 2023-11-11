package main

import (
	"fmt"
	"os"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/clientpool"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplenosqldb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/thrift"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workload"
	"golang.org/x/exp/slog"
)

func serviceDefaults(spec wiring.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("p%s", serviceName)
	clientpool.Create(spec, serviceName, 5)
	thrift.Deploy(spec, serviceName)
	return goproc.CreateProcess(spec, procName, serviceName)
}

func main() {
	fmt.Println("Constructing Wiring Spec")

	spec := wiring.NewWiringSpec("leaf_example")

	// Create the wiring spec
	workflow.Init("../workflow")

	// b_cache := memcached.PrebuiltProcess(spec, "b_cache")
	b_database := simplenosqldb.Define(spec, "b_database")
	b_cache := simplecache.Define(spec, "b_cache")
	b := workflow.Define(spec, "b", "LeafServiceImpl", b_cache, b_database)

	// b := workflow.Define(spec, "b", "LeafServiceImpl")
	// a := workflow.Define(spec, "a", "NonLeafServiceImpl", b) // Will fail, because no constructors returning the impl directly
	a := workflow.Define(spec, "a", "NonLeafService", b)

	pa := serviceDefaults(spec, a)
	pb := serviceDefaults(spec, b)
	// proc := goproc.CreateProcess(spec, "proc", a, b)

	// client := goproc.CreateClientProcess(spec, "client", a)
	client := workload.Generator(spec, a)

	// Let's print out all of the nodes currently defined in the wiring spec
	slog.Info("Wiring Spec: \n" + spec.String())

	// Build the IR for our specific nodes
	nodesToInstantiate := []string{pa, pb, client}
	application, err := spec.BuildIR(nodesToInstantiate...)
	if err != nil {
		slog.Error("Unable to build blueprint, exiting", "error", err)
		slog.Info("Application: \n" + application.String())
		os.Exit(1)
	}

	slog.Info("Application: \n" + application.String())

	err = application.GenerateArtifacts("tmp")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	fmt.Println("Exiting")
}
