package leaf

import (
	"context"
)

type NonLeafService interface {
	Hello(ctx context.Context, a int64) (int64, error)
}

type NonLeafServiceImpl struct {
	NonLeafService
	leafService LeafService
}

func NewNonLeafServiceImpl(leafService LeafService) (NonLeafService, error) {
	nonleaf := &NonLeafServiceImpl{}
	nonleaf.leafService = leafService
	return nonleaf, nil
}

func (nl *NonLeafServiceImpl) Hello(ctx context.Context, a int64) (int64, error) {
	return nl.leafService.HelloInt(ctx, a)
}
