package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/gotests"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplenosqldb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplequeue"
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

	cart_db := simplenosqldb.Define(spec, "cart_db")
	cart_service := workflow.Define(spec, "cart_service", "CartService", cart_db)
	cart_proc := applyGrpcDefaults(spec, cart_service, "cart_proc")

	shipqueue := simplequeue.Define(spec, "shipping_queue")
	shipdb := simplenosqldb.Define(spec, "shipping_db")
	shipping_service := workflow.Define(spec, "shipping_service", "ShippingService", shipqueue, shipdb)
	shipping_proc := applyGrpcDefaults(spec, shipping_service, "shipping_proc")

	// Deploy queue master to the same process as the shipping proc
	queue_master := workflow.Define(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)
	goproc.AddChildToProcess(spec, shipping_proc, queue_master)

	order_db := simplenosqldb.Define(spec, "order_db")
	order_service := workflow.Define(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)
	order_proc := applyGrpcDefaults(spec, order_service, "order_proc")

	tests := gotests.Test(spec, user_service, payment_service, cart_service, shipping_service, order_service)

	return []string{user_proc, payment_proc, cart_proc, shipping_proc, order_proc, tests}, nil
}

func applyGrpcDefaults(spec wiring.WiringSpec, serviceName string, procName string) string {
	grpc.Deploy(spec, serviceName)
	return goproc.CreateProcess(spec, procName, serviceName)
}
