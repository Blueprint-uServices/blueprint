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
// TODO: add declineOverAmount param after implementing config nodes
func NewPaymentService(ctx context.Context /*, declineOverAmount string*/) (PaymentService, error) {
	declineOverAmount := "500"
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
	authorised := false
	message := "Payment declined"
	if amount <= s.declineOverAmount {
		authorised = true
		message = "Payment authorised"
	} else {
		message = fmt.Sprintf("Payment declined: amount exceeds %.2f", s.declineOverAmount)
	}
	return Authorisation{
		Authorised: authorised,
		Message:    message,
	}, nil
}
