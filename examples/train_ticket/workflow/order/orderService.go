// Package order implements ts-order and ts-orderOther services from the original TrainTicket applications
package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/station"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type OrderService interface {
	GetTicketListByDateAndTripId(ctx context.Context, travelDate string, trainNumber string) ([]Ticket, error)
	CreateNewOrder(ctx context.Context, order Order) (Order, error)
	AddCreateNewOrder(ctx context.Context, order Order) (Order, error)
	QueryOrders(ctx context.Context, orderInfo OrderInfo, accountId string) ([]Order, error)
	QueryOrdersForRefresh(ctx context.Context, orderInfo OrderInfo, accountId string) ([]Order, error)
	CalculateSoldTicket(ctx context.Context, travelDate string, trainNumber string) (SoldTicket, error)
	GetOrderPrice(ctx context.Context, orderId string) (float32, error)
	PayOrder(ctx context.Context, orderId string) (Order, error)
	GetOrderById(ctx context.Context, orderId string) (Order, error)
	ModifyOrder(ctx context.Context, orderId string, status uint16) (Order, error)
	SecurityInfoCheck(ctx context.Context, checkDate string, accountId string) (map[string]uint16, error)
	SaveOrderInfo(ctx context.Context, order Order) (Order, error)
	UpdateOrder(ctx context.Context, order Order) (Order, error)
	DeleteOrder(ctx context.Context, orderId string) (string, error)
	FindAllOrder(ctx context.Context) ([]Order, error)
}

type OrderServiceImpl struct {
	db             backend.NoSQLDatabase
	stationService station.StationService
}

func NewOrderService(ctx context.Context, stationService station.StationService, db backend.NoSQLDatabase) (*OrderServiceImpl, error) {
	return &OrderServiceImpl{db: db, stationService: stationService}, nil
}

func (osi *OrderServiceImpl) GetTicketListByDateAndTripId(ctx context.Context, travelDate string, trainNumber string) ([]Ticket, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"TravelDate": %s, "TrainNumber": %s}`, travelDate, trainNumber)
	res, err := collection.FindMany(query)
	if err != nil {
		return nil, err
	}

	var orders []Order
	var tickets []Ticket

	err = res.All(&orders)
	if err != nil {
		return nil, err
	}

	for _, order := range orders {
		tickets = append(tickets, Ticket{
			SeatNo:       order.SeatNumber,
			StartStation: order.From,
			DestStation:  order.To,
		})
	}

	return tickets, nil
}

func (osi *OrderServiceImpl) CreateNewOrder(ctx context.Context, order Order) (Order, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"AccountId": %s}`, order.AccountId)
	res, err := collection.FindOne(query)
	if err == nil {
		var exOrder Order
		res.Decode(&exOrder)
		if exOrder.Id != "" {
			return Order{}, errors.New("Order already exists for this account.")
		}
	}

	order.Id = uuid.New().String()
	err = collection.InsertOne(order)
	if err != nil {
		return Order{}, nil
	}

	return order, nil
}

