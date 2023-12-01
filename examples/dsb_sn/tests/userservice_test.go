package tests

import (
	"context"

	"gitlab.mpi-sws.org/cld/blueprint/examples/DSB_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
)

var userServiceRegistry = registry.NewServiceRegistry[socialnetwork.UserService]("user_service")

func init() {
	userServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.UserService, error) {
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}

		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		socialgraphservice, err := socialGraphServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewUserServiceImpl(ctx, cache, db, socialgraphservice, "secret")
	})
}
