package leaf

import (
	"context"
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

type NonLeafService interface {
	Hello(ctx context.Context, a int64) (int64, error)
}

type NonLeafServiceImpl struct {
	NonLeafService
	leafService LeafService
}

func NewNonLeafServiceImpl(ctx context.Context, leafService LeafService) (NonLeafService, error) {
	nonleaf := &NonLeafServiceImpl{}
	nonleaf.leafService = leafService
	return nonleaf, nil
}

func (nl *NonLeafServiceImpl) Hello(ctx context.Context, a int64) (int64, error) {
	a, err := nl.leafService.HelloInt(ctx, a)
	if err != nil {
		return a, err
	}

	err = nl.leafService.HelloNothing(ctx)
	if err != nil {
		return 0, err
	}

	var b int32
	b = int32(a * 2)

	c := fmt.Sprintf("hello %v", a)

	d := make(map[string]LeafObject)
	dc := LeafObject{
		ID:    a,
		Name:  c,
		Props: make(map[string]NestedLeafObject),
	}
	d[c] = dc
	d[c].Props["hello"] = NestedLeafObject{
		Key:   "greetings",
		Value: "mate",
		Props: []string{"cool", "beans"},
	}
	// string, []string, int32, int, map[string]LeafObject, error)
	ra, rb, rc, rd, re, err := nl.leafService.HelloMate(ctx, int(a), b, c, d, []string{"hi", "bye"}, nil)
	if err != nil {
		return a, err
	}

	ctx, _ = backend.GetLogger().Info(ctx, ra)
	ctx, _ = backend.GetLogger().Info(ctx, fmt.Sprintf("%v", rb))
	ctx, _ = backend.GetLogger().Info(ctx, fmt.Sprintf("%v", rc))
	ctx, _ = backend.GetLogger().Info(ctx, fmt.Sprintf("%v", rd))
	ctx, _ = backend.GetLogger().Info(ctx, fmt.Sprintf("%v", re))

	return a, nil
}
