// Package order implements the SockShop orders microservice.
//
// The service calls other services to collect information and then
// submits the order to the shipping service
package order

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/cart"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/payment"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/shipping"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/user"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	// OrderService is for users to place orders.

	// The service calls other services to collect information and then
	// submits the order to the shipping service
	OrderService interface {
		// Place an order for the specified items
		NewOrder(ctx context.Context, customerID, addressID, cardID, cartID string) (Order, error)

		// Get all orders for a customer, sorted by date
		GetOrders(ctx context.Context, customerID string) ([]Order, error)

		// Get an order by ID
		GetOrder(ctx context.Context, orderID string) (Order, error)
	}

	// A successfully placed order
	Order struct {
		ID         string
		CustomerID string
		Customer   user.User
		Address    user.Address
		Card       user.Card
		Items      []cart.Item
		Shipment   shipping.Shipment
		Date       string
		Total      float32
	}
)

// Creates a new [OrderService] instance.
// Customer, Address, and Card information will be looked up in the provided userService
// Successfully placed orders will be stored in [orderDB]
func NewOrderService(ctx context.Context, userService user.UserService, cartService cart.CartService, payments payment.PaymentService, shipping shipping.ShippingService, orderDB backend.NoSQLDatabase) (OrderService, error) {
	collection, err := orderDB.GetCollection(ctx, "order_service", "orders")
	if err != nil {
		return nil, err
	}
	return &orderImpl{
		users:    userService,
		carts:    cartService,
		payments: payments,
		shipping: shipping,
		db:       collection,
	}, nil
}

type orderImpl struct {
	users    user.UserService
	carts    cart.CartService
	payments payment.PaymentService
	shipping shipping.ShippingService
	db       backend.NoSQLCollection
}

// GetOrder implements OrderService.
func (s *orderImpl) GetOrder(ctx context.Context, orderID string) (Order, error) {
	filter := bson.D{{"id", orderID}}
	cursor, err := s.db.FindOne(ctx, filter)
	if err != nil {
		return Order{}, err
	}
	var order Order
	hasResult, err := cursor.One(ctx, &order)
	if err != nil {
		return Order{}, err
	}
	if !hasResult {
		return Order{}, fmt.Errorf("order %v does not exist", orderID)
	}
	return order, nil
}

// GetOrders implements OrderService.
func (s *orderImpl) GetOrders(ctx context.Context, customerID string) ([]Order, error) {
	filter := bson.D{{"customerid", customerID}}
	cursor, err := s.db.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	var orders []Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

// NewOrder implements OrderService.
func (s *orderImpl) NewOrder(ctx context.Context, customerID, addressID, cardID, cartID string) (Order, error) {
	// All arguments must be provided
	if customerID == "" {
		return Order{}, fmt.Errorf("missing customerID")
	} else if addressID == "" {
		return Order{}, fmt.Errorf("missing addressID")
	} else if cardID == "" {
		return Order{}, fmt.Errorf("missing cardID")
	} else if cartID == "" {
		return Order{}, fmt.Errorf("missing cartID")
	}

	// Fetch data concurrently
	var wg sync.WaitGroup
	wg.Add(4)

	var items []cart.Item
	var users []user.User
	var addresses []user.Address
	var cards []user.Card
	var err1, err2, err3, err4 error

	go func() {
		defer wg.Done()
		items, err1 = s.carts.GetCart(ctx, cartID)
	}()
	go func() {
		defer wg.Done()
		users, err2 = s.users.GetUsers(ctx, customerID)
	}()
	go func() {
		defer wg.Done()
		addresses, err3 = s.users.GetAddresses(ctx, addressID)
	}()
	go func() {
		defer wg.Done()
		cards, err4 = s.users.GetCards(ctx, cardID)
	}()

	// Await completion and validate responses
	wg.Wait()

	if err := any(err1, err2, err3, err4); err != nil {
		return Order{}, err
	}
	if len(items) == 0 {
		return Order{}, fmt.Errorf("no items in cart")
	} else if len(users) == 0 {
		return Order{}, fmt.Errorf("unknown customer %v", customerID)
	} else if len(addresses) == 0 {
		return Order{}, fmt.Errorf("invalid address %v", addressID)
	} else if len(cards) == 0 {
		return Order{}, fmt.Errorf("invalid card %v", cardID)
	}

	// Calculate total and authorize payment.
	amount := calculateTotal(items)
	auth, err := s.payments.Authorise(ctx, amount)
	if err != nil {
		return Order{}, err
	} else if !auth.Authorised {
		return Order{}, fmt.Errorf("payment not authorized due to %v", auth.Message)
	}

	// Submit the shipment
	shipment := shipping.Shipment{
		ID:     uuid.NewString(),
		Name:   customerID,
		Status: "awaiting shipment",
	}
	shipment, err = s.shipping.PostShipping(ctx, shipment)
	if err != nil {
		return Order{}, err
	}

	// Save the order
	order := Order{
		ID:         shipment.ID,
		CustomerID: customerID,
		Address:    addresses[0],
		Card:       cards[0],
		Items:      items,
		Shipment:   shipment,
		Date:       time.Now().String(),
		Total:      amount,
	}
	err = s.db.InsertOne(ctx, order)
	if err != nil {
		return Order{}, err
	}

	// Delete the cart
	return order, s.carts.DeleteCart(ctx, customerID)
}

func calculateTotal(items []cart.Item) float32 {
	amount := float32(0)
	shipping := float32(4.99)
	for _, item := range items {
		amount += float32(item.Quantity) * item.UnitPrice
	}
	amount += shipping
	return amount
}

func any(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
