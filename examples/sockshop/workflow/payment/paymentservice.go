// Package payment implements the SockShop payment microservice.
//
// The service fakes payments, implementing simple logic whereby payments
// are authorized when they're below a predefined threshold, and rejected
// when they are above that threshold.
package payment

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

// PaymentService provides payment services
type PaymentService interface {
	Authorise(ctx context.Context, amount float32) (Authorisation, error)
}

type Authorisation struct {
	Authorised bool   `json:"authorised"`
	Message    string `json:"message"`
}

// Returns a payment service where any transaction above the preconfigured
// threshold will return an invalid payment amount
func NewPaymentService(ctx context.Context, declineOverAmount string) (PaymentService, error) {
	amount, err := strconv.ParseFloat(declineOverAmount, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid declineOverAmount %v; expected a float32", declineOverAmount)
	}
	return &paymentImpl{
		declineOverAmount: float32(amount),
	}, nil
}

type paymentImpl struct {
	declineOverAmount float32
}

var ErrInvalidPaymentAmount = errors.New("invalid payment amount")

func (s *paymentImpl) Authorise(ctx context.Context, amount float32) (Authorisation, error) {
	if amount == 0 {
		return Authorisation{}, ErrInvalidPaymentAmount
	}
	if amount < 0 {
		return Authorisation{}, ErrInvalidPaymentAmount
	}
	if amount <= s.declineOverAmount {
		return Authorisation{
			Authorised: true,
			Message:    "Payment authorised",
		}, nil
	}
	return Authorisation{
		Authorised: false,
		Message:    fmt.Sprintf("Payment declined: amount exceeds %.2f", s.declineOverAmount),
	}, nil
}
