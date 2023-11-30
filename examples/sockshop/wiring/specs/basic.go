package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/gotests"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// Used by main.go
var Basic = wiringcmd.SpecOption{
	Name:        "basic",
	Description: "A basic single-process wiring spec with no modifiers",
	Build:       makeBasicSpec,
}

// Creates a basic sockshop wiring spec.
// Returns the names of the nodes to instantiate or an error
func makeBasicSpec(spec wiring.WiringSpec) ([]string, error) {
	user_db := simple.NoSQLDB(spec, "user_db")
	user_service := workflow.Define(spec, "user_service", "UserService", user_db)

	payment_service := workflow.Define(spec, "payment_service", "PaymentService")

	cart_db := simple.NoSQLDB(spec, "cart_db")
	cart_service := workflow.Define(spec, "cart_service", "CartService", cart_db)

	shipqueue := simple.Queue(spec, "shipping_queue")
	shipdb := simple.NoSQLDB(spec, "shipping_db")
	shipping_service := workflow.Define(spec, "shipping_service", "ShippingService", shipqueue, shipdb)

	queue_master := workflow.Define(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)

	order_db := simple.NoSQLDB(spec, "order_db")
	order_service := workflow.Define(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)

	catalogue_db := simple.RelationalDB(spec, "catalogue_db")
	catalogue_service := workflow.Define(spec, "catalogue_service", "CatalogueService", catalogue_db)

	tests := gotests.Test(spec, user_service, payment_service, cart_service, shipping_service, order_service, catalogue_service)

	return []string{user_service, payment_service, cart_service, shipping_service, queue_master, order_service, catalogue_service, tests}, nil
}
