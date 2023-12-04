package socialnetwork

import (
	"context"
	"errors"
)

type MediaService interface {
	ComposeMedia(ctx context.Context, reqID int64, mediaTypes []string, mediaIds []int64) ([]Media, error)
}

type MediaServiceImpl struct{}

func NewMediaServiceImpl(ctx context.Context) (MediaService, error) {
	return &MediaServiceImpl{}, nil
}

func (m *MediaServiceImpl) ComposeMedia(ctx context.Context, reqID int64, mediaTypes []string, mediaIds []int64) ([]Media, error) {
	var medias []Media

	if len(mediaTypes) != len(mediaIds) {
		return medias, errors.New("The lengths of media_id list and media_type list are not equal")
	}

	for idx, mediaId := range mediaIds {
		media := Media{MediaID: mediaId, MediaType: mediaTypes[idx]}
		medias = append(medias, media)
	}

	return medias, nil
}
