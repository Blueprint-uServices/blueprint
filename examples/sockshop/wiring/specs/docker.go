package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/gotests"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// Used by main.go
var Docker = wiringcmd.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with gRPC, and uses mongodb as NoSQL database backends.",
	Build:       makeDockerSpec,
}

// Creates a basic sockshop wiring spec.
// Returns the names of the nodes to instantiate or an error
func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	user_db := mongodb.PrebuiltContainer(spec, "user_db")
	user_service := workflow.Define(spec, "user_service", "UserService", user_db)
	user_ctr := applyDockerDefaults(spec, user_service, "user_proc", "user_container")

	payment_service := workflow.Define(spec, "payment_service", "PaymentService")
	payment_ctr := applyDockerDefaults(spec, payment_service, "payment_proc", "payment_container")

	tests := gotests.Test(spec, user_service, payment_service)

	return []string{user_ctr, payment_ctr, tests}, nil
}

func applyDockerDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
