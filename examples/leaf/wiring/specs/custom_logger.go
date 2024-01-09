package specs

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/xtrace"
)

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
