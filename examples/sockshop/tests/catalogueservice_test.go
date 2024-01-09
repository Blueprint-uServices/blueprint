package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/catalogue"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/sqlitereldb"
	"github.com/stretchr/testify/require"
)

// Tests acquire a CatalogueService instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var catalogueRegistry = registry.NewServiceRegistry[catalogue.CatalogueService]("catalogue_service")

func init() {
	// If the tests are run locally, we fall back to this CatalogueService implementation
	catalogueRegistry.Register("local", func(ctx context.Context) (catalogue.CatalogueService, error) {
		db, err := sqlitereldb.NewSqliteRelDB(ctx)
		if err != nil {
			return nil, err
		}

		return catalogue.NewCatalogueService(ctx, db)
	})

	// // Manually switch over to this implementation to test against a locally-deployed mysql server
	// // To run using mysql, start a mysql docker container with:
	// //   docker run -p 3306:3306 --env MYSQL_ROOT_PASSWORD=pass
	// catalogueRegistry.Register("mysql", func(ctx context.Context) (catalogue.CatalogueService, error) {
	// 	db, err := mysql.NewMySqlDB(ctx, "localhost:3306", "catalogue_db", "root", "pass")
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return catalogue.NewCatalogueService(ctx, db)
	// })
}

func TestCatalogueService(t *testing.T) {
	ctx := context.Background()
	service, err := catalogueRegistry.Get(ctx)
	require.NoError(t, err)

	{
		catalogueTags, err := service.Tags(ctx)
		require.NoError(t, err)
		require.Empty(t, catalogueTags)
	}

	tags := []string{"brown", "blue", "green"}

	{
		// Add new tags
		err := service.AddTags(ctx, tags)
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
		err := service.AddTags(ctx, tags)
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
		err := service.AddTags(ctx, tags2)
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
		// List socks; should be empty
		socks, err := service.List(ctx, tags2, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 0)
	}

	{
		// Count socks, should be 0
		count, err := service.Count(ctx, tags2)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	}

	sock := catalogue.Sock{
		Name:        "mysock",
		Description: "A Sock",
		Price:       1.99,
		Quantity:    5,
		Tags:        []string{"blue", "red"},
	}

	{
		// Add a sock
		id, err := service.AddSock(ctx, sock)
		require.NoError(t, err)

		res, err := service.Get(ctx, id)
		require.NoError(t, err)
		requireSocksEqual(t, sock, res)
	}

	{
		// List blue socks; should have 1
		socks, err := service.List(ctx, []string{"blue"}, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 1)
		requireSocksEqual(t, sock, socks[0])
	}

	{
		// Count blue socks, should be 1
		count, err := service.Count(ctx, []string{"blue"})
		require.NoError(t, err)
		require.Equal(t, 1, count)
	}

	{
		// List blue, brown, or green socks; should have 1
		socks, err := service.List(ctx, tags, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 1)
		requireSocksEqual(t, sock, socks[0])
	}

	{
		// Count blue, brown, or green socks, should be 1
		count, err := service.Count(ctx, tags)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	}

	{
		// List green socks; should have 0
		socks, err := service.List(ctx, []string{"green"}, "", 1, 1000)
		require.NoError(t, err)
		require.Empty(t, socks)
	}

	{
		// Count green socks; should have 0
		count, err := service.Count(ctx, []string{"green"})
		require.NoError(t, err)
		require.Equal(t, 0, count)
	}

	{
		// List all socks; should have 1
		socks, err := service.List(ctx, nil, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 1)
		requireSocksEqual(t, sock, socks[0])
	}

	sock2 := catalogue.Sock{
		Name:        "my second sock",
		Description: "B Sock",
		Price:       3.49,
		Quantity:    11,
		Tags:        []string{"red", "green"},
	}

	{
		// Add another sock
		id, err := service.AddSock(ctx, sock2)
		require.NoError(t, err)

		res, err := service.Get(ctx, id)
		require.NoError(t, err)
		requireSocksEqual(t, sock2, res)
	}

	{
		// List blue socks; should have 1
		socks, err := service.List(ctx, []string{"blue"}, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 1)
		requireSocksEqual(t, sock, socks[0])
	}

	{
		// List green socks; should have 1
		socks, err := service.List(ctx, []string{"green"}, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 1)
		requireSocksEqual(t, sock2, socks[0])
	}

	{
		// List all socks; should have 2
		socks, err := service.List(ctx, nil, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 2)
		requireSock(t, sock, socks)
		requireSock(t, sock2, socks)
	}

	{
		// List blue, brown, or green socks; should have 2
		socks, err := service.List(ctx, tags, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 2)
		requireSock(t, sock, socks)
		requireSock(t, sock2, socks)
	}

	{
		// Count blue, brown, or green socks, should be 2
		count, err := service.Count(ctx, tags)
		require.NoError(t, err)
		require.Equal(t, 2, count)
	}

	{
		// List red socks; should have 2
		socks, err := service.List(ctx, []string{"red"}, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 2)
		requireSock(t, sock, socks)
		requireSock(t, sock2, socks)
	}

	{
		socks, err := service.List(ctx, nil, "", 1, 1000)
		require.NoError(t, err)
		require.Len(t, socks, 2)
		requireSock(t, sock, socks)
		requireSock(t, sock2, socks)

		for i, sock := range socks {
			{
				// Delete the i'th sock
				err := service.DeleteSock(ctx, sock.ID)
				require.NoError(t, err, "sock %v (%v)", i, sock)
			}

			{
				// Check remaining socks
				remaining, err := service.List(ctx, nil, "", 1, 1000)
				require.NoError(t, err)
				require.Len(t, remaining, len(socks)-i-1)
				require.ElementsMatch(t, socks[i+1:], remaining)
			}
		}
	}

}

func requireSock(t *testing.T, a catalogue.Sock, bs []catalogue.Sock) {
	require.True(t, hasSock(a, bs))
}

func hasSock(a catalogue.Sock, bs []catalogue.Sock) bool {
	for _, b := range bs {
		if socksEqual(a, b) {
			return true
		}
	}
	return false
}

func requireSocksEqual(t *testing.T, a, b catalogue.Sock) {
	require.Equal(t, a.Name, b.Name)
	require.Equal(t, a.Description, b.Description)
	require.Equal(t, a.Price, b.Price)
	require.Equal(t, a.Quantity, b.Quantity)
	require.Subset(t, a.Tags, b.Tags)
	require.Subset(t, b.Tags, a.Tags)
}

func socksEqual(a, b catalogue.Sock) bool {
	return a.Name == b.Name && a.Description == b.Description && a.Price == b.Price && a.Quantity == b.Quantity && tagsEqual(a.Tags, b.Tags)
}

func tagsEqual(as, bs []string) bool {
	am := tomap(as)
	bm := tomap(bs)
	for a := range am {
		if _, inB := bm[a]; !inB {
			return false
		}
	}
	for b := range bm {
		if _, inA := am[b]; !inA {
			return false
		}
	}
	return true
}

func tomap(elems []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, elem := range elems {
		m[elem] = struct{}{}
	}
	return m
}
