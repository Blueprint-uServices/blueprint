package simplequeue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPushPop(t *testing.T) {
	ctx := context.Background()

	q := newSimpleQueueWithCapacity(1)

	snd := "hello"
	{
		// Send an item should succeed
		err := q.Push(ctx, snd)
		require.NoError(t, err)
	}

	{
		// Try pop should return the item
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.True(t, success)
		require.Equal(t, snd, rcv)
	}
}

func TestPushTryPop(t *testing.T) {
	ctx := context.Background()

	q := newSimpleQueueWithCapacity(1)

	{
		// When queue is empty, try pop should return and fail
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.False(t, success)
	}

	snd := "hello"
	{
		// Send an item should succeed
		err := q.Push(ctx, snd)
		require.NoError(t, err)
	}

	{
		// Try pop should return the item
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.True(t, success)
		require.Equal(t, snd, rcv)
	}

	{
		// When queue is empty, try pop should return and fail
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.False(t, success)
	}
}

func TestTryPush(t *testing.T) {
	ctx := context.Background()

	q := newSimpleQueueWithCapacity(1)

	{
		// When queue is empty, try pop should return and fail
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.False(t, success)
	}

	first := "first"
	{
		// trypush an item should succeed
		success, err := q.TryPush(ctx, first)
		require.True(t, success)
		require.NoError(t, err)
	}

	second := "second"
	{
		// Subsequent trypush should fail
		success, err := q.TryPush(ctx, second)
		require.False(t, success)
		require.NoError(t, err)
	}

	{
		// Try pop should return the first item
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.True(t, success)
		require.Equal(t, first, rcv)
	}

	{
		// Subsequent trypop should fail
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.False(t, success)
	}

	third := "third"
	{
		// trypush an item should succeed
		success, err := q.TryPush(ctx, third)
		require.True(t, success)
		require.NoError(t, err)
	}

	{
		// Try pop should return the third item and not the second
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.True(t, success)
		require.Equal(t, third, rcv)
	}

}

func TestPushBlocksAndOrder(t *testing.T) {

	ctx := context.Background()

	q := newSimpleQueueWithCapacity(1)

	items := []string{"hello", "world", "goodbye"}
	sending := int32(0)
	sent := int32(0)

	go func() {
		// Send items; this should block each time
		for i := range items {
			atomic.AddInt32(&sending, 1)
			err := q.Push(ctx, items[i])
			require.NoError(t, err)
			atomic.AddInt32(&sent, 1)
		}
	}()

	// Pop items slowly
	for i := range items {
		time.Sleep(10 * time.Millisecond)
		if i < len(items)-1 {
			require.Equal(t, int32(i+2), atomic.LoadInt32(&sending))
		}
		require.Equal(t, int32(i+1), atomic.LoadInt32(&sent))

		var rcv string
		err := q.Pop(ctx, &rcv)
		require.NoError(t, err)
		require.Equal(t, items[i], rcv)
	}
}

func TestTryTimeout(t *testing.T) {

	ctx := context.Background()

	q := newSimpleQueueWithCapacity(1)

	first := "hello"
	second := "world"

	go func() {
		time.Sleep(10 * time.Millisecond)
		{
			// Send an item should succeed
			err := q.Push(ctx, first)
			require.NoError(t, err)
		}
		time.Sleep(10 * time.Millisecond)
		{
			// Send an item should succeed
			err := q.Push(ctx, second)
			require.NoError(t, err)
		}
	}()

	{
		// No first item yet
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.False(t, success)
	}

	{
		// Item eventually received
		var rcv string
		success, err := q.TryPop(ctx, &rcv, 20*time.Millisecond)
		require.NoError(t, err)
		require.True(t, success)
		require.Equal(t, first, rcv)
	}

	{
		// No second item yet
		var rcv string
		success, err := q.TryPop(ctx, &rcv)
		require.NoError(t, err)
		require.False(t, success)
	}

	{
		// Item eventually received
		var rcv string
		success, err := q.TryPop(ctx, &rcv, 20*time.Millisecond)
		require.NoError(t, err)
		require.True(t, success)
		require.Equal(t, second, rcv)
	}
}
