package specs

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/clientpool"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/thrift"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

var Thrift = wiringcmd.SpecOption{
	Name:        "thrift",
	Description: "Deploys each service in a separate process, communicating using Thrift.",
	Build:       makeThriftSpec,
}

func makeThriftSpec(spec wiring.WiringSpec) ([]string, error) {
	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_proc := applyThriftDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_proc := applyThriftDefaults(spec, nonleaf_service)

	return []string{leaf_proc, nonleaf_proc}, nil
}

func applyThriftDefaults(spec wiring.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	clientpool.Create(spec, serviceName, 5)
	thrift.Deploy(spec, serviceName)
	return goproc.CreateProcess(spec, procName, serviceName)
}
