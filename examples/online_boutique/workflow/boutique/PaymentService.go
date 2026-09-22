package boutique

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

var (
	ErrInvalidCreditCard = errors.New("invalid credit card")
	ErrExpiredCreditCard = errors.New("credit card expired")
)

type PaymentService interface {
	Charge(ctx context.Context, amount Money, card CreditCardInfo) (string, error)
}

type PaymentServiceImpl struct {
	logger backend.Logger
}

func NewPaymentServiceImpl(ctx context.Context) (PaymentService, error) {
	return &PaymentServiceImpl{logger: backend.GetLogger()}, nil
}

func (s *PaymentServiceImpl) Charge(ctx context.Context, amount Money, card CreditCardInfo) (string, error) {
	number := strings.ReplaceAll(card.Number, "-", "")
	var company string
	switch {
	case len(number) < 4:
		return "", ErrInvalidCreditCard
	case number[0] == '4':
		company = "Visa"
	case number[0] == '5':
		company = "MasterCard"
	default:
		return "", ErrInvalidCreditCard
	}
	if card.CVV < 100 || card.CVV > 9999 {
		return "", ErrInvalidCreditCard
	}
	if card.ExpirationMonth < time.January || card.ExpirationMonth > time.December {
		return "", ErrInvalidCreditCard
	}
	expires := time.Date(card.ExpirationYear, card.ExpirationMonth+1, 1, 0, 0, 0, 0, time.Local)
	if !time.Now().Before(expires) {
		return "", ErrExpiredCreditCard
	}
	ctx, _ = s.logger.Info(ctx, "Transaction processed: company=%s last_four=%s currency=%s amount=%s", company, number[len(number)-4:], amount.CurrencyCode, fmt.Sprintf("%d.%09d", amount.Units, amount.Nanos))
	return newUUID()
}
