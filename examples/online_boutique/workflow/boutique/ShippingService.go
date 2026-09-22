package boutique

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type ShippingService interface {
	GetQuote(ctx context.Context, addr Address, items []CartItem) (Money, error)
	ShipOrder(ctx context.Context, addr Address, items []CartItem) (string, error)
}

type ShippingServiceImpl struct {
	logger backend.Logger
}

func NewShippingServiceImpl(ctx context.Context) (ShippingService, error) {
	return &ShippingServiceImpl{logger: backend.GetLogger()}, nil
}

func (s *ShippingServiceImpl) GetQuote(ctx context.Context, addr Address, items []CartItem) (Money, error) {
	ctx, _ = s.logger.Info(ctx, "GetQuote received: items=%d", len(items))
	q := createQuoteFromCount(len(items))
	return Money{CurrencyCode: "USD", Units: int64(q.Dollars), Nanos: int32(q.Cents * 10_000_000)}, nil
}

func (s *ShippingServiceImpl) ShipOrder(ctx context.Context, addr Address, items []CartItem) (string, error) {
	ctx, _ = s.logger.Info(ctx, "ShipOrder received: items=%d", len(items))
	baseAddress := fmt.Sprintf("%s, %s, %s", addr.StreetAddress, addr.City, addr.State)
	return createTrackingID(baseAddress), nil
}

func createTrackingID(salt string) string {
	return fmt.Sprintf("%c%c-%d%s-%d%s", randomLetter(), randomLetter(), len(salt), randomDigits(3), len(salt)/2, randomDigits(7))
}

func randomLetter() byte { return byte('A' + rand.Intn(25)) }

func randomDigits(digits int) string {
	var b strings.Builder
	b.Grow(digits)
	for i := 0; i < digits; i++ {
		b.WriteByte(byte('0' + rand.Intn(10)))
	}
	return b.String()
}
