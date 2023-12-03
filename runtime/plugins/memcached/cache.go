package memcached

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

type Memcached struct {
	backend.Cache
	Client *memcache.Client
}

func NewMemcachedClient(ctx context.Context, serverAddress string) (*Memcached, error) {
	cache := &Memcached{}
	client := memcache.New(serverAddress)
	client.MaxIdleConns = 1000
	cache.Client = client
	return cache, nil
}

func (m *Memcached) Put(ctx context.Context, key string, value interface{}) error {
	marshaled_val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.Client.Set(&memcache.Item{Key: key, Value: marshaled_val})
}

func (m *Memcached) Get(ctx context.Context, key string, value interface{}) (bool, error) {
	it, err := m.Client.Get(key)
	if err == memcache.ErrCacheMiss {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(it.Value, value)
}

func (m *Memcached) Incr(ctx context.Context, key string) (int64, error) {
	val, err := m.Client.Increment(key, 1)
	return int64(val), err
}

func (m *Memcached) Delete(ctx context.Context, key string) error {
	return m.Client.Delete(key)
}

func (m *Memcached) Mget(ctx context.Context, keys []string, values []interface{}) error {
	val_map, err := m.Client.GetMulti(keys)
	if err != nil {
		return err
	}
	for idx, key := range keys {
		if val, ok := val_map[key]; ok {
			err := json.Unmarshal(val.Value, values[idx])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Memcached) Mset(ctx context.Context, keys []string, values []interface{}) error {
	var wg sync.WaitGroup
	wg.Add(len(keys))
	err_chan := make(chan error, len(keys))
	for idx, key := range keys {
		go func(key string, val interface{}) {
			defer wg.Done()
			err_chan <- m.Put(ctx, key, val)
		}(key, values[idx])
	}
	wg.Wait()
	close(err_chan)
	for err := range err_chan {
		if err != nil {
			return err
		}
	}
	return nil
}
