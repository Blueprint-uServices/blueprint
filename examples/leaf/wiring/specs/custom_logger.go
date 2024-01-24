package specs

import (
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/jaeger"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/opentelemetry"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/xtrace"
)

// [Xtrace_Logger] demonstrates how to use a custom logger for Blueprint applications.
// The wiring spec uses xtrace logger as the custom logger for demonstration.
// Each service is deployed in a separate container with services communicating using HTTP.
// Launches an XTrace server with each service wrapped in xtrace tracing.
var Xtrace_Logger = cmdbuilder.SpecOption{
	Name:        "xtrace_logger",
	Description: "Deploys each service in a separate container, communicating using HTTP. Wraps each service in XTrace tracing and sets the XTraceLogger for each process.",
	Build:       makeXTraceLoggerSpec,
}

// [OT_Logger] demonstrates how to use a custom logger for Blueprint applications.
// The wiring spec uses opentelemetry logger as the custom logger for demonstration.
// Each service is deployed in a separate container with services communicating using HTTP.
// Launches an jaeger server which collects spans generated from each service wrapped in opentelemetry tracing.
var OT_Logger = cmdbuilder.SpecOption{
	Name:        "ot_logger",
	Description: "Deploys each service in a separate container, communicating using HTTP. Wraps each service in opentelemetry tracing and sets the OTLogger for each process. All spans are collected by the jaeger collector.",
	Build:       makeOTLoggerSpec,
}

func makeOTLoggerSpec(spec wiring.WiringSpec) ([]string, error) {
	return makeCustomLoggerSpec(spec, "ot")
}

func makeXTraceLoggerSpec(spec wiring.WiringSpec) ([]string, error) {
	return makeCustomLoggerSpec(spec, "xtrace")
}

func makeCustomLoggerSpec(spec wiring.WiringSpec, logger_type string) ([]string, error) {
	var collector string
	if logger_type == "ot" {
		collector = jaeger.Collector(spec, "jaeger")
	}
	applyLoggerDefaults := func(service_name string) string {

		procName := strings.ReplaceAll(service_name, "service", "process")
		if logger_type == "xtrace" {
			xtrace.Instrument(spec, service_name)
		} else if logger_type == "ot" {
			opentelemetry.Instrument(spec, service_name, collector)
		}
		cntrName := strings.ReplaceAll(service_name, "service", "container")
		http.Deploy(spec, service_name)
		goproc.CreateProcess(spec, procName, service_name)
		if logger_type == "xtrace" {
			xtrace.Logger(spec, procName)
		} else if logger_type == "ot" {
			opentelemetry.Logger(spec, procName)
		}
		return linuxcontainer.CreateContainer(spec, cntrName, procName)
	}
	leaf_db := simple.NoSQLDB(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service[*leaf.LeafServiceImpl](spec, "leaf_service", leaf_cache, leaf_db)
	leaf_proc := applyLoggerDefaults(leaf_service)

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_service)
	nonleaf_proc := applyLoggerDefaults(nonleaf_service)

	return []string{leaf_proc, nonleaf_proc}, nil
}
