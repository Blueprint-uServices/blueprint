package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var checkoutServiceRegistry = registry.NewServiceRegistry[boutique.CheckoutService]("checkout_service")

func init() {
	checkoutServiceRegistry.Register("local", func(ctx context.Context) (boutique.CheckoutService, error) {
		setupLogger(ctx)
		catalog, err := productCatalogServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		cart, err := cartServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		currency, err := currencyServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		shipping, err := shippingServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		email, err := emailServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		payment, err := paymentServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return boutique.NewCheckoutServiceImpl(ctx, catalog, cart, currency, shipping, email, payment)
	})
}

func TestCheckoutService(t *testing.T) {
	ctx := context.Background()
	service, err := checkoutServiceRegistry.Get(ctx)
	require.NoError(t, err)
	cart, err := cartServiceRegistry.Get(ctx)
	require.NoError(t, err)
	userID := "checkout-service-user"
	require.NoError(t, cart.AddItem(ctx, userID, boutique.CartItem{ProductID: "OLJCESPC7Z", Quantity: 2}))

	order, err := service.PlaceOrder(ctx, boutique.PlaceOrderRequest{
		UserID: userID, UserCurrency: "USD", Email: "buyer@example.com", CreditCard: validCreditCard(),
		Address: boutique.Address{StreetAddress: "1 Main Street", City: "Berlin", State: "BE", Country: "Germany", ZipCode: 10115},
	})
	require.NoError(t, err)
	require.NotEmpty(t, order.OrderID)
	require.NotEmpty(t, order.ShippingTrackingID)
	require.Len(t, order.Items, 1)
	require.Equal(t, int32(2), order.Items[0].Item.Quantity)

	remaining, err := cart.GetCart(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, remaining)
}
