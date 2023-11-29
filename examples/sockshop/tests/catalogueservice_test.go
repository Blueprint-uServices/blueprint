package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/catalogue"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplereldb"
)

// Tests acquire a CatalogueService instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var catalogueRegistry = registry.NewServiceRegistry[catalogue.CatalogueService]("catalogue_service")

func init() {
	// If the tests are run locally, we fall back to this CatalogueService implementation
	catalogueRegistry.Register("local", func(ctx context.Context) (catalogue.CatalogueService, error) {
		db, err := simplereldb.NewSimpleRelDB(ctx)
		if err != nil {
			return nil, err
		}

		return catalogue.NewCatalogueService(ctx, db)
	})
}

func TestCatalogueService(t *testing.T) {
	ctx := context.Background()
	service, err := catalogueRegistry.Get(ctx)
	require.NoError(t, err)

	{
		catalogueTags, err := service.Tags(ctx)
		require.NoError(t, err)
		require.Equal(t, []string{}, catalogueTags)
	}

	tags := []string{"blue", "brown", "green"}

	{
		// Add new tags
		err := service.AddTags(ctx, tags...)
		require.NoError(t, err)
	}

	{
		// Check they exist
		catalogueTags, err := service.Tags(ctx)
		require.NoError(t, err)
		require.Equal(t, tags, catalogueTags)
	}

	{
		// Try re-adding tags
		err := service.AddTags(ctx, tags...)
		require.NoError(t, err)
	}

	{
		// Check no duplicates
		catalogueTags, err := service.Tags(ctx)
		require.NoError(t, err)
		require.Equal(t, tags, catalogueTags)
	}

	tags2 := []string{"red"}

	{
		// Add new tags
		err := service.AddTags(ctx, tags2...)
		require.NoError(t, err)
	}

	{
		// Check they exist
		catalogueTags, err := service.Tags(ctx)
		require.NoError(t, err)
		require.Len(t, catalogueTags, 4)
		require.Equal(t, tags, catalogueTags[:3])
		require.Equal(t, tags2, catalogueTags[3:])
	}

	{
		// Add a sock
		sock := catalogue.Sock{
			Name:        "mysock",
			Description: "A Sock",
			Price:       1.99,
			Quantity:    5,
			Tags:        []string{"blue", "red"},
		}
		id, err := service.AddSock(ctx, sock)
		require.NoError(t, err)
		fmt.Printf("sock ID is %v\n", id)

		sock2, err := service.Get(ctx, id)
		require.NoError(t, err)
		fmt.Printf("received sock %v\n", sock2)
	}
}
