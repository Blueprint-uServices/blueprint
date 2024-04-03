package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var userTimelineServiceRegistry = registry.NewServiceRegistry[socialnetwork.UserTimelineService]("usertimeline_service")
var userTimelineCacheRegistry = registry.NewServiceRegistry[backend.Cache]("usertimeline_cache")
var userTimelineDBRegistry = registry.NewServiceRegistry[backend.NoSQLDatabase]("usertimeline_db")

func init() {

	userTimelineCacheRegistry.Register("local", func(ctx context.Context) (backend.Cache, error) {
		return simplecache.NewSimpleCache(ctx)
	})

	// Simplenosqldb doesn't support the operators required for UserTimelineService
	userTimelineDBRegistry.Register("local", func(ctx context.Context) (backend.NoSQLDatabase, error) {
		return simplenosqldb.NewSimpleNoSQLDB(ctx)
	})

	// Requires that the mongodb server is running.
	/*
		userTimelineDBRegistry.Register("local", func(ctx context.Context) (backend.NoSQLDatabase, error) {
			return mongodb.NewMongoDB(ctx, "localhost:27017")
		})
	*/

	userTimelineServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.UserTimelineService, error) {
		cache, err := userTimelineCacheRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		db, err := userTimelineDBRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		postStorageService, err := postStorageServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewUserTimelineServiceImpl(ctx, cache, db, postStorageService)
	})
}

func TestWriteUserTimeline(t *testing.T) {
	ctx := context.Background()
	service, err := userTimelineServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := userTimelineDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "usertimeline", "usertimeline")
	require.NoError(t, err)

	defer func() {
		ids := []int64{vaastav.UserID}
		cleanup_usertimeline_backends(t, ctx, ids)
	}()

	cache, err := userTimelineCacheRegistry.Get(ctx)
	require.NoError(t, err)

	postIDs := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for idx, pid := range postIDs {
		err = service.WriteUserTimeline(ctx, int64(idx), pid, vaastav.UserID, time.Now().UnixNano())
		require.NoError(t, err)
	}

	// Check cache contents & cleanup cache
	var cache_infos []socialnetwork.PostInfo
	id_str := strconv.FormatInt(vaastav.UserID, 10)
	exists, err := cache.Get(ctx, id_str, &cache_infos)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, cache_infos, 10)

	err = cache.Delete(ctx, id_str)

	// Check database contents
	var user_posts socialnetwork.UserPosts
	res, err := coll.FindOne(ctx, bson.D{})
	require.NoError(t, err)

	exists, err = res.One(ctx, &user_posts)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, vaastav.UserID, user_posts.UserID)
	require.Len(t, user_posts.Posts, 10)

}

func TestReadUserTimeline(t *testing.T) {

	// Populate database
	ctx := context.Background()
	service, err := userTimelineServiceRegistry.Get(ctx)
	require.NoError(t, err)

	defer func() {

		ids := []int64{vaastav.UserID}
		cleanup_usertimeline_backends(t, ctx, ids)
	}()
	cache, err := userTimelineCacheRegistry.Get(ctx)
	require.NoError(t, err)

	postIDs := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for idx, pid := range postIDs {
		err = service.WriteUserTimeline(ctx, int64(idx), pid, vaastav.UserID, time.Now().UnixNano())
		require.NoError(t, err)
	}

	// Test ReadUserTimeline
	// Check out of bound start, stop
	read_ids, err := service.ReadUserTimeline(ctx, 1000, vaastav.UserID, -1, 10)
	require.NoError(t, err)
	require.Len(t, read_ids, 0)
	read_ids, err = service.ReadUserTimeline(ctx, 1001, vaastav.UserID, 7, 5)
	require.NoError(t, err)
	require.Len(t, read_ids, 0)

	// Read cached version of the posts.
	read_ids, err = service.ReadUserTimeline(ctx, 1002, vaastav.UserID, 0, 10)
	require.NoError(t, err)
	require.Len(t, read_ids, 10)

	// Cleanup cache
	id_str := strconv.FormatInt(vaastav.UserID, 10)
	err = cache.Delete(ctx, id_str)
	require.NoError(t, err)

	// Read from the database
	read_ids, err = service.ReadUserTimeline(ctx, 1002, vaastav.UserID, 0, 10)
	require.NoError(t, err)
	require.Len(t, read_ids, 10)

}

func cleanup_usertimeline_backends(t *testing.T, ctx context.Context, ids []int64) {
	db, err := userTimelineDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "usertimeline", "usertimeline")
	require.NoError(t, err)
	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
	cache, err := userTimelineCacheRegistry.Get(ctx)
	require.NoError(t, err)
	for _, id := range ids {
		idstr := strconv.FormatInt(id, 10)
		cache.Delete(ctx, idstr)
	}
}
