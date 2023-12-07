package socialnetwork

import (
	"context"
	"strconv"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

type HomeTimelineService interface {
	ReadHomeTimeline(ctx context.Context, reqID int64, userID int64, start int64, stop int64) ([]int64, error)
	WriteHomeTimeline(ctx context.Context, reqID int64, postID int64, userID int64, timestamp int64, userMentionIDs []int64) error
}

type HomeTimelineServiceImpl struct {
	homeTimelineCache  backend.Cache
	postStorageService PostStorageService
	socialGraphService SocialGraphService
}

func NewHomeTimelineServiceImpl(ctx context.Context, homeTimelineCache backend.Cache, postStorageService PostStorageService, socialGraphService SocialGraphService) (HomeTimelineService, error) {
	return &HomeTimelineServiceImpl{homeTimelineCache: homeTimelineCache, postStorageService: postStorageService, socialGraphService: socialGraphService}, nil
}

func (h *HomeTimelineServiceImpl) WriteHomeTimeline(ctx context.Context, reqID int64, postID int64, userID int64, timestamp int64, userMentionIDs []int64) error {
	followers, err := h.socialGraphService.GetFollowers(ctx, reqID, userID)
	if err != nil {
		return err
	}
	followers_set := make(map[int64]bool)
	for _, follower := range followers {
		followers_set[follower] = true
	}
	for _, um := range userMentionIDs {
		followers_set[um] = true
	}
	for id, _ := range followers_set {
		id_str := strconv.FormatInt(id, 10)
		var posts []PostInfo
		_, err = h.homeTimelineCache.Get(ctx, id_str, &posts)
		if err != nil {
			return err
		}
		posts = append(posts, PostInfo{PostID: postID, Timestamp: timestamp})
		err = h.homeTimelineCache.Put(ctx, id_str, posts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *HomeTimelineServiceImpl) ReadHomeTimeline(ctx context.Context, reqID int64, userID int64, start int64, stop int64) ([]int64, error) {
	if stop <= start || start < 0 {
		return []int64{}, nil
	}
	userIDStr := strconv.FormatInt(userID, 10)
	var postIDs []int64
	var postInfos []PostInfo
	_, err := h.homeTimelineCache.Get(ctx, userIDStr, &postInfos)
	if err != nil {
		return []int64{}, err
	}
	for _, pinfo := range postInfos {
		postIDs = append(postIDs, pinfo.PostID)
	}
	if start < int64(len(postIDs)) {
		minstop := stop
		if stop > int64(len(postIDs)) {
			minstop = int64(len(postIDs))
		}
		postIDs = postIDs[start:minstop]
	}
	_, err = h.postStorageService.ReadPosts(ctx, reqID, postIDs)
	if err != nil {
		return postIDs, err
	}
	return postIDs, nil
}
