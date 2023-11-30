package socialnetwork

import (
	"context"
	"errors"
	"log"
	"strconv"
	"sync"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type PostStorageService interface {
	StorePost(ctx context.Context, reqID int64, post Post) error
	ReadPost(ctx context.Context, reqID int64, postID int64) (Post, error)
	ReadPosts(ctx context.Context, reqID int64, postIDs []int64) ([]Post, error)
}

type PostStorageServiceImpl struct {
	postStorageCache backend.Cache
	postStorageDB    backend.NoSQLDatabase
	CacheHits        int64
	NumReqs          int64
	CacheMiss        int64
}

func NewPostStorageServiceImpl(ctx context.Context, postStorageCache backend.Cache, postStorageDB backend.NoSQLDatabase) (PostStorageService, error) {
	p := &PostStorageServiceImpl{postStorageCache: postStorageCache, postStorageDB: postStorageDB}
	return p, nil
}

func (p *PostStorageServiceImpl) StorePost(ctx context.Context, reqID int64, post Post) error {
	collection, err := p.postStorageDB.GetCollection(ctx, "post", "post")
	if err != nil {
		return err
	}
	return collection.InsertOne(ctx, post)
}

func (p *PostStorageServiceImpl) ReadPost(ctx context.Context, reqID int64, postID int64) (Post, error) {
	var post Post
	err := p.postStorageCache.Get(ctx, strconv.FormatInt(postID, 10), &post)
	if err != nil {
		// Post was not in Cache, check DB!
		collection, err := p.postStorageDB.GetCollection(ctx, "post", "post")
		if err != nil {
			return post, err
		}
		query := bson.D{{"postid", postID}}
		result, err := collection.FindOne(ctx, query)
		if err != nil {
			return post, err
		}
		res, err := result.One(ctx, &post)
		if !res || err != nil {
			return post, errors.New("Post doesn't exist in MongoDB")
		}
	}
	return post, nil
}

func (p *PostStorageServiceImpl) ReadPosts(ctx context.Context, reqID int64, postIDs []int64) ([]Post, error) {
	unique_post_ids := make(map[int64]bool)
	for _, pid := range postIDs {
		unique_post_ids[pid] = true
	}
	//if len(unique_post_ids) != len(postIDs) {
	//	return []Post{}, errors.New("Post Ids are duplicated")
	//}
	var keys []string
	for _, pid := range postIDs {
		keys = append(keys, strconv.FormatInt(pid, 10))
	}
	values := make([]Post, len(keys))
	var retvals []interface{}
	for idx, _ := range values {
		retvals = append(retvals, &values[idx])
	}
	err := p.postStorageCache.Mget(ctx, keys, retvals)
	for _, post := range values {
		delete(unique_post_ids, post.PostID)
	}
	if err != nil {
		log.Println(err)
		//log.Println("Length of uniqueIDs", len(unique_post_ids), " ", len(postIDs))
	}
	p.NumReqs += 1
	if len(unique_post_ids) != 0 {
		p.CacheMiss += 1
		//log.Println("Current Cache Miss",p.CacheMiss)
		var new_posts []Post
		var unique_pids []int64
		for k := range unique_post_ids {
			unique_pids = append(unique_pids, k)
		}
		collection, err := p.postStorageDB.GetCollection(ctx, "post", "post")
		if err != nil {
			return []Post{}, err
		}
		//delim := ","
		//query := `{"PostID": {"$in": ` + strings.Join(strings.Fields(fmt.Sprint(unique_pids)), delim) + `}}`
		query := bson.D{} //TODO: Fix this
		vals, err := collection.FindMany(ctx, query)
		if err != nil {
			log.Println(err)
			return []Post{}, err
		}
		vals.All(ctx, &new_posts)
		values = append(values, new_posts...)
		var wg sync.WaitGroup
		for _, new_post := range new_posts {
			wg.Add(1)

			go func() {
				defer wg.Done()
				p.postStorageCache.Put(ctx, strconv.FormatInt(new_post.PostID, 10), new_post)
			}()
		}
		wg.Wait()
	}
	return values, nil
}
