package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/clientpool"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/gotests"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/retries"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/workload"
)

// A wiring spec that deploys each service to a separate process, with services communicating over GRPC.
// The user, cart, shipping, and order services use simple in-memory NoSQL databases to store their data.
// The catalogue service uses a simple in-memory sqlite database to store its data.
// The shipping service and queue master service run within the same process (TODO: separate processes)
var GRPC = cmdbuilder.SpecOption{
	Name:        "grpc",
	Description: "Deploys each service in a separate process with gRPC.",
	Build:       makeGrpcSpec,
}

func makeGrpcSpec(spec wiring.WiringSpec) ([]string, error) {

	// Modifiers that will be applied to all services
	applyDefaults := func(serviceName string, useHTTP ...bool) {
		// Golang-level modifiers that add functionality
		retries.AddRetries(spec, serviceName, 3)
		clientpool.Create(spec, serviceName, 10)
		if len(useHTTP) > 0 && useHTTP[0] {
			http.Deploy(spec, serviceName)
		} else {
			grpc.Deploy(spec, serviceName)
		}

		// Deploying to namespaces
		goproc.Deploy(spec, serviceName)

		// Also add to tests
		gotests.Test(spec, serviceName)
	}

	user_db := simple.NoSQLDB(spec, "user_db")
	user_service := workflow.Service(spec, "user_service", "UserService", user_db)
	applyDefaults(user_service)

	payment_service := workflow.Service(spec, "payment_service", "PaymentService", "500")
	applyDefaults(payment_service)

	cart_db := simple.NoSQLDB(spec, "cart_db")
	cart_service := workflow.Service(spec, "cart_service", "CartService", cart_db)
	applyDefaults(cart_service)

	shipqueue := simple.Queue(spec, "shipping_queue")
	shipdb := simple.NoSQLDB(spec, "shipping_db")
	shipping_service := workflow.Service(spec, "shipping_service", "ShippingService", shipqueue, shipdb)
	applyDefaults(shipping_service)

	// Deploy queue master to the same process as the shipping proc
	queue_master := workflow.Service(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)
	goproc.AddToProcess(spec, "shipping_proc", queue_master)

	order_db := simple.NoSQLDB(spec, "order_db")
	order_service := workflow.Service(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)
	applyDefaults(order_service)

	catalogue_db := simple.RelationalDB(spec, "catalogue_db")
	catalogue_service := workflow.Service(spec, "catalogue_service", "CatalogueService", catalogue_db)
	applyDefaults(catalogue_service)

	frontend := workflow.Service(spec, "frontend", "Frontend", user_service, catalogue_service, cart_service, order_service)
	applyDefaults(frontend)

	wlgen := workload.Generator(spec, "wlgen", "SimpleWorkload", frontend)

	// Instantiate starting with the frontend which will trigger all other services to be instantiated
	// Also include the tests and wlgen
	return []string{"frontend_proc", wlgen, "gotests"}, nil
}
