package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/loadbalancer"
	"github.com/blueprint-uservices/blueprint/plugins/memcached"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

var HTTP_LoadBalancer = cmdbuilder.SpecOption{
	Name:        "http_lb",
	Description: "Deploys each service in a separate process, communicating using HTTP. Leaf service has 2 replicas and NonLeafService chooses between the two at random.",
	Build:       makeHTTPLbSpec,
}

func makeHTTPLbSpec(spec wiring.WiringSpec) ([]string, error) {
	leaf_cache := memcached.Container(spec, "leaf_cache")
	leaf_db := mongodb.Container(spec, "leaf_db")

	leaf_service1 := workflow.Service[*leaf.LeafServiceImpl](spec, "leaf1_service", leaf_cache, leaf_db)
	leaf_proc1 := leaf_service1 + "_process"
	http.Deploy(spec, leaf_service1)
	leaf_proc1 = goproc.CreateProcess(spec, leaf_proc1, leaf_service1)
	leaf_cntr1 := linuxcontainer.Deploy(spec, leaf_service1)
	leaf_service2 := workflow.Service[*leaf.LeafServiceImpl](spec, "leaf2_service", leaf_cache, leaf_db)
	leaf_proc2 := leaf_service2 + "_process"
	http.Deploy(spec, leaf_service2)
	leaf_proc2 = goproc.CreateProcess(spec, leaf_proc2, leaf_service2)
	leaf_cntr2 := linuxcontainer.Deploy(spec, leaf_service2)

	leaf_lb := loadbalancer.Create(spec, "LeafServices", []string{leaf_service1, leaf_service2})

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_lb)
	nonleaf_proc := nonleaf_service + "_process"
	http.Deploy(spec, nonleaf_service)
	nonleaf_proc = goproc.CreateProcess(spec, nonleaf_proc, nonleaf_service)
	nonleaf_cntr := linuxcontainer.Deploy(spec, nonleaf_service)
	return []string{leaf_cntr1, leaf_cntr2, nonleaf_cntr}, nil
}
