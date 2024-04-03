package socialnetwork

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// The UserTimelineService interface
// The full Timeline of a user is represented as an array of post ids: post_ids[id_0 ,..., id_n].
type UserTimelineService interface {
	// Reads the timeline of the user that has the id `userID`.
	// The return value is represented by the slice: post_ids[start:stop].
	ReadUserTimeline(ctx context.Context, reqID int64, userID int64, start int64, stop int64) ([]int64, error)
	// Adds a new post to the user timeline of the user that has the id `userID`
	// The new post ID is placed at the 0th position in the post ids array.
	//    post_ids = []int64{`postID`, post_ids...)
	WriteUserTimeline(ctx context.Context, reqID int64, postID int64, userID int64, timestamp int64) error
	// CleanupUTimelineBackends caches and databases
	CleanupUTimelineBackends(ctx context.Context) error
}

// The format of a single post in a user's timeline stored in the backend.
type PostInfo struct {
	PostID    int64
	Timestamp int64
}

// The format of a user's timeline stored in the backend.
type UserPosts struct {
	UserID int64
	Posts  []PostInfo
}

// Implementation of [UserTimelineService]
type UserTimelineServiceImpl struct {
	userTimelineCache  backend.Cache
	userTimelineDB     backend.NoSQLDatabase
	postStorageService PostStorageService
	CacheHits          int64
	CacheMiss          int64
	NumRequests        int64
}

// Creates a [UserTimelineService] instance for managing the user timelines for the various users.
func NewUserTimelineServiceImpl(ctx context.Context, userTimelineCache backend.Cache, userTimelineDB backend.NoSQLDatabase, postStorageService PostStorageService) (UserTimelineService, error) {
	u := &UserTimelineServiceImpl{userTimelineCache: userTimelineCache, userTimelineDB: userTimelineDB, postStorageService: postStorageService}
	return u, nil
}

// Implements UserTimelineService interface
func (u *UserTimelineServiceImpl) ReadUserTimeline(ctx context.Context, reqID int64, userID int64, start int64, stop int64) ([]int64, error) {
	u.NumRequests += 1
	if stop <= start || start < 0 {
		return []int64{}, nil
	}

	userIDStr := strconv.FormatInt(userID, 10)
	var post_infos []PostInfo
	exists, err := u.userTimelineCache.Get(ctx, userIDStr, &post_infos)
	if err != nil {
		return []int64{}, err
	}
	if exists {
		u.CacheHits += 1
	} else {
		u.CacheMiss += 1
	}
	var post_ids []int64
	seen_posts := make(map[int64]bool)
	for _, post_info := range post_infos {
		post_ids = append(post_ids, post_info.PostID)
		seen_posts[post_info.PostID] = true
	}
	db_start := start + int64(len(post_ids))
	var new_post_ids []int64
	if db_start < stop {
		collection, err := u.userTimelineDB.GetCollection(ctx, "usertimeline", "usertimeline")
		if err != nil {
			return []int64{}, err
		}
		query := fmt.Sprintf(`{"UserID": %[1]d}`, userID)
		projection := fmt.Sprintf(`{"posts": {"$slice": [0, %[1]d]}}`, stop)
		query_d, err := parseNoSQLDBQuery(query)
		if err != nil {
			return []int64{}, err
		}
		projection_d, err := parseNoSQLDBQuery(projection)
		if err != nil {
			return []int64{}, err
		}
		post_db_val, err := collection.FindOne(ctx, query_d, projection_d)
		if err != nil {
			return []int64{}, err
		}
		var user_posts UserPosts
		exists, err = post_db_val.One(ctx, &user_posts)
		if err != nil {
			return []int64{}, err
		}
		if !exists {
			return []int64{}, errors.New("Failed to find posts in database")
		}
		for _, post := range user_posts.Posts {
			// Avoid duplicated post_ids
			if _, ok := seen_posts[post.PostID]; ok {
				continue
			}
			new_post_ids = append(new_post_ids, post.PostID)
		}
	}

	post_ids = append(new_post_ids, post_ids...)
	fmt.Println(post_ids)
	post_channel := make(chan bool)
	err_post_channel := make(chan error)
	//var posts []Post
	go func() {
		var err error
		_, err = u.postStorageService.ReadPosts(ctx, reqID, post_ids)
		if err != nil {
			log.Println(err)
			err_post_channel <- err
			return
		}
		post_channel <- true
	}()

	if len(new_post_ids) > 0 {
		err := u.userTimelineCache.Put(ctx, userIDStr, post_ids)
		if err != nil {
			return []int64{}, err
		}
	}
	select {
	case <-post_channel:
		break
	case err := <-err_post_channel:
		return []int64{}, err
	}
	return post_ids, nil
}

// Implements UserTimelineService interface
func (u *UserTimelineServiceImpl) WriteUserTimeline(ctx context.Context, reqID int64, postID int64, userID int64, timestamp int64) error {
	collection, err := u.userTimelineDB.GetCollection(ctx, "usertimeline", "usertimeline")
	if err != nil {
		return err
	}

	query := bson.D{{"userid", userID}}
	results, err := collection.FindMany(ctx, query)
	var userPosts []UserPosts
	if err != nil {
		return err
	}
	results.All(ctx, &userPosts)

	if len(userPosts) == 0 {
		fmt.Println("Inserting new entry for", userID)
		userPosts := UserPosts{UserID: userID, Posts: []PostInfo{PostInfo{PostID: postID, Timestamp: timestamp}}}
		err := collection.InsertOne(ctx, userPosts)
		if err != nil {
			return errors.New("Failed to insert user timeline user to Database")
		}
	} else {
		fmt.Println("Adding a new post for user", userID)
		postIDstr := strconv.FormatInt(postID, 10)
		timestampstr := strconv.FormatInt(timestamp, 10)
		update := fmt.Sprintf(`{"$push": {"Posts": {"PostID": %s, "Timestamp": %s}}}`, postIDstr, timestampstr)
		update_d, err := parseNoSQLDBQuery(update)
		if err != nil {
			return err
		}
		_, err = collection.UpdateMany(ctx, query, update_d)
		if err != nil {
			log.Println(err)
			return errors.New("Failed to update user timeline user to Database")
		}
	}
	var postInfo []PostInfo
	userIDStr := strconv.FormatInt(userID, 10)
	// Ignore error check for Get!
	_, err = u.userTimelineCache.Get(ctx, userIDStr, &postInfo)
	if err != nil {
		return err
	}
	postInfo = append(postInfo, PostInfo{PostID: postID, Timestamp: timestamp})
	return u.userTimelineCache.Put(ctx, userIDStr, postInfo)
}

// Implements UserTimelineService
func (u *UserTimelineServiceImpl) CleanupUTimelineBackends(ctx context.Context) error {
	collection, err := u.userTimelineDB.GetCollection(ctx, "usertimeline", "usertimeline")
	if err != nil {
		return err
	}
	err = collection.DeleteMany(ctx, bson.D{})
	if err != nil {
		return err
	}
	return u.userTimelineCache.DeleteAll(ctx)
}
