package tests

import (
	"context"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/stretchr/testify/require"
)

var homeTimelineServiceRegistry = registry.NewServiceRegistry[socialnetwork.HomeTimelineService]("hometimeline_service")

func init() {

	homeTimelineServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.HomeTimelineService, error) {
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}

		postStorageService, err := postStorageServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		socialGraphService, err := socialGraphServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewHomeTimelineServiceImpl(ctx, cache, postStorageService, socialGraphService)
	})
}

func TestWriteHomeTimeline(t *testing.T) {
	ctx := context.Background()
	service, err := homeTimelineServiceRegistry.Get(ctx)
	require.NoError(t, err)
	// Add some followers using social graph service!
	setup_social(t, ctx)

	ids := []int64{jcmace.UserID, antoinek.UserID, dg.UserID}
	defer func() {
		cleanup_social_database(t, ctx)
		ids = append(ids, vaastav.UserID)
		cleanup_hometimeline_cache(t, ctx)
	}()

	err = service.WriteHomeTimeline(ctx, 1001, post1.PostID, vaastav.UserID, time.Now().UnixNano(), []int64{dg.UserID})
	require.NoError(t, err)

	// Add another post
	err = service.WriteHomeTimeline(ctx, 1002, post2.PostID, vaastav.UserID, time.Now().UnixNano(), []int64{jcmace.UserID})
	require.NoError(t, err)

	// Cleanup databases and caches

	cleanup_hometimeline_cache(t, ctx)

	cleanup_social_database(t, ctx)
}

func TestReadHomeTimeline(t *testing.T) {
	ctx := context.Background()
	service, err := homeTimelineServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Add some followers using social graph service!
	setup_social(t, ctx)
	ids := []int64{jcmace.UserID, antoinek.UserID, dg.UserID}
	defer func() {
		cleanup_social_database(t, ctx)
		ids = append(ids, vaastav.UserID)
		cleanup_hometimeline_cache(t, ctx)
	}()

	err = service.WriteHomeTimeline(ctx, 1001, post1.PostID, vaastav.UserID, time.Now().UnixNano(), []int64{dg.UserID})
	require.NoError(t, err)

	// Check cache entries to ensure the posts are loaded.

	// Add another post
	err = service.WriteHomeTimeline(ctx, 1002, post2.PostID, vaastav.UserID, time.Now().UnixNano(), []int64{jcmace.UserID})
	require.NoError(t, err)

	// Time to test the read
	postids, err := service.ReadHomeTimeline(ctx, 1003, jcmace.UserID, -1, 5)
	require.NoError(t, err)
	require.Len(t, postids, 0)
	postids, err = service.ReadHomeTimeline(ctx, 1004, jcmace.UserID, 3, 2)
	require.NoError(t, err)
	require.Len(t, postids, 0)

	postids, err = service.ReadHomeTimeline(ctx, 1005, jcmace.UserID, 0, 5)
	require.NoError(t, err)
	require.Len(t, postids, 2)

	postids, err = service.ReadHomeTimeline(ctx, 1005, antoinek.UserID, 0, 5)
	require.NoError(t, err)
	require.Len(t, postids, 2)

	postids, err = service.ReadHomeTimeline(ctx, 1005, dg.UserID, 0, 5)
	require.NoError(t, err)
	require.Len(t, postids, 1)

	// Cleanup databases and caches

	cleanup_hometimeline_cache(t, ctx)

	cleanup_social_database(t, ctx)
}

func setup_social(t *testing.T, ctx context.Context) {
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	err = service.InsertUser(ctx, 1000, vaastav.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1001, jcmace.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1002, antoinek.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1005, dg.UserID)
	require.NoError(t, err)

	err = service.Follow(ctx, 1003, jcmace.UserID, vaastav.UserID)
	require.NoError(t, err)
	err = service.Follow(ctx, 1004, antoinek.UserID, vaastav.UserID)
	require.NoError(t, err)
}

func cleanup_social_database(t *testing.T, ctx context.Context) {
	social_service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)
	err = social_service.CleanupSocialBackends(ctx)
	require.NoError(t, err)

}

func cleanup_hometimeline_cache(t *testing.T, ctx context.Context) {
	service, err := homeTimelineServiceRegistry.Get(ctx)
	require.NoError(t, err)
	err = service.CleanupHTimelineBackends(ctx)
	require.NoError(t, err)
}
