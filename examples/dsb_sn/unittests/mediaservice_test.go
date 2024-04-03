package unittests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var mediaServiceRegistry = registry.NewServiceRegistry[socialnetwork.MediaService]("media_service")

func init() {
	mediaServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.MediaService, error) {
		return socialnetwork.NewMediaServiceImpl(ctx)
	})
}

func TestComposeMedia(t *testing.T) {
	ctx := context.Background()
	service, err := mediaServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Equal length of media types
	media, err := service.ComposeMedia(ctx, 1000, []string{"video", "pic", "audio"}, []int64{0, 1, 2})
	require.NoError(t, err)
	require.Len(t, media, 3)

	// Non-equal length
	media, err = service.ComposeMedia(ctx, 1000, []string{}, []int64{0, 1, 2})
	require.Error(t, err)
}
