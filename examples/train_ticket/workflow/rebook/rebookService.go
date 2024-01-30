// Package rebook implements ts-rebook-service from the original train ticket application
package rebook

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
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

func isTripGD(tripid string) bool {
	if tripid[0:1] == "G" || tripid[0:1] == "D" {
		return true
	}
	return false
}

func (r *RebookServiceImpl) UpdateOrder(ctx context.Context, o order.Order, info RebookInfo, resp travel.TripResponse, ticketPrice float64, trip common.Trip) (order.Order, error) {
	newOrder := order.Order{}
	newOrder.TrainNumber = info.TripID
	newOrder.BoughtDate = time.Now().Format("Sat Jul 26 00:00:00 2025")
	newOrder.Status = order.Change
	newOrder.SeatClass = uint16(info.SeatType)
	newOrder.TravelDate = info.Date
	newOrder.Price = ticketPrice

	rt, err := r.routeService.GetRouteById(ctx, trip.RouteID)
	if err != nil {
		return o, err
	}
	tt, err := r.trainService.RetrieveByName(ctx, trip.TrainTypeName)
	if err != nil {
		return o, err
	}
	var s seat.Seat
	s.TrainNumber = info.TripID
	s.Stations = rt.Stations
	s.SeatType = int(info.SeatType)
	s.StartStation = o.From
	s.DstStation = o.To
	s.TravelDate = info.Date
	if info.SeatType == seat.FIRSTCLASS {
		s.TotalNum = tt.ComfortClass
	} else if info.SeatType == seat.SECONDCLASS {
		s.TotalNum = tt.EconomyClass
	}
	ticket, err := r.seatService.DistributeSeat(ctx, s)
	if err != nil {
		return o, err
	}
	newOrder.SeatClass = uint16(s.SeatType)
	newOrder.SeatNumber = fmt.Sprintf("%d", ticket.SeatNo)

	res1 := isTripGD(info.OldTripID)
	res2 := isTripGD(info.TripID)
	if res1 == res2 {
		newOrder.Id = o.Id
		if res1 {
			return r.orderService.UpdateOrder(ctx, newOrder)
		} else {
			return r.orderOtherService.UpdateOrder(ctx, newOrder)
		}
	} else {
		if res1 {
			_, err := r.orderService.DeleteOrder(ctx, o.Id)
			if err != nil {
				return o, err
			}
			return r.orderOtherService.CreateNewOrder(ctx, newOrder)
		}
		if res2 {
			_, err := r.orderOtherService.DeleteOrder(ctx, o.Id)
			if err != nil {
				return o, err
			}
			return r.orderService.CreateNewOrder(ctx, newOrder)
		}
	}
	return o, errors.New("Failed to update order")
}

