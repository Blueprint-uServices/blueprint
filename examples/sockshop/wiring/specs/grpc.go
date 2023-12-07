package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/gotests"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// A wiring spec that deploys each service to a separate process, with services communicating over GRPC.
// The user, cart, shipping, and order services use simple in-memory NoSQL databases to store their data.
// The catalogue service uses a simple in-memory sqlite database to store its data.
// The shipping service and queue master service run within the same process (TODO: separate processes)
var GRPC = wiringcmd.SpecOption{
	Name:        "grpc",
	Description: "Deploys each service in a separate process with gRPC.",
	Build:       makeGrpcSpec,
}

func makeGrpcSpec(spec wiring.WiringSpec) ([]string, error) {
	user_db := simple.NoSQLDB(spec, "user_db")
	user_service := workflow.Service(spec, "user_service", "UserService", user_db)
	user_proc := applyGrpcDefaults(spec, user_service, "user_proc")

	payment_service := workflow.Service(spec, "payment_service", "PaymentService", "500")
	payment_proc := applyGrpcDefaults(spec, payment_service, "payment_proc")

	cart_db := simple.NoSQLDB(spec, "cart_db")
	cart_service := workflow.Service(spec, "cart_service", "CartService", cart_db)
	cart_proc := applyGrpcDefaults(spec, cart_service, "cart_proc")

	shipqueue := simple.Queue(spec, "shipping_queue")
	shipdb := simple.NoSQLDB(spec, "shipping_db")
	shipping_service := workflow.Service(spec, "shipping_service", "ShippingService", shipqueue, shipdb)
	shipping_proc := applyGrpcDefaults(spec, shipping_service, "shipping_proc")

	// Deploy queue master to the same process as the shipping proc
	queue_master := workflow.Service(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)
	goproc.AddChildToProcess(spec, shipping_proc, queue_master)

	order_db := simple.NoSQLDB(spec, "order_db")
	order_service := workflow.Service(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)
	order_proc := applyGrpcDefaults(spec, order_service, "order_proc")

	catalogue_db := simple.RelationalDB(spec, "catalogue_db")
	catalogue_service := workflow.Service(spec, "catalogue_service", "CatalogueService", catalogue_db)
	catalogue_proc := applyGrpcDefaults(spec, catalogue_service, "catalogue_proc")

	frontend := workflow.Service(spec, "frontend", "Frontend", user_service, catalogue_service, cart_service, order_service)
	frontend_proc := applyGrpcDefaults(spec, frontend, "frontend_proc")

	tests := gotests.Test(spec, user_service, payment_service, cart_service, shipping_service, order_service, catalogue_service, frontend)

	return []string{user_proc, payment_proc, cart_proc, shipping_proc, order_proc, catalogue_proc, frontend_proc, tests}, nil
}

func applyGrpcDefaults(spec wiring.WiringSpec, serviceName string, procName string) string {
	grpc.Deploy(spec, serviceName)
	return goproc.CreateProcess(spec, procName, serviceName)
}
