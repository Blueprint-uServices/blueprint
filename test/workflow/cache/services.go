package cache

import (
	ctxx "context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/blueprint-uservices/blueprint/test/workflow/workflow"
)

/*
Implements the services from ../workflow using a cache
*/

/*
Service implementation structs
*/
type (
	TestLeafServiceImplWithCache struct {
		workflow.TestLeafService
		Cache backend.Cache
	}
)

/*
Constructors
*/

func NewTestLeafServiceImplWithCache(ctx ctxx.Context, cache backend.Cache) (*TestLeafServiceImplWithCache, error) {
	return &TestLeafServiceImplWithCache{Cache: cache}, nil
}

/*
Interface method bodies
*/

func (l *TestLeafServiceImplWithCache) HelloNothing(ctx ctxx.Context) error {
	return nil
}

func (l *TestLeafServiceImplWithCache) HelloInt(ctx ctxx.Context, a int16) (int32, error) {
	err := l.Cache.Put(ctx, "myint", int32(a))
	if err != nil {
		return 0, err
	}
	var myint int32
	_, err = l.Cache.Get(ctx, "myint", &myint)
	return myint, err
}

func (l *TestLeafServiceImplWithCache) HelloObject(ctx ctxx.Context, obj workflow.TestLeafObject) (*workflow.TestLeafObject, error) {
	var count int64
	_, err := l.Cache.Get(ctx, "objectcount", &count)
	if err != nil {
		return nil, err
	}
	count += 10
	obj.Count = int(count)
	return &obj, l.Cache.Put(ctx, "objectcount", count)
}
