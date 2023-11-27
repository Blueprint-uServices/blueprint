package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/shipping"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplequeue"
)

// Tests acquire a ShippingService instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var shippingRegistry = registry.NewServiceRegistry[shipping.ShippingService]("shipping_service")

func init() {
	// If the tests are run locally, we fall back to this ShippingService implementation
	shippingRegistry.Register("local", func(ctx context.Context) (shipping.ShippingService, error) {
		queue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		return shipping.NewShippingService(ctx, queue)
	})
}

// We write the service test as a single test because we don't want to tear down and
// spin up the Mongo backends between tests, so state will persist in the database
// between tests.
func TestShippingService(t *testing.T) {
	ctx := context.Background()
	service, err := shippingRegistry.Get(ctx)
	require.NoError(t, err)

	shipment := shipping.Shipment{
		ID:   "hello",
		Name: "world",
	}
	sent, err := service.PostShipping(ctx, shipment)
	require.NoError(t, err)
	require.Equal(t, shipment, sent)
}
