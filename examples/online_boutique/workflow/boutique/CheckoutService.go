package boutique

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type PlaceOrderRequest struct {
	UserID       string
	UserCurrency string
	Address      Address
	Email        string
	CreditCard   CreditCardInfo
}

type CheckoutService interface {
	PlaceOrder(ctx context.Context, req PlaceOrderRequest) (Order, error)
}

type CheckoutServiceImpl struct {
	catalogService  ProductCatalogService
	cartService     CartService
	currencyService CurrencyService
	shippingService ShippingService
	emailService    EmailService
	paymentService  PaymentService
	logger          backend.Logger
}

func NewCheckoutServiceImpl(
	ctx context.Context,
	catalogService ProductCatalogService,
	cartService CartService,
	currencyService CurrencyService,
	shippingService ShippingService,
	emailService EmailService,
	paymentService PaymentService,
) (CheckoutService, error) {
	return &CheckoutServiceImpl{
		catalogService: catalogService, cartService: cartService, currencyService: currencyService,
		shippingService: shippingService, emailService: emailService, paymentService: paymentService,
		logger: backend.GetLogger(),
	}, nil
}

func (s *CheckoutServiceImpl) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (Order, error) {
	ctx, _ = s.logger.Info(ctx, "PlaceOrder: user_id=%s user_currency=%s", req.UserID, req.UserCurrency)
	prep, err := s.prepareOrder(ctx, req.UserID, req.UserCurrency, req.Address)
	if err != nil {
		return Order{}, err
	}
	total := Money{CurrencyCode: req.UserCurrency}
	if total, err = Sum(total, prep.shippingCost); err != nil {
		return Order{}, err
	}
	for _, item := range prep.orderItems {
		if item.Item.Quantity < 0 {
			return Order{}, fmt.Errorf("invalid item quantity %d", item.Item.Quantity)
		}
		total, err = Sum(total, MultiplySlow(item.Cost, uint32(item.Item.Quantity)))
		if err != nil {
			return Order{}, err
		}
	}
	transactionID, err := s.paymentService.Charge(ctx, total, req.CreditCard)
	if err != nil {
		return Order{}, fmt.Errorf("failed to charge card: %w", err)
	}
	ctx, _ = s.logger.Info(ctx, "payment went through: transaction_id=%s", transactionID)
	trackingID, err := s.shippingService.ShipOrder(ctx, req.Address, prep.cartItems)
	if err != nil {
		return Order{}, fmt.Errorf("shipping error: %w", err)
	}
	if err := s.cartService.EmptyCart(ctx, req.UserID); err != nil {
		ctx, _ = s.logger.Warn(ctx, "failed to empty cart: %v", err)
	}
	orderID, err := newUUID()
	if err != nil {
		return Order{}, fmt.Errorf("failed to create order id: %w", err)
	}
	order := Order{
		OrderID: orderID, ShippingTrackingID: trackingID, ShippingCost: prep.shippingCost,
		ShippingAddress: req.Address, Items: prep.orderItems,
	}
	if err := s.emailService.SendOrderConfirmation(ctx, req.Email, order); err != nil {
		ctx, _ = s.logger.Error(ctx, "failed to send order confirmation: email=%s err=%v", req.Email, err)
	}
	return order, nil
}

type orderPreparation struct {
	orderItems   []OrderItem
	cartItems    []CartItem
	shippingCost Money
}

func (s *CheckoutServiceImpl) prepareOrder(ctx context.Context, userID, currency string, address Address) (orderPreparation, error) {
	var out orderPreparation
	cart, err := s.cartService.GetCart(ctx, userID)
	if err != nil {
		return out, fmt.Errorf("failed to get user cart during checkout: %w", err)
	}
	out.orderItems = make([]OrderItem, len(cart))
	for i, item := range cart {
		product, err := s.catalogService.GetProduct(ctx, item.ProductID)
		if err != nil {
			return out, fmt.Errorf("failed to get product #%q: %w", item.ProductID, err)
		}
		price, err := s.currencyService.Convert(ctx, product.PriceUSD, currency)
		if err != nil {
			return out, fmt.Errorf("failed to convert price of %q to %s: %w", item.ProductID, currency, err)
		}
		out.orderItems[i] = OrderItem{Item: item, Cost: price}
	}
	shippingUSD, err := s.shippingService.GetQuote(ctx, address, cart)
	if err != nil {
		return out, fmt.Errorf("failed to get shipping quote: %w", err)
	}
	out.shippingCost, err = s.currencyService.Convert(ctx, shippingUSD, currency)
	if err != nil {
		return out, fmt.Errorf("failed to convert shipping cost to currency: %w", err)
	}
	out.cartItems = cart
	return out, nil
}
