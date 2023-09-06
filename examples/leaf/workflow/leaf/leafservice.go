package leaf

import (
	ctxx "context"
	"fmt"
)

type MyInt int64

type NestedLeafObject struct {
	Key   string
	Value string
}

type LeafObject struct {
	ID    int64
	Name  string
	Props map[string]NestedLeafObject
}

type LeafService interface {
	HelloInt(ctx ctxx.Context, a int64) (int64, error)
	HelloObject(ctx ctxx.Context, obj *LeafObject) (*LeafObject, error)
}

type LeafServiceImpl struct {
	LeafService
}

func (l *LeafServiceImpl) HelloInt(ctx ctxx.Context, a int64) (int64, error) {
	fmt.Println("hello")
	return a, nil
}

func (l *LeafServiceImpl) HelloObject(ctx ctxx.Context, obj *LeafObject) (*LeafObject, error) {
	return obj, nil
}

func (l *LeafServiceImpl) NonServiceFunction() int64 {
	return 3
}

func NewLeafServiceImpl() (*LeafServiceImpl, error) {
	return &LeafServiceImpl{}, nil
}
