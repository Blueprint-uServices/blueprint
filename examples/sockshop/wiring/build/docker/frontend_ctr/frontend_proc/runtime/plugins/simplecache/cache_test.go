package simplecache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	ctx := context.Background()
	cache, _ := NewSimpleCache(ctx)

	err := cache.Put(ctx, "hello", "world")
	assert.NoError(t, err)

	var v string
	exists, err := cache.Get(ctx, "hello", &v)
	assert.True(t, exists)
	assert.NoError(t, err)

	assert.Equal(t, v, "world")

	var j int
	exists, err = cache.Get(ctx, "nonexistent", &j)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Can't cast string to int
	var i int
	exists, err = cache.Get(ctx, "hello", &i)
	assert.True(t, exists)
	assert.Error(t, err)
}

func TestIncr(t *testing.T) {
	ctx := context.Background()
	cache, _ := NewSimpleCache(ctx)

	for j := 0; j < 10; j++ {
		jv, err := cache.Incr(ctx, "nothing")
		assert.NoError(t, err)
		assert.Equal(t, jv, int64(j+1))
	}
}

func TestMget(t *testing.T) {
	ctx := context.Background()
	cache, _ := NewSimpleCache(ctx)

	keys := []string{"a", "b"}
	values := []interface{}{int64(5), "hello"}
	err := cache.Mset(ctx, keys, values)
	assert.NoError(t, err)

	var va int64
	var vb string
	getvalues := []interface{}{&va, &vb}
	err = cache.Mget(ctx, keys, getvalues)
	assert.NoError(t, err)
	assert.Equal(t, va, int64(5))
	assert.Equal(t, vb, "hello")

	err = cache.Mset(ctx, keys, []interface{}{})
	assert.Error(t, err)

	err = cache.Mset(ctx, []string{}, values)
	assert.Error(t, err)

	err = cache.Mget(ctx, keys, []interface{}{})
	assert.Error(t, err)

	err = cache.Mget(ctx, []string{}, getvalues)
	assert.Error(t, err)
}
