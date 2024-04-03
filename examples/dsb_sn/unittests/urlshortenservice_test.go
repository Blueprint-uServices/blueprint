package unittests

import (
	"context"
	"strings"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var urlShortenServiceRegistry = registry.NewServiceRegistry[socialnetwork.UrlShortenService]("urlShorten_service")

func init() {
	urlShortenServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.UrlShortenService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewUrlShortenServiceImpl(ctx, db)
	})
}

func TestComposeUrls(t *testing.T) {
	ctx := context.Background()
	service, err := urlShortenServiceRegistry.Get(ctx)
	require.NoError(t, err)

	urls, err := service.ComposeUrls(ctx, 1000, []string{"http://localhost:9000/hello", "http://localhost:9000/world"})
	require.NoError(t, err)
	require.Len(t, urls, 2)

	require.True(t, strings.HasPrefix(urls[0].ShortenedUrl, "http://short-url/"))
	require.True(t, strings.HasPrefix(urls[1].ShortenedUrl, "http://short-url/"))
	require.Equal(t, "http://localhost:9000/hello", urls[0].ExpandedUrl)
	require.Equal(t, "http://localhost:9000/world", urls[1].ExpandedUrl)
}

func TestGetExtendedUrls(t *testing.T) {
	ctx := context.Background()
	service, err := urlShortenServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// API is not currently implemented, so we should get blank values and no error
	extended_urls, err := service.GetExtendedUrls(ctx, 1000, []string{"http://short-url/blah"})
	require.NoError(t, err)
	require.Len(t, extended_urls, 0)
}
