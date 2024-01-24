package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/clientpool"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/healthchecker"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

// [Docker] demonstrates how to deploy a service in a Docker container using the [docker] plugin.
// It also wraps services in clientpools using the [clientpool] plugin and and adds a health
// check API using the [healthchecker] plugin.  Services are deployed over RPC using the [grpc] plugin.
//
// [docker]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/docker
// [clientpool]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/clientpool
// [grpc]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/grpc
// [healthchecker]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/healthchecker
var Docker = cmdbuilder.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with gRPC, uses mongodb as NoSQL database backends, and applies a number of modifiers.",
	Build:       makeDockerSpec,
}

func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {

	applyDockerDefaults := func(spec wiring.WiringSpec, serviceName string) string {
		clientpool.Create(spec, serviceName, 5)
		healthchecker.AddHealthCheckAPI(spec, serviceName)
		grpc.Deploy(spec, serviceName)
		goproc.Deploy(spec, serviceName)
		return linuxcontainer.Deploy(spec, serviceName)
	}

	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service[*leaf.LeafServiceImpl](spec, "leaf_service", leaf_cache, leaf_db)
	leaf_ctr := applyDockerDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_service)
	nonleaf_ctr := applyDockerDefaults(spec, nonleaf_service)

	return []string{leaf_ctr, nonleaf_ctr}, nil
}
