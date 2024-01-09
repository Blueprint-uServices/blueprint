package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/payment"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var paymentServiceRegistry = registry.NewServiceRegistry[payment.PaymentService]("payment_service")

func init() {
	paymentServiceRegistry.Register("local", func(ctx context.Context) (payment.PaymentService, error) {
		payDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		moneyDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return payment.NewPaymentServiceImpl(ctx, payDB, moneyDB)
	})
}

func genTestPaymentData() []payment.Payment {
	res := []payment.Payment{}
	for i := 0; i < 10; i++ {
		p := payment.Payment{
			ID:      fmt.Sprintf("ID%d", i),
			OrderID: fmt.Sprintf("Order%d", i),
			UserID:  fmt.Sprintf("User%d", i),
			Price:   fmt.Sprintf("%d", (i+1)*100),
		}
		res = append(res, p)
	}
	return res
}

func TestPaymentService(t *testing.T) {
	ctx := context.Background()
	service, err := paymentServiceRegistry.Get(ctx)
	require.NoError(t, err)

	testData := genTestPaymentData()

	// Test init payment
	for _, d := range testData {
		err = service.InitPayment(ctx, d)
		require.NoError(t, err)
	}

	// Test query
	all_payments, err := service.Query(ctx)
	require.NoError(t, err)
	require.Len(t, all_payments, len(testData))

	// Test AddMoney
	for _, d := range testData {
		err = service.AddMoney(ctx, d)
		require.NoError(t, err)
	}

	// Test Pay
	for _, d := range testData {
		err = service.AddMoney(ctx, d)
		require.NoError(t, err)
	}

	// Delete all
	err = service.Cleanup(ctx)
	require.NoError(t, err)
}

func requirePayment(t *testing.T, expected payment.Payment, actual payment.Payment) {
	require.Equal(t, expected.ID, actual.OrderID)
	require.Equal(t, expected.OrderID, actual.OrderID)
	require.Equal(t, expected.Price, actual.Price)
	require.Equal(t, expected.UserID, actual.UserID)
}