func (osi *OrderServiceImpl) AddCreateNewOrder(ctx context.Context, order Order) (Order, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"AccountId": %s}`, order.AccountId)
	res, err := collection.FindOne(query)
	if err == nil {
		var exOrder Order
		res.Decode(&exOrder)
		if exOrder.Id != "" {
			return Order{}, errors.New("Order already exists for this account.")
		}
	}

	order.Id = uuid.New().String()
	err = collection.InsertOne(order)
	if err != nil {
		return Order{}, nil
	}

	return order, nil
}

func (osi *OrderServiceImpl) QueryOrders(ctx context.Context, orderInfo OrderInfo, accountId string) ([]Order, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"AccountId": %s}`, accountId)
	res, err := collection.FindMany(query)
	if err != nil {
		return nil, err
	}

	var orderList []Order
	err = res.All(&orderList)
	if err != nil {
		return nil, err
	}

	var finalList []Order

	if orderInfo.EnableTravelDateQuery || orderInfo.EnableBoughtDateQuery || orderInfo.EnableStateQuery {

		statePassFlag := false
		travelDatePassFlag := false
		boughtDatePassFlag := false

		for _, order := range orderList {

			if orderInfo.EnableStateQuery {
				if order.Status == orderInfo.State {
					statePassFlag = true
				}
			}

			if orderInfo.EnableTravelDateQuery {
				t1, _ := time.Parse(time.ANSIC, order.TravelDate)
				t2, _ := time.Parse(time.ANSIC, orderInfo.TravelDateEnd)
				t3, _ := time.Parse(time.ANSIC, order.TravelDate)
				t4, _ := time.Parse(time.ANSIC, orderInfo.TravelDateStart)

				if t1.Before(t2) && t3.Before(t4) {
					travelDatePassFlag = true
				}
			}

			if orderInfo.EnableBoughtDateQuery {
				t1, _ := time.Parse(time.ANSIC, order.BoughtDate)
				t2, _ := time.Parse(time.ANSIC, orderInfo.BoughtDateEnd)
				t3, _ := time.Parse(time.ANSIC, order.BoughtDate)
				t4, _ := time.Parse(time.ANSIC, orderInfo.BoughtDateStart)

				if t1.Before(t2) && t3.Before(t4) {
					travelDatePassFlag = true
				}
			}

			if statePassFlag && travelDatePassFlag && boughtDatePassFlag {
				finalList = append(finalList, order)
			}
		}
	} else {
		for _, order := range orderList {
			finalList = append(finalList, order)
		}
	}

	return finalList, nil
}

func (osi *OrderServiceImpl) QueryOrdersForRefresh(ctx context.Context, orderInfo OrderInfo, accountId string) ([]Order, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"AccountId": %s}`, accountId)
	res, err := collection.FindMany(query)
	if err != nil {
		return []Order{}, err
	}

	var orders []Order
	err = res.All(&orders)
	if err != nil {
		return orders, err
	}

	var stationIds []string
	for _, order := range orders {
		stationIds = append(stationIds, order.From)
		stationIds = append(stationIds, order.To)
	}

	names, err := osi.stationService.QueryForNameBatch(ctx, stationIds, token)
	if err != nil {
		return orders, err
	}

	for idx, _ := range names {
		orders[idx].From = names[idx*2]
		orders[idx].To = names[idx*2+1]
	}

	return orders, nil
}

func (osi *OrderServiceImpl) CalculateSoldTicket(ctx context.Context, travelDate string, trainNumber string) (SoldTicket, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"TravelDate": %s, "TrainNumber": %s}`, travelDate, trainNumber)
	res, err := collection.FindMany(query)
	if err != nil {
		return SoldTicket{}, err
	}

	var orders []Order
	err = res.All(&orders)
	if err != nil {
		return SoldTicket{}, err
	}

	soldTicket := SoldTicket{}

	for _, order := range orders {
		if order.Status == uint16(Change) {
			continue
		}

		switch order.SeatClass {
		case None:
			soldTicket.NoSeat += 1
		case Business:
			soldTicket.BusinessSeat += 1
		case FirstClass:
			soldTicket.FirstClassSeat += 1
		case SecondClass:
			soldTicket.SecondClassSeat += 1
		case HardSeat:
			soldTicket.HardSeat += 1
		case SoftSeat:
			soldTicket.SoftSeat += 1
		case HardBed:
			soldTicket.HardBed += 1
		case SoftBed:
			soldTicket.SoftBed += 1
		case HighSoftBed:
			soldTicket.HighSoftBed += 1

		default:
			continue
		}
	}

	return soldTicket, nil
}

