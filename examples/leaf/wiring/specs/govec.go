package specs

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/govector"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

var GoVec = wiringcmd.SpecOption{
	Name:        "govec",
	Description: "Deploys each service in a separate process, communicating using HTTP.  Wraps each service with GoVector instrumentation.",
	Build:       makeGoVecSpec,
}

func makeGoVecSpec(spec wiring.WiringSpec) ([]string, error) {

	leaf_db := mongodb.Container(spec, "leaf_db")
	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_service := workflow.Service(spec, "leaf_service", "LeafServiceImpl", leaf_cache, leaf_db)
	leaf_proc := applyGoVecDefaults(spec, leaf_service)

	nonleaf_service := workflow.Service(spec, "nonleaf_service", "NonLeafService", leaf_service)
	nonleaf_proc := applyGoVecDefaults(spec, nonleaf_service)

	return []string{leaf_proc, nonleaf_proc}, nil
}

func applyGoVecDefaults(spec wiring.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	//cntrName := fmt.Sprintf("%s_cntr", serviceName)
	logger_name := fmt.Sprintf("%s_logger", serviceName)
	logger := govector.DefineLogger(spec, logger_name)
	govector.Instrument(spec, serviceName)
	http.Deploy(spec, serviceName)
	return goproc.CreateProcess(spec, procName, serviceName, logger)
	//return linuxcontainer.CreateContainer(spec, cntrName, procName)
}
