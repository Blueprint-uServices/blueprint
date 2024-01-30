// Package avatar implements the ts-avatar-service from the original application
package avatar

import "context"

type AvatarService interface {
	Hello(ctx context.Context, img []byte) ([]byte, error)
}

type AvatarServiceImpl struct{}

func NewAvatarServiceImpl(ctx context.Context) (*AvatarServiceImpl, error) {
	return &AvatarServiceImpl{}, nil
}

func (a *AvatarServiceImpl) Hello(ctx context.Context, img []byte) ([]byte, error) {
	// TODO: Find a face detection library and use that to detect faces
	// In the meantime, return the received image as is.
	return img, nil
}
