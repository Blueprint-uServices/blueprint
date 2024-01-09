package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var uniqueIdServiceRegistry = registry.NewServiceRegistry[socialnetwork.UniqueIdService]("uniqueId_service")

func init() {

	uniqueIdServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.UniqueIdService, error) {
		return socialnetwork.NewUniqueIdServiceImpl(ctx)
	})
}

func TestComposeUniqueId(t *testing.T) {
	ctx := context.Background()
	service, err := uniqueIdServiceRegistry.Get(ctx)
	require.NoError(t, err)

	id, err := service.ComposeUniqueId(ctx, 1000, socialnetwork.POST)
	require.NoError(t, err)
	require.Positive(t, id)
}
