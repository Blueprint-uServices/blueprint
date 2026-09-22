package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var frontendRegistry = registry.NewServiceRegistry[boutique.Frontend]("frontend")

func init() {
	frontendRegistry.Register("local", func(ctx context.Context) (boutique.Frontend, error) {
		setupLogger(ctx)
		catalog, err := productCatalogServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		currency, err := currencyServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		cart, err := cartServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		recommendation, err := recommendationServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		checkout, err := checkoutServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		shipping, err := shippingServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		ad, err := adServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return boutique.NewFrontendImpl(ctx, catalog, currency, cart, recommendation, checkout, shipping, ad)
	})
}

func TestFrontend(t *testing.T) {
	ctx := context.Background()
	service, err := frontendRegistry.Get(ctx)
	require.NoError(t, err)
	userID := "frontend-service-user"

	home, err := service.Home(ctx, userID, "USD")
	require.NoError(t, err)
	require.Len(t, home.Products, 9)
	require.Equal(t, int32(0), home.CartSize)

	require.NoError(t, service.AddToCart(ctx, userID, "OLJCESPC7Z", 2))
	cart, err := service.GetCart(ctx, userID, "USD")
	require.NoError(t, err)
	require.Len(t, cart.Items, 1)
	require.Equal(t, int32(2), cart.Items[0].Quantity)
	require.Len(t, cart.Recommendations, 5)

	product, err := service.GetProduct(ctx, userID, "OLJCESPC7Z", "EUR")
	require.NoError(t, err)
	require.Equal(t, "Sunglasses", product.Product.Item.Name)
	require.Equal(t, "EUR", product.Product.Price.CurrencyCode)

	order, err := service.PlaceOrder(ctx, boutique.PlaceOrderRequest{
		UserID: userID, UserCurrency: "USD", Email: "buyer@example.com", CreditCard: validCreditCard(),
		Address: boutique.Address{StreetAddress: "1 Main Street", City: "Berlin", State: "BE", Country: "Germany", ZipCode: 10115},
	})
	require.NoError(t, err)
	require.NotEmpty(t, order.OrderID)
}
