package backend

import "context"

// A Queue backend is used for pushing and popping elements.
type Queue interface {
	// Attempt to push an item to the tail of the queue.  Does not block;
	// for example, if the queue is full then the method returns immediately
	// with a value of false.  An error will be returned only when an
	// erroneous state is encountered.
	//
	// Reports whether the push was successful, and possibly an error
	TryPush(ctx context.Context, item interface{}) (bool, error)

	// Pushes an item to the tail of the queue, blocking until it can do so.
	Push(ctx context.Context, item interface{}) error

	// Attempt to pop an item from the head of the queue.  Does not block;
	// for example, if the queue is empty then the method returns immediately
	// with a value of false.  An error will only be returned when an erroneous
	// state is encountered.
	//
	// dst must be a pointer to a receiver struct type.
	//
	// Reports whether the pop was successful, and possibly an error.  If the
	// pop was successful then the result is set in dst
	TryPop(ctx context.Context, dst interface{}) (bool, error)

	// Pops an item from the front of the queue, blocking until it can do so,
	// and placing the result in dst.
	//
	// dst must be a pointer
	Pop(ctx context.Context, dst interface{}) error
}
