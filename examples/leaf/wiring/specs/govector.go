package specs

import (
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/govector"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

// [Govector] demonstrates how to instrument an application using GoVector to propagate vector clocks and to create logs with vector clocks.
// The wiring spec uses govector logger as the custom logger for processes.
// Each service is deployed in a separate process with services communicating using HTTP.
var Govector = cmdbuilder.SpecOption{
	Name:        "govector",
	Description: "Deploys each service in a separate process, communicating using HTTP. Wraps each service in GoVector vector clocks and sets the GoVectorLogger for each process.",
	Build:       makeGoVectorLoggerSpec,
}

func makeGoVectorLoggerSpec(spec wiring.WiringSpec) ([]string, error) {
	applyLoggerDefaults := func(service_name string) string {

		procName := strings.ReplaceAll(service_name, "service", "process")
		govector.Instrument(spec, service_name)

		http.Deploy(spec, service_name)
		proc := goproc.CreateProcess(spec, procName, service_name)
		govector.Logger(spec, procName)
		return proc
	}
	leaf_db := simple.NoSQLDB(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service[*leaf.LeafService](spec, "leaf_service", leaf_cache, leaf_db)
	leaf_proc := applyLoggerDefaults(leaf_service)

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_service)
	nonleaf_proc := applyLoggerDefaults(nonleaf_service)

	return []string{leaf_proc, nonleaf_proc}, nil
}
