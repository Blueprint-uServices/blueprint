// Package tests implements unit tests for the SockShop application that are compatible with the Blueprint gotests plugin.
//
// After compiling the SockShop application, tests can be found in the `gotests` subdirectory of the output folder.
package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/cart"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

// Tests acquire a CartService instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var cartRegistry = registry.NewServiceRegistry[cart.CartService]("cart_service")

func init() {
	// If the tests are run locally, we fall back to this CartService implementation
	cartRegistry.Register("local", func(ctx context.Context) (cart.CartService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		return cart.NewCartService(ctx, db)
	})
}

func TestNonExistentCart(t *testing.T) {
	ctx := context.Background()
	service, err := cartRegistry.Get(ctx)
	require.NoError(t, err)

	items, err := service.GetCart(ctx, "TestNonExistentCart")
	require.NoError(t, err)
	require.Len(t, items, 0)
}

var myitem = cart.Item{
	ID:        "myitem",
	Quantity:  5,
	UnitPrice: 37.75,
}

func TestAddItemToNonExistentCart(t *testing.T) {
	customerID := "TestAddItemToNonExistentCart"
	item := myitem

	ctx := context.Background()
	service, err := cartRegistry.Get(ctx)
	require.NoError(t, err)

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No item should exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, cart.Item{}, item2)
	}

	{
		// Add an item
		item2, err := service.AddItem(ctx, customerID, item)
		require.NoError(t, err)
		require.Equal(t, item, item2)
	}

	{
		// The cart should now exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, item, items[0])
	}

	{
		// The item should now exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, item, item2)
	}

	{
		// Delete the customer
		err := service.DeleteCart(ctx, customerID)
		require.NoError(t, err)
	}

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No item should exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, cart.Item{}, item2)
	}
}

func TestAddRemoveItems(t *testing.T) {
	customerID := "TestAddRemoveItems"
	items := []cart.Item{
		cart.Item{ID: "firstitem", Quantity: 5, UnitPrice: 37.75},
		cart.Item{ID: "seconditem", Quantity: 12, UnitPrice: 12.25},
	}

	ctx := context.Background()
	service, err := cartRegistry.Get(ctx)
	require.NoError(t, err)

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No items should exist
		for _, item := range items {
			item2, err := service.GetItem(ctx, customerID, item.ID)
			require.NoError(t, err)
			require.Equal(t, cart.Item{}, item2)
		}
	}

	{
		// Add all the items
		for _, item := range items {
			// Add an item
			item2, err := service.AddItem(ctx, customerID, item)
			require.NoError(t, err)
			require.Equal(t, item, item2)
		}
	}

	{
		// The cart should now exist with all items
		items2, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items2, len(items))
		for i := 0; i < len(items); i++ {
			require.Equal(t, items[i], items2[i])
		}
	}

	{
		// Remove one item
		service.RemoveItem(ctx, customerID, items[0].ID)
		require.NoError(t, err)
	}

	{
		// The first item should no longer exist
		item, err := service.GetItem(ctx, customerID, items[0].ID)
		require.NoError(t, err)
		require.Equal(t, cart.Item{}, item)
	}

	{
		// The second item should exist
		item, err := service.GetItem(ctx, customerID, items[1].ID)
		require.NoError(t, err)
		require.Equal(t, items[1], item)
	}

	{
		// The cart should exist with one item
		items2, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items2, 1)
		require.Equal(t, items[1], items2[0])
	}

	{
		// Remove all items including non-existent
		for _, item := range items {
			service.RemoveItem(ctx, customerID, item.ID)
			require.NoError(t, err)
		}
	}

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No items should exist
		for _, item := range items {
			item2, err := service.GetItem(ctx, customerID, item.ID)
			require.NoError(t, err)
			require.Equal(t, cart.Item{}, item2)
		}
	}
}

func TestUpdateItem(t *testing.T) {
	customerID := "TestUpdateItem"
	item := cart.Item{
		ID:        "myitem",
		Quantity:  5,
		UnitPrice: 37.75,
	}

	ctx := context.Background()
	service, err := cartRegistry.Get(ctx)
	require.NoError(t, err)

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No item should exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, cart.Item{}, item2)
	}

	{
		// Update the item should insert it
		err := service.UpdateItem(ctx, customerID, item)
		require.NoError(t, err)
	}

	{
		// The cart should now exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, item, items[0])
	}

	{
		// The item should now exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, item, item2)
	}

	// The item is on sale, add more to the cart, because it's cheaper
	itemUpdate := cart.Item{ID: item.ID, Quantity: 1, UnitPrice: 30}

	{
		// Update the cart
		err := service.UpdateItem(ctx, customerID, itemUpdate)
		require.NoError(t, err)
	}

	{
		// The cart should still exist with new quantity
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, item.ID, items[0].ID)
		require.Equal(t, itemUpdate.Quantity, items[0].Quantity)
		require.Equal(t, itemUpdate.UnitPrice, items[0].UnitPrice)
	}

	{
		// The item should now exist with increased quantity
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, item.ID, item2.ID)
		require.Equal(t, itemUpdate.Quantity, item2.Quantity)
		require.Equal(t, itemUpdate.UnitPrice, item2.UnitPrice)
	}

	{
		// Delete the customer
		err := service.DeleteCart(ctx, customerID)
		require.NoError(t, err)
	}

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}
}

