package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/stretchr/testify/require"
)

var cartServiceRegistry = registry.NewServiceRegistry[boutique.CartService]("cart_service")

func init() {
	cartServiceRegistry.Register("local", func(ctx context.Context) (boutique.CartService, error) {
		setupLogger(ctx)
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}
		return boutique.NewCartServiceImpl(ctx, cache)
	})
}

func TestCartService(t *testing.T) {
	ctx := context.Background()
	service, err := cartServiceRegistry.Get(ctx)
	require.NoError(t, err)
	userID := "cart-service-user"

	cart, err := service.GetCart(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, cart)

	require.NoError(t, service.AddItem(ctx, userID, boutique.CartItem{ProductID: "product", Quantity: 2}))
	require.NoError(t, service.AddItem(ctx, userID, boutique.CartItem{ProductID: "product", Quantity: 3}))
	cart, err = service.GetCart(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, []boutique.CartItem{{ProductID: "product", Quantity: 5}}, cart)

	require.NoError(t, service.EmptyCart(ctx, userID))
	cart, err = service.GetCart(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, cart)
}
