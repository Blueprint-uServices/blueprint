// Package queuemaster implements the queue-master SockShop service, responsible for
// pulling and "processing" shipments from the shipment queue.
package queuemaster

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/shipping"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
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

// Creates a new QueueMaster service.
//
// New: once an order is shipped, it will update the order status in the orderservice.
func NewQueueMaster(ctx context.Context, queue backend.Queue, shipping shipping.ShippingService) (QueueMaster, error) {
	return newQueueMasterImpl(queue, shipping, false), nil
}

func newQueueMasterImpl(queue backend.Queue, shipping shipping.ShippingService, exitOnError bool) *queueMasterImpl {
	return &queueMasterImpl{
		q:           queue,
		shipping:    shipping,
		exitOnError: exitOnError,
		processed:   0,
	}
}

type queueMasterImpl struct {
	q           backend.Queue
	shipping    shipping.ShippingService
	exitOnError bool
	processed   int32
}

// Starts a processing loop that continually pulls elements from the queue.
// Does not exit when an error is encountered; only when ctx is cancelled
func (q *queueMasterImpl) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var shipment shipping.Shipment
			didPop, err := q.q.Pop(ctx, &shipment)
			if err != nil {
				if q.exitOnError {
					return err
				} else {
					slog.Error(fmt.Sprintf("QueueMaster unable to pull order from shipping queue due to %v", err))
					continue
				}
			}
			if didPop {
				msgNumber := atomic.AddInt32(&q.processed, 1)
				slog.Info(fmt.Sprintf("Received shipment task %v %v: %v", msgNumber, shipment.ID, shipment.Name))

				// Keep attempting to update shipping status
				for {
					err := q.shipping.UpdateStatus(ctx, shipment.ID, "shipped")
					if err != nil {
						if q.exitOnError {
							return err
						} else {
							slog.Error(fmt.Sprintf("Unable to send shipment %v due to %v; waiting 1 second then retrying", shipment.ID, err))
							time.Sleep(1 * time.Second)
						}
					} else {
						break
					}
				}
			}
		}
	}
}
