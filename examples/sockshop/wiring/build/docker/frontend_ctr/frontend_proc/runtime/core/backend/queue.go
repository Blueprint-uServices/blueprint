package backend

import (
	"context"
)

// A Queue backend is used for pushing and popping elements.
type Queue interface {

	// Pushes an item to the tail of the queue.
	//
	// This call will block until the item is successfully pushed, or until the context
	// is cancelled.
	//
	// Reports whether the item was pushed to the queue, or if an error was encountered.
	// A context cancellation/timeout is not considered an error.
	Push(ctx context.Context, item interface{}) (bool, error)

	// Pops an item from the front of the queue.
	//
	// This call will block until an item is successfully popped, or until the context
	// is cancelled.
	//
	// dst must be a pointer type that can receive the item popped from the queue.
	//
	// Reports whether the item was pushed to the queue, or if an error was encountered.
	// A context cancellation/timeout is not considered an error.
	Pop(ctx context.Context, dst interface{}) (bool, error)
}
