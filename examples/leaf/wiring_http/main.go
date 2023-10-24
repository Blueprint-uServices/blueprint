package main

import (
	"fmt"
	"os"

	"golang.org/x/exp/slog"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/dockerdeployment"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/opentelemetry"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func serviceDefaults(wiring blueprint.WiringSpec, serviceName string, collectorName string) string {
	procName := fmt.Sprintf("p%s", serviceName)
	//retries.AddRetries(wiring, serviceName, 10)
	//healthchecker.AddHealthCheckAPI(wiring, serviceName)
	//circuitbreaker.AddCircuitBreaker(wiring, serviceName, 1000, 0.1, "1s")
	//xtrace.Instrument(wiring, serviceName)
	opentelemetry.Instrument(wiring, serviceName)
	//opentelemetry.InstrumentUsingCustomCollector(wiring, serviceName, collectorName)
	http.Deploy(wiring, serviceName)
	return goproc.CreateProcess(wiring, procName, serviceName)
}

func main() {
	slog.Info("Constructing Wiring Spec")

	// Initialize blueprint compiler
	linuxcontainer.RegisterBuilders()
	dockerdeployment.RegisterBuilders()

	wiring := blueprint.NewWiringSpec("leaf_example")

	workflow.Init("../workflow")

	//b_database := simplenosqldb.Define(wiring, "b_database")
	b_database := mongodb.PrebuiltProcess(wiring, "b_database")
	b_cache := simplecache.Define(wiring, "b_cache")
	//b_cache := memcached.PrebuiltProcess(wiring, "b_cache")
	//b_cache := redis.PrebuiltProcess(wiring, "b_cache")
	//trace_collector := zipkin.DefineZipkinCollector(wiring, "zipkin")
	trace_collector := ""
	b := workflow.Define(wiring, "b", "LeafServiceImpl", b_cache, b_database)

	a := workflow.Define(wiring, "a", "NonLeafService", b)
	pa := serviceDefaults(wiring, a, trace_collector)
	pb := serviceDefaults(wiring, b, trace_collector)

	slog.Info("Wiring Spec: \n" + wiring.String())

	bp, err := wiring.GetBlueprint()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	bp.Instantiate(pa, pb)

	application, err := bp.BuildIR()
	if err != nil {
		slog.Error("Unable to build blueprint, exiting", "error", err)
		slog.Info("Application: \n" + application.String())
		os.Exit(1)
	}

	slog.Info("Application: \n" + application.String())

	// Below here is a WIP on generating code

	err = application.Compile("tmp")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	fmt.Println("Exiting")
}
