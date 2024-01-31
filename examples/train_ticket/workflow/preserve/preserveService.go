// Package preserve implements ts-preserve-service from the original Train Ticket application
package preserve

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/assurance"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/basic"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/consign"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/contacts"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/food"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/order"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/seat"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/security"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/station"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/travel"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/user"
	"github.com/google/uuid"
)

type PreserveService interface {
	Preserve(ctx context.Context, oti common.OrderTicketsInfo) (string, error)
}

type PreserveServiceImpl struct {
	assuranceService assurance.AssuranceService
	basicService     basic.BasicService
	consignService   consign.ConsignService
	contactService   contacts.ContactsService
	foodService      food.FoodService
	orderService     order.OrderService
	seatService      seat.SeatService
	securityService  security.SecurityService
	stationService   station.StationService
	travelService    travel.TravelService
	userService      user.UserService
}

func NewPreserveServiceImpl(ctx context.Context, assuranceService assurance.AssuranceService, basicService basic.BasicService, consignService consign.ConsignService, contactsService contacts.ContactsService, foodService food.FoodService, orderService order.OrderService, seatService seat.SeatService, stationService station.StationService, securityService security.SecurityService, travelService travel.TravelService, userService user.UserService) (*PreserveServiceImpl, error) {
	return &PreserveServiceImpl{assuranceService, basicService, consignService, contactsService, foodService, orderService, seatService, securityService, stationService, travelService, userService}, nil
}

func (p *PreserveServiceImpl) Preserve(ctx context.Context, oti common.OrderTicketsInfo) (string, error) {
	// Check Security
	ok, err := p.securityService.Check(ctx, oti.AccountID)
	if err != nil {
		return "Check Failure!", err
	}
	if !ok {
		return "Check Failure!", errors.New("Security check failure!")
	}

	// Get contacts
	c, err := p.contactService.FindContactsById(ctx, oti.ContactID)
	if err != nil {
		return "", errors.New("Can't find contact")
	}

	// Get trip information
	tr, _, err := p.travelService.GetTripAllDetailInfo(ctx, oti.TripID, oti.From, oti.To, oti.Date)
	if err != nil {
		return "", errors.New("Failed to get trip information")
	}

	o := order.Order{}
	o.Id = uuid.New().String()
	o.AccountId = oti.AccountID
	o.TrainNumber = oti.TripID
	o.From = oti.From
	o.To = oti.To
	o.BoughtDate = time.Now().Format("Sat Jul 26 00:00:00 2025")
	o.Status = order.NotPaid
	o.ContactsDocumentNumber = c.DocumentNumber
	o.DocumentType = uint16(c.DocumentType)
	o.ContactsName = c.Name

	var travel common.Travel
	travel.T = tr
	travel.StartPlace = oti.From
	travel.EndPlace = oti.To
	travel.DepartureTime = oti.Date

	travel_res, err := p.basicService.QueryForTravel(ctx, travel)
	if err != nil {
		return "", err
	}
	o.SeatClass = uint16(oti.SeatType)
	o.TravelDate = oti.Date

	stations := travel_res.Route.Stations
	var totalNum int64
	var price_str string
	if oti.SeatType == seat.FIRSTCLASS {
		totalNum = travel_res.TType.ComfortClass
		price_str = travel_res.Prices["comfortClass"]
	} else if oti.SeatType == seat.SECONDCLASS {
		totalNum = travel_res.TType.EconomyClass
		price_str = travel_res.Prices["economyClass"]
	}
	var s seat.Seat
	s.Stations = stations
	s.StartStation = oti.From
	s.DstStation = oti.To
	s.TotalNum = totalNum
	s.SeatType = int(oti.SeatType)
	s.TrainNumber = o.TrainNumber
	s.TravelDate = oti.Date
	ticket, err := p.seatService.DistributeSeat(ctx, s)
	if err != nil {
		return "", err
	}
	o.SeatNumber = fmt.Sprintf("%d", ticket.SeatNo)
	price, err := strconv.ParseFloat(price_str, 64)
	if err != nil {
		return "", err
	}
	o.Price = price

	o, err = p.orderService.AddCreateNewOrder(ctx, o)
	if err != nil {
		return "", err
	}

	if oti.Assurance != 0 {
		// Don't return error as original app doesn't return error
		_, err := p.assuranceService.Create(ctx, oti.Assurance, o.Id)
		if err != nil {
			return "Success but no assurance", nil
		}
	}
	if oti.FoodType != 0 {
		var fo food.FoodOrder
		fo.OrderID = o.Id
		fo.FoodType = oti.FoodType
		fo.FoodName = oti.FoodName
		fo.Price = oti.FoodPrice
		if oti.FoodType == 2 {
			fo.StationName = oti.StationName
			fo.StoreName = oti.StoreName
		}
		_, err := p.foodService.CreateFoodOrder(ctx, fo)
		if err != nil {
			return "Success but no food", nil
		}
	}

	if oti.ConsigneeName != "" {
		c := consign.Consign{}
		c.Consignee = oti.ConsigneeName
		c.AccountId = o.AccountId
		c.HandleDate = oti.HandleDate
		c.TargetDate = o.TravelDate
		c.From = o.From
		c.To = o.To
		c.OrderId = o.Id
		c.Phone = oti.ConsigneePhone
		c.Within = oti.WithinRegion
		c.Weight = oti.ConsigneeWeight
		_, err := p.consignService.InsertConsign(ctx, c)
		if err != nil {
			return "Success but consign failed", nil
		}
	}

	// User notification is disabled in the original app but preserving the call
	_, err = p.userService.FindByUserID(ctx, oti.AccountID)
	if err != nil {
		return "Couldn't find user for notification", nil
	}

	return "Success!", nil
}
