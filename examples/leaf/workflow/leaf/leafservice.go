package leaf

import (
	ctxx "context"
)

type LeafObject struct {
	ID   int64
	Name string
}

type LeafService interface {
	HelloInt(ctx ctxx.Context, a int64) (int64, error)
	HelloObject(ctx *ctxx.Context, obj LeafObject) (LeafObject, error)
}

type LeafServiceImpl struct {
	LeafService
}

func (l *LeafServiceImpl) HelloInt(ctx ctxx.Context, a int64) (int64, error) {
	return a, nil
}

func (l *LeafServiceImpl) HelloObject(ctx *ctxx.Context, obj LeafObject) (LeafObject, error) {
	return obj, nil
}

func NewLeafServiceImpl() *LeafServiceImpl {
	return &LeafServiceImpl{}
}
