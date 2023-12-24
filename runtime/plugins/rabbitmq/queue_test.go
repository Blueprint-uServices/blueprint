package rabbitmq

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPushPop(t *testing.T) {
	ctx := context.Background()

	q, err := NewRabbitMQ(ctx, "localhost:5672", "queue")
	require.NoError(t, err)

	snd := "hello"
	{
		// Send an item should succeed
		success, err := q.Push(ctx, snd)
		require.NoError(t, err)
		require.True(t, success)
	}

	{
		// Pop should return the item
		var rcv string
		success, err := q.Pop(ctx, &rcv)
		require.NoError(t, err)
		require.True(t, success)
		require.Equal(t, snd, rcv)
	}
}

func TestTryTimeout(t *testing.T) {

	ctx := context.Background()

	q, err := NewRabbitMQ(ctx, "localhost:5672", "queue")
	require.NoError(t, err)

	first := "hello"
	second := "world"

	go func() {
		time.Sleep(10 * time.Millisecond)
		{
			// Send an item should succeed
			success, err := q.Push(ctx, first)
			require.NoError(t, err)
			require.True(t, success)
		}
		time.Sleep(10 * time.Millisecond)
		{
			// Send an item should succeed
			success, err := q.Push(ctx, second)
			require.NoError(t, err)
			require.True(t, success)
		}
	}()

	{
		// No first item yet
		var rcv string
		timeoutCtx, _ := context.WithTimeout(ctx, 0*time.Second)
		success, err := q.Pop(timeoutCtx, &rcv)
		require.NoError(t, err)
		require.False(t, success)
	}

	{
		// Item eventually received
		var rcv string
		timeoutCtx, _ := context.WithTimeout(ctx, 20*time.Millisecond)
		success, err := q.Pop(timeoutCtx, &rcv)
		require.NoError(t, err)
		require.True(t, success)
		require.Equal(t, first, rcv)
	}

	{
		// No second item yet
		var rcv string
		timeoutCtx, _ := context.WithTimeout(ctx, 0*time.Second)
		success, err := q.Pop(timeoutCtx, &rcv)
		require.NoError(t, err)
		require.False(t, success)
	}

	{
		// Item eventually received
		var rcv string
		timeoutCtx, _ := context.WithTimeout(ctx, 20*time.Millisecond)
		success, err := q.Pop(timeoutCtx, &rcv)
		require.NoError(t, err)
		require.True(t, success)
		require.Equal(t, second, rcv)
	}
}
