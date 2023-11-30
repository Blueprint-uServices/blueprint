package socialnetwork

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"
)

type Wrk2APIService interface {
	ReadHomeTimeline(ctx context.Context, user_id int64, start int64, stop int64) ([]int64, error)
	ReadUserTimeline(ctx context.Context, user_id int64, start int64, stop int64) ([]int64, error)
	Follow(ctx context.Context, username string, followeeName string, user_id int64, followeeID int64) error
	Unfollow(ctx context.Context, username string, followeeName string, user_id int64, followeeID int64) error
	Register(ctx context.Context, firstName string, lastName string, username string, password string, user_id int64) error
	ComposePost(ctx context.Context, user_id int64, username string, post_type string, text string, media_types []string, media_ids []int64) (int64, []int64, error)
}

type Wrk2APIServiceImpl struct {
	userService         UserService
	composePostService  ComposePostService
	userTimelineService UserTimelineService
	homeTimelineService HomeTimelineService
	socialGraphService  SocialGraphService
}

func NewWrk2APIServiceImpl(ctx context.Context, userService UserService, composePostService ComposePostService, userTimelineService UserTimelineService, homeTimelineService HomeTimelineService, socialGraphService SocialGraphService) (Wrk2APIService, error) {
	rand.Seed(time.Now().UnixNano())
	return &Wrk2APIServiceImpl{userService: userService, composePostService: composePostService, userTimelineService: userTimelineService, homeTimelineService: homeTimelineService, socialGraphService: socialGraphService}, nil
}

func (w *Wrk2APIServiceImpl) ReadUserTimeline(ctx context.Context, user_id int64, start int64, stop int64) ([]int64, error) {
	reqID := rand.Int63()
	return w.userTimelineService.ReadUserTimeline(ctx, reqID, user_id, start, stop)
}

func (w *Wrk2APIServiceImpl) ReadHomeTimeline(ctx context.Context, user_id int64, start int64, stop int64) ([]int64, error) {
	reqID := rand.Int63()
	return w.homeTimelineService.ReadHomeTimeline(ctx, reqID, user_id, start, stop)
}

func (w *Wrk2APIServiceImpl) Follow(ctx context.Context, username string, followeeName string, user_id int64, followeeID int64) error {
	reqID := rand.Int63()
	if user_id != 0 && followeeID != 0 {
		return w.socialGraphService.Follow(ctx, reqID, user_id, followeeID)
	} else if username != "" && followeeName != "" {
		return w.socialGraphService.FollowWithUsername(ctx, reqID, username, followeeName)
	}
	return errors.New("Invalid Arguments")
}

func (w *Wrk2APIServiceImpl) Unfollow(ctx context.Context, username string, followeeName string, user_id int64, followeeID int64) error {
	reqID := rand.Int63()
	if user_id != 0 && followeeID != 0 {
		return w.socialGraphService.Unfollow(ctx, reqID, user_id, followeeID)
	} else if username != "" && followeeName != "" {
		return w.socialGraphService.UnfollowWithUsername(ctx, reqID, username, followeeName)
	}
	return errors.New("Invalid Arguments")
}

func (w *Wrk2APIServiceImpl) Register(ctx context.Context, firstName string, lastName string, username string, password string, user_id int64) error {
	if firstName == "" || lastName == "" || username == "" || password == "" {
		return errors.New("Incomplete Arguments")
	}
	reqID := rand.Int63()
	return w.userService.RegisterUserWithId(ctx, reqID, firstName, lastName, username, password, user_id)
}

func (w *Wrk2APIServiceImpl) ComposePost(ctx context.Context, user_id int64, username string, post_type string, text string, media_types []string, media_ids []int64) (int64, []int64, error) {
	if user_id == 0 || username == "" || post_type == "" || text == "" {
		return -1, []int64{}, errors.New("Incomplete Arguments")
	}
	reqID := rand.Int63()
	postInt, err := strconv.ParseInt(post_type, 10, 64)
	if err != nil {
		return -1, []int64{}, err
	}
	postType := PostType(postInt)
	return w.composePostService.ComposePost(ctx, reqID, username, user_id, text, media_ids, media_types, postType)
}
