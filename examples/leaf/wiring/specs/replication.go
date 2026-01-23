package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/memcached"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/replication"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

var Replication = cmdbuilder.SpecOption{
	Name:        "replication",
	Description: "Deploys each service in a separate container with http, uses mongodb as NoSQL database backends, and applies a number of modifiers",
	Build:       makeLoadBalancerSpec,
}

func makeLoadBalancerSpec(spec wiring.WiringSpec) ([]string, error) {
	applyLoadBalancerDefaults := func(spec wiring.WiringSpec, serviceName string) string {
		http.Deploy(spec, serviceName)
		goproc.Deploy(spec, serviceName)
		return linuxcontainer.Deploy(spec, serviceName)
	}

	cntrs := []string{}

	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := memcached.Container(spec, "leaf_cache")
	leaf_services, leaf_service_lb := replication.Replicate[*leaf.LeafServiceImpl](spec, "leaf_service", 2, leaf_cache, leaf_db)
	for _, leaf_service := range leaf_services {
		leaf_ctr := applyLoadBalancerDefaults(spec, leaf_service)
		cntrs = append(cntrs, leaf_ctr)
	}
	lb_ctr := applyLoadBalancerDefaults(spec, leaf_service_lb)

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_service_lb)
	nonleaf_ctr := applyLoadBalancerDefaults(spec, nonleaf_service)
	cntrs = append(cntrs, nonleaf_ctr, lb_ctr)

	return cntrs, nil
}
