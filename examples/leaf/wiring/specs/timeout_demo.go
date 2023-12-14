package specs

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/healthchecker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/latencyinjector"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
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

func makeDockerTimeoutSpec(spec wiring.WiringSpec) ([]string, error) {
	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_ctr := applyDockerDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_ctr := applyDockerTimeoutDefaults(spec, nonleaf_service)

	return []string{leaf_ctr, nonleaf_ctr}, nil
}

func applyDockerTimeoutDefaults(spec wiring.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	ctrName := fmt.Sprintf("%s_container", serviceName)
	// opentelemetry.Instrument(spec, serviceName)
	timeouts.AddTimeouts(spec, serviceName, "100ms")
	latencyinjector.AddFixedLatency(spec, serviceName, "200ms")
	healthchecker.AddHealthCheckAPI(spec, serviceName)
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
