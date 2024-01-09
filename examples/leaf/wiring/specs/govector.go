package specs

import (
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/govector"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/wiringcmd"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

var Govector = wiringcmd.SpecOption{
	Name:        "govector",
	Description: "Deploys each service in a separate process, communicating using HTTP. Wraps each service in GoVector vector clocks and sets the GoVectorLogger for each process.",
	Build:       makeGoVectorLoggerSpec,
}

func makeGoVectorLoggerSpec(spec wiring.WiringSpec) ([]string, error) {
	applyLoggerDefaults := func(service_name string) string {

		procName := strings.ReplaceAll(service_name, "service", "process")
		var logger string
		govector.Instrument(spec, service_name)
		logger = govector.DefineLogger(spec, procName+"_logger")

		http.Deploy(spec, service_name)
		proc := goproc.CreateProcess(spec, procName, service_name)
		if logger != "" {
			goproc.SetLogger(spec, procName, logger)
		}
		return proc
	}
	leaf_db := simple.NoSQLDB(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_proc := applyLoggerDefaults(leaf_service)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_proc := applyLoggerDefaults(nonleaf_service)

	return []string{leaf_proc, nonleaf_proc}, nil
}
