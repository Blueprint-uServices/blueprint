package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var recommendationServiceRegistry = registry.NewServiceRegistry[boutique.RecommendationService]("recommendation_service")

func init() {
	recommendationServiceRegistry.Register("local", func(ctx context.Context) (boutique.RecommendationService, error) {
		setupLogger(ctx)
		catalog, err := productCatalogServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return boutique.NewRecommendationServiceImpl(ctx, catalog)
	})
}

func TestRecommendationService(t *testing.T) {
	ctx := context.Background()
	service, err := recommendationServiceRegistry.Get(ctx)
	require.NoError(t, err)

	recommendations, err := service.ListRecommendations(ctx, "recommendation-user", []string{"OLJCESPC7Z", "66VCHSJNUP"})
	require.NoError(t, err)
	require.Len(t, recommendations, 5)
	require.NotContains(t, recommendations, "OLJCESPC7Z")
	require.NotContains(t, recommendations, "66VCHSJNUP")
}
