package queuemaster

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/shipping"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplequeue"
	"github.com/stretchr/testify/require"
)

// Unit tests that don't use gotests plugin

func TestQueueMaster(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	q, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)

	db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	require.NoError(t, err)

	shipService, err := shipping.NewShippingService(ctx, q, db)
	require.NoError(t, err)

	qMaster := newQueueMasterImpl(q, shipService, true)
	require.Equal(t, int32(0), qMaster.processed)

	exitCount := int32(0)
	go func() {
		err = qMaster.Run(ctx)
		require.NoError(t, err)
		atomic.AddInt32(&exitCount, 1)
	}()

	shipment := shipping.Shipment{
		ID:     "first",
		Name:   "my first shipment",
		Status: "unshipped",
	}
	_, err = shipService.PostShipping(ctx, shipment)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	require.Equal(t, int32(1), atomic.LoadInt32(&qMaster.processed))

	shipment2, err := shipService.GetShipment(ctx, shipment.ID)
	require.NoError(t, err)
	require.NotEqual(t, shipment, shipment2)
	require.Equal(t, "shipped", shipment2.Status)

	cancel()

	time.Sleep(10 * time.Millisecond)
	require.Equal(t, int32(1), atomic.LoadInt32(&exitCount))
}
