package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/clientpool"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/gotests"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/mysql"
	"github.com/blueprint-uservices/blueprint/plugins/opentelemetry"
	"github.com/blueprint-uservices/blueprint/plugins/retries"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/wiringcmd"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/workload"
	"github.com/blueprint-uservices/blueprint/plugins/zipkin"
)

// A wiring spec that deploys each service into its own Docker container and using gRPC to communicate between services.
//
// All RPC calls are retried up to 3 times.
// RPC clients use a client pool with 10 clients.
// All services are instrumented with OpenTelemetry and traces are exported to Zipkin
//
// The user, cart, shipping, and orders services using separate MongoDB instances to store their data.
// The catalogue service uses MySQL to store catalogue data.
// The shipping service and queue master service run within the same process.
var Docker = wiringcmd.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with gRPC, and uses mongodb as NoSQL database backends.",
	Build:       makeDockerSpec,
}

func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	// Define the trace collector, which will be used by all services
	trace_collector := zipkin.Collector(spec, "zipkin")

	// Modifiers that will be applied to all services
	applyDockerDefaults := func(serviceName string) {
		// Golang-level modifiers that add functionality
		retries.AddRetries(spec, serviceName, 3)
		clientpool.Create(spec, serviceName, 10)
		opentelemetry.Instrument(spec, serviceName, trace_collector)
		grpc.Deploy(spec, serviceName)

		// Deploying to namespaces
		goproc.Deploy(spec, serviceName)
		linuxcontainer.Deploy(spec, serviceName)

		// Also add to tests
		gotests.Test(spec, serviceName)
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
	queue_master := workflow.Service(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)
	goproc.AddToProcess(spec, "shipping_proc", queue_master)

	order_db := mongodb.Container(spec, "order_db")
	order_service := workflow.Service(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)
	applyDockerDefaults(order_service)

	catalogue_db := mysql.Container(spec, "catalogue_db")
	catalogue_service := workflow.Service(spec, "catalogue_service", "CatalogueService", catalogue_db)
	applyDockerDefaults(catalogue_service)

	frontend := workflow.Service(spec, "frontend", "Frontend", user_service, catalogue_service, cart_service, order_service)
	applyDockerDefaults(frontend)

	wlgen := workload.Generator(spec, "wlgen", "SimpleWorkload", frontend)

	// Instantiate starting with the frontend which will trigger all other services to be instantiated
	// Also include the tests
	return []string{frontend, wlgen, "gotests"}, nil
}
