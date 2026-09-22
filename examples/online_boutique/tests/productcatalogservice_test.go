package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var productCatalogServiceRegistry = registry.NewServiceRegistry[boutique.ProductCatalogService]("product_catalog_service")

func init() {
	productCatalogServiceRegistry.Register("local", func(ctx context.Context) (boutique.ProductCatalogService, error) {
		setupLogger(ctx)
		return boutique.NewProductCatalogServiceImpl(ctx)
	})
}

func TestProductCatalogService(t *testing.T) {
	ctx := context.Background()
	service, err := productCatalogServiceRegistry.Get(ctx)
	require.NoError(t, err)

	products, err := service.ListProducts(ctx)
	require.NoError(t, err)
	require.Len(t, products, 9)

	product, err := service.GetProduct(ctx, "OLJCESPC7Z")
	require.NoError(t, err)
	require.Equal(t, "Sunglasses", product.Name)

	matches, err := service.SearchProducts(ctx, "kitchen")
	require.NoError(t, err)
	require.NotEmpty(t, matches)

	_, err = service.GetProduct(ctx, "missing")
	require.ErrorIs(t, err, boutique.ErrProductNotFound)
}
