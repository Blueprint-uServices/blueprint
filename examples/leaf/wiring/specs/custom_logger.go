package specs

import (
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/wiringcmd"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/xtrace"
)

// Wiring specification demonstrates how to use a custom logger for Blueprint applications.
// The wiring spec uses xtrace logger as the custom logger for demonstration.
// Each service is deployed in a separate container with services communicating using HTTP.
// Launches an XTrace server with each service wrapped in xtrace tracing.
var Xtrace_Logger = wiringcmd.SpecOption{
	Name:        "xtrace_logger",
	Description: "Deploys each service in a separate container, communicating using HTTP. Wraps each service in XTrace tracing and sets the XTraceLogger for each process.",
	Build:       makeXTraceLoggerSpec,
}

func makeXTraceLoggerSpec(spec wiring.WiringSpec) ([]string, error) {
	return makeCustomLoggerSpec(spec, "xtrace")
}

func makeCustomLoggerSpec(spec wiring.WiringSpec, logger_type string) ([]string, error) {
	applyLoggerDefaults := func(service_name string) string {

		procName := strings.ReplaceAll(service_name, "service", "process")
		var logger string
		if logger_type == "xtrace" {
			xtrace.Instrument(spec, service_name)
			logger = xtrace.DefineXTraceLogger(spec, procName)
		}
		cntrName := strings.ReplaceAll(service_name, "service", "container")
		http.Deploy(spec, service_name)
		goproc.CreateProcess(spec, procName, service_name)
		if logger != "" {
			goproc.SetLogger(spec, procName, logger)
		}
		return linuxcontainer.CreateContainer(spec, cntrName, procName)
	}
	leaf_db := simple.NoSQLDB(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_proc := applyLoggerDefaults(leaf_service)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_proc := applyLoggerDefaults(nonleaf_service)

	return []string{leaf_proc, nonleaf_proc}, nil
}
