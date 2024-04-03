package socialnetwork

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// The SocialGraphService interface
type SocialGraphService interface {
	// Returns the user IDs of all the followers of the user with user id `userID`.
	// Returns an error if user with `userID` doesn't exist in database.
	GetFollowers(ctx context.Context, reqID int64, userID int64) ([]int64, error)
	// Returns the user IDs of all the followees of the user with user id `userID`.
	// Returns an error if user with `userID` doesn't exist in database.
	GetFollowees(ctx context.Context, reqID int64, userID int64) ([]int64, error)
	// Creates a follower-followee relationship between users with IDs `userID`-`followeeID`.
	Follow(ctx context.Context, reqID int64, userID int64, followeeID int64) error
	// Removes the follower-followee relationship between users with IDs `userID`-`followeeID`.
	Unfollow(ctx context.Context, reqID int64, userID int64, followeeID int64) error
	// Creates a follower-followee relationship between users with usernames `userUsername`-`followeeUsername`.
	FollowWithUsername(ctx context.Context, reqID int64, userUsername string, followeeUsername string) error
	// Removes the follower-followee relationship between users with usernames `userUsername`-`followeeUsername`.
	UnfollowWithUsername(ctx context.Context, reqID int64, userUsername string, followeeUsername string) error
	// Inserts a new user with `userID` in the database.
	InsertUser(ctx context.Context, reqID int64, userID int64) error
	// Cleansup the database
	CleanupSocialBackends(ctx context.Context) error
}

// The format of a follower's info stored in the user info in the social-graph
type FollowerInfo struct {
	FollowerID int64
	Timestamp  int64
}

// The format of a followee's info stored in the user info in the social-graph
type FolloweeInfo struct {
	FolloweeID int64
	Timestamp  int64
}

// The format of a user's info stored in the social-graph
type UserInfo struct {
	UserID    int64
	Followers []FollowerInfo
	Followees []FolloweeInfo
}

// Implementation of [SocialGraphService]
type SocialGraphServiceImpl struct {
	socialGraphCache backend.Cache
	socialGraphDB    backend.NoSQLDatabase
	userIDService    UserIDService
}

// Creates a [SocialGraphService] instance that maintains the social graph backends.
func NewSocialGraphServiceImpl(ctx context.Context, socialGraphCache backend.Cache, socialGraphDB backend.NoSQLDatabase, userIDService UserIDService) (SocialGraphService, error) {
	return &SocialGraphServiceImpl{socialGraphCache: socialGraphCache, socialGraphDB: socialGraphDB, userIDService: userIDService}, nil
}

// Implements SocialGraphService interface
func (s *SocialGraphServiceImpl) GetFollowers(ctx context.Context, reqID int64, userID int64) ([]int64, error) {
	var followers []int64
	var followerInfos []FollowerInfo
	userIDstr := strconv.FormatInt(userID, 10)
	exists, err := s.socialGraphCache.Get(ctx, userIDstr+":followers", &followerInfos)
	if err != nil {
		return followers, err
	}
	if !exists {
		collection, err := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
		if err != nil {
			return followers, err
		}
		query := `{"UserID":` + userIDstr + `}`
		query_d, err := parseNoSQLDBQuery(query)
		val, err := collection.FindOne(ctx, query_d)
		if err != nil {
			return followers, err
		}
		var userInfo UserInfo
		in_db, err := val.One(ctx, &userInfo)
		if err != nil {
			return followers, err
		}
		if !in_db {
			return followers, errors.New("User with " + userIDstr + " not found in db")
		}
		for _, follower := range userInfo.Followers {
			followers = append(followers, follower.FollowerID)
		}
		err = s.socialGraphCache.Put(ctx, userIDstr+":followers", userInfo.Followers)
	} else {
		for _, f := range followerInfos {
			followers = append(followers, f.FollowerID)
		}
	}
	return followers, nil
}

