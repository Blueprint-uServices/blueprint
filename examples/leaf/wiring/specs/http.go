package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/opentelemetry"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/zipkin"
)

// [HTTP] demonstrates how to deploy a service as an HTTP webserver using the [http] plugin.
// The wiring spec also instruments services with distributed tracing using the [opentelemetry] plugin.
//
// [http]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/http
// [opentelemetry]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/opentelemetry
var HTTP = cmdbuilder.SpecOption{
	Name:        "http",
	Description: "Deploys each service in a separate process, communicating using HTTP.  Wraps each service in Zipkin tracing.",
	Build:       makeHTTPSpec,
}

func makeHTTPSpec(spec wiring.WiringSpec) ([]string, error) {
	trace_collector := zipkin.Collector(spec, "zipkin")

	applyHTTPDefaults := func(spec wiring.WiringSpec, serviceName string) string {
		opentelemetry.Instrument(spec, serviceName, trace_collector)
		http.Deploy(spec, serviceName)
		return goproc.Deploy(spec, serviceName)
	}

	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service[*leaf.LeafServiceImpl](spec, "leaf_service", leaf_cache, leaf_db)
	leaf_proc := applyHTTPDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_service)
	nonleaf_proc := applyHTTPDefaults(spec, nonleaf_service)

	return []string{leaf_proc, nonleaf_proc}, nil
}
