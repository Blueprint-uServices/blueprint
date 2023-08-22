package leaf

import (
	"context"
)

type LeafObject struct {
	ID   int64
	Name string
}

type LeafService interface {
	HelloInt(ctx context.Context, a int64) (int64, error)
	HelloObject(ctx context.Context, obj LeafObject) (LeafObject, error)
}

type LeafServiceImpl struct {
	LeafService
}

func (l *LeafServiceImpl) HelloInt(ctx context.Context, a int64) (int64, error) {
	return a, nil
}

func (l *LeafServiceImpl) HelloObject(ctx context.Context, obj LeafObject) (LeafObject, error) {
	return obj, nil
}

func NewLeafServiceImpl() *LeafServiceImpl {
	return &LeafServiceImpl{}
}
