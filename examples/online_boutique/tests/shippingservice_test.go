package tests

import (
	"context"
	"regexp"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var shippingServiceRegistry = registry.NewServiceRegistry[boutique.ShippingService]("shipping_service")

func init() {
	shippingServiceRegistry.Register("local", func(ctx context.Context) (boutique.ShippingService, error) {
		setupLogger(ctx)
		return boutique.NewShippingServiceImpl(ctx)
	})
}

func TestShippingService(t *testing.T) {
	ctx := context.Background()
	service, err := shippingServiceRegistry.Get(ctx)
	require.NoError(t, err)
	address := boutique.Address{StreetAddress: "1 Main Street", City: "Berlin", State: "BE", Country: "Germany", ZipCode: 10115}
	items := []boutique.CartItem{{ProductID: "OLJCESPC7Z", Quantity: 1}}

	quote, err := service.GetQuote(ctx, address, items)
	require.NoError(t, err)
	require.Equal(t, boutique.Money{CurrencyCode: "USD", Units: 8, Nanos: 990000000}, quote)

	trackingID, err := service.ShipOrder(ctx, address, items)
	require.NoError(t, err)
	require.Regexp(t, regexp.MustCompile(`^[A-Z]{2}-[0-9]+-[0-9]+$`), trackingID)
}
