package tests

import (
	"context"

	"gitlab.mpi-sws.org/cld/blueprint/examples/DSB_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
)

var socialGraphServiceRegistry = registry.NewServiceRegistry[socialnetwork.SocialGraphService]("socialgraph_service")

func init() {
	socialGraphServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.SocialGraphService, error) {
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}

		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		userIDService, err := userIDServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewSocialGraphServiceImpl(ctx, cache, db, userIDService)
	})
}
