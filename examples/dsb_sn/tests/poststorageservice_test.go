package tests

import (
	"context"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/Blueprint-uServices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/Blueprint-uServices/blueprint/runtime/core/backend"
	"github.com/Blueprint-uServices/blueprint/runtime/core/registry"
	"github.com/Blueprint-uServices/blueprint/runtime/plugins/simplecache"
	"github.com/Blueprint-uServices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var postStorageServiceRegistry = registry.NewServiceRegistry[socialnetwork.PostStorageService]("postStorage_service")
var postCacheRegistry = registry.NewServiceRegistry[backend.Cache]("post_cache")
var postDBRegistry = registry.NewServiceRegistry[backend.NoSQLDatabase]("post_db")

func init() {
	postCacheRegistry.Register("local", func(ctx context.Context) (backend.Cache, error) {
		return simplecache.NewSimpleCache(ctx)
	})

	postDBRegistry.Register("local", func(ctx context.Context) (backend.NoSQLDatabase, error) {
		return simplenosqldb.NewSimpleNoSQLDB(ctx)
	})

	postStorageServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.PostStorageService, error) {
		cache, err := postCacheRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		db, err := postDBRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewPostStorageServiceImpl(ctx, cache, db)
	})
}

var post1 = socialnetwork.Post{
	PostID:  1,
	Creator: socialnetwork.Creator{UserID: 5, Username: "vaastav"},
	ReqID:   1000,
	Text:    "Hello World! Check out Blueprint(http://blueprint-uservices.github.io) by @vaastav, @jcmace, @antoinek, @dg !",
	UserMentions: []socialnetwork.UserMention{
		{UserID: 5, Username: "vaastav"},
		{UserID: 2, Username: "jcmace"},
		{UserID: 1, Username: "antoinek"},
		{UserID: 3, Username: "dg"},
	},
	Medias: []socialnetwork.Media{},
	Urls: []socialnetwork.URL{
		{ShortenedUrl: "http://short-url/hello", ExpandedUrl: "http://blueprint-uservices.github.io"},
	},
	Timestamp: time.Now().Unix(),
	PostType:  socialnetwork.POST,
}

var post2 = socialnetwork.Post{
	PostID:  2,
	Creator: socialnetwork.Creator{UserID: 5, Username: "vaastav"},
	ReqID:   1001,
	Text:    "Hello World Again! Check out Blueprint(http://blueprint-uservices.github.io) by @vaastav, @jcmace, @antoinek, @dg !",
	UserMentions: []socialnetwork.UserMention{
		{UserID: 5, Username: "vaastav"},
		{UserID: 2, Username: "jcmace"},
		{UserID: 1, Username: "antoinek"},
		{UserID: 3, Username: "dg"},
	},
	Medias: []socialnetwork.Media{},
	Urls: []socialnetwork.URL{
		{ShortenedUrl: "http://short-url/helloagain", ExpandedUrl: "http://blueprint-uservices.github.io"},
	},
	Timestamp: time.Now().Unix(),
	PostType:  socialnetwork.POST,
}

func cleanup_post_backends(t *testing.T, ctx context.Context, pids []int64) {
	cache, err := postCacheRegistry.Get(ctx)
	require.NoError(t, err)
	db, err := postDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "post", "post")
	require.NoError(t, err)
	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
	for _, pid := range pids {
		pid_string := strconv.FormatInt(pid, 10)
		cache.Delete(ctx, pid_string)
	}
}

func TestStorePost(t *testing.T) {
	ctx := context.Background()
	service, err := postStorageServiceRegistry.Get(ctx)
	require.NoError(t, err)

	defer func() {
		cleanup_post_backends(t, ctx, []int64{post1.PostID})
	}()

	err = service.StorePost(ctx, 1002, post1)
	require.NoError(t, err)
}

func TestReadPost(t *testing.T) {
	ctx := context.Background()
	service, err := postStorageServiceRegistry.Get(ctx)
	require.NoError(t, err)

	defer func() {
		cleanup_post_backends(t, ctx, []int64{post1.PostID})
	}()

	err = service.StorePost(ctx, 1002, post1)
	require.NoError(t, err)

	post, err := service.ReadPost(ctx, 1003, 1)
	require.NoError(t, err)
	requirePostEqual(t, post1, post)
}

func TestReadPosts(t *testing.T) {
	ctx := context.Background()
	service, err := postStorageServiceRegistry.Get(ctx)
	require.NoError(t, err)

	post1_c := post1
	post1_c.PostID = 3
	post2_c := post2
	post2_c.PostID = 4

	defer func() {
		cleanup_post_backends(t, ctx, []int64{post1_c.PostID, post2_c.PostID})
	}()

	err = service.StorePost(ctx, 1002, post1_c)
	require.NoError(t, err)
	err = service.StorePost(ctx, 1003, post2_c)
	require.NoError(t, err)

	posts, err := service.ReadPosts(ctx, 1004, []int64{post1_c.PostID, post2_c.PostID})
	require.NoError(t, err)
	require.Len(t, posts, 2)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PostID < posts[j].PostID
	})
	requirePostEqual(t, post1_c, posts[0])
	requirePostEqual(t, post2_c, posts[1])
}

func requirePostEqual(t *testing.T, p1, p2 socialnetwork.Post) {
	require.Equal(t, p1.PostID, p2.PostID)
	require.Equal(t, p1.ReqID, p2.ReqID)
	require.Equal(t, p1.Timestamp, p2.Timestamp)
	require.Equal(t, p1.Text, p2.Text)
	require.True(t, p1.PostType == p2.PostType)
	requireCreatorEqual(t, p1.Creator, p2.Creator)
	require.True(t, len(p1.Urls) == len(p2.Urls))
	for i := 0; i < len(p1.Urls); i++ {
		requireUrlEqual(t, p1.Urls[i], p2.Urls[i])
	}
	require.True(t, len(p1.UserMentions) == len(p2.UserMentions))
	for i := 0; i < len(p1.UserMentions); i++ {
		requireUserMentionEqual(t, p1.UserMentions[i], p2.UserMentions[i])
	}
}

func requireCreatorEqual(t *testing.T, c1, c2 socialnetwork.Creator) {
	require.Equal(t, c1.UserID, c2.UserID)
	require.Equal(t, c1.Username, c2.Username)
}

func requireUrlEqual(t *testing.T, u1, u2 socialnetwork.URL) {
	require.Equal(t, u1.ShortenedUrl, u2.ShortenedUrl)
	require.Equal(t, u1.ExpandedUrl, u2.ExpandedUrl)
}

func requireUserMentionEqual(t *testing.T, u1, u2 socialnetwork.UserMention) {
	require.Equal(t, u1.UserID, u2.UserID)
	require.Equal(t, u1.Username, u2.Username)
}
