package specs

import (
	"flag"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/kubernetes"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

var Kubernetes = cmdbuilder.SpecOption{
	Name:        "kubernetes",
	Description: "Deploys each service as a Kubernetes Service connected with gRPC, uses mongodb as NoSQL database backends, and applies a number of modifiers",
	Build:       makeKubernetesSpec,
}

var regAddr = flag.String("registry", "", "Address at which docker registry is hosted that will be used by Kubernetes")

func makeKubernetesSpec(spec wiring.WiringSpec) ([]string, error) {
	applyKubeDefaults := func(spec wiring.WiringSpec, serviceName string) string {
		grpc.Deploy(spec, serviceName)
		goproc.Deploy(spec, serviceName)
		return linuxcontainer.Deploy(spec, serviceName)
	}

	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service[*leaf.LeafServiceImpl](spec, "leaf_service", leaf_cache, leaf_db)
	leaf_ctr := applyKubeDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_service)
	nonleaf_ctr := applyKubeDefaults(spec, nonleaf_service)

	kube_app := kubernetes.NewApplication(spec, "leaf", *regAddr, leaf_db, leaf_ctr, nonleaf_ctr)
	return []string{kube_app}, nil
}
