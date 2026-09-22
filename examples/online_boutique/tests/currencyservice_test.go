package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var currencyServiceRegistry = registry.NewServiceRegistry[boutique.CurrencyService]("currency_service")

func init() {
	currencyServiceRegistry.Register("local", func(ctx context.Context) (boutique.CurrencyService, error) {
		setupLogger(ctx)
		return boutique.NewCurrencyServiceImpl(ctx)
	})
}

func TestCurrencyService(t *testing.T) {
	ctx := context.Background()
	service, err := currencyServiceRegistry.Get(ctx)
	require.NoError(t, err)

	currencies, err := service.GetSupportedCurrencies(ctx)
	require.NoError(t, err)
	require.Contains(t, currencies, "USD")
	require.Contains(t, currencies, "EUR")

	converted, err := service.Convert(ctx, boutique.Money{CurrencyCode: "EUR", Units: 10}, "USD")
	require.NoError(t, err)
	require.Equal(t, "USD", converted.CurrencyCode)
	require.Equal(t, int64(11), converted.Units)
	require.InDelta(t, 305000000, converted.Nanos, 1)

	_, err = service.Convert(ctx, boutique.Money{CurrencyCode: "BAD", Units: 1}, "USD")
	require.Error(t, err)
}
