// Package notification implements ts-notification service from the original TrainTicket application
// Currently does not have functionality to send emails.
package notification

import (
	"context"
	"fmt"
)

type NotificationService interface {
	PreserveSuccess(ctx context.Context, info NotificationInfo) error
	OrderCreateSuccess(ctx context.Context, info NotificationInfo) error
	OrderChangedSuccess(ctx context.Context, info NotificationInfo) error
	OrderCancelSuccess(ctx context.Context, info NotificationInfo) error
}

type NotificationServiceImpl struct {
	emailSender string
}

func NewNotificationServiceImpl(ctx context.Context) (*NotificationServiceImpl, error) {

	return &NotificationServiceImpl{emailSender: "train-ticket@mpi-sws.org"}, nil
}

func (nsi *NotificationServiceImpl) PreserveSuccess(ctx context.Context, info NotificationInfo) error {

	mail := map[string]interface{}{
		"mailFrom": nsi.emailSender,
		"mailTo":   info.Email,
		"subject":  "Reservation successful",
		"model": map[string]interface{}{
			"username":      info.Username,
			"startingPlace": info.StartingPlace,
			"endPlace":      info.EndPlace,
			"startingTime":  info.StartingTime,
			"date":          info.Date,
			"seatClass":     info.SeatClass,
			"seatNumber":    info.SeatNumber,
			"price":         info.Price,
		},
	}

	fmt.Println(mail)
	return nil
}

func (nsi *NotificationServiceImpl) OrderCreateSuccess(ctx context.Context, info NotificationInfo) error {
	mail := map[string]interface{}{
		"mailFrom": nsi.emailSender,
		"mailTo":   info.Email,
		"subject":  "Successful order creation",
		"model": map[string]interface{}{
			"username":      info.Username,
			"startingPlace": info.StartingPlace,
			"endPlace":      info.EndPlace,
			"startingTime":  info.StartingTime,
			"date":          info.Date,
			"seatClass":     info.SeatClass,
			"seatNumber":    info.SeatNumber,
			"orderNumber":   info.OrderNumber,
		},
	}

	fmt.Println(mail)
	return nil
}

func (nsi *NotificationServiceImpl) OrderChangedSuccess(ctx context.Context, info NotificationInfo) error {
	mail := map[string]interface{}{
		"mailFrom": nsi.emailSender,
		"mailTo":   info.Email,
		"subject":  "Successful order update",
		"model": map[string]interface{}{
			"username":      info.Username,
			"startingPlace": info.StartingPlace,
			"endPlace":      info.EndPlace,
			"startingTime":  info.StartingTime,
			"date":          info.Date,
			"seatClass":     info.SeatClass,
			"seatNumber":    info.SeatNumber,
			"orderNumber":   info.OrderNumber,
		},
	}

	fmt.Println(mail)
	return nil
}

func (nsi *NotificationServiceImpl) OrderCancelSuccess(ctx context.Context, info NotificationInfo) error {
	mail := map[string]interface{}{
		"mailFrom": nsi.emailSender,
		"mailTo":   info.Email,
		"subject":  "Successful order cancelation",
		"model": map[string]interface{}{
			"username": info.Username,
			"price":    info.Price,
		},
	}

	fmt.Println(mail)
	return nil
}
