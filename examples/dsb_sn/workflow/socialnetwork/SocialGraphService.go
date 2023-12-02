package socialnetwork

import (
	"context"
	"strconv"
	"sync"
	"time"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

type SocialGraphService interface {
	GetFollowers(ctx context.Context, reqID int64, userID int64) ([]int64, error)
	GetFollowees(ctx context.Context, reqID int64, userID int64) ([]int64, error)
	Follow(ctx context.Context, reqID int64, userID int64, followeeID int64) error
	Unfollow(ctx context.Context, reqID int64, userID int64, followeeID int64) error
	FollowWithUsername(ctx context.Context, reqID int64, userUsername string, followeeUsername string) error
	UnfollowWithUsername(ctx context.Context, reqID int64, userUsername string, followeeUsername string) error
	InsertUser(ctx context.Context, reqID int64, userID int64) error
}

type FollowerInfo struct {
	FollowerID int64
	Timestamp  int64
}

type FolloweeInfo struct {
	FolloweeID int64
	Timestamp  int64
}

type UserInfo struct {
	UserID    int64
	Followers []FollowerInfo
	Followees []FolloweeInfo
}

type SocialGraphServiceImpl struct {
	socialGraphCache backend.Cache
	socialGraphDB    backend.NoSQLDatabase
	userIDService    UserIDService
}

func NewSocialGraphServiceImpl(ctx context.Context, socialGraphCache backend.Cache, socialGraphDB backend.NoSQLDatabase, userIDService UserIDService) (SocialGraphService, error) {
	return &SocialGraphServiceImpl{socialGraphCache: socialGraphCache, socialGraphDB: socialGraphDB, userIDService: userIDService}, nil
}

func (s *SocialGraphServiceImpl) GetFollowers(ctx context.Context, reqID int64, userID int64) ([]int64, error) {
	var followers []int64
	var followerInfos []FollowerInfo
	userIDstr := strconv.FormatInt(userID, 10)
	err := s.socialGraphCache.Get(ctx, userIDstr+":followers", &followerInfos)
	if err != nil {
		collection, err := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
		if err != nil {
			return followers, err
		}
		query := `{"UserID":` + userIDstr + `}`
		query_d, err := backend.ParseNoSQLDBQuery(query)
		val, err := collection.FindOne(ctx, query_d)
		if err != nil {
			return followers, err
		}
		var userInfo UserInfo
		val.One(ctx, &userInfo)
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

func (s *SocialGraphServiceImpl) GetFollowees(ctx context.Context, reqID int64, userID int64) ([]int64, error) {
	var followees []int64
	var followeeInfos []FolloweeInfo
	userIDstr := strconv.FormatInt(userID, 10)
	err := s.socialGraphCache.Get(ctx, userIDstr+":followees", &followeeInfos)
	if err != nil {
		collection, err := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
		if err != nil {
			return followees, err
		}
		query := `{"UserID":` + userIDstr + `}`
		query_d, err := backend.ParseNoSQLDBQuery(query)
		if err != nil {
			return followees, err
		}
		val, err := collection.FindOne(ctx, query_d)
		if err != nil {
			return followees, err
		}
		var userInfo UserInfo
		val.One(ctx, &userInfo)
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
		query_d, err_internal := backend.ParseNoSQLDBQuery(query)
		if err_internal != nil {
			err1 = err_internal
			return
		}
		update_d, err_internal := backend.ParseNoSQLDBQuery(update)
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
		query_d, err_internal := backend.ParseNoSQLDBQuery(query)
		if err_internal != nil {
			err2 = err_internal
			return
		}
		update_d, err_internal := backend.ParseNoSQLDBQuery(update)
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
		query_d, err_internal := backend.ParseNoSQLDBQuery(query)
		if err_internal != nil {
			err1 = err_internal
			return
		}
		update_d, err_internal := backend.ParseNoSQLDBQuery(update)
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
		update := `{"$pull": {"followees": {"UserID": ` + userIDstr + `}}}`
		query_d, err_internal := backend.ParseNoSQLDBQuery(query)
		if err_internal != nil {
			err2 = err_internal
			return
		}
		update_d, err_internal := backend.ParseNoSQLDBQuery(update)
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

func (s *SocialGraphServiceImpl) InsertUser(ctx context.Context, reqID int64, userID int64) error {
	collection, err := s.socialGraphDB.GetCollection(ctx, "social-graph", "social-graph")
	if err != nil {
		return err
	}
	user := UserInfo{UserID: userID, Followers: []FollowerInfo{}, Followees: []FolloweeInfo{}}
	return collection.InsertOne(ctx, user)
}