func (r *RebookServiceImpl) Rebook(ctx context.Context, info RebookInfo) (order.Order, error) {
	var o order.Order
	var err error
	if isTripGD(info.OldTripID) {
		o, err = r.orderService.GetOrderById(ctx, info.OrderID)
	} else {
		o, err = r.orderOtherService.GetOrderById(ctx, info.OrderID)
	}
	if err != nil {
		return o, err
	}
	if o.Status == order.NotPaid {
		return o, errors.New("Original ticket was not paid. Rebooking not possible.")
	} else if o.Status == order.Change {
		return o, errors.New("Ticket was already changed once. It's cant be changed again.")
	} else if o.Status == order.Collected {
		return o, errors.New("Ticket was already collected. It can't be collected.")
	} else if o.Status != order.Paid {
		return o, errors.New("Ticket can't be changed.")
	}

	now := time.Now()
	dateFormat := "Sat Jul 26 00:00:00 2025"
	t, err := time.Parse(dateFormat, info.Date)
	if err != nil {
		return o, err
	}
	diff := now.Sub(t)
	if diff.Hours() > 2 {
		return o, errors.New("Ticket can only be changed up to 2 hours after travel time")
	}

	var trip common.Trip
	var trip_resp travel.TripResponse
	if isTripGD(info.OldTripID) {
		trip, trip_resp, err = r.travelService.GetTripAllDetailInfo(ctx, info.TripID, o.From, o.To, info.Date)
	} else {
		trip, trip_resp, err = r.travel2Service.GetTripAllDetailInfo(ctx, info.TripID, o.From, o.To, info.Date)
	}
	if err != nil {
		return o, err
	}
	var ticketPrice float64
	if info.SeatType == seat.FIRSTCLASS {
		if trip_resp.ComfortClass <= 0 {
			return o, errors.New("No more seats available")
		}
		ticketPrice = trip_resp.PriceForComfortClass
	} else if info.SeatType == seat.SECONDCLASS {
		if trip_resp.EconomyClass <= 0 {
			return o, errors.New("No more seats available")
		}
		ticketPrice = trip_resp.PriceForEconomyClass
	}
	if ticketPrice < o.Price {
		difference := fmt.Sprintf("%f", o.Price-ticketPrice)
		_, err = r.insidePaymentService.DrawBack(ctx, info.LoginID, difference)
		if err != nil {
			return o, err
		}
	} else if ticketPrice > o.Price {
		copy_o := order.Order{}
		copy_o.Price = ticketPrice - o.Price
		return copy_o, errors.New("Please pay the difference")
	}
	return r.UpdateOrder(ctx, o, info, trip_resp, ticketPrice, trip)
}

func (r *RebookServiceImpl) PayDifference(ctx context.Context, info RebookInfo) (bool, error) {
	var o order.Order
	var err error
	if isTripGD(info.OldTripID) {
		o, err = r.orderService.GetOrderById(ctx, info.OrderID)
	} else {
		o, err = r.orderOtherService.GetOrderById(ctx, info.OrderID)
	}
	if err != nil {
		return false, err
	}
	if o.Status == order.NotPaid {
		return false, errors.New("Original ticket was not paid. Rebooking not possible.")
	} else if o.Status == order.Change {
		return false, errors.New("Ticket was already changed once. It's cant be changed again.")
	} else if o.Status == order.Collected {
		return false, errors.New("Ticket was already collected. It can't be collected.")
	} else if o.Status != order.Paid {
		return false, errors.New("Ticket can't be changed.")
	}

	now := time.Now()
	dateFormat := "Sat Jul 26 00:00:00 2025"
	t, err := time.Parse(dateFormat, info.Date)
	if err != nil {
		return false, err
	}
	diff := now.Sub(t)
	if diff.Hours() > 2 {
		return false, errors.New("Ticket can only be changed up to 2 hours after travel time")
	}

	var trip common.Trip
	var trip_resp travel.TripResponse
	if isTripGD(info.OldTripID) {
		trip, trip_resp, err = r.travelService.GetTripAllDetailInfo(ctx, info.TripID, o.From, o.To, info.Date)
	} else {
		trip, trip_resp, err = r.travel2Service.GetTripAllDetailInfo(ctx, info.TripID, o.From, o.To, info.Date)
	}
	if err != nil {
		return false, err
	}
	var ticketPrice float64
	if info.SeatType == seat.FIRSTCLASS {
		if trip_resp.ComfortClass <= 0 {
			return false, errors.New("No more seats available")
		}
		ticketPrice = trip_resp.PriceForComfortClass
	} else if info.SeatType == seat.SECONDCLASS {
		if trip_resp.EconomyClass <= 0 {
			return false, errors.New("No more seats available")
		}
		ticketPrice = trip_resp.PriceForEconomyClass
	}
	_, err = r.insidePaymentService.PayDifference(ctx, o.Id, info.LoginID, fmt.Sprintf("%f", ticketPrice-o.Price))
	if err != nil {
		return false, err
	}
	_, err = r.UpdateOrder(ctx, o, info, trip_resp, ticketPrice, trip)
	if err != nil {
		return false, err
	}
	return true, nil
}
