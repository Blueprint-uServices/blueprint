package insidepayment

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/order"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/payment"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type InsidePaymentService interface {
	Pay(ctx context.Context, tripId string, userId string, orderId string) (string, error)
	CreateAccount(ctx context.Context, money string, userId string) (string, error)
	AddMoney(ctx context.Context, userId string, money string) (string, error)
	QueryPayment(ctx context.Context) ([]payment.Payment, error)
	QueryAccount(ctx context.Context) ([]Balance, error)
	DrawBack(ctx context.Context, userId string, money string) (string, error)
	PayDifference(ctx context.Context, orderId string, userId string, price string) (string, error)
	QueryAddMoney(ctx context.Context) ([]payment.Money, error)
}

type InsidePaymentServiceImpl struct {
	db                backend.NoSQLDatabase
	orderService      order.OrderService
	orderOtherService order.OrderService
	paymentService    payment.PaymentService
}

func NewInsidePaymentServiceImpl(ctx context.Context, db backend.NoSQLDatabase, orderService order.OrderService, orderOtherService order.OrderService, paymentService payment.PaymentService) (*InsidePaymentServiceImpl, error) {
	return &InsidePaymentServiceImpl{db: db, orderService: orderService, orderOtherService: orderOtherService, paymentService: paymentService}, nil
}

func (ipsi *InsidePaymentServiceImpl) Pay(ctx context.Context, tripId string, userId string, orderId string) (string, error) {
	var o order.Order
	var err error
	if tripId[0:1] == "G" || tripId[0:1] == "D" {
		o, err = ipsi.orderService.GetOrderById(ctx, orderId)
	} else {
		o, err = ipsi.orderOtherService.GetOrderById(ctx, orderId)
	}

	newPayment := payment.Payment{
		ID:      uuid.New().String(),
		OrderID: orderId,
		UserID:  userId,
		Price:   fmt.Sprintf("%f", o.Price),
	}

	query := bson.D{{"userid", userId}}
	collection, err := ipsi.db.GetCollection(ctx, "payment", "payment")
	if err != nil {
		return "", err
	}
	res, err := collection.FindMany(ctx, query)
	if err != nil {
		return "", err
	}
	var payments []payment.Payment
	err = res.All(ctx, &payments)
	if err != nil {
		return "", err
	}
	totalExpand := o.Price
	for _, p := range payments {
		price, _ := strconv.ParseFloat(p.Price, 32)
		totalExpand += price
	}

	amCollection, err := ipsi.db.GetCollection(ctx, "payment", "money")
	if err != nil {
		return "", err
	}
	res, err = amCollection.FindMany(ctx, query)
	if err != nil {
		return "", err
	}
	var addMoney []payment.Money
	err = res.All(ctx, &addMoney)
	if err != nil {
		return "", err
	}

	totalMoney := float64(0.0)
	for _, am := range addMoney {
		money, _ := strconv.ParseFloat(am.Price, 64)
		if am.Type == "Add" {
			totalMoney += money
		} else if am.Type == "DrawBack" {
			totalExpand += money
		}
	}

	if totalExpand > totalMoney {
		err = ipsi.paymentService.Pay(ctx, newPayment)
		if err != nil {
			return "", err
		}
		newPayment.Type = "OutsidePayment"
	} else {
		newPayment.Type = "NormalPayment"
	}

	err = collection.InsertOne(ctx, newPayment)
	if err != nil {
		return "", err
	}

	if tripId[0:1] == "G" || tripId[0:1] == "D" {
		_, err = ipsi.orderService.ModifyOrder(ctx, orderId, uint16(order.Paid))
	} else {
		_, err = ipsi.orderOtherService.ModifyOrder(ctx, orderId, uint16(order.Paid))
	}

	if err != nil {
		return "", err
	}

	return "Payment successful", nil
}

func (ipsi *InsidePaymentServiceImpl) CreateAccount(ctx context.Context, money string, userId string) (string, error) {
	collection, err := ipsi.db.GetCollection(ctx, "payment", "money")
	if err != nil {
		return "", err
	}
	query := bson.D{{"userid", userId}}

	res, err := collection.FindOne(ctx, query)
	if err != nil {
		return "", err
	}
	var acc payment.Money
	exists, err := res.One(ctx, &acc)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("Account already exists for user " + userId)
	}

	err = collection.InsertOne(ctx, payment.Money{
		ID:     uuid.New().String(),
		Price:  money,
		UserID: userId,
		Type:   "Add",
	})
	if err != nil {
		return "", err
	}

	return "Account created successfully.", nil
}

func (ipsi *InsidePaymentServiceImpl) AddMoney(ctx context.Context, userId string, money string) (string, error) {
	collection, err := ipsi.db.GetCollection(ctx, "payment", "money")
	if err != nil {
		return "", err
	}
	query := bson.D{{"userid", userId}}
	res, err := collection.FindOne(ctx, query)
	if err != nil {
		return "", err
	}
	var existing_acc payment.Money
	ok, err := res.One(ctx, &existing_acc)
	if err != nil {
		return "", err
	}
	if ok {
		return "", errors.New("Failed to add money as account doesn't exist for user " + userId)
	}
	acc := payment.Money{ID: uuid.New().String(), Price: money, UserID: userId, Type: "Add"}
	err = collection.InsertOne(ctx, acc)
	if err != nil {
		return "", err
	}
	return "Add money success", nil
}

