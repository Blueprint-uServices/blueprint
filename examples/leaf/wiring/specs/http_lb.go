package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/loadbalancer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/memcached"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

var HTTP_LoadBalancer = wiringcmd.SpecOption{
	Name:        "http_lb",
	Description: "Deploys each service in a separate process, communicating using HTTP. Leaf service has 2 replicas and NonLeafService chooses between the two at random.",
	Build:       makeHTTPLbSpec,
}

func makeHTTPLbSpec(spec wiring.WiringSpec) ([]string, error) {
	leaf_cache := memcached.PrebuiltContainer(spec, "leaf_cache")
	leaf_db := mongodb.PrebuiltContainer(spec, "leaf_db")

	leaf_service1 := workflow.Define(spec, "leaf_service1", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_proc1 := leaf_service1 + "_process"
	http.Deploy(spec, leaf_service1)
	leaf_proc1 = goproc.CreateProcess(spec, leaf_proc1, leaf_service1)
	leaf_service2 := workflow.Define(spec, "leaf_service2", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_proc2 := leaf_service2 + "_process"
	http.Deploy(spec, leaf_service2)
	leaf_proc2 = goproc.CreateProcess(spec, leaf_proc2, leaf_service2)

	leaf_lb := loadbalancer.Create(spec, []string{leaf_service1, leaf_service2}, "LeafService")

	nonleaf_service := workflow.Define(spec, "nonleaf_service", "NonLeafService", leaf_lb)
	nonleaf_proc := nonleaf_service + "_process"
	http.Deploy(spec, nonleaf_service)
	nonleaf_proc = goproc.CreateProcess(spec, nonleaf_proc, nonleaf_service)
	return []string{leaf_proc1, leaf_proc2, nonleaf_proc}, nil
}
