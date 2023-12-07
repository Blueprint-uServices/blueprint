package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/clientpool"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/gotests"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mysql"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/opentelemetry"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/retries"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// A wiring spec that deploys each service into its own Docker container and using gRPC to communicate between services.
// The user, cart, shipping, and orders services using separate MongoDB instances to store their data.
// The catalogue service uses MySQL to store catalogue data.
// The shipping service and queue master service run within the same process (TODO: separate processes)
var Docker = wiringcmd.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with gRPC, and uses mongodb as NoSQL database backends.",
	Build:       makeDockerSpec,
}

// Creates a basic sockshop wiring spec.
// Returns the names of the nodes to instantiate or an error
func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	user_db := mongodb.Container(spec, "user_db")
	user_service := workflow.Service(spec, "user_service", "UserService", user_db)
	user_ctr := applyDockerDefaults(spec, user_service, "user_proc", "user_container")

	payment_service := workflow.Service(spec, "payment_service", "PaymentService", "500")
	payment_ctr := applyDockerDefaults(spec, payment_service, "payment_proc", "payment_container")

	cart_db := mongodb.Container(spec, "cart_db")
	cart_service := workflow.Service(spec, "cart_service", "CartService", cart_db)
	cart_ctr := applyDockerDefaults(spec, cart_service, "cart_proc", "cart_ctr")

	shipqueue := simple.Queue(spec, "shipping_queue")
	shipdb := mongodb.Container(spec, "shipping_db")
	shipping_service := workflow.Service(spec, "shipping_service", "ShippingService", shipqueue, shipdb)
	shipping_ctr := applyDockerDefaults(spec, shipping_service, "shipping_proc", "shipping_ctr")

	// Deploy queue master to the same process as the shipping proc
	// TODO: after distributed queue is supported, move to separate containers
	queue_master := workflow.Service(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)
	goproc.AddChildToProcess(spec, "shipping_proc", queue_master)

	order_db := mongodb.Container(spec, "order_db")
	order_service := workflow.Service(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)
	order_ctr := applyDockerDefaults(spec, order_service, "order_proc", "order_ctr")

	catalogue_db := mysql.Container(spec, "catalogue_db")
	catalogue_service := workflow.Service(spec, "catalogue_service", "CatalogueService", catalogue_db)
	catalogue_ctr := applyDockerDefaults(spec, catalogue_service, "catalogue_proc", "catalogue_ctr")

	frontend := workflow.Service(spec, "frontend", "Frontend", user_service, catalogue_service, cart_service, order_service)
	frontend_ctr := applyDockerDefaults(spec, frontend, "frontend_proc", "frontend_ctr")

	tests := gotests.Test(spec, user_service, payment_service, cart_service, shipping_service, order_service, catalogue_service, frontend)

	return []string{user_ctr, payment_ctr, cart_ctr, shipping_ctr, order_ctr, catalogue_ctr, frontend_ctr, tests}, nil
}

func applyDockerDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	retries.AddRetries(spec, serviceName, 3)
	clientpool.Create(spec, serviceName, 10)
	opentelemetry.Instrument(spec, serviceName)
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
