// Package simplequeue implements an simple in-memory [backend.Queue] that internally
// uses a golang channel of capacity 10 for passing items from producer to consumer.
//
// Calls to [backend.Queue.Push] will block once the queue capacity reaches 10.
package simplequeue

import (
	"context"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

// A simple chan-based queue that implements the [backend.Queue] interface
type simpleQueue struct {
	q chan any
}

// Instantiates a [backend.Queue] that internally uses a golang channel of capacity 10.
//
// Calls to [q.Push] will block once the queue capacity reaches 10.
func NewSimpleQueue(ctx context.Context) (q backend.Queue, err error) {
	return newSimpleQueueWithCapacity(10), nil
}

// Instantiates a [simpleQueue] with the specified capacity.
func newSimpleQueueWithCapacity(capacity int) *simpleQueue {
	return &simpleQueue{
		q: make(chan any, capacity),
	}
}

// Pop implements backend.Queue.
func (q *simpleQueue) Pop(ctx context.Context, dst interface{}) (bool, error) {
	select {
	case v := <-q.q:
		return true, backend.CopyResult(v, dst)
	default:
		{
			select {
			case v := <-q.q:
				return true, backend.CopyResult(v, dst)
			case <-ctx.Done():
				return false, nil
			}
		}
	}
}

// Push implements backend.Queue.
func (q *simpleQueue) Push(ctx context.Context, item interface{}) (bool, error) {
	select {
	case q.q <- item:
		return true, nil
	default:
		{
			select {
			case q.q <- item:
				return true, nil
			case <-ctx.Done():
				return false, nil
			}
		}
	}
}
