package specs

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/latency"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/retries"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/timeouts"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

var TimeoutDemo = wiringcmd.SpecOption{
	Name:        "timeout_demo",
	Description: "Deploys each service in a separate container with gRPC and configures the clients with timeouts and the servers with latency injectors to demonstrate timeouts in blueprint",
	Build:       makeDockerTimeoutSpec,
}

var TimeoutRetriesDemo = wiringcmd.SpecOption{
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
		timeouts.AddTimeouts(spec, serviceName, "100ms")
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
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_ctr := applyDockerTimeoutDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_ctr := applyDockerTimeoutDefaults(spec, nonleaf_service)

	return []string{leaf_ctr, nonleaf_ctr}, nil
}
