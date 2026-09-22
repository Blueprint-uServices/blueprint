package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/petshop/workflow/petshop"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var petFoodAgentRegistry = registry.NewServiceRegistry[petshop.PetFoodAgentService]("pet_food_agent_service")

func init() {
	petFoodAgentRegistry.Register("local", func(ctx context.Context) (petshop.PetFoodAgentService, error) {
		search, err := petSearchRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		food, err := petFoodRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return petshop.NewPetFoodAgentServiceImpl(ctx, search, food)
	})
}

func TestPetFoodAgentService(t *testing.T) {
	ctx := context.Background()
	service, err := petFoodAgentRegistry.Get(ctx)
	require.NoError(t, err)
	recommendations, err := service.Recommend(ctx, "agent-user", "What should I feed my kitten?", "")
	require.NoError(t, err)
	require.NotEmpty(t, recommendations)
	for _, recommendation := range recommendations {
		require.Equal(t, "kitten", recommendation.Food.PetType)
		require.NotEmpty(t, recommendation.Reason)
	}
}
