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
	"gitlab.mpi-sws.org/cld/blueprint/plugins/rabbitmq"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/retries"
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
var DockerRabbit = wiringcmd.SpecOption{
	Name:        "rabbit",
	Description: "Deploys each service in a separate container with gRPC, and uses mongodb as NoSQL database backends and rabbitmq as the queue backend.",
	Build:       makeDockerRabbitSpec,
}

func makeDockerRabbitSpec(spec wiring.WiringSpec) ([]string, error) {
	// Define the trace collector, which will be used by all services
	trace_collector := zipkin.Collector(spec, "zipkin")

	// Modifiers that will be applied to all services
	applyDockerDefaults := func(serviceName string) {
		// Golang-level modifiers that add functionality
		retries.AddRetries(spec, serviceName, 3)
		clientpool.Create(spec, serviceName, 10)
		opentelemetry.InstrumentUsingCustomCollector(spec, serviceName, trace_collector)
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

	shipqueue := rabbitmq.Container(spec, "shipping_queue", "shippingq")
	shipdb := mongodb.Container(spec, "shipping_db")
	shipping_service := workflow.Service(spec, "shipping_service", "ShippingService", shipqueue, shipdb)
	applyDockerDefaults(shipping_service)

	queue_master := workflow.Service(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)
	applyDockerDefaults(queue_master)

	order_db := mongodb.Container(spec, "order_db")
	order_service := workflow.Service(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)
	applyDockerDefaults(order_service)

	catalogue_db := mysql.Container(spec, "catalogue_db")
	catalogue_service := workflow.Service(spec, "catalogue_service", "CatalogueService", catalogue_db)
	applyDockerDefaults(catalogue_service)

	frontend := workflow.Service(spec, "frontend", "Frontend", user_service, catalogue_service, cart_service, order_service)
	applyDockerDefaults(frontend)

	// Instantiate starting with the frontend which will trigger all other services to be instantiated
	// Also include the tests
	return []string{frontend, "gotests"}, nil
}