// Implements SocialGraphService interface
func (s *SocialGraphServiceImpl) GetFollowees(ctx context.Context, reqID int64, userID int64) ([]int64, error) {
	var followees []int64
	var followeeInfos []FolloweeInfo
	userIDstr := strconv.FormatInt(userID, 10)
	exists, err := s.socialGraphCache.Get(ctx, userIDstr+":followees", &followeeInfos)
	if err != nil {
		return followees, err
	}
	if !exists {
		collection, err := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
		if err != nil {
			return followees, err
		}
		query := `{"UserID":` + userIDstr + `}`
		query_d, err := parseNoSQLDBQuery(query)
		if err != nil {
			return followees, err
		}
		val, err := collection.FindOne(ctx, query_d)
		if err != nil {
			return followees, err
		}
		var userInfo UserInfo
		in_db, err := val.One(ctx, &userInfo)
		if err != nil {
			return followees, err
		}
		if !in_db {
			return followees, errors.New("User wtih " + userIDstr + " doesn't exist in db")
		}
		for _, followee := range userInfo.Followees {
			followees = append(followees, followee.FolloweeID)
		}
		err = s.socialGraphCache.Put(ctx, userIDstr+":followees", userInfo.Followees)
	} else {
		for _, f := range followeeInfos {
			followees = append(followees, f.FolloweeID)
		}
	}
	return followees, nil
}

// Implements SocialGraphService interface
func (s *SocialGraphServiceImpl) Follow(ctx context.Context, reqID int64, userID int64, followeeID int64) error {
	now := time.Now().UnixNano()
	timestamp := strconv.FormatInt(now, 10)
	userIDstr := strconv.FormatInt(userID, 10)
	followeeIDstr := strconv.FormatInt(followeeID, 10)
	var err1, err2, err3 error
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		collection, err_internal := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
		if err_internal != nil {
			err1 = err_internal
			return
		}
		//query := `{"$and": [{"UserID":` + userIDstr + `}, {"followees.userid" : {"$ne":` + followeeIDstr + `}}] }`
		query := `{"UserID": ` + userIDstr + `}`
		update := `{"$push": {"followees": {"UserID": ` + followeeIDstr + `,"Timestamp": ` + timestamp + `}}}`
		query_d, err_internal := parseNoSQLDBQuery(query)
		if err_internal != nil {
			err1 = err_internal
			return
		}
		update_d, err_internal := parseNoSQLDBQuery(update)
		if err_internal != nil {
			err1 = err_internal
			return
		}
		_, err1 = collection.UpdateOne(ctx, query_d, update_d)
	}()
	go func() {
		defer wg.Done()
		collection, err_internal := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
		if err_internal != nil {
			err1 = err_internal
			return
		}
		//query := `{"$and": [{"UserID":` + followeeIDstr + `}, {"followers.userid" : {"$ne":` + userIDstr + `}}] }`
		query := `{"UserID": ` + followeeIDstr + `}`
		update := `{"$push": {"followers": {"UserID": ` + userIDstr + `,"Timestamp": ` + timestamp + `}}}`
		query_d, err_internal := parseNoSQLDBQuery(query)
		if err_internal != nil {
			err2 = err_internal
			return
		}
		update_d, err_internal := parseNoSQLDBQuery(update)
		if err_internal != nil {
			err2 = err_internal
			return
		}
		_, err2 = collection.UpdateOne(ctx, query_d, update_d)
	}()
	go func() {
		defer wg.Done()
		var followees []FolloweeInfo
		var followers []FollowerInfo
		s.socialGraphCache.Get(ctx, userIDstr+":followees", &followees)
		s.socialGraphCache.Get(ctx, followeeIDstr+":followers", &followers)
		followees = append(followees, FolloweeInfo{FolloweeID: followeeID, Timestamp: now})
		followers = append(followers, FollowerInfo{FollowerID: userID, Timestamp: now})
		err3 = s.socialGraphCache.Put(ctx, userIDstr+":followees", followees)
		if err3 != nil {
			return
		}
		err3 = s.socialGraphCache.Put(ctx, followeeIDstr+":followers", followers)
	}()
	wg.Wait()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return err3
}

