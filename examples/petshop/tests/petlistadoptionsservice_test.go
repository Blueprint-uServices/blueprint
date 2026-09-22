package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/petshop/workflow/petshop"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var petListAdoptionsRegistry = registry.NewServiceRegistry[petshop.PetListAdoptionsService]("pet_list_adoptions_service")

func init() {
	petListAdoptionsRegistry.Register("local", func(ctx context.Context) (petshop.PetListAdoptionsService, error) {
		db, _, _, err := localBackends(ctx)
		if err != nil {
			return nil, err
		}
		search, err := petSearchRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return petshop.NewPetListAdoptionsServiceImpl(ctx, db, search)
	})
}

func TestPetListAdoptionsService(t *testing.T) {
	ctx := context.Background()
	payment, err := payForAdoptionRegistry.Get(ctx)
	require.NoError(t, err)
	_, err = payment.CompleteAdoption(ctx, "002", "puppy", "history-user")
	require.NoError(t, err)
	service, err := petListAdoptionsRegistry.Get(ctx)
	require.NoError(t, err)
	history, err := service.ListAdoptions(ctx, "history-user", 25)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, "002", history[0].Pet.PetID)
}
