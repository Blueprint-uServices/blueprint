package media

import (
	"context"
	"strconv"
	"sync"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type UserReview struct {
	UserID      int
	ReviewInfos []ReviewInfo
}

type UserReviewService interface {
	UploadUserReview(ctx context.Context, reqID int, userID int, reviewID int, timestamp string) error
	ReadUserReviews(ctx context.Context, reqID int, userID int, start int, stop int) ([]Review, error)
}

type UserReviewServiceImpl struct {
	reviewStorageService ReviewStorageService
	userReviewDB         backend.NoSQLDatabase
	userReviewCache      backend.Cache
	mu                   sync.Mutex
}

func NewUserReviewServiceImpl(reviewStorageService ReviewStorageService, userReviewDB backend.NoSQLDatabase, userReviewCache backend.Cache) (UserReviewService, error) {
	return &UserReviewServiceImpl{reviewStorageService: reviewStorageService, userReviewDB: userReviewDB, userReviewCache: userReviewCache}, nil
}

func (u *UserReviewServiceImpl) UploadUserReview(ctx context.Context, reqID int, userID int, reviewID int, timestamp string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	collection, err := u.userReviewDB.GetCollection(ctx, "user-review", "user-review")
	if err != nil {
		return err
	}
	query := bson.D{{"userid", userID}}
	result, err := collection.FindOne(ctx, query)
	if err != nil {
		return err
	}
	var userReview UserReview
	found, err := result.One(ctx, &userReview)
	if err != nil {
		return err
	}
	userReview.UserID = userID
	userReview.ReviewInfos = append([]ReviewInfo{{ReviewID: reviewID, Timestamp: timestamp}}, userReview.ReviewInfos...)
	if found {
		_, err = collection.ReplaceOne(ctx, query, userReview)
	} else {
		err = collection.InsertOne(ctx, userReview)
	}
	if err != nil {
		return err
	}
	return u.userReviewCache.Put(ctx, strconv.Itoa(userID), userReview.ReviewInfos)
}

func (u *UserReviewServiceImpl) ReadUserReviews(ctx context.Context, reqID int, userID int, start int, stop int) ([]Review, error) {
	if start < 0 || stop <= start {
		return []Review{}, nil
	}
	infos, err := u.readReviewInfos(ctx, userID)
	if err != nil {
		return nil, err
	}
	if start >= len(infos) {
		return []Review{}, nil
	}
	if stop > len(infos) {
		stop = len(infos)
	}
	ids := make([]int, 0, stop-start)
	for _, info := range infos[start:stop] {
		ids = append(ids, info.ReviewID)
	}
	return u.reviewStorageService.ReadReviews(ctx, reqID, ids)
}

func (u *UserReviewServiceImpl) readReviewInfos(ctx context.Context, userID int) ([]ReviewInfo, error) {
	key := strconv.Itoa(userID)
	var infos []ReviewInfo
	found, err := u.userReviewCache.Get(ctx, key, &infos)
	if err != nil || found {
		return infos, err
	}
	collection, err := u.userReviewDB.GetCollection(ctx, "user-review", "user-review")
	if err != nil {
		return nil, err
	}
	result, err := collection.FindOne(ctx, bson.D{{"userid", userID}})
	if err != nil {
		return nil, err
	}
	var userReview UserReview
	found, err = result.One(ctx, &userReview)
	if err != nil {
		return nil, err
	}
	if !found {
		return []ReviewInfo{}, nil
	}
	return userReview.ReviewInfos, u.userReviewCache.Put(ctx, key, userReview.ReviewInfos)
}