func (osi *OrderServiceImpl) GetOrderPrice(ctx context.Context, orderId string) (float32, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"Id": %s}`, orderId)
	res, err := collection.FindOne(query)
	if err != nil {
		return 0.0, err
	}

	var order Order
	err = res.Decode(&order)
	if err != nil {
		return 0.0, err
	}

	return order.Price, nil
}

func (osi *OrderServiceImpl) PayOrder(ctx context.Context, orderId string) (Order, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"Id": %s}`, orderId)
	res, err := collection.FindOne(query)
	if err != nil {
		return Order{}, err
	}
	var order Order
	res.Decode(&order)
	update := fmt.Sprintf(`{"$set": {"Status": %d}}`, Paid)
	err = collection.UpdateOne(query, update)
	if err != nil {
		return Order{}, err
	}

	order.Status = uint16(Paid)
	return order, nil
}

func (osi *OrderServiceImpl) GetOrderById(ctx context.Context, orderId string) (Order, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"Id": %s}`, orderId)
	res, err := collection.FindOne(query)
	if err != nil {
		return Order{}, err
	}

	var order Order
	err = res.Decode(&order)
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

func (osi *OrderServiceImpl) ModifyOrder(ctx context.Context, orderId string, status uint16) (Order, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"Id": %s}`, orderId)
	res, err := collection.FindOne(query)
	if err != nil {
		return Order{}, err
	}
	var order Order
	res.Decode(&order)
	update := fmt.Sprintf(`{"$set": {"Status": %d}}`, status)
	err = collection.UpdateOne(query, update)
	if err != nil {
		return Order{}, err
	}

	order.Status = status
	return order, nil
}

func (osi *OrderServiceImpl) SecurityInfoCheck(ctx context.Context, checkDate string, accountId string) (map[string]uint16, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"AccountId": %s}`, accountId)
	res, err := collection.FindMany(query) //TODO verify this query works
	if err != nil {
		return nil, err
	}

	var orders []Order
	res.All(&orders)
	countTotalValidOrder := uint16(0)
	countOrderInOneHour := uint16(0)

	dateFrom, _ := time.Parse(time.ANSIC, checkDate)

	for _, order := range orders {

		if order.Status == uint16(NotPaid) || order.Status == uint16(Paid) || order.Status == uint16(Collected) {
			countTotalValidOrder += 1
		}

		t1, _ := time.Parse(time.ANSIC, order.BoughtDate)

		if t1.After(dateFrom) {
			countOrderInOneHour += 1
		}
	}

	return map[string]uint16{
		"OrderNumInLastHour":   countOrderInOneHour,
		"OrderNumOfValidOrder": countTotalValidOrder,
	}, nil
}

func (osi *OrderServiceImpl) SaveOrderInfo(ctx context.Context, order Order) (Order, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"Id": %s}`, order.Id)
	_, err = collection.FindOne(query)
	if err != nil {
		return Order{}, err
	}

	err = collection.ReplaceOne(query, order)
	if err != nil {
		return Order{}, nil
	}

	return order, nil
}

func (osi *OrderServiceImpl) UpdateOrder(ctx context.Context, order Order) (Order, error) {
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := fmt.Sprintf(`{"Id": %s}`, order.Id)
	_, err = collection.FindOne(query)
	if err != nil {
		return Order{}, err
	}

	err = collection.ReplaceOne(query, order)
	if err != nil {
		return Order{}, nil
	}

	return order, nil
}

func (osi *OrderServiceImpl) DeleteOrder(ctx context.Context, orderId string) (string, error) {

	collection, err := osi.db.GetCollection(ctx, "orders", "orders")

	query := bson.D{{"id", orderId}}
	err = collection.DeleteOne(ctx, query)
	if err != nil {
		return "", err
	}

	return "Order deleted.", nil
}

func (osi *OrderServiceImpl) FindAllOrder(ctx context.Context) ([]Order, error) {
	var orders []Order
	collection, err := osi.db.GetCollection(ctx, "orders", "orders")
	if err != nil {
		return orders, err
	}

	res, err := collection.FindMany(ctx, bson.D{})
	if err != nil {
		return orders, err
	}

	err = res.All(ctx, &orders)
	if err != nil {
		return orders, err
	}

	return orders, nil
}
