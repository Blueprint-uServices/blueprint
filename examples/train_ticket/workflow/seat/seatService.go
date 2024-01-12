// Package seat implements ts-seat-service from the original TrainTicket application
package seat

import (
	"context"
	"math/rand"
	"strconv"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/config"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/order"
	"golang.org/x/exp/slices"
)

type SeatService interface {
	// Selects a seat number to issue a ticket based on the provided info
	DistributeSeat(ctx context.Context, s Seat) (common.Ticket, error)
	// Returns the number of tickets left
	GetLeftTicketOfInterval(ctx context.Context, s Seat) (int64, error)
}

type SeatServiceImpl struct {
	configService     config.ConfigService
	orderService      order.OrderService
	orderOtherService order.OrderService
}

func NewSeatServiceImpl(ctx context.Context, configService config.ConfigService, orderService order.OrderService, orderOtherService order.OrderService) (*SeatServiceImpl, error) {
	return &SeatServiceImpl{configService: configService, orderService: orderService, orderOtherService: orderOtherService}, nil
}

func (s *SeatServiceImpl) DistributeSeat(ctx context.Context, st Seat) (common.Ticket, error) {
	var purchased_tickets []common.Ticket
	var err error
	if st.TrainNumber[0:1] == "G" || st.TrainNumber[0:1] == "D" {
		purchased_tickets, err = s.orderService.GetAllSoldTickets(ctx, st.TravelDate, st.TrainNumber)
	} else {
		purchased_tickets, err = s.orderOtherService.GetAllSoldTickets(ctx, st.TravelDate, st.TrainNumber)
	}
	if err != nil {
		return common.Ticket{}, err
	}
	var ticket common.Ticket
	ticket.StartStation = st.StartStation
	ticket.DestStation = st.DstStation
	seat_range := int64(len(purchased_tickets))
	seat_num := rand.Int63n(seat_range) + 1

	// Check if we can allocate a seat from one of the previously sold tickets
	all_allocated_seats := make(map[int64]bool)
	for _, t := range purchased_tickets {
		all_allocated_seats[t.SeatNo] = true
		if slices.Index(st.Stations, t.DestStation) <= slices.Index(st.Stations, t.StartStation) {
			ticket.SeatNo = t.SeatNo
			return ticket, nil
		}
	}
	for {
		if _, ok := all_allocated_seats[seat_num]; !ok {
			ticket.SeatNo = seat_num
			break
		} else {
			seat_num = rand.Int63n(seat_range) + 1
		}
	}
	return ticket, nil
}

func (s *SeatServiceImpl) GetLeftTicketOfInterval(ctx context.Context, st Seat) (int64, error) {
	var sold order.SoldTicket
	var err error
	if st.TrainNumber[0:1] == "G" || st.TrainNumber[0:1] == "D" {
		sold, err = s.orderService.CalculateSoldTicket(ctx, st.TravelDate, st.TrainNumber)
	} else {
		sold, err = s.orderOtherService.CalculateSoldTicket(ctx, st.TravelDate, st.TrainNumber)
	}
	if err != nil {
		return -1, err
	}
	sold_tickets := int64(sold.NoSeat + sold.BusinessSeat + sold.FirstClassSeat + sold.SecondClassSeat + sold.HardSeat + sold.SoftSeat + sold.HardBed + sold.SoftBed + sold.HighSoftBed)

	direct_proportion, err := s.configService.Find(ctx, "DirectTicketAllocationProportion")
	if err != nil {
		return -1, err
	}
	dir_prop, _ := strconv.ParseFloat(direct_proportion.Value, 64)
	if st.StartStation != st.Stations[0] || st.DstStation != st.Stations[len(st.Stations)-1] {
		dir_prop = 1.0 - dir_prop
	}
	unsold_tickets := int64(float64(st.TotalNum)*dir_prop) - sold_tickets

	return unsold_tickets, nil
}
