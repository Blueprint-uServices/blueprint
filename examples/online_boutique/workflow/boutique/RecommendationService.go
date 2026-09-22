package boutique

import (
	"context"
	"math/rand"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type RecommendationService interface {
	ListRecommendations(ctx context.Context, userID string, productIDs []string) ([]string, error)
}

type RecommendationServiceImpl struct {
	catalogService ProductCatalogService
	logger         backend.Logger
}

func NewRecommendationServiceImpl(ctx context.Context, catalogService ProductCatalogService) (RecommendationService, error) {
	return &RecommendationServiceImpl{catalogService: catalogService, logger: backend.GetLogger()}, nil
}

func (s *RecommendationServiceImpl) ListRecommendations(ctx context.Context, userID string, productIDs []string) ([]string, error) {
	products, err := s.catalogService.ListProducts(ctx)
	if err != nil {
		return nil, err
	}
	excluded := make(map[string]struct{}, len(productIDs))
	for _, id := range productIDs {
		excluded[id] = struct{}{}
	}
	filtered := make([]string, 0, len(products))
	for _, product := range products {
		if _, ok := excluded[product.ID]; !ok {
			filtered = append(filtered, product.ID)
		}
	}
	perm := rand.Perm(len(filtered))
	if len(perm) > 5 {
		perm = perm[:5]
	}
	recommendations := make([]string, len(perm))
	for i, index := range perm {
		recommendations[i] = filtered[index]
	}
	return recommendations, nil
}
