package simplecache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	cache, _ := NewSimpleCache()
	ctx := context.Background()

	err := cache.Put(ctx, "hello", "world")
	assert.NoError(t, err)

	var v string
	err = cache.Get(ctx, "hello", &v)
	assert.NoError(t, err)

	assert.Equal(t, v, "world")

	var j int
	err = cache.Get(ctx, "nonexistent", &j)
	assert.NoError(t, err)
	assert.Equal(t, j, 0)

	// Can't cast string to int
	var i int
	err = cache.Get(ctx, "hello", &i)
	assert.Error(t, err)
}

func TestIncr(t *testing.T) {
	cache, _ := NewSimpleCache()
	ctx := context.Background()

	for j := 0; j < 10; j++ {
		jv, err := cache.Incr(ctx, "nothing")
		assert.NoError(t, err)
		assert.Equal(t, jv, int64(j+1))
	}
}

func TestMget(t *testing.T) {
	cache, _ := NewSimpleCache()
	ctx := context.Background()

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
