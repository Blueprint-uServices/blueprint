package specs

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/healthchecker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/xtrace"
)

var XtraceLogging = wiringcmd.SpecOption{
	Name:        "xtlogging",
	Description: "Showcases how to use xtrace logger with Blueprint application. Deploys each service in a separate container with http, uses mongodb as NoSQL database backends, and applies a number of modifiers.",
	Build:       makextLoggerSpec,
}

func makextLoggerSpec(spec wiring.WiringSpec) ([]string, error) {
	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_ctr := applyXTLoggerDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_ctr := applyXTLoggerDefaults(spec, nonleaf_service)

	return []string{leaf_ctr, nonleaf_ctr}, nil
}

func applyXTLoggerDefaults(spec wiring.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	ctrName := fmt.Sprintf("%s_container", serviceName)
	// opentelemetry.Instrument(spec, serviceName)
	healthchecker.AddHealthCheckAPI(spec, serviceName)
	xtrace.Instrument(spec, serviceName)
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	logger := xtrace.DefineXTraceLogger(spec, procName)
	goproc.AddToProcess(spec, procName, logger)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
