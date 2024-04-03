package tests

import (
	"context"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var userTimelineServiceRegistry = registry.NewServiceRegistry[socialnetwork.UserTimelineService]("usertimeline_service")

func init() {

	userTimelineServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.UserTimelineService, error) {
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
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

	defer func() {
		cleanup_usertimeline_backends(t, ctx)
	}()

	postIDs := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for idx, pid := range postIDs {
		err = service.WriteUserTimeline(ctx, int64(idx), pid, vaastav.UserID, time.Now().UnixNano())
		require.NoError(t, err)
	}
}

func TestReadUserTimeline(t *testing.T) {

	// Populate database
	ctx := context.Background()
	service, err := userTimelineServiceRegistry.Get(ctx)
	require.NoError(t, err)

	defer func() {
		cleanup_usertimeline_backends(t, ctx)
	}()

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

	read_ids, err = service.ReadUserTimeline(ctx, 1002, vaastav.UserID, 0, 10)
	require.NoError(t, err)
	require.Len(t, read_ids, 10)

}

func cleanup_usertimeline_backends(t *testing.T, ctx context.Context) {
	service, err := userTimelineServiceRegistry.Get(ctx)
	require.NoError(t, err)
	err = service.CleanupUTimelineBackends(ctx)
	require.NoError(t, err)
}
