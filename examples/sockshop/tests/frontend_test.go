package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/frontend"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
)

// Tests acquire a Frontend instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var frontendRegistry = registry.NewServiceRegistry[frontend.Frontend]("frontend")

func init() {
	// If the tests are run locally, we fall back to this Frontend implementation
	frontendRegistry.Register("local", func(ctx context.Context) (frontend.Frontend, error) {
		user, err := userServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		cart, err := cartRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		catalogue, err := catalogueRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		order, err := ordersRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return frontend.NewFrontend(ctx, user, catalogue, cart, order)
	})
}

func TestFrontend(t *testing.T) {
	ctx := context.Background()
	_, err := frontendRegistry.Get(ctx)
	require.NoError(t, err)
}
