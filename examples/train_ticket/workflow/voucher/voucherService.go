// Package voucher implements ts-voucher-service from the original Train Ticket application
package voucher

import (
	"context"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/order"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

var TABLE_QUERY = `CREATE TABLE IF NOT EXISTS voucher (
	voucher_id INT NOT NULL AUTO_INCREMENT,
	order_id VARCHAR(1024) NOT NULL,
	travelDate DATE NOT NULL,
	travelTime VARCHAR(1024) NOT NULL,
	contactName VARCHAR(1024) NOT NULL,
	trainNumber VARCHAR(1024) NOT NULL,
	seatClass INT NOT NULL,
	seatNumber VARCHAR(1024) NOT NULL,
	startStation VARCHAR(1024) NOT NULL,
	destStation VARCHAR(1024) NOT NULL,
	price FLOAT NOT NULL,
	PRIMARY KEY (voucher_id));`

type VoucherService interface {
	// Add a new Voucher to the order
	Post(ctx context.Context, orderId string, typ string) (Voucher, error)
	// Get the voucher applied to the order
	GetVoucher(ctx context.Context, orderId string) (Voucher, error)
}

type VoucherServiceImpl struct {
	db                backend.RelationalDB
	orderService      order.OrderService
	orderOtherService order.OrderService
}

func NewVoucherServiceImpl(ctx context.Context, db backend.RelationalDB, orderService order.OrderService, orderOtherService order.OrderService) (*VoucherServiceImpl, error) {
	return &VoucherServiceImpl{db, orderService, orderOtherService}, nil
}

func (vsi *VoucherServiceImpl) Post(ctx context.Context, orderId string, typ string, token string) (Voucher, error) {

	findQuery := "SELECT * FROM voucher where order_id = ? LIMIT 1"
	res, err := vsi.db.Query(ctx, findQuery, orderId)
	if err != nil {
		return Voucher{}, err
	}

	var v Voucher

	//* only get one voucher
	if res.Next() {
		// Voucher already exists!
		res.Scan(&v.VoucherId, &v.OrderId, &v.TravelDate, &v.ContactName, &v.TrainNumber, &v.SeatClass, &v.SeatNumber, &v.StartStation, &v.DestStation, &v.Price)
		return v, nil
	}

	//* Insert

	var o order.Order
	if typ == "0" {
		o, err = vsi.orderService.GetOrderById(ctx, orderId)
	} else {
		o, err = vsi.orderOtherService.GetOrderById(ctx, orderId)
	}

	query := "INSERT INTO voucher (order_id,travelDate,contactName,trainNumber,seatClass,seatNumber,startStation,destStation,price)VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?);"
	v.OrderId = o.Id
	v.TravelDate = o.TravelDate
	v.ContactName = o.ContactsName
	v.TrainNumber = o.TrainNumber
	v.SeatClass = o.SeatClass
	v.SeatNumber = o.SeatNumber
	v.StartStation = o.From
	v.DestStation = o.To
	v.Price = o.Price

	// Ignore result as not every driver might support populating the result with meaningful values.
	_, err = vsi.db.Exec(ctx, query, o.Id, o.TravelDate, o.ContactsName, o.TrainNumber, o.SeatClass, o.SeatNumber, o.From, o.To, o.Price)
	if err != nil {
		return Voucher{}, err
	}

	return v, nil
}

func (vsi *VoucherServiceImpl) GetVoucher(ctx context.Context, orderId, token string) (Voucher, error) {

	findQuery := "SELECT * FROM voucher where order_id = ? LIMIT 1"

	res, err := vsi.db.Query(ctx, findQuery, orderId)
	if err != nil {
		return Voucher{}, err
	}

	var v Voucher

	if res.Next() {
		res.Scan(&v.VoucherId, &v.OrderId, &v.TravelDate, &v.ContactName, &v.TrainNumber, &v.SeatClass, &v.SeatNumber, &v.StartStation, &v.DestStation, &v.Price)
		return v, nil
	}

	return Voucher{}, nil
}
