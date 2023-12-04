package socialnetwork

import (
	"context"
	"log"
	"sync"
	"time"
)

type ComposePostService interface {
	ComposePost(ctx context.Context, reqID int64, username string, userID int64, text string, mediaIDs []int64, mediaTypes []string, post_type int64) (int64, []int64, error)
}

type ComposePostServiceImpl struct {
	postStorageService  PostStorageService
	userTimelineService UserTimelineService
	userService         UserService
	uniqueIDService     UniqueIdService
	mediaService        MediaService
	textService         TextService
	homeTimelineService HomeTimelineService
}

func NewComposePostServiceImpl(ctx context.Context, postStorageService PostStorageService, userTimelineService UserTimelineService, userService UserService, uniqueIDService UniqueIdService, mediaService MediaService, textService TextService, homeTimelineService HomeTimelineService) (ComposePostService, error) {
	return &ComposePostServiceImpl{postStorageService: postStorageService, userTimelineService: userTimelineService, userService: userService, uniqueIDService: uniqueIDService, mediaService: mediaService, textService: textService, homeTimelineService: homeTimelineService}, nil
}

func (c *ComposePostServiceImpl) ComposePost(ctx context.Context, reqID int64, username string, userID int64, text string, mediaIDs []int64, mediaTypes []string, post_type int64) (int64, []int64, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	var err1, err2, err3, err4 error
	var uniqueID int64
	var creator Creator
	var up_text string
	var medias []Media
	var urls []URL
	var usermentions []UserMention
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		up_text, usermentions, urls, err1 = c.textService.ComposeText(ctx, reqID, text)
	}()
	go func() {
		defer wg.Done()
		medias, err2 = c.mediaService.ComposeMedia(ctx, reqID, mediaTypes, mediaIDs)
	}()
	go func() {
		defer wg.Done()
		uniqueID, err3 = c.uniqueIDService.ComposeUniqueId(ctx, reqID, post_type)
	}()
	go func() {
		defer wg.Done()
		creator, err4 = c.userService.ComposeCreatorWithUserId(ctx, reqID, userID, username)
	}()
	wg.Wait()

	if err1 != nil {
		return -1, []int64{}, err1
	}
	if err2 != nil {
		return -1, []int64{}, err2
	}
	if err3 != nil {
		return -1, []int64{}, err3
	}
	if err4 != nil {
		return -1, []int64{}, err4
	}
	var post Post
	post.PostID = uniqueID
	post.Creator = creator
	post.Medias = medias
	post.Text = up_text
	post.Urls = urls
	post.UserMentions = usermentions
	post.ReqID = reqID
	post.PostType = post_type

	var usermentionIds []int64
	for _, um := range usermentions {
		usermentionIds = append(usermentionIds, um.UserID)
	}
	var wg2 sync.WaitGroup
	wg2.Add(3)
	go func() {
		defer wg2.Done()
		err1 = c.postStorageService.StorePost(ctx, reqID, post)
	}()
	go func() {
		defer wg2.Done()
		err2 = c.userTimelineService.WriteUserTimeline(ctx, reqID, uniqueID, userID, timestamp)
		log.Println(err2)
	}()
	go func() {
		defer wg2.Done()
		err3 = c.homeTimelineService.WriteHomeTimeline(ctx, reqID, uniqueID, userID, timestamp, usermentionIds)
		log.Println(err3)
	}()
	wg2.Wait()
	if err1 != nil {
		return uniqueID, usermentionIds, err1
	}
	if err2 != nil {
		return uniqueID, usermentionIds, err2
	}
	return uniqueID, usermentionIds, err3
}
