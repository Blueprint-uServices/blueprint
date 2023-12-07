package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/DSB_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
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
