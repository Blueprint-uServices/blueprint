package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/petshop/workflow/petshop"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var petSearchRegistry = registry.NewServiceRegistry[petshop.PetSearchService]("pet_search_service")

func init() {
	petSearchRegistry.Register("local", func(ctx context.Context) (petshop.PetSearchService, error) {
		db, _, _, err := localBackends(ctx)
		if err != nil {
			return nil, err
		}
		return petshop.NewPetSearchServiceImpl(ctx, db, "https://images.example.com")
	})
}

func TestPetSearchService(t *testing.T) {
	ctx := context.Background()
	service, err := petSearchRegistry.Get(ctx)
	require.NoError(t, err)
	pets, err := service.Search(ctx, petshop.PetSearchQuery{PetType: "puppy", PetColor: "black"})
	require.NoError(t, err)
	require.NotEmpty(t, pets)
	for _, pet := range pets {
		require.Equal(t, "puppy", pet.PetType)
		require.Equal(t, "black", pet.PetColor)
		require.Contains(t, pet.PetURL, "/puppies/")
	}
	_, err = service.Search(ctx, petshop.PetSearchQuery{PetType: "dragon"})
	require.ErrorIs(t, err, petshop.ErrInvalidPetType)
}
