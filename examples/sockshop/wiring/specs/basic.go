// Package specs implements wiring specs for the SockShop application.
//
// The wiring spec can be specified using the -w option when running wiring/main.go
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
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

// A simple wiring spec that compiles all services to a single process and therefore directly invoke each other.
// No RPC, containers, processes etc. are used.
var Basic = cmdbuilder.SpecOption{
	Name:        "basic",
	Description: "A basic single-process wiring spec with no modifiers",
	Build:       makeBasicSpec,
}

func makeBasicSpec(spec wiring.WiringSpec) ([]string, error) {
	user_db := simple.NoSQLDB(spec, "user_db")
	user_service := workflow.Service[user.UserService](spec, "user_service", user_db)

	payment_service := workflow.Service[payment.PaymentService](spec, "payment_service", "500")

	cart_db := simple.NoSQLDB(spec, "cart_db")
	cart_service := workflow.Service[cart.CartService](spec, "cart_service", cart_db)

	shipqueue := simple.Queue(spec, "shipping_queue")
	shipdb := simple.NoSQLDB(spec, "shipping_db")
	shipping_service := workflow.Service[shipping.ShippingService](spec, "shipping_service", shipqueue, shipdb)

	queue_master := workflow.Service[queuemaster.QueueMaster](spec, "queue_master", shipqueue, shipping_service)

	order_db := simple.NoSQLDB(spec, "order_db")
	order_service := workflow.Service[order.OrderService](spec, "order_service", user_service, cart_service, payment_service, shipping_service, order_db)

	catalogue_db := simple.RelationalDB(spec, "catalogue_db")
	catalogue_service := workflow.Service[catalogue.CatalogueService](spec, "catalogue_service", catalogue_db)

	frontend_service := workflow.Service[frontend.Frontend](spec, "frontend", user_service, catalogue_service, cart_service, order_service)

	return []string{user_service, payment_service, cart_service, shipping_service, queue_master, order_service, catalogue_service, frontend_service}, nil
}
