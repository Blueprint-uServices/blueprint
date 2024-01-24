package specs

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/latency"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/retries"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/timeouts"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

// A wiring spec that demonstrates how to add timeouts to a blueprint application.
// The spec deploys each service in a separate container.
// The services use GRPC to communicate with each other.
// Server side of each service is configured with a latency injector which adds a fixed amount of latency for every request.
// Client side for each service is configured with timeouts.
// All requests in the generated system with this wiring specification result in a TimeOut error.
var TimeoutDemo = cmdbuilder.SpecOption{
	Name:        "timeout_demo",
	Description: "Deploys each service in a separate container with gRPC and configures the clients with timeouts and the servers with latency injectors to demonstrate timeouts in blueprint",
	Build:       makeDockerTimeoutSpec,
}

// A wiring spec that demonstrates how to add timeouts with retries to a blueprint application.
// The spec deploys each service in a separate container.
// The services use GRPC to communicate with each other.
// Server side of each service is configured with a latency injector which adds a fixed amount of latency for every request.
// Client side for each service is configured with retries where each separate request results in a timeout.
// All requests in the generated system with this wiring specification result in a TimeOut error.
var TimeoutRetriesDemo = cmdbuilder.SpecOption{
	Name:        "timeout_retries_demo",
	Description: "Deploys each service in a separate container with gRPC and configures the clients with both retries and timeouts and the servers with latency injectors to demonstrate timeouts in blueprint",
	Build:       makeDockerTimeoutRetriesSpec,
}

func makeDockerTimeoutSpec(spec wiring.WiringSpec) ([]string, error) {
	return makeDockerTimeoutSpecGeneric(spec, false)
}

func makeDockerTimeoutRetriesSpec(spec wiring.WiringSpec) ([]string, error) {
	return makeDockerTimeoutSpecGeneric(spec, true)
}

func makeDockerTimeoutSpecGeneric(spec wiring.WiringSpec, use_retries bool) ([]string, error) {
	applyDockerTimeoutDefaults := func(spec wiring.WiringSpec, serviceName string) string {
		procName := fmt.Sprintf("%s_process", serviceName)
		ctrName := fmt.Sprintf("%s_container", serviceName)
		timeouts.Add(spec, serviceName, "100ms")
		if use_retries {
			retries.AddRetriesWithTimeouts(spec, serviceName, 10, "100ms")
		}
		latency.AddFixed(spec, serviceName, "200ms")
		http.Deploy(spec, serviceName)
		goproc.CreateProcess(spec, procName, serviceName)
		return linuxcontainer.CreateContainer(spec, ctrName, procName)
	}
	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service[*leaf.LeafServiceImpl](spec, "leaf_service", leaf_cache, leaf_db)
	leaf_ctr := applyDockerTimeoutDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_service)
	nonleaf_ctr := applyDockerTimeoutDefaults(spec, nonleaf_service)

	return []string{leaf_ctr, nonleaf_ctr}, nil
}
