package specs

import (
	"fmt"

	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/wiring"
	"github.com/Blueprint-uServices/blueprint/plugins/clientpool"
	"github.com/Blueprint-uServices/blueprint/plugins/goproc"
	"github.com/Blueprint-uServices/blueprint/plugins/grpc"
	"github.com/Blueprint-uServices/blueprint/plugins/healthchecker"
	"github.com/Blueprint-uServices/blueprint/plugins/linuxcontainer"
	"github.com/Blueprint-uServices/blueprint/plugins/mongodb"
	"github.com/Blueprint-uServices/blueprint/plugins/simple"
	"github.com/Blueprint-uServices/blueprint/plugins/wiringcmd"
	"github.com/Blueprint-uServices/blueprint/plugins/workflow"
)

var Docker = wiringcmd.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with gRPC, uses mongodb as NoSQL database backends, and applies a number of modifiers.",
	Build:       makeDockerSpec,
}

func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_ctr := applyDockerDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_ctr := applyDockerDefaults(spec, nonleaf_service)

	return []string{leaf_ctr, nonleaf_ctr}, nil
}

func applyDockerDefaults(spec wiring.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	ctrName := fmt.Sprintf("%s_container", serviceName)
	// opentelemetry.Instrument(spec, serviceName)
	clientpool.Create(spec, serviceName, 5)
	healthchecker.AddHealthCheckAPI(spec, serviceName)
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
