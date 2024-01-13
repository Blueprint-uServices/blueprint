// Package cancel implements ts-cancel service from original Train Ticket application
package cancel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/insidepayment"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/notification"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/order"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/user"
)

type CancelService interface {
	CalculateRefund(ctx context.Context, orderId string) (float64, error)
	CancelTicket(ctx context.Context, orderId string, userId string) (string, error)
}

type CancelServiceImpl struct {
	insidePaymentService insidepayment.InsidePaymentService
	notificationService  notification.NotificationService
	orderService         order.OrderService
	orderOtherService    order.OrderService
	userService          user.UserService
}

func NewCancelServiceImpl(ctx context.Context, insidePaymentService insidepayment.InsidePaymentService, notificationService notification.NotificationService, orderService order.OrderService, orderOtherService order.OrderService, userService user.UserService) (*CancelServiceImpl, error) {
	return &CancelServiceImpl{insidePaymentService, notificationService, orderService, orderOtherService, userService}, nil
}

func (csi *CancelServiceImpl) Calculate(ctx context.Context, orderId string) (float64, error) {

	var o order.Order
	o, err := csi.orderService.GetOrderById(ctx, orderId)
	if err != nil {
		o, err = csi.orderOtherService.GetOrderById(ctx, orderId)
	}

	if err != nil {
		return 0.0, err
	}

	if o.Status == order.NotPaid {
		return 0.0, errors.New("Nothing to refund")
	} else if o.Status == order.Paid {
		nowDate := time.Now()
		trDate, _ := time.Parse(time.ANSIC, o.TravelDate)

		if nowDate.After(trDate) {
			return 0.0, nil
		} else {
			price := o.Price * 0.8
			return price, nil
		}
	} else {
		return 0.0, errors.New("Refund not permitted")
	}
}

func (csi *CancelServiceImpl) CancelTicket(ctx context.Context, orderId string, userId string) (string, error) {
	var o order.Order
	o, err := csi.orderService.GetOrderById(ctx, orderId)
	if err != nil {
		o, err = csi.orderOtherService.GetOrderById(ctx, orderId)
	}

	if err != nil {
		return "", err
	}

	status := o.Status
	if status != order.Paid && status != order.NotPaid && status != order.Change {
		return "", errors.New("Cancelation not permitted.")
	}

	nowDate := time.Now()
	trDate, _ := time.Parse(time.ANSIC, o.TravelDate)

	var refund string
	if nowDate.After(trDate) {
		refund = "0"
	} else {
		refund = fmt.Sprintf("%f", o.Price*0.8)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	var err1, err2 error

	var u user.User
	go func() {
		defer wg.Done()
		_, err1 = csi.insidePaymentService.DrawBack(ctx, userId, refund)
	}()
	go func() {
		defer wg.Done()
		u, err2 = csi.userService.FindByUserID(ctx, o.AccountId)
	}()
	wg.Wait()
	if err1 != nil {
		return "", err1
	}
	if err2 != nil {
		return "", err2
	}

	err = csi.notificationService.OrderCancelSuccess(ctx, notification.NotificationInfo{
		Email:         u.Email,
		OrderNumber:   o.Id,
		Username:      u.Username,
		StartingPlace: o.From,
		EndPlace:      o.To,
		StartingTime:  o.TravelDate,
		SeatClass:     fmt.Sprintf("%d", o.SeatClass),
		Price:         fmt.Sprintf("%f", o.Price),
	})

	if err != nil {
		return "", nil
	}

	return "Cancelation successful", nil
}
