// Package rebook implements ts-rebook-service from the original train ticket application
package rebook

import (
	"context"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/insidepayment"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/order"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/route"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/seat"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/train"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/travel"
)

type RebookService interface {
	Rebook(ctx context.Context, info RebookInfo) (order.Order, error)
	PayDifference(ctx context.Context, info RebookInfo) (bool, error)
}

type RebookServiceImpl struct {
	insidePaymentService insidepayment.InsidePaymentService
	routeService         route.RouteService
	seatService          seat.SeatService
	trainService         train.TrainService
	travelService        travel.TravelService
	travel2Service       travel.TravelService
	orderService         order.OrderService
	orderOtherService    order.OrderService
}

func NewRebookServiceImpl(ctx context.Context, insidePaymentService insidepayment.InsidePaymentService, routeService route.RouteService, seatService seat.SeatService, trainService train.TrainService, travelService travel.TravelService, travel2Service travel.TravelService, orderService order.OrderService, orderOtherService order.OrderService) (*RebookServiceImpl, error) {
	return &RebookServiceImpl{insidePaymentService, routeService, seatService, trainService, travelService, travel2Service, orderService, orderOtherService}, nil
}
