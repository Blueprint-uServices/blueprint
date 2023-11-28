package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/shipping"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplequeue"
)

// Tests acquire a ShippingService instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var shippingRegistry = registry.NewServiceRegistry[shipping.ShippingService]("shipping_service")
var queueRegistry = registry.NewServiceRegistry[backend.Queue]("shipping_queue")

func init() {
	queueRegistry.Register("local", func(ctx context.Context) (backend.Queue, error) {
		return simplequeue.NewSimpleQueue(ctx)
	})

	// If the tests are run locally, we fall back to this ShippingService implementation
	shippingRegistry.Register("local", func(ctx context.Context) (shipping.ShippingService, error) {
		queue, err := queueRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		ship, err := shipping.NewShippingService(ctx, queue, db)
		if err != nil {
			return nil, err
		}

		return ship, nil
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
		ID:     "hello",
		Name:   "world",
		Status: "awaiting shipment",
	}

	{
		sent, err := service.PostShipping(ctx, shipment)
		require.NoError(t, err)
		require.Equal(t, shipment, sent)
	}

	// Start the queue master if not already started
	_, err = queuemasterRegistry.Get(ctx)
	require.NoError(t, err)

	// Sleep for up to 30 seconds checking shipment status
	for i := 0; i < 30; i++ {
		shipment2, err := service.GetShipment(ctx, shipment.ID)
		require.NoError(t, err)
		if shipment2.Status == "awaiting shipment" {
			time.Sleep(1 * time.Second)
			continue
		}
		require.Equal(t, "shipped", shipment2.Status)
	}

}
