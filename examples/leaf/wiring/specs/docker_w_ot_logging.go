package specs

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/healthchecker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/opentelemetry"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/zipkin"
)

var OTtraceLogging = wiringcmd.SpecOption{
	Name:        "otlogging",
	Description: "Showcases how to use opentelemetry trace logger with Blueprint applications. Deploys each service in a separate container with http, uses mongodb as NoSQL database backends, and applies a number of modifiers.",
	Build:       makeotLoggerSpec,
}

func makeotLoggerSpec(spec wiring.WiringSpec) ([]string, error) {
	collector := zipkin.Collector(spec, "zipkin")
	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_ctr := applyOTLoggerDefaults(spec, leaf_service, collector)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_ctr := applyOTLoggerDefaults(spec, nonleaf_service, collector)

	return []string{leaf_ctr, nonleaf_ctr}, nil
}

func applyOTLoggerDefaults(spec wiring.WiringSpec, serviceName string, collectorName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	ctrName := fmt.Sprintf("%s_container", serviceName)
	healthchecker.AddHealthCheckAPI(spec, serviceName)
	opentelemetry.InstrumentUsingCustomCollector(spec, serviceName, collectorName)
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	logger := opentelemetry.DefineOTTraceLogger(spec, procName)
	goproc.AddToProcess(spec, procName, logger)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
