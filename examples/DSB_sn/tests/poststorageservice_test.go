package tests

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/DSB_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
)

var postStorageServiceRegistry = registry.NewServiceRegistry[socialnetwork.PostStorageService]("postStorage_service")

func init() {
	postStorageServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.PostStorageService, error) {
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}

		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
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

func TestStorePost(t *testing.T) {
	ctx := context.Background()
	service, err := postStorageServiceRegistry.Get(ctx)
	require.NoError(t, err)

	err = service.StorePost(ctx, 1002, post1)
	require.NoError(t, err)
}

func TestReadPost(t *testing.T) {
	ctx := context.Background()
	service, err := postStorageServiceRegistry.Get(ctx)
	require.NoError(t, err)

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

	err = service.StorePost(ctx, 1002, post1)
	require.NoError(t, err)
	err = service.StorePost(ctx, 1003, post2)
	require.NoError(t, err)

	posts, err := service.ReadPosts(ctx, 1004, []int64{post1.PostID, post2.PostID})
	require.NoError(t, err)
	t.Log(posts)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PostID < posts[j].PostID
	})
	requirePostEqual(t, post1, posts[0])
	requirePostEqual(t, post2, posts[1])
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
