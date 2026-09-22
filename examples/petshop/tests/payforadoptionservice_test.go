package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/petshop/workflow/petshop"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var payForAdoptionRegistry = registry.NewServiceRegistry[petshop.PayForAdoptionService]("pay_for_adoption_service")

func init() {
	payForAdoptionRegistry.Register("local", func(ctx context.Context) (petshop.PayForAdoptionService, error) {
		db, _, queue, err := localBackends(ctx)
		if err != nil {
			return nil, err
		}
		search, err := petSearchRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return petshop.NewPayForAdoptionServiceImpl(ctx, db, search, queue)
	})
}

func TestPayForAdoptionService(t *testing.T) {
	ctx := context.Background()
	service, err := payForAdoptionRegistry.Get(ctx)
	require.NoError(t, err)
	adoption, err := service.CompleteAdoption(ctx, "001", "puppy", "adopter")
	require.NoError(t, err)
	require.NotEmpty(t, adoption.TransactionID)
	_, err = service.CompleteAdoption(ctx, "001", "puppy", "other")
	require.ErrorIs(t, err, petshop.ErrPetUnavailable)
	require.NoError(t, service.CleanupAdoptions(ctx, "adopter"))
}
