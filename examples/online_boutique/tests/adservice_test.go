package tests

import (
	"context"
	"sync"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow/boutique"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/slogger"
	"github.com/stretchr/testify/require"
)

var adServiceRegistry = registry.NewServiceRegistry[boutique.AdService]("ad_service")
var loggerOnce sync.Once

func setupLogger(ctx context.Context) {
	loggerOnce.Do(func() { _, _ = slogger.NewSLogger(ctx) })
}

func init() {
	adServiceRegistry.Register("local", func(ctx context.Context) (boutique.AdService, error) {
		setupLogger(ctx)
		return boutique.NewAdServiceImpl(ctx)
	})
}

func TestAdService(t *testing.T) {
	ctx := context.Background()
	service, err := adServiceRegistry.Get(ctx)
	require.NoError(t, err)

	ads, err := service.GetAds(ctx, []string{"kitchen"})
	require.NoError(t, err)
	require.Equal(t, []boutique.Ad{{RedirectURL: "/product/9SIQT8TOJO", Text: "Bamboo glass jar for sale. 10% off."}}, ads)

	ads, err = service.GetAds(ctx, []string{"not-a-category"})
	require.NoError(t, err)
	require.Len(t, ads, 2)
}
