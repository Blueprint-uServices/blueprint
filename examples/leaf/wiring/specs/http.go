package specs

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/opentelemetry"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/wiringcmd"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/zipkin"
)

var HTTP = wiringcmd.SpecOption{
	Name:        "http",
	Description: "Deploys each service in a separate process, communicating using HTTP.  Wraps each service in Zipkin tracing.",
	Build:       makeHTTPSpec,
}

func makeHTTPSpec(spec wiring.WiringSpec) ([]string, error) {
	trace_collector := zipkin.Collector(spec, "zipkin")

	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_proc := applyHTTPDefaults(spec, leaf_service, trace_collector)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_proc := applyHTTPDefaults(spec, nonleaf_service, trace_collector)

	return []string{leaf_proc, nonleaf_proc}, nil
}

func applyHTTPDefaults(spec wiring.WiringSpec, serviceName string, collectorName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	// ctrName := fmt.Sprintf("%s_container", serviceName)
	opentelemetry.Instrument(spec, serviceName, collectorName)
	http.Deploy(spec, serviceName)
	return goproc.CreateProcess(spec, procName, serviceName)
	//return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
