package main

import (
	"fmt"
	"os"

	"golang.org/x/exp/slog"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/dockerdeployment"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/memcached"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/opentelemetry"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/zipkin"
)

func serviceDefaults(spec wiring.WiringSpec, serviceName string, collectorName string) string {
	procName := fmt.Sprintf("p%s", serviceName)
	//retries.AddRetries(spec, serviceName, 10)
	//healthchecker.AddHealthCheckAPI(spec, serviceName)
	//circuitbreaker.AddCircuitBreaker(spec, serviceName, 1000, 0.1, "1s")
	//xtrace.Instrument(spec, serviceName)
	//opentelemetry.Instrument(spec, serviceName)
	opentelemetry.InstrumentUsingCustomCollector(spec, serviceName, collectorName)
	http.Deploy(spec, serviceName)
	return goproc.CreateProcess(spec, procName, serviceName)
}

func main() {
	slog.Info("Constructing Wiring Spec")

	// Initialize blueprint compiler
	linuxcontainer.RegisterBuilders()
	dockerdeployment.RegisterBuilders()

	spec := wiring.NewWiringSpec("leaf_example")

	workflow.Init("../workflow")

	//b_database := simplenosqldb.Define(spec, "b_database")
	b_database := mongodb.PrebuiltContainer(spec, "b_database")
	//b_cache := simplecache.Define(spec, "b_cache")
	b_cache := memcached.PrebuiltContainer(spec, "b_cache")
	//b_cache := redis.PrebuiltProcess(spec, "b_cache")
	trace_collector := zipkin.DefineZipkinCollector(spec, "zipkin")
	//trace_collector := ""
	b := workflow.Define(spec, "b", "LeafServiceImpl", b_cache, b_database)

	a := workflow.Define(spec, "a", "NonLeafService", b)
	pa := serviceDefaults(spec, a, trace_collector)
	pb := serviceDefaults(spec, b, trace_collector)

	slog.Info("Wiring Spec: \n" + spec.String())

	// Build the IR for our specific nodes
	nodesToInstantiate := []string{pa, pb}
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
