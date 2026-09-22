package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var emailServiceRegistry = registry.NewServiceRegistry[boutique.EmailService]("email_service")

func init() {
	emailServiceRegistry.Register("local", func(ctx context.Context) (boutique.EmailService, error) {
		setupLogger(ctx)
		return boutique.NewEmailServiceImpl(ctx)
	})
}

func TestEmailService(t *testing.T) {
	ctx := context.Background()
	service, err := emailServiceRegistry.Get(ctx)
	require.NoError(t, err)

	order := boutique.Order{
		OrderID: "order-1", ShippingTrackingID: "AB-123-456",
		ShippingCost: boutique.Money{CurrencyCode: "USD", Units: 8, Nanos: 990000000},
		Items:        []boutique.OrderItem{{Item: boutique.CartItem{ProductID: "OLJCESPC7Z", Quantity: 1}, Cost: boutique.Money{CurrencyCode: "USD", Units: 19, Nanos: 990000000}}},
	}
	require.NoError(t, service.SendOrderConfirmation(ctx, "buyer@example.com", order))
}
