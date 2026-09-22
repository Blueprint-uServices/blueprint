package tests

import (
	"context"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var paymentServiceRegistry = registry.NewServiceRegistry[boutique.PaymentService]("payment_service")

func init() {
	paymentServiceRegistry.Register("local", func(ctx context.Context) (boutique.PaymentService, error) {
		setupLogger(ctx)
		return boutique.NewPaymentServiceImpl(ctx)
	})
}

func validCreditCard() boutique.CreditCardInfo {
	return boutique.CreditCardInfo{Number: "4111-1111-1111-1111", CVV: 123, ExpirationYear: time.Now().Year() + 1, ExpirationMonth: time.December}
}

func TestPaymentService(t *testing.T) {
	ctx := context.Background()
	service, err := paymentServiceRegistry.Get(ctx)
	require.NoError(t, err)

	transactionID, err := service.Charge(ctx, boutique.Money{CurrencyCode: "USD", Units: 25}, validCreditCard())
	require.NoError(t, err)
	require.NotEmpty(t, transactionID)

	invalid := validCreditCard()
	invalid.Number = "1234"
	_, err = service.Charge(ctx, boutique.Money{CurrencyCode: "USD", Units: 25}, invalid)
	require.ErrorIs(t, err, boutique.ErrInvalidCreditCard)
}
