// package news implements the ts-news-service from the TrainTicket application
package news

import "context"

// News Service provides the latest news about the application
type NewsService interface {
	Hello(ctx context.Context, val string) (string, error)
}

// News Service Implementation
type NewsServiceImpl struct{}

func NewNewsServiceImpl(ctx context.Context) (*NewsServiceImpl, error) {
	return &NewsServiceImpl{}, nil
}

func (n *NewsServiceImpl) Hello(ctx context.Context, val string) (string, error) {
	var str = []byte(`[
                       {"Title": "News Service Complete", "Content": "Congratulations:Your News Service Complete"},
                       {"Title": "Total Ticket System Complete", "Content": "Just a total test"}
                    ]`)
	return string(str), nil
}
