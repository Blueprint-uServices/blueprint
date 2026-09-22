package tests

import (
	"context"
	"sync"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplequeue"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/slogger"
)

var local struct {
	sync.Mutex
	initialized bool
	db          backend.NoSQLDatabase
	cache       backend.Cache
	queue       backend.Queue
}

func localBackends(ctx context.Context) (backend.NoSQLDatabase, backend.Cache, backend.Queue, error) {
	local.Lock()
	defer local.Unlock()
	if local.initialized {
		return local.db, local.cache, local.queue, nil
	}
	if _, err := slogger.NewSLogger(ctx); err != nil {
		return nil, nil, nil, err
	}
	db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	cache, err := simplecache.NewSimpleCache(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	queue, err := simplequeue.NewSimpleQueue(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	local.db, local.cache, local.queue, local.initialized = db, cache, queue, true
	return db, cache, queue, nil
}
