package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/petshop/workflow/petshop"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var petSiteRegistry = registry.NewServiceRegistry[petshop.PetSiteService]("pet_site_service")

func init() {
	petSiteRegistry.Register("local", func(ctx context.Context) (petshop.PetSiteService, error) {
		search, err := petSearchRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		adoption, err := payForAdoptionRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		history, err := petListAdoptionsRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		food, err := petFoodRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		agent, err := petFoodAgentRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return petshop.NewPetSiteServiceImpl(ctx, search, adoption, history, food, agent)
	})
}

func TestPetSiteService(t *testing.T) {
	ctx := context.Background()
	service, err := petSiteRegistry.Get(ctx)
	require.NoError(t, err)
	home, err := service.Home(ctx, "site-user", petshop.PetSearchQuery{PetType: "bunny"})
	require.NoError(t, err)
	require.NotEmpty(t, home.Pets)
	require.NotEmpty(t, home.Foods)
	adoption, err := service.Adopt(ctx, "site-user", "bunny", "023")
	require.NoError(t, err)
	require.NotEmpty(t, adoption.TransactionID)
	history, err := service.AdoptionHistory(ctx, "site-user")
	require.NoError(t, err)
	require.Len(t, history, 1)
}
