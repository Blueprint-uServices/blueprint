package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
	"go.mongodb.org/mongo-driver/bson"
)

var homeTimelineServiceRegistry = registry.NewServiceRegistry[socialnetwork.HomeTimelineService]("hometimeline_service")
var homeTimelineCacheRegistry = registry.NewServiceRegistry[backend.Cache]("hometimeline_cache")

func init() {
	homeTimelineCacheRegistry.Register("local", func(ctx context.Context) (backend.Cache, error) {
		return simplecache.NewSimpleCache(ctx)
	})

	homeTimelineServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.HomeTimelineService, error) {
		cache, err := homeTimelineCacheRegistry.Get(ctx)
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

	cache, err := homeTimelineCacheRegistry.Get(ctx)
	require.NoError(t, err)

	// Add some followers using social graph service!
	setup_social(t, ctx)

	err = service.WriteHomeTimeline(ctx, 1001, post1.PostID, vaastav.UserID, time.Now().UnixNano(), []int64{dg.UserID})
	require.NoError(t, err)

	// Check cache entries to ensure the posts are loaded.
	ids := []int64{jcmace.UserID, antoinek.UserID, dg.UserID}
	for _, id := range ids {
		idstr := strconv.FormatInt(id, 10)
		var posts []socialnetwork.PostInfo
		exists, err := cache.Get(ctx, idstr, &posts)
		require.NoError(t, err)
		require.True(t, exists)
		require.Len(t, posts, 1)
	}

	// Add another post
	err = service.WriteHomeTimeline(ctx, 1002, post2.PostID, vaastav.UserID, time.Now().UnixNano(), []int64{jcmace.UserID})
	require.NoError(t, err)

	// Check cache entries again
	for _, id := range ids {
		idstr := strconv.FormatInt(id, 10)
		var posts []socialnetwork.PostInfo
		exists, err := cache.Get(ctx, idstr, &posts)
		require.NoError(t, err)
		require.True(t, exists)
		if id == dg.UserID {
			require.Len(t, posts, 1)
		} else {
			require.Len(t, posts, 2)
		}
	}

	// Cleanup databases and caches

	for _, id := range ids {
		idstr := strconv.FormatInt(id, 10)
		err = cache.Delete(ctx, idstr)
		require.NoError(t, err)
	}

	socialcache, err := socialGraphCacheRegistry.Get(ctx)
	require.NoError(t, err)
	ids = []int64{jcmace.UserID, antoinek.UserID}
	for _, id := range ids {
		idstr := strconv.FormatInt(id, 10)
		err = socialcache.Delete(ctx, idstr+":followees")
		require.NoError(t, err)
	}

	vaasIDstr := strconv.FormatInt(vaastav.UserID, 10)
	err = socialcache.Delete(ctx, vaasIDstr+":followers")
	require.NoError(t, err)

	cleanup_social_database(t, ctx)
}

func TestReadHomeTimeline(t *testing.T) {
	ctx := context.Background()
	service, err := homeTimelineServiceRegistry.Get(ctx)
	require.NoError(t, err)

	cache, err := homeTimelineCacheRegistry.Get(ctx)
	require.NoError(t, err)

	// Add some followers using social graph service!
	setup_social(t, ctx)

	err = service.WriteHomeTimeline(ctx, 1001, post1.PostID, vaastav.UserID, time.Now().UnixNano(), []int64{dg.UserID})
	require.NoError(t, err)

	// Check cache entries to ensure the posts are loaded.
	ids := []int64{jcmace.UserID, antoinek.UserID, dg.UserID}

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

	for _, id := range ids {
		idstr := strconv.FormatInt(id, 10)
		err = cache.Delete(ctx, idstr)
		require.NoError(t, err)
	}

	socialcache, err := socialGraphCacheRegistry.Get(ctx)
	require.NoError(t, err)
	ids = []int64{jcmace.UserID, antoinek.UserID}
	for _, id := range ids {
		idstr := strconv.FormatInt(id, 10)
		err = socialcache.Delete(ctx, idstr+":followees")
		require.NoError(t, err)
	}

	vaasIDstr := strconv.FormatInt(vaastav.UserID, 10)
	err = socialcache.Delete(ctx, vaasIDstr+":followers")
	require.NoError(t, err)

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
	socialdb, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	socialcoll, err := socialdb.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)
	err = socialcoll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}
