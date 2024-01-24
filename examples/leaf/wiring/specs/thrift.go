package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/clientpool"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/thrift"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

// [Thrift] demonstrates how to deploy a service as over RPC using the [thrift] plugin.
//
// [thrift]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/thrift
var Thrift = cmdbuilder.SpecOption{
	Name:        "thrift",
	Description: "Deploys each service in a separate process, communicating using Thrift.",
	Build:       makeThriftSpec,
}

func makeThriftSpec(spec wiring.WiringSpec) ([]string, error) {

	applyThriftDefaults := func(spec wiring.WiringSpec, serviceName string) string {
		clientpool.Create(spec, serviceName, 5)
		thrift.Deploy(spec, serviceName)
		return goproc.Deploy(spec, serviceName)
	}

	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service[*leaf.LeafServiceImpl](spec, "leaf_service", leaf_cache, leaf_db)
	leaf_proc := applyThriftDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_service)
	nonleaf_proc := applyThriftDefaults(spec, nonleaf_service)

	return []string{leaf_proc, nonleaf_proc}, nil
}
