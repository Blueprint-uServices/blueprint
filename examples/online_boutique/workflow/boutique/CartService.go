package boutique

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type CartService interface {
	AddItem(ctx context.Context, userID string, item CartItem) error
	GetCart(ctx context.Context, userID string) ([]CartItem, error)
	EmptyCart(ctx context.Context, userID string) error
}

type CartServiceImpl struct {
	cache  backend.Cache
	logger backend.Logger
}

func NewCartServiceImpl(ctx context.Context, cache backend.Cache) (CartService, error) {
	return &CartServiceImpl{cache: cache, logger: backend.GetLogger()}, nil
}

func (s *CartServiceImpl) AddItem(ctx context.Context, userID string, item CartItem) error {
	ctx, _ = s.logger.Info(ctx, "AddItem called: userID=%s productID=%s quantity=%d", userID, item.ProductID, item.Quantity)
	if userID == "" || item.ProductID == "" || item.Quantity <= 0 {
		return fmt.Errorf("invalid cart item")
	}
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}
	out := make([]CartItem, 0, len(cart)+1)
	found := false
	for _, current := range cart {
		if current.ProductID == item.ProductID {
			current.Quantity += item.Quantity
			found = true
		}
		out = append(out, current)
	}
	if !found {
		out = append(out, item)
	}
	return s.cache.Put(ctx, userID, out)
}

func (s *CartServiceImpl) GetCart(ctx context.Context, userID string) ([]CartItem, error) {
	ctx, _ = s.logger.Info(ctx, "GetCart called: userID=%s", userID)
	var cart []CartItem
	found, err := s.cache.Get(ctx, userID, &cart)
	if err != nil {
		return nil, err
	}
	if !found {
		return []CartItem{}, nil
	}
	return append([]CartItem(nil), cart...), nil
}

func (s *CartServiceImpl) EmptyCart(ctx context.Context, userID string) error {
	ctx, _ = s.logger.Info(ctx, "EmptyCart called: userID=%s", userID)
	return s.cache.Delete(ctx, userID)
}
