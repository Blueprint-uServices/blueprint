package socialnetwork

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type UserTimelineService interface {
	ReadUserTimeline(ctx context.Context, reqID int64, userID int64, start int64, stop int64) ([]int64, error)
	WriteUserTimeline(ctx context.Context, reqID int64, postID int64, userID int64, timestamp int64) error
}

type PostInfo struct {
	PostID    int64
	Timestamp int64
}

type UserPosts struct {
	UserID int64
	Posts  []PostInfo
}

type UserTimelineServiceImpl struct {
	userTimelineCache  backend.Cache
	userTimelineDB     backend.NoSQLDatabase
	postStorageService PostStorageService
	CacheHits          int64
	CacheMiss          int64
	NumRequests        int64
}

func NewUserTimelineServiceImpl(ctx context.Context, userTimelineCache backend.Cache, userTimelineDB backend.NoSQLDatabase, postStorageService PostStorageService) (UserTimelineService, error) {
	u := &UserTimelineServiceImpl{userTimelineCache: userTimelineCache, userTimelineDB: userTimelineDB, postStorageService: postStorageService}
	return u, nil
}

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
	for _, post_info := range post_infos {
		post_ids = append(post_ids, post_info.PostID)
	}
	db_start := start + int64(len(post_ids))
	var new_post_ids []int64
	if db_start < stop {
		collection, err := u.userTimelineDB.GetCollection(ctx, "user-timeline", "user-timeline")
		if err != nil {
			return []int64{}, err
		}
		//query := fmt.Sprintf(`{"UserID": %[1]d}`, userID)
		//projection := fmt.Sprintf(`{"projection": {"posts": {"$slice": [0, %[1]d]}}}`, stop)
		query := bson.D{{}}      // TODO: Fix this
		projection := bson.D{{}} // TODO: Fix this
		post_db_val, err := collection.FindMany(ctx, query, projection)
		if err != nil {
			return []int64{}, err
		}
		post_db_val.All(ctx, &new_post_ids)
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

func (u *UserTimelineServiceImpl) WriteUserTimeline(ctx context.Context, reqID int64, postID int64, userID int64, timestamp int64) error {
	collection, err := u.userTimelineDB.GetCollection(ctx, "user-timeline", "user-timeline")
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
		update := fmt.Sprintf(`{"$push": {"Posts": {"$each": [{"PostID": %s, "Timestamp": %s}], "$position": 0}}}`, postIDstr, timestampstr)
		update_d, err := backend.ParseNoSQLDBQuery(update)
		if err != nil {
			return err
		}
		_, err = collection.UpdateMany(ctx, query, update_d)
		if err != nil {
			return errors.New("Failed to insert user timeline user to Database")
		}
	}
	var postInfo []PostInfo
	userIDStr := strconv.FormatInt(userID, 10)
	// Ignore error check for Get!
	u.userTimelineCache.Get(ctx, userIDStr, &postInfo)
	postInfo = append(postInfo, PostInfo{PostID: postID, Timestamp: timestamp})
	return u.userTimelineCache.Put(ctx, userIDStr, postInfo)
}
