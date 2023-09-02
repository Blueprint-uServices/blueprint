package leaf

import (
	ctxx "context"
	"fmt"

	. "gitlab.mpi-sws.org/cld/blueprint/examples/leaf/workflow/leaf/example"
)

type MyInt int64

type LeafObject struct {
	ID   int64
	Name string
}

type LeafService interface {
	HelloInt(ctx ctxx.Context, a int64) (int64, error)
	HelloObject(ctx ctxx.Context, obj *LeafObject) (*LeafObject, error)
}

type LeafServiceImpl struct {
	LeafService
}

func (l *LeafServiceImpl) HelloInt(ctx ctxx.Context, a int64) (int64, error) {
	Hi()
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