// Implements SocialGraphService interface
func (s *SocialGraphServiceImpl) Unfollow(ctx context.Context, reqID int64, userID int64, followeeID int64) error {
	userIDstr := strconv.FormatInt(userID, 10)
	followeeIDstr := strconv.FormatInt(followeeID, 10)
	var err1, err2, err3 error
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		collection, err_internal := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
		if err_internal != nil {
			err1 = err_internal
			return
		}
		query := `{"UserID": ` + userIDstr + `}`
		update := `{"$pull": {"followees": {"UserID": ` + followeeIDstr + `}}}`
		query_d, err_internal := parseNoSQLDBQuery(query)
		if err_internal != nil {
			err1 = err_internal
			return
		}
		update_d, err_internal := parseNoSQLDBQuery(update)
		if err_internal != nil {
			err1 = err_internal
			return
		}
		_, err1 = collection.UpdateOne(ctx, query_d, update_d)
	}()
	go func() {
		defer wg.Done()
		collection, err_internal := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
		if err_internal != nil {
			err2 = err_internal
			return
		}
		query := `{"UserID": ` + followeeIDstr + `}`
		update := `{"$pull": {"followers": {"UserID": ` + userIDstr + `}}}`
		query_d, err_internal := parseNoSQLDBQuery(query)
		if err_internal != nil {
			err2 = err_internal
			return
		}
		update_d, err_internal := parseNoSQLDBQuery(update)
		if err_internal != nil {
			err2 = err_internal
			return
		}
		_, err2 = collection.UpdateOne(ctx, query_d, update_d)
	}()
	go func() {
		defer wg.Done()
		var followees []FolloweeInfo
		var followers []FollowerInfo
		var new_followers []FollowerInfo
		var new_followees []FolloweeInfo
		s.socialGraphCache.Get(ctx, userIDstr+":followees", &followees)
		s.socialGraphCache.Get(ctx, followeeIDstr+":followers", &followers)
		for _, followee := range followees {
			if followee.FolloweeID != followeeID {
				new_followees = append(new_followees, followee)
			}
		}
		for _, follower := range followers {
			if follower.FollowerID != userID {
				new_followers = append(new_followers, follower)
			}
		}
		err3 = s.socialGraphCache.Put(ctx, userIDstr+":followees", new_followees)
		if err3 != nil {
			return
		}
		err3 = s.socialGraphCache.Put(ctx, followeeIDstr+":followers", new_followers)
	}()
	wg.Wait()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return err3
}

// Implements SocialGraphService interface
func (s *SocialGraphServiceImpl) FollowWithUsername(ctx context.Context, reqID int64, username string, followee_name string) error {
	var id int64
	var followee_id int64
	var err1 error
	var err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		id, err1 = s.userIDService.GetUserId(ctx, reqID, username)
	}()
	go func() {
		defer wg.Done()
		followee_id, err2 = s.userIDService.GetUserId(ctx, reqID, followee_name)
	}()
	wg.Wait()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return s.Follow(ctx, reqID, id, followee_id)
}

// Implements SocialGraphService interface
func (s *SocialGraphServiceImpl) UnfollowWithUsername(ctx context.Context, reqID int64, username string, followee_name string) error {
	var id int64
	var followee_id int64
	var err1 error
	var err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		id, err1 = s.userIDService.GetUserId(ctx, reqID, username)
	}()
	go func() {
		defer wg.Done()
		followee_id, err2 = s.userIDService.GetUserId(ctx, reqID, followee_name)
	}()
	wg.Wait()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return s.Unfollow(ctx, reqID, id, followee_id)
}

// Implements SocialGraphService interface
func (s *SocialGraphServiceImpl) InsertUser(ctx context.Context, reqID int64, userID int64) error {
	collection, err := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
	if err != nil {
		return err
	}
	user := UserInfo{UserID: userID, Followers: []FollowerInfo{}, Followees: []FolloweeInfo{}}
	return collection.InsertOne(ctx, user)
}

func (s *SocialGraphServiceImpl) CleanupSocialBackends(ctx context.Context) error {
	socialcoll, err := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
	if err != nil {
		return err
	}
	err = socialcoll.DeleteMany(ctx, bson.D{})
	if err != nil {
		return err
	}
	return s.socialGraphCache.DeleteAll(ctx)
}
