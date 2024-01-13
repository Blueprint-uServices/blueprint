package admin

import (
	"context"
	"errors"
	"sync"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/order"
)

func CreateAdminOrderService(ctx context.Context, orderService order.OrderService, orderOtherService order.OrderService) (*AdminOrderServiceImpl, error) {
	return &AdminOrderServiceImpl{
		orderService:      orderService,
		orderOtherService: orderOtherService,
	}, nil
}

type AdminOrderService interface {
	GetAllOrders(ctx context.Context) ([]order.Order, error)
	DeleteOrder(ctx context.Context, orderId string) (string, error)
	UpdateOrder(ctx context.Context, o order.Order) (order.Order, error)
	AddOrder(ctx context.Context, o order.Order) (order.Order, error)
}

type AdminOrderServiceImpl struct {
	orderService      order.OrderService
	orderOtherService order.OrderService
}

func (aosi *AdminOrderServiceImpl) GetAllOrders(ctx context.Context, token string) ([]order.Order, error) {

	var err1, err2 error
	var ordersFirstBatch, ordersSecondBatch []order.Order
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		ordersFirstBatch, err1 = aosi.orderService.FindAllOrder(ctx)
	}()

	go func() {
		defer wg.Done()
		ordersSecondBatch, err2 = aosi.orderOtherService.FindAllOrder(ctx)
	}()
	wg.Wait()

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	orders := append(ordersFirstBatch, ordersSecondBatch...)

	if len(orders) == 0 {
		return nil, errors.New("No orders found")
	}

	return orders, nil
}

func (aosi *AdminOrderServiceImpl) DeleteOrder(ctx context.Context, orderId string, trainNumber string) (string, error) {
	var msg string
	var err error
	if trainNumber[0:1] == "D" || trainNumber[0:1] == "G" {
		msg, err = aosi.orderService.DeleteOrder(ctx, orderId)

	} else {
		msg, err = aosi.orderOtherService.DeleteOrder(ctx, orderId)
	}

	if err != nil {
		return "", err
	}

	return msg, nil
}

func (aosi *AdminOrderServiceImpl) UpdateOrder(ctx context.Context, o order.Order) (order.Order, error) {
	var err error
	if o.TrainNumber[0:1] == "D" || o.TrainNumber[0:1] == "G" {
		o, err = aosi.orderService.UpdateOrder(ctx, o)
	} else {
		o, err = aosi.orderOtherService.UpdateOrder(ctx, o)
	}

	if err != nil {
		return order.Order{}, err
	}

	return o, nil
}

func (aosi *AdminOrderServiceImpl) AddOrder(ctx context.Context, o order.Order) (order.Order, error) {
	var err error
	if o.TrainNumber[0:1] == "D" || o.TrainNumber[0:1] == "G" {
		o, err = aosi.orderService.AddCreateNewOrder(ctx, o)
	} else {
		o, err = aosi.orderOtherService.AddCreateNewOrder(ctx, o)
	}

	if err != nil {
		return order.Order{}, err
	}

	return o, nil
}
