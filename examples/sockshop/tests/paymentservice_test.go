package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/payment"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/assert"
)

// Tests acquire a PaymentService instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var paymentServiceRegistry = registry.NewServiceRegistry[payment.PaymentService]("payment_service")

func init() {
	// If the tests are run locally, we fall back to this PaymentService implementation
	paymentServiceRegistry.Register("local", func(ctx context.Context) (payment.PaymentService, error) {
		return payment.NewPaymentService(ctx, "500")
	})
}

// We write the service test as a single test because we don't want to tear down and
// spin up the Mongo backends between tests, so state will persist in the database
// between tests.
func TestPaymentService(t *testing.T) {
	ctx := context.Background()
	service, err := paymentServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	rsp, err := service.Authorise(ctx, 1000)
	assert.NoError(t, err)
	assert.False(t, rsp.Authorised)

	rsp, err = service.Authorise(ctx, 100)
	assert.NoError(t, err)
	assert.True(t, rsp.Authorised)
}
