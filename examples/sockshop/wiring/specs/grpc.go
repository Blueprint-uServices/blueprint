package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/cart"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/catalogue"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/frontend"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/order"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/payment"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/queuemaster"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/shipping"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/user"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workload/workloadgen"
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
	user_service := workflow.Service[user.UserService](spec, "user_service", user_db)
	applyDefaults(user_service)

	payment_service := workflow.Service[payment.PaymentService](spec, "payment_service", "500")
	applyDefaults(payment_service)

	cart_db := simple.NoSQLDB(spec, "cart_db")
	cart_service := workflow.Service[cart.CartService](spec, "cart_service", cart_db)
	applyDefaults(cart_service)

	shipqueue := simple.Queue(spec, "shipping_queue")
	shipdb := simple.NoSQLDB(spec, "shipping_db")
	shipping_service := workflow.Service[shipping.ShippingService](spec, "shipping_service", shipqueue, shipdb)
	applyDefaults(shipping_service)

	// Deploy queue master to the same process as the shipping proc
	queue_master := workflow.Service[queuemaster.QueueMaster](spec, "queue_master", shipqueue, shipping_service)
	goproc.AddToProcess(spec, "shipping_proc", queue_master)

	order_db := simple.NoSQLDB(spec, "order_db")
	order_service := workflow.Service[order.OrderService](spec, "order_service", user_service, cart_service, payment_service, shipping_service, order_db)
	applyDefaults(order_service)

	catalogue_db := simple.RelationalDB(spec, "catalogue_db")
	catalogue_service := workflow.Service[catalogue.CatalogueService](spec, "catalogue_service", catalogue_db)
	applyDefaults(catalogue_service)

	frontend_service := workflow.Service[frontend.Frontend](spec, "frontend", user_service, catalogue_service, cart_service, order_service)
	applyDefaults(frontend_service)

	wlgen := workload.Generator[workloadgen.SimpleWorkload](spec, "wlgen", frontend_service)

	// Instantiate starting with the frontend which will trigger all other services to be instantiated
	// Also include the tests and wlgen
	return []string{"frontend_proc", wlgen, "gotests"}, nil
}
