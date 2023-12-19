package specs

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/clientpool"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/opentelemetry"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/retries"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/zipkin"
)

// A wiring spec that deploys each service into its own Docker container and using gRPC to communicate between services.
// All RPC calls are retried up to 3 times.  RPC clients use a client pool with 10 clients.
// All services are instrumented with OpenTelemetry and traces are exported to Zipkin
// The user, cart, shipping, and orders services using separate MongoDB instances to store their data.
// The catalogue service uses MySQL to store catalogue data.
// The shipping service and queue master service run within the same process (TODO: separate processes)
var Docker = wiringcmd.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with gRPC, and uses mongodb as NoSQL database backends.",
	Build:       makeDockerSpec,
}

func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	var allServices []string
	var allCtrs []string

	// Define the trace collector, which will be used by all services
	trace_collector := zipkin.Collector(spec, "zipkin")

	// Modifiers that will be applied to all services
	applyDockerDefaults := func(serviceName string) string {
		name, _ := strings.CutSuffix(serviceName, "_service")

		// Apply Blueprint modifiers
		retries.AddRetries(spec, serviceName, 3)
		clientpool.Create(spec, serviceName, 10)
		opentelemetry.InstrumentUsingCustomCollector(spec, serviceName, trace_collector)
		grpc.Deploy(spec, serviceName)
		proc := goproc.CreateProcess(spec, name+"_proc", serviceName)
		ctr := linuxcontainer.CreateContainer(spec, name+"_ctr", proc)

		// Save the service and ctr for later instantiation
		allServices = append(allServices, serviceName)
		allCtrs = append(allCtrs, ctr)
		return ctr
	}

	user_db := mongodb.Container(spec, "user_db")
	user_service := workflow.Service(spec, "user_service", "UserService", user_db)
	applyDockerDefaults(user_service)

	payment_service := workflow.Service(spec, "payment_service", "PaymentService", "500")
	applyDockerDefaults(payment_service)

	cart_db := mongodb.Container(spec, "cart_db")
	cart_service := workflow.Service(spec, "cart_service", "CartService", cart_db)
	applyDockerDefaults(cart_service)

	shipqueue := simple.Queue(spec, "shipping_queue")
	shipdb := mongodb.Container(spec, "shipping_db")
	shipping_service := workflow.Service(spec, "shipping_service", "ShippingService", shipqueue, shipdb)
	applyDockerDefaults(shipping_service)

	// Deploy queue master to the same process as the shipping proc
	// TODO: after distributed queue is supported, move to separate containers
	// queue_master := workflow.Service(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)
	// goproc.AddToProcess(spec, "shipping_proc", queue_master)

	order_db := mongodb.Container(spec, "order_db")
	order_service := workflow.Service(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)
	applyDockerDefaults(order_service)

	// catalogue_db := mysql.Container(spec, "catalogue_db")
	// catalogue_service := workflow.Service(spec, "catalogue_service", "CatalogueService", catalogue_db)
	// applyDockerDefaults(catalogue_service)

	// frontend := workflow.Service(spec, "frontend", "Frontend", user_service, catalogue_service, cart_service, order_service)
	// applyDockerDefaults(frontend)

	// tests := gotests.Test(spec, allServices...)
	// allCtrs = append(allCtrs, tests)

	return allCtrs, nil
}
