// Package clientpool implements the runtime components of Blueprint's ClientPool plugin.
//
// ClientPools do not need to be used directly by application workflow specs.  Instead, this
// code is included in a compiled application by applying the ClientPool modifier to the wiring spec.
package clientpool

import (
	"context"
	"sync/atomic"

	"errors"
)

// A ClientPool that contains up to Capacity clients. Clients are acquired with
// Pop and returned with Push.
type ClientPool[T any] struct {
	clients   chan T
	build     func() (T, error)
	capacity  int64
	size      int64
	available int64
	waiting   int64
}

// Instantiates a [ClientPool] that will have up to maxClients client instances.
// The provided function fn is used to instantiate clients.
//
// Callers acquire a client instance by calling [ClientPool.Pop], and when they
// are finished with a client, return it to the pool by calling [ClientPool.Push]
func NewClientPool[T any](capacity int, build func() (T, error)) *ClientPool[T] {
	pool := &ClientPool[T]{
		clients:   make(chan T, capacity),
		build:     build,
		capacity:  int64(capacity),
		size:      0,
		available: 0,
		waiting:   0,
	}
	return pool
}

// Acquires a client from the pool, blocking if all clients are currently in use.
//
// When a caller has finished using a client, it *must* call [ClientPool.Push] to
// return the client to the pool.
func (pool *ClientPool[T]) Pop(ctx context.Context) (client T, err error) {
	// Attempt to immediately reuse an existing client
	select {
	case <-ctx.Done():
		err = errors.New("timeout before client was available")
		return
	case client = <-pool.clients:
		atomic.AddInt64(&pool.available, -1)
		return client, nil
	default:
	}

	// If the pool isn't at capacity, we can instantiate a new client
	for curSize := pool.size; curSize < pool.capacity; {
		if atomic.CompareAndSwapInt64(&pool.size, curSize, curSize+1) {
			client, err = pool.build()
			if err != nil {
				atomic.AddInt64(&pool.size, -1)
			}
			return
		}
		curSize = pool.size
	}

	// Pool is at capacity; wait to reuse an existing client
	atomic.AddInt64(&pool.waiting, 1)
	select {
	case <-ctx.Done():
		err = errors.New("timeout before client was available")
	case client = <-pool.clients:
		atomic.AddInt64(&pool.available, -1)
	}
	atomic.AddInt64(&pool.waiting, -1)
	return
}

// Returns a client to the pool.
func (pool *ClientPool[T]) Push(client T) {
	atomic.AddInt64(&pool.available, 1)
	pool.clients <- client
}

// Returns the capacity of the client pool
func (pool *ClientPool[T]) Capacity() int {
	return int(pool.capacity)
}

// Returns the current size of the client pool
func (pool *ClientPool[T]) Size() int {
	return int(pool.size)
}

// Returns the current number of available clients in the client pool
func (pool *ClientPool[T]) Available() int {
	return int(pool.available)
}
