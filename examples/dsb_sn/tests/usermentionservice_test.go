package tests

import (
	"context"
	"testing"

	"github.com/Blueprint-uServices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/Blueprint-uServices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var userMentionServiceRegistry = registry.NewServiceRegistry[socialnetwork.UserMentionService]("userMention_service")

func init() {

	userMentionServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.UserMentionService, error) {
		cache, err := userCacheRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		db, err := userDBRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewUserMentionServiceImpl(ctx, cache, db)
	})
}

func TestComposeUserMentionsCache(t *testing.T) {
	ctx := context.Background()
	service, err := userMentionServiceRegistry.Get(ctx)
	require.NoError(t, err)

	cache, err := userCacheRegistry.Get(ctx)
	require.NoError(t, err)

	err = cache.Put(ctx, "vaastav:UserID", int64(5))
	require.NoError(t, err)
	err = cache.Put(ctx, "jcmace:UserID", int64(2))
	require.NoError(t, err)
	err = cache.Put(ctx, "antoinek:UserID", int64(1))
	require.NoError(t, err)
	err = cache.Put(ctx, "dg:UserID", int64(3))
	require.NoError(t, err)

	mentions, err := service.ComposeUserMentions(ctx, 1000, []string{"vaastav", "jcmace", "antoinek", "dg"})

	require.NoError(t, err)
	require.Len(t, mentions, 4)
	require.Equal(t, int64(5), mentions[0].UserID)
	require.Equal(t, int64(2), mentions[1].UserID)
	require.Equal(t, int64(1), mentions[2].UserID)
	require.Equal(t, int64(3), mentions[3].UserID)

	// Cleanup cache
	err = cache.Delete(ctx, "vaastav:UserID")
	require.NoError(t, err)
	err = cache.Delete(ctx, "jcmace:UserID")
	require.NoError(t, err)
	err = cache.Delete(ctx, "antoinek:UserID")
	require.NoError(t, err)
	err = cache.Delete(ctx, "dg:UserID")
	require.NoError(t, err)
}

func TestComposeUserMentionsNoCache(t *testing.T) {
	ctx := context.Background()
	service, err := userMentionServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	err = coll.InsertMany(ctx, []interface{}{vaastav, jcmace, antoinek, dg})
	require.NoError(t, err)

	mentions, err := service.ComposeUserMentions(ctx, 1000, []string{"vaastav", "jcmace", "antoinek", "dg"})

	require.NoError(t, err)
	require.Len(t, mentions, 4)
	require.Equal(t, int64(5), mentions[0].UserID)
	require.Equal(t, int64(2), mentions[1].UserID)
	require.Equal(t, int64(1), mentions[2].UserID)
	require.Equal(t, int64(3), mentions[3].UserID)

	// cleanup database after the test
	err = coll.DeleteMany(ctx, bson.D{{"userid", vaastav.UserID}})
	require.NoError(t, err)
	err = coll.DeleteMany(ctx, bson.D{{"userid", jcmace.UserID}})
	require.NoError(t, err)
	err = coll.DeleteMany(ctx, bson.D{{"userid", antoinek.UserID}})
	require.NoError(t, err)
	err = coll.DeleteMany(ctx, bson.D{{"userid", dg.UserID}})
	require.NoError(t, err)
}
