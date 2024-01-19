// Package waitlistorder implements ts-wait-order-service from the original Train Ticket implementation
package waitlistorder

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type WaitlistOrderService interface {
	FindOrderById(ctx context.Context, id string) (WaitlistOrder, error)
	Create(ctx context.Context, vo WaitlistOrderVO) (WaitlistOrder, error)
	GetAllOrders(ctx context.Context) ([]WaitlistOrder, error)
	UpdateOrder(ctx context.Context, o WaitlistOrder) error
	ModifyWaitlistOrderStatus(ctx context.Context, orderId string, status int) error
	GetAllWaitListOrders(ctx context.Context) ([]WaitlistOrder, error)
}

type WaitlistOrderServiceImpl struct {
	db backend.NoSQLDatabase
}

func NewWaitlistOrderServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*WaitlistOrderServiceImpl, error) {
	return &WaitlistOrderServiceImpl{db}, nil
}

func (w *WaitlistOrderServiceImpl) GetAllOrders(ctx context.Context) ([]WaitlistOrder, error) {
	var response []WaitlistOrder
	coll, err := w.db.GetCollection(ctx, "waitlist", "waitlist")
	if err != nil {
		return response, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return response, err
	}
	err = res.All(ctx, &response)
	return response, err
}

func (w *WaitlistOrderServiceImpl) GetAllWaitListOrders(ctx context.Context) ([]WaitlistOrder, error) {
	orders, err := w.GetAllOrders(ctx)
	if err != nil {
		return orders, err
	}
	var filtered_orderes []WaitlistOrder
	for _, o := range orders {
		if o.Status == PAID || o.Status == NOTPAID {
			filtered_orderes = append(filtered_orderes, o)
		}
	}
	return filtered_orderes, nil
}

func (w *WaitlistOrderServiceImpl) FindOrderById(ctx context.Context, id string) (WaitlistOrder, error) {
	var response WaitlistOrder
	coll, err := w.db.GetCollection(ctx, "waitlist", "waitlist")
	if err != nil {
		return response, err
	}
	res, err := coll.FindOne(ctx, bson.D{{"id", id}})
	if err != nil {
		return response, err
	}
	ok, err := res.One(ctx, &response)
	if err != nil {
		return response, err
	}
	if !ok {
		return response, errors.New("Order with ID " + id + " does not exist")
	}
	return response, nil
}

func (w *WaitlistOrderServiceImpl) Create(ctx context.Context, vo WaitlistOrderVO) (WaitlistOrder, error) {
	coll, err := w.db.GetCollection(ctx, "waitlist", "waitlist")
	if err != nil {
		return WaitlistOrder{}, err
	}
	query := bson.D{{"$and", bson.A{
		bson.D{{"accountid", vo.AccountID}},
		bson.D{{"seattype", vo.SeatType}},
		bson.D{{"trainnumber", vo.TripID}},
		bson.D{{"traveltime", vo.Date}},
		bson.D{{"from", vo.From}},
		bson.D{{"to", vo.To}},
		bson.D{{"contactid", vo.ContactID}},
	}}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return WaitlistOrder{}, err
	}
	var o WaitlistOrder
	ok, err := res.One(ctx, &o)
	if err != nil {
		return WaitlistOrder{}, err
	}
	if ok {
		return o, nil
	}
	o.From = vo.From
	o.To = vo.To
	o.Price = vo.Price
	o.ContactID = vo.ContactID
	o.AccountID = vo.AccountID
	o.SeatType = vo.SeatType
	o.TravelTime = vo.Date
	return o, coll.InsertOne(ctx, o)
}

func (w *WaitlistOrderServiceImpl) UpdateOrder(ctx context.Context, o WaitlistOrder) error {
	coll, err := w.db.GetCollection(ctx, "waitlist", "waitlist")
	if err != nil {
		return err
	}
	ok, err := coll.Upsert(ctx, bson.D{{"id", o.ID}}, o)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Failed to update order")
	}
	return nil
}

func (w *WaitlistOrderServiceImpl) ModifyWaitlistOrderStatus(ctx context.Context, id string, status int) error {
	coll, err := w.db.GetCollection(ctx, "waitlist", "waitlist")
	if err != nil {
		return err
	}
	query := bson.D{{"id", id}}
	update := bson.D{{"$set", bson.D{{"status", status}}}}
	n, err := coll.UpdateOne(ctx, query, update)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("Failed to update waitlistorder")
	}
	return nil
}
