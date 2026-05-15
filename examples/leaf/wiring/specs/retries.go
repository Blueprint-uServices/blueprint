package specs

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/faultinjector"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/retries"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

var RetriesDemo = cmdbuilder.SpecOption{
	Name:        "retries",
	Description: "Deploys each service in a separate container with http and configures the clients with retries token buckets in blueprint",
	Build:       makeRetriesSpec,
}

func makeRetriesSpec(spec wiring.WiringSpec) ([]string, error) {
	applyRetryDefaults := func(spec wiring.WiringSpec, serviceName string) string {
		procName := fmt.Sprintf("%s_process", serviceName)
		ctrName := fmt.Sprintf("%s_container", serviceName)
		retries.AddRetriesTokenBucket(spec, serviceName, 10.0, 1.0, 0.5)
		if serviceName == "leaf_service" {
			faultinjector.AddProbabilisticFailures(spec, serviceName, 90)
		}
		http.Deploy(spec, serviceName)
		goproc.CreateProcess(spec, procName, serviceName)
		return linuxcontainer.CreateContainer(spec, ctrName, procName)
	}

	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service[*leaf.LeafServiceImpl](spec, "leaf_service", leaf_cache, leaf_db)
	leaf_ctr := applyRetryDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service[leaf.NonLeafService](spec, "nonleaf_service", leaf_service)
	nonleaf_ctr := applyRetryDefaults(spec, nonleaf_service)

	return []string{leaf_ctr, nonleaf_ctr}, nil
}
