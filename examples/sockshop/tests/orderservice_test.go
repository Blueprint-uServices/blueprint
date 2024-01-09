package tests

import (
	"context"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/order"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

// Tests acquire an OrderService instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var ordersRegistry = registry.NewServiceRegistry[order.OrderService]("order_service")

func init() {
	// If the tests are run locally, we fall back to this OrderService implementation
	ordersRegistry.Register("local", func(ctx context.Context) (order.OrderService, error) {
		user, err := userServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		cart, err := cartRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		payment, err := paymentServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		shipping, err := shippingRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		orderdb, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		return order.NewOrderService(ctx, user, cart, payment, shipping, orderdb)
	})
}

func TestOrderService(t *testing.T) {

	ctx := context.Background()

	// Get the orders service
	orderService, err := ordersRegistry.Get(ctx)

	// Try placing an empty order
	_, err = orderService.NewOrder(ctx, "", "", "", "")
	require.Error(t, err)

	// Try placing an order without a user
	_, err = orderService.NewOrder(ctx, "jon", "jonsaddress", "jonscard", "jon")
	require.Error(t, err)

	// Add our user
	user, err := userServiceRegistry.Get(ctx)
	require.NoError(t, err)
	userId, err := user.PostUser(ctx, deepak)
	require.NoError(t, err)
	defer user.Delete(ctx, "customers", userId)

	// Get the card and address IDs
	users, err := user.GetUsers(ctx, userId)
	require.NoError(t, err)
	require.Len(t, users, 1)
	cardId := users[0].Cards[0].ID
	addressId := users[0].Addresses[0].ID

	// Try placing an order without an item
	_, err = orderService.NewOrder(ctx, userId, addressId, cardId, userId)
	require.Error(t, err)

	// Put some items in the cart
	cart, err := cartRegistry.Get(ctx)
	require.NoError(t, err)
	cart.AddItem(ctx, userId, myitem)

	// Place the order
	require.NoError(t, err)
	order, err := orderService.NewOrder(ctx, userId, addressId, cardId, userId)
	require.NoError(t, err)
	require.Equal(t, userId, order.CustomerID)

	// Check we can look up the order
	order2, err := orderService.GetOrder(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, order, order2)

	// Wait up to 30 seconds for the status to change
	shipping, err := shippingRegistry.Get(ctx)
	require.NoError(t, err)
	for i := 0; i < 30; i++ {
		shipment2, err := shipping.GetShipment(ctx, order2.ID)
		require.NoError(t, err)
		if shipment2.Status == "awaiting shipment" {
			time.Sleep(1 * time.Second)
			continue
		}
		require.Equal(t, "shipped", shipment2.Status)
	}

}
