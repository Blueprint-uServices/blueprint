// Package execute implements ts-execute service from the original TrainTicket application
package execute

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/order"
)

type ExecuteService interface {
	ExecuteTicket(ctx context.Context, orderId string) (string, error)
	CollectTicket(ctx context.Context, orderId string) (string, error)
}

type ExecuteServiceImpl struct {
	orderService      order.OrderService
	orderOtherService order.OrderService
}

func NewExecuteServiceImpl(ctx context.Context, orderService order.OrderService, orderOtherService order.OrderService) (*ExecuteServiceImpl, error) {
	return &ExecuteServiceImpl{orderService, orderOtherService}, nil
}

func (esi *ExecuteServiceImpl) ExecuteTicket(ctx context.Context, orderId string) (string, error) {
	var o order.Order
	var err error
	first := true
	o, err = esi.orderService.GetOrderById(ctx, orderId)
	if err == nil {

		o, err = esi.orderOtherService.GetOrderById(ctx, orderId)
		if err != nil {
			return "", err
		}
		first = false
	}

	if o.Status != order.Paid && o.Status != order.Change {
		return "", errors.New("Order cannot be collected!")
	}

	if first {
		_, err = esi.orderService.ModifyOrder(ctx, orderId, order.Collected)
	} else {
		_, err = esi.orderOtherService.ModifyOrder(ctx, orderId, order.Collected)
	}

	if err != nil {
		return "", err
	}

	return "Order collected successfully", nil
}

func (esi *ExecuteServiceImpl) CollectTicket(ctx context.Context, orderId string, userId string) (string, error) {
	var err error
	first := true
	_, err = esi.orderService.GetOrderById(ctx, orderId)
	if err == nil {

		_, err = esi.orderOtherService.GetOrderById(ctx, orderId)
		if err != nil {
			return "", err
		}
		first = false
	}

	if first {
		_, err = esi.orderService.ModifyOrder(ctx, orderId, order.Used)
	} else {
		_, err = esi.orderOtherService.ModifyOrder(ctx, orderId, order.Used)
	}

	if err != nil {
		return "", err
	}

	return "Order executed successfully", nil

}
