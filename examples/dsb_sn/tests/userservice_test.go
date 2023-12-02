package tests

import (
	"context"

	"gitlab.mpi-sws.org/cld/blueprint/examples/DSB_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
)

var userServiceRegistry = registry.NewServiceRegistry[socialnetwork.UserService]("user_service")

var userCacheRegistry = registry.NewServiceRegistry[backend.Cache]("user_cache")

var userDBRegistry = registry.NewServiceRegistry[backend.NoSQLDatabase]("user_db")

var vaastav = socialnetwork.User{
	FirstName: "Vaastav",
	LastName:  "Anand",
	UserID:    5,
	Username:  "vaastav",
}

var jcmace = socialnetwork.User{
	FirstName: "Jonathan",
	LastName:  "Mace",
	UserID:    2,
	Username:  "jcmace",
}

var antoinek = socialnetwork.User{
	FirstName: "Antoine",
	LastName:  "Kaufmann",
	UserID:    1,
	Username:  "antoinek",
}

var dg = socialnetwork.User{
	FirstName: "Deepak",
	LastName:  "Garg",
	UserID:    3,
	Username:  "dg",
}

func init() {

	userCacheRegistry.Register("local", func(ctx context.Context) (backend.Cache, error) {
		return simplecache.NewSimpleCache(ctx)
	})

	userDBRegistry.Register("local", func(ctx context.Context) (backend.NoSQLDatabase, error) {
		return simplenosqldb.NewSimpleNoSQLDB(ctx)
	})

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