func TestNegativeUpdateItem(t *testing.T) {
	customerID := "TestNegativeUpdateItem"
	item := cart.Item{
		ID:        "myitem",
		Quantity:  5,
		UnitPrice: 37.75,
	}
	doubleItem := cart.Item{
		ID:        "myitem",
		Quantity:  10,
		UnitPrice: 37.75,
	}
	negativeItem := cart.Item{
		ID:        "myitem",
		Quantity:  -5,
		UnitPrice: 37.75,
	}

	ctx := context.Background()
	service, err := cartRegistry.Get(ctx)
	require.NoError(t, err)

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No item should exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, cart.Item{}, item2)
	}

	{
		// Try to do negative update
		err := service.UpdateItem(ctx, customerID, negativeItem)
		require.NoError(t, err)
	}

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No item should exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, cart.Item{}, item2)
	}

	{
		// Update the item should insert it
		err := service.UpdateItem(ctx, customerID, item)
		require.NoError(t, err)
	}

	{
		// The cart should now exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, item, items[0])
	}

	{
		// The item should now exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, item, item2)
	}

	{
		// Update the item should increment quantity
		err := service.UpdateItem(ctx, customerID, doubleItem)
		require.NoError(t, err)
	}

	{
		// The cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, doubleItem, items[0])
	}

	{
		// The doubleItem should now exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, doubleItem, item2)
	}

	{
		// Do negative update
		err := service.UpdateItem(ctx, customerID, negativeItem)
		require.NoError(t, err)
	}

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No item should exist
		item2, err := service.GetItem(ctx, customerID, item.ID)
		require.NoError(t, err)
		require.Equal(t, cart.Item{}, item2)
	}
}

func TestMergeCarts(t *testing.T) {
	customerID := "TestMergeCarts_Customer"
	sessionID := "TestMergeCarts_Session"
	item := cart.Item{
		ID:        "myitem",
		Quantity:  5,
		UnitPrice: 37.75,
	}

	ctx := context.Background()
	service, err := cartRegistry.Get(ctx)
	require.NoError(t, err)

	{
		// No customer cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No session cart should exist
		items, err := service.GetCart(ctx, sessionID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// Add an item to the customer cart
		item2, err := service.AddItem(ctx, customerID, item)
		require.NoError(t, err)
		require.Equal(t, item, item2)
	}

	{
		// Add an item to the session cart
		item2, err := service.AddItem(ctx, sessionID, item)
		require.NoError(t, err)
		require.Equal(t, item, item2)
	}

	{
		// The customer cart should now exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, item, items[0])
	}

	{
		// The session cart should now exist
		items, err := service.GetCart(ctx, sessionID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, item, items[0])
	}

	{
		// Merge the carts
		err := service.MergeCarts(ctx, customerID, sessionID)
		require.NoError(t, err)
	}

	{
		// Session cart should not exist
		items, err := service.GetCart(ctx, sessionID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// The customer cart should now exist with double quantity
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, item.ID, items[0].ID)
		require.Equal(t, item.Quantity*2, items[0].Quantity)
		require.Equal(t, item.UnitPrice, items[0].UnitPrice)
	}

	{
		// Delete the customer
		err := service.DeleteCart(ctx, customerID)
		require.NoError(t, err)
	}

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}
}

func TestMergeCartsConcat(t *testing.T) {
	customerID := "TestMergeCartsConcat_Customer"
	sessionID := "TestMergeCartsConcat_Session"
	firstitem := cart.Item{
		ID:        "firstitem",
		Quantity:  5,
		UnitPrice: 37.75,
	}
	seconditem := cart.Item{
		ID:        "seconditem",
		Quantity:  7,
		UnitPrice: 48,
	}

	ctx := context.Background()
	service, err := cartRegistry.Get(ctx)
	require.NoError(t, err)

	{
		// No customer cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// No session cart should exist
		items, err := service.GetCart(ctx, sessionID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// Add firstitem to the customer cart
		item2, err := service.AddItem(ctx, customerID, firstitem)
		require.NoError(t, err)
		require.Equal(t, firstitem, item2)
	}

	{
		// Add seconditem to the session cart
		item2, err := service.AddItem(ctx, sessionID, seconditem)
		require.NoError(t, err)
		require.Equal(t, seconditem, item2)
	}

	{
		// The customer cart should now exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, firstitem, items[0])
	}

	{
		// The session cart should now exist
		items, err := service.GetCart(ctx, sessionID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, seconditem, items[0])
	}

	{
		// Merge the carts
		err := service.MergeCarts(ctx, customerID, sessionID)
		require.NoError(t, err)
	}

	{
		// Session cart should not exist
		items, err := service.GetCart(ctx, sessionID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}

	{
		// The customer cart should now exist with two items
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 2)
		require.Equal(t, firstitem, items[0])
		require.Equal(t, seconditem, items[1])
	}

	{
		// Delete the customer
		err := service.DeleteCart(ctx, customerID)
		require.NoError(t, err)
	}

	{
		// No cart should exist
		items, err := service.GetCart(ctx, customerID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	}
}

// We write the service test as a single test because we don't want to tear down and
// spin up the Mongo backends between tests, so state will persist in the database
// between tests.
func TestCartService(t *testing.T) {
	ctx := context.Background()
	service, err := cartRegistry.Get(ctx)
	require.NoError(t, err)

	// Get a non-existent cart
	items, err := service.GetCart(ctx, "nobody")
	require.NoError(t, err)
	require.Len(t, items, 0)
}
