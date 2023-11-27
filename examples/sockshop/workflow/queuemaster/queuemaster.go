// Package queuemaster implements the queue-master SockShop service, responsible for
// pulling and "processing" shipments from the shipment queue.
package queuemaster

import (
	"context"
	"fmt"
	"sync/atomic"

	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/shipping"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"golang.org/x/exp/slog"
)

// QueueMaster implements the SockShop queue-master microservice.
//
// It is not a service that can be called; instead it pulls shipments from
// the shipments queue
type QueueMaster interface {
	// Runs the background goroutine that continually pulls elements from
	// the queue.  Does not return until ctx is cancelled or an error is
	// encountered
	Run(ctx context.Context) error
}

func NewQueueMaster(ctx context.Context, queue backend.Queue) (QueueMaster, error) {
	return newQueueMasterImpl(queue), nil
}

func newQueueMasterImpl(queue backend.Queue) *queueMasterImpl {
	return &queueMasterImpl{
		q:         queue,
		processed: 0,
	}
}

type queueMasterImpl struct {
	q         backend.Queue
	processed int32
}

// Starts a processing loop that continually pulls elements from the queue.
// Does not return until ctx is cancelled or an error is encountered
func (q *queueMasterImpl) Run(ctx context.Context) error {
	for {
		var shipment shipping.Shipment
		err := q.q.Pop(ctx, &shipment)
		if err != nil {
			return err
		}
		msgNumber := atomic.AddInt32(&q.processed, 1)

		slog.Info(fmt.Sprintf("Received shipment task %v %v: %v", msgNumber, shipment.ID, shipment.Name))
	}
}
