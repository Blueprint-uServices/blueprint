package clientpool_test

import (
	"context"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/plugins/clientpool"
	"github.com/stretchr/testify/require"
)

type element struct {
	i int
}

func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func TestClientPool(t *testing.T) {

	i := 0
	build := func() (*element, error) {
		e := &element{i: i}
		i += 1
		return e, nil
	}

	cap := 5
	pool := clientpool.NewClientPool[*element](cap, build)

	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	es := []*element{}
	for i := 0; i < 3; i++ {
		require.Equal(t, i, pool.Size(), "iteration %v", i)
		e, err := pool.Pop(ctx)
		require.NoError(t, err)
		require.Equal(t, i, e.i)
		es = append(es, e)
		require.Equal(t, 0, pool.Available(), "iteration %v", i)
		require.Equal(t, i+1, pool.Size(), "iteration %v", i)
		require.Equal(t, cap, pool.Capacity(), "iteration %v", i)
		require.False(t, isDone(ctx), "iteration %v", i)
	}

	for i := range es {
		require.Equal(t, 3, pool.Size(), "iteration %v", i)
		pool.Push(es[i])
		require.Equal(t, i+1, pool.Available(), "iteration %v", i)
		require.Equal(t, 3, pool.Size(), "iteration %v", i)
		require.False(t, isDone(ctx), "iteration %v", i)
	}

	es = []*element{}
	for i := 0; i < 3; i++ {
		require.Equal(t, 3-i, pool.Available(), "iteration %v", i)
		require.Equal(t, 3, pool.Size(), "iteration %v", i)
		require.Equal(t, cap, pool.Capacity(), "iteration %v", i)
		e, err := pool.Pop(ctx)
		require.NoError(t, err)
		require.Equal(t, i, e.i)
		es = append(es, e)
		require.Equal(t, 3-i-1, pool.Available(), "iteration %v", i)
		require.Equal(t, 3, pool.Size(), "iteration %v", i)
		require.Equal(t, cap, pool.Capacity(), "iteration %v", i)
		require.False(t, isDone(ctx), "iteration %v", i)
	}

	for i := 3; i < 5; i++ {
		require.Equal(t, i, pool.Size(), "iteration %v", i)
		e, err := pool.Pop(ctx)
		require.NoError(t, err)
		require.Equal(t, i, e.i)
		es = append(es, e)
		require.Equal(t, 0, pool.Available(), "iteration %v", i)
		require.Equal(t, i+1, pool.Size(), "iteration %v", i)
		require.Equal(t, cap, pool.Capacity(), "iteration %v", i)
		require.False(t, isDone(ctx), "iteration %v", i)
	}

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Millisecond)

	{
		require.Equal(t, cap, pool.Size())
		_, err := pool.Pop(ctx)
		require.Error(t, err)
		require.True(t, isDone(ctx))
	}

	ctx, _ = context.WithTimeout(context.Background(), 1*time.Second)

	go func() {
		for _, e := range es {
			time.Sleep(10 * time.Millisecond)
			pool.Push(e)
		}
	}()

	es2 := []*element{}
	for i := 0; i < 5; i++ {
		require.Equal(t, cap, pool.Size(), "iteration %v", i)
		require.Equal(t, cap, pool.Capacity(), "iteration %v", i)
		require.Equal(t, 0, pool.Available(), "iteration %v", i)
		e, err := pool.Pop(ctx)
		require.NoError(t, err)
		require.Equal(t, i, e.i)
		es2 = append(es2, e)
		require.Equal(t, cap, pool.Size(), "iteration %v", i)
		require.Equal(t, cap, pool.Capacity(), "iteration %v", i)
		require.Equal(t, 0, pool.Available(), "iteration %v", i)
		require.False(t, isDone(ctx), "iteration %v", i)
	}
}
