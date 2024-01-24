package memcached

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type someData struct {
	ID   int64
	Name string
}

func equal(d1 someData, d2 someData) bool {
	return (d1.ID == d2.ID && d1.Name == d2.Name)
}

func TestMemcachedPut(t *testing.T) {
	ctx := context.Background()
	memcached, err := NewMemcachedClient(ctx, "localhost:11211")
	if err != nil {
		t.Error(err)
	}
	data := someData{ID: 5, Name: "Vaastav"}
	err = memcached.Put(ctx, "testData", data)
	if err != nil {
		t.Error(err)
	}
	var resultData someData
	exists, err := memcached.Get(ctx, "testData", &resultData)
	assert.True(t, exists)
	if err != nil {
		t.Error(err)
	}
	if !equal(data, resultData) {
		t.Errorf("Incorrect data received from server: Expected: %v, Actual: %v", data, resultData)
	}
}

func TestMemcachedGet(t *testing.T) {
	ctx := context.Background()
	memcached, err := NewMemcachedClient(ctx, "localhost:11211")
	if err != nil {
		t.Error(err)
	}
	var resultData someData
	exists, err := memcached.Get(ctx, "testData", &resultData)
	assert.True(t, exists)
	if err != nil {
		t.Error(err)
	}
	if resultData.ID != 5 || resultData.Name != "Vaastav" {
		t.Errorf("Incorrect data received from server")
	}
}

func TestMemcachedIncr(t *testing.T) {
	ctx := context.Background()
	memcached, err := NewMemcachedClient(ctx, "localhost:11211")
	err = memcached.Put(ctx, "intKey", 5)
	if err != nil {
		t.Error(err)
	}
	val, err := memcached.Incr(ctx, "intKey")
	if err != nil {
		t.Error(err)
	}
	if val != 6 {
		t.Errorf("Incorrect data received. Expected: 6, Actual %d", val)
	}
}

func TestMemcachedDelete(t *testing.T) {
	ctx := context.Background()
	memcached, err := NewMemcachedClient(ctx, "localhost:11211")
	if err != nil {
		t.Error(err)
	}
	err = memcached.Put(ctx, "deleteKey", 6)
	if err != nil {
		t.Error(err)
	}
	var val int
	exists, err := memcached.Get(ctx, "deleteKey", &val)
	assert.True(t, exists)
	if err != nil {
		t.Error(err)
	}
	if val != 6 {
		t.Errorf("Setup failed")
	}
	err = memcached.Delete(ctx, "deleteKey")
	if err != nil {
		t.Error(err)
	}
	var newval int
	exists, err = memcached.Get(ctx, "deleteKey", &newval)
	assert.False(t, exists)
	assert.NoError(t, err)
	if newval != 0 {
		t.Errorf("Delete followed by a Get returned non-zero value")
	}
}

func TestMemcachedMget(t *testing.T) {
	var val1 someData
	var val2 int
	ctx := context.Background()
	keys := []string{"testData", "intKey"}
	vals := []interface{}{&val1, &val2}
	memcached, err := NewMemcachedClient(ctx, "localhost:11211")
	if err != nil {
		t.Error(err)
	}
	err = memcached.Mget(ctx, keys, vals)
	if err != nil {
		t.Error(err)
	}
	if val2 != 6 {
		t.Errorf("Incorrect value received from server. Expected: 6, Actual: %d", val2)
	}
	if val1.ID != 5 || val1.Name != "Vaastav" {
		t.Errorf("Incorrect value received from server. Expected: {5 Vaastav}, Actual: %v", val1)
	}
}

func TestMemcachedMset(t *testing.T) {
	ctx := context.Background()
	memcached, err := NewMemcachedClient(ctx, "localhost:11211")
	if err != nil {
		t.Error(err)
	}
	keys := []string{"newKey", "testData", "intKey"}
	testData := someData{ID: 7, Name: "NotVaastav"}
	new_vals := []interface{}{6, testData, 5}

	err = memcached.Mset(ctx, keys, new_vals)
	if err != nil {
		t.Error(err)
	}

	var val0 int
	var val1 someData
	var val2 int

	vals := []interface{}{&val0, &val1, &val2}
	err = memcached.Mget(ctx, keys, vals)
	if err != nil {
		t.Error(err)
	}
	if val0 != 6 {
		t.Errorf("Incorrect value received from server. Expected: 6, Actual: %d", val0)
	}
	if val2 != 5 {
		t.Errorf("Incorrect value received from server. Expected: 5, Actual: %d", val2)
	}
	if val1.ID != 7 || val1.Name != "NotVaastav" {
		t.Errorf("Incorrect value received from server. Expected: {7 NotVaastav}, Actual: %v", val1)
	}
}
