// Package security implements ts-security service from the original Train Ticket application.
package security

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/order"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type SecurityService interface {
	FindAllSecurityConfigs(ctx context.Context) ([]SecurityConfig, error)
	Create(ctx context.Context, name string, value string, description string) (SecurityConfig, error)
	Update(ctx context.Context, sc SecurityConfig) (SecurityConfig, error)
	Delete(ctx context.Context, id string) (bool, error)
	Check(ctx context.Context, accountId string) (bool, error)
}

type SecurityServiceImpl struct {
	db                backend.NoSQLDatabase
	orderService      order.OrderService
	orderOtherService order.OrderService
}

func NewSecurityServiceImpl(ctx context.Context, db backend.NoSQLDatabase, orderService order.OrderService, orderOtherService order.OrderService) (*SecurityServiceImpl, error) {
	return &SecurityServiceImpl{db: db, orderService: orderService, orderOtherService: orderOtherService}, nil
}

func (s *SecurityServiceImpl) FindAllSecurityConfigs(ctx context.Context) ([]SecurityConfig, error) {
	var configs []SecurityConfig
	coll, err := s.db.GetCollection(ctx, "security", "security")
	if err != nil {
		return configs, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return configs, err
	}
	err = res.All(ctx, &configs)
	return configs, err
}

func (s *SecurityServiceImpl) Create(ctx context.Context, name string, value string, description string) (SecurityConfig, error) {
	c := SecurityConfig{}
	coll, err := s.db.GetCollection(ctx, "security", "security")
	if err != nil {
		return c, err
	}
	c.Description = description
	c.Value = value
	c.Name = name
	c.Id = uuid.New().String()
	err = coll.InsertOne(ctx, c)
	if err != nil {
		return SecurityConfig{}, err
	}
	return c, nil
}

func (s *SecurityServiceImpl) Delete(ctx context.Context, id string) (bool, error) {
	coll, err := s.db.GetCollection(ctx, "security", "security")
	if err != nil {
		return false, err
	}
	query := bson.D{{"id", id}}
	err = coll.DeleteOne(ctx, query)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SecurityServiceImpl) Update(ctx context.Context, sc SecurityConfig) (SecurityConfig, error) {
	coll, err := s.db.GetCollection(ctx, "security", "security")
	if err != nil {
		return SecurityConfig{}, err
	}
	query := bson.D{{"id", sc.Id}}
	ok, err := coll.Upsert(ctx, query, sc)
	if err != nil {
		return SecurityConfig{}, err
	}
	if !ok {
		return SecurityConfig{}, errors.New("Failed to update security config")
	}

	return sc, nil
}

func (s *SecurityServiceImpl) Check(ctx context.Context, accountId string) (bool, error) {
	dateFormat := "Sat Jul 26 00:00:00 2025"
	dtNow := time.Now().Format(dateFormat)

	orderResult, err := s.orderService.SecurityInfoCheck(ctx, dtNow, accountId)
	if err != nil {
		return false, err
	}

	orderOtherResult, err := s.orderOtherService.SecurityInfoCheck(ctx, dtNow, accountId)
	if err != nil {
		return false, err
	}

	coll, err := s.db.GetCollection(ctx, "security", "security")
	if err != nil {
		return false, err
	}

	orderInOneHour := orderResult["OrderNumInLastHour"] + orderOtherResult["OrderNumInLastHour"]
	totalValidOrders := orderResult["OrderNumOfValidOrder"] + orderOtherResult["OrderNumOfValidOrder"]

	query := bson.D{{"name", "max_order_1_hour"}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return false, err
	}
	var max_1_hour_conf SecurityConfig
	ok, err := res.One(ctx, &max_1_hour_conf)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, errors.New("Config max_order_1_hour was not found")
	}

	query = bson.D{{"name", "max_order_not_use"}}
	res, err = coll.FindOne(ctx, query)
	if err != nil {
		return false, err
	}
	var max_order_not_use_conf SecurityConfig
	ok, err = res.One(ctx, &max_order_not_use_conf)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, errors.New("Config max_order_1_hour was not found")
	}

	oneHourLine, _ := strconv.ParseUint(max_1_hour_conf.Value, 10, 32)
	totalValidLine, _ := strconv.ParseFloat(max_order_not_use_conf.Value, 32)

	if orderInOneHour > uint16(oneHourLine) || totalValidOrders > uint16(totalValidLine) {
		return false, errors.New("Too many orders in one hour or too many valid orders in total.")
	}

	return true, nil
}
