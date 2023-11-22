package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/gotests"
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
	user_proc := applyGrpcDefaults(spec, user_service, "user_proc")

	payment_service := workflow.Define(spec, "payment_service", "PaymentService")
	payment_proc := applyGrpcDefaults(spec, payment_service, "payment_proc")

	tests := gotests.Test(spec, user_service, payment_service)

	return []string{user_proc, payment_proc, tests}, nil
}

func applyGrpcDefaults(spec wiring.WiringSpec, serviceName string, procName string) string {
	grpc.Deploy(spec, serviceName)
	return goproc.CreateProcess(spec, procName, serviceName)
}
