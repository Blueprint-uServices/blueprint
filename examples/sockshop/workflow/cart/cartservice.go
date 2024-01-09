// Package cart implements the SockShop cart microservice.
package cart

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	// The CartService interface
	CartService interface {
		// Get all items in a customer's cart.  A customer might not have a cart,
		// in which case the empty list is returned.  customerID can be a userID
		// for a logged in user, or a sessionID for an anonymous user.
		GetCart(ctx context.Context, customerID string) ([]Item, error)

		// Delete a customer's cart
		DeleteCart(ctx context.Context, customerID string) error

		// Merge two carts.  Used when an anonymous customer logs in
		MergeCarts(ctx context.Context, customerID, sessionID string) error

		// Get a specific item from a customer's cart
		GetItem(ctx context.Context, customerID string, itemID string) (Item, error)

		// Add an item to a customer's cart.
		// If the item already exists in the cart, then the total quantity is
		// updated to reflect the combined total.
		// Returns the current state of the item in the customer's cart.
		AddItem(ctx context.Context, customerID string, item Item) (Item, error)

		// Remove an item from the customer's cart
		RemoveItem(ctx context.Context, customerID, itemID string) error

		// Updates an item in the customer's cart to the value provided.
		UpdateItem(ctx context.Context, customerID string, item Item) error
	}

	// A cart belongs to either a customer or a session
	cart struct {
		ID    string
		Items []Item
	}

	// A cart item is just an item ID and a quantity.  The catalogue service is responsible
	// for managing the actual items.
	Item struct {
		ID        string  // Item ID will correspond to the ID used by the catalogue service
		Quantity  int     // The quantity of this item in the car
		UnitPrice float32 // The price of the item
	}
)

// Implementation of [CartService]
type cartImpl struct {
	db backend.NoSQLCollection
}

// Creates a [CartService] instance that persists cart data in the provided db
func NewCartService(ctx context.Context, db backend.NoSQLDatabase) (CartService, error) {
	collection, err := db.GetCollection(ctx, "cart", "carts")
	return &cartImpl{db: collection}, err
}

// AddItem implements CartService.
func (s *cartImpl) AddItem(ctx context.Context, customerID string, item Item) (Item, error) {
	cart, err := s.getCart(ctx, customerID)
	if err != nil {
		return item, err
	}

	if existingItem := findItem(cart, item.ID); existingItem != nil {
		existingItem.Quantity += item.Quantity
		item = *existingItem
	} else {
		cart.Items = append(cart.Items, item)
	}

	_, err = s.db.Upsert(ctx, bson.D{{"id", customerID}}, cart)
	return item, err
}

// DeleteCart implements CartService.
func (s *cartImpl) DeleteCart(ctx context.Context, customerID string) error {
	return s.db.DeleteMany(ctx, bson.D{{"id", customerID}})
}

// GetCart implements CartService.
func (s *cartImpl) GetCart(ctx context.Context, customerID string) ([]Item, error) {
	cart, err := s.getCart(ctx, customerID)
	if err != nil {
		return nil, err
	}
	return cart.Items, nil
}

// GetItem implements CartService.
func (s *cartImpl) GetItem(ctx context.Context, customerID string, itemID string) (Item, error) {
	cart, err := s.getCart(ctx, customerID)
	if err == nil {
		if item := findItem(cart, itemID); item != nil {
			return *item, nil
		}
	}
	return Item{}, err
}

// MergeCarts implements CartService.
func (s *cartImpl) MergeCarts(ctx context.Context, customerID string, sessionID string) error {
	sessionCart, err := s.getCart(ctx, sessionID)
	if err != nil {
		return err
	}
	customerCart, err := s.getCart(ctx, customerID)
	if err != nil {
		return err
	}

	if len(sessionCart.Items) == 0 {
		// No update to perform
		return nil
	}

	// Update quantity of existing items; append new items

	customerCartItems := make(map[string]*Item)
	for i := 0; i < len(customerCart.Items); i++ {
		customerCartItems[customerCart.Items[i].ID] = &customerCart.Items[i]
	}

	for _, item := range sessionCart.Items {
		if existing, exists := customerCartItems[item.ID]; exists {
			existing.Quantity += item.Quantity
			existing.UnitPrice = item.UnitPrice
		} else {
			customerCart.Items = append(customerCart.Items, item)
		}
	}

	_, err = s.db.Upsert(ctx, bson.D{{"id", customerID}}, customerCart)
	if err != nil {
		return err
	}

	// Only delete the session after successfully merging over to customer
	return s.db.DeleteOne(ctx, bson.D{{"id", sessionID}})
}

// RemoveItem implements CartService.
func (s *cartImpl) RemoveItem(ctx context.Context, customerID string, itemID string) error {
	c, err := s.getCart(ctx, customerID)
	if err != nil {
		return err
	}

	if removed := removeItem(c, itemID); !removed {
		return nil
	}

	if len(c.Items) == 0 {
		return s.DeleteCart(ctx, customerID)
	} else {
		_, err := s.db.ReplaceOne(ctx, bson.D{{"id", customerID}}, c)
		return err
	}
}

// UpdateItem implements CartService.
func (s *cartImpl) UpdateItem(ctx context.Context, customerID string, item Item) error {
	cart, err := s.getCart(ctx, customerID)
	if err != nil {
		return err
	}

	if existing := findItem(cart, item.ID); existing != nil {
		// Item exists in the cart, update the quantity
		existing.Quantity = item.Quantity
		existing.UnitPrice = item.UnitPrice

		// After updating, item quantity is gone, so remove item from cart
		if existing.Quantity <= 0 {
			removeItem(cart, item.ID)
		}

		// If no items left in cart, delete cart
		if len(cart.Items) == 0 {
			return s.DeleteCart(ctx, customerID)
		}
	} else {
		// Item doesn't exist in cart and no items added, so do nothing
		if item.Quantity <= 0 {
			return nil
		}

		// Item needs to be added to cart
		cart.Items = append(cart.Items, item)
	}

	_, err = s.db.Upsert(ctx, bson.D{{"id", customerID}}, cart)
	return err
}

func (s *cartImpl) getCart(ctx context.Context, id string) (*cart, error) {
	filter := bson.D{{"id", id}}
	cursor, err := s.db.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	c := cart{ID: id}
	_, err = cursor.One(ctx, &c)
	return &c, err
}

func findItem(c *cart, itemID string) *Item {
	if c == nil {
		return nil
	}
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			return &c.Items[i]
		}
	}
	return nil
}

func removeItem(c *cart, itemID string) bool {
	removed := false
	for i := 0; i < len(c.Items); i++ {
		if c.Items[i].ID == itemID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			i--
			removed = true
		}
	}
	return removed
}
