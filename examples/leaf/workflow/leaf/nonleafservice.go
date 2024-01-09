package leaf

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type NonLeafService interface {
	Hello(ctx context.Context, a int64) (int64, error)
}

type NonLeafServiceImpl struct {
	NonLeafService
	leafService LeafService
	logger      backend.Logger
}

func NewNonLeafServiceImpl(ctx context.Context, leafService LeafService) (NonLeafService, error) {
	nonleaf := &NonLeafServiceImpl{}
	nonleaf.leafService = leafService
	nonleaf.logger = backend.GetLogger()
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

	ctx, _ = nl.logger.Info(ctx, "ra: %s", ra)
	ctx, _ = nl.logger.Info(ctx, "rb: %v", rb)
	ctx, _ = nl.logger.Info(ctx, "rc: %d", rc)
	ctx, _ = nl.logger.Info(ctx, "rd: %d", rd)
	ctx, _ = nl.logger.Info(ctx, "re: %v", re)

	return a, nil
}
