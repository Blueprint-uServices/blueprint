package media

import "context"

type TextService interface {
	UploadText(ctx context.Context, reqID int, text string) error
}

type TextServiceImpl struct{ composeReviewService ComposeReviewService }

func NewTextServiceImpl(composeReviewService ComposeReviewService) (TextService, error) {
	return &TextServiceImpl{composeReviewService: composeReviewService}, nil
}

func (t *TextServiceImpl) UploadText(ctx context.Context, reqID int, text string) error {
	return t.composeReviewService.UploadText(ctx, reqID, text)
}
