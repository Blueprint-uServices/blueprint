package queuemaster

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/shipping"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplequeue"
)

// Unit tests that don't use gotests plugin

func TestQueueMaster(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	q, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)

	shipService, err := shipping.NewShippingService(ctx, q)
	require.NoError(t, err)

	qMaster := newQueueMasterImpl(q)
	require.Equal(t, int32(0), qMaster.processed)

	exitCount := int32(0)
	go func() {
		err = qMaster.Run(ctx)
		require.Error(t, err)
		atomic.AddInt32(&exitCount, 1)
	}()

	shipment := shipping.Shipment{
		ID:   "first",
		Name: "my first shipment",
	}
	_, err = shipService.PostShipping(ctx, shipment)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	require.Equal(t, int32(1), atomic.LoadInt32(&qMaster.processed))

	cancel()

	time.Sleep(10 * time.Millisecond)
	require.Equal(t, int32(1), atomic.LoadInt32(&exitCount))
}
