package specs

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplenosqldb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// Used by main.go
var GRPC = wiringcmd.SpecOption{
	Name:        "grpc",
	Description: "Deploys each service in a separate process with gRPC.",
	Build:       makeGrpcSpec,
}

// Creates a basic sockshop wiring spec.
// Returns the names of the nodes to instantiate or an error
func makeGrpcSpec(spec wiring.WiringSpec) ([]string, error) {
	user_db := simplenosqldb.Define(spec, "user_db")
	user_service := workflow.Define(spec, "user_service", "UserService", user_db)
	user_service_proc := applyGrpcDefaults(spec, user_service)

	return []string{user_service_proc}, nil
}

func applyGrpcDefaults(spec wiring.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	grpc.Deploy(spec, serviceName)
	return goproc.CreateProcess(spec, procName, serviceName)
}