func (ipsi *InsidePaymentServiceImpl) QueryPayment(ctx context.Context) ([]payment.Payment, error) {
	var payments []payment.Payment
	coll, err := ipsi.db.GetCollection(ctx, "payment", "payment")
	if err != nil {
		return payments, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return payments, err
	}
	err = res.All(ctx, &payments)
	return payments, err
}

func (ipsi *InsidePaymentServiceImpl) QueryAccount(ctx context.Context) ([]Balance, error) {
	var result []Balance
	coll, err := ipsi.db.GetCollection(ctx, "payment", "money")
	if err != nil {
		return result, err
	}
	var all_money []payment.Money
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return result, err
	}
	err = res.All(ctx, &all_money)
	if err != nil {
		return result, err
	}

	user_accounts := make(map[string]float64)
	for _, money := range all_money {
		val, err := strconv.ParseFloat(money.Price, 64)
		if err != nil {
			return result, err
		}
		if v, ok := user_accounts[money.UserID]; ok {
			if money.Type == "Add" {
				user_accounts[money.UserID] = v + val
			} else if money.Type == "DrawBack" {
				user_accounts[money.UserID] = v - val
			}
		} else {
			user_accounts[money.UserID] = val
		}
	}

	all_payments, err := ipsi.QueryPayment(ctx)
	if err != nil {
		return result, err
	}

	for _, p := range all_payments {
		val, err := strconv.ParseFloat(p.Price, 64)
		if err != nil {
			return result, err
		}
		if v, ok := user_accounts[p.UserID]; ok {
			user_accounts[p.UserID] = v - val
		} else {
			user_accounts[p.UserID] = -val
		}
	}

	for user, state := range user_accounts {
		b := Balance{UserId: user, Amount: state}
		result = append(result, b)
	}

	return result, nil
}

func (ipsi *InsidePaymentServiceImpl) DrawBack(ctx context.Context, userId string, money string) (string, error) {
	collection, err := ipsi.db.GetCollection(ctx, "payment", "money")
	if err != nil {
		return "", err
	}
	query := bson.D{{"userid", userId}}
	res, err := collection.FindOne(ctx, query)
	if err != nil {
		return "", err
	}
	var existing_acc payment.Money
	ok, err := res.One(ctx, &existing_acc)
	if err != nil {
		return "", err
	}
	if ok {
		return "", errors.New("Failed to drawback money as account doesn't exist for user " + userId)
	}
	acc := payment.Money{ID: uuid.New().String(), Price: money, UserID: userId, Type: "DrawBack"}
	err = collection.InsertOne(ctx, acc)
	if err != nil {
		return "", err
	}
	return "Drawback money success", nil
}

func (ipsi *InsidePaymentServiceImpl) QueryAddMoney(ctx context.Context) ([]payment.Money, error) {
	var moneys []payment.Money
	coll, err := ipsi.db.GetCollection(ctx, "payment", "money")
	if err != nil {
		return moneys, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return moneys, err
	}
	err = res.All(ctx, &moneys)
	return moneys, err
}

func (ipsi *InsidePaymentServiceImpl) PayDifference(ctx context.Context, orderId string, userId string, price string) (string, error) {
	newPayment := payment.Payment{
		ID:      uuid.New().String(),
		OrderID: orderId,
		UserID:  userId,
		Price:   price,
	}
	price_parsed, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return "", err
	}

	query := bson.D{{"userid", userId}}
	collection, err := ipsi.db.GetCollection(ctx, "payment", "payment")
	if err != nil {
		return "", err
	}
	res, err := collection.FindMany(ctx, query)
	if err != nil {
		return "", err
	}
	var payments []payment.Payment
	err = res.All(ctx, &payments)
	if err != nil {
		return "", err
	}
	totalExpand := price_parsed
	for _, p := range payments {
		price, _ := strconv.ParseFloat(p.Price, 32)
		totalExpand += price
	}

	amCollection, err := ipsi.db.GetCollection(ctx, "payment", "money")
	if err != nil {
		return "", err
	}
	res, err = amCollection.FindMany(ctx, query)
	if err != nil {
		return "", err
	}
	var addMoney []payment.Money
	err = res.All(ctx, &addMoney)
	if err != nil {
		return "", err
	}

	totalMoney := float64(0.0)
	for _, am := range addMoney {
		money, _ := strconv.ParseFloat(am.Price, 64)
		if am.Type == "Add" {
			totalMoney += money
		} else if am.Type == "DrawBack" {
			totalExpand += money
		}
	}

	if totalExpand > totalMoney {
		err = ipsi.paymentService.Pay(ctx, newPayment)
		if err != nil {
			return "", err
		}
		newPayment.Type = "OutsidePayment"
	} else {
		newPayment.Type = "NormalPayment"
	}

	err = collection.InsertOne(ctx, newPayment)
	if err != nil {
		return "", err
	}
	return "Payment success", nil
}
