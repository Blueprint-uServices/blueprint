package specs

import (
	"strings"

	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/wiring"
	"github.com/Blueprint-uServices/blueprint/plugins/clientpool"
	"github.com/Blueprint-uServices/blueprint/plugins/goproc"
	"github.com/Blueprint-uServices/blueprint/plugins/gotests"
	"github.com/Blueprint-uServices/blueprint/plugins/grpc"
	"github.com/Blueprint-uServices/blueprint/plugins/simple"
	"github.com/Blueprint-uServices/blueprint/plugins/wiringcmd"
	"github.com/Blueprint-uServices/blueprint/plugins/workflow"
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
	var allServices []string
	var allProcs []string

	applyDefaults := func(serviceName string) string {
		name, _ := strings.CutSuffix(serviceName, "service")

		// Apply Blueprint modifiers
		clientpool.Create(spec, serviceName, 10)
		grpc.Deploy(spec, serviceName)
		proc := goproc.CreateProcess(spec, name+"Proc", serviceName)

		// Save the service and proc for later instantiation
		allServices = append(allServices, serviceName)
		allProcs = append(allProcs, proc)
		return proc
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
	shipping_proc := applyDefaults(shipping_service)

	// Deploy queue master to the same process as the shipping proc
	queue_master := workflow.Service(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)
	goproc.AddToProcess(spec, shipping_proc, queue_master)

	order_db := simple.NoSQLDB(spec, "order_db")
	order_service := workflow.Service(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)
	applyDefaults(order_service)

	catalogue_db := simple.RelationalDB(spec, "catalogue_db")
	catalogue_service := workflow.Service(spec, "catalogue_service", "CatalogueService", catalogue_db)
	applyDefaults(catalogue_service)

	frontend := workflow.Service(spec, "frontend", "Frontend", user_service, catalogue_service, cart_service, order_service)
	applyDefaults(frontend)

	tests := gotests.Test(spec, allServices...)
	allProcs = append(allProcs, tests)

	return allProcs, nil
}
