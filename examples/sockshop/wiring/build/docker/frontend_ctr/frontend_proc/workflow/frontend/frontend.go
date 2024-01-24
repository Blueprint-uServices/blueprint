// Package frontend implements the SockShop frontend service, typically deployed via HTTP
package frontend

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/cart"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/catalogue"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/order"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/user"
	"github.com/google/uuid"
)

type (
	// The SockShop Frontend receives requests from users and proxies them to the application's other services
	Frontend interface {
		// List items in cart for current logged in user, or for the current session if not logged in.
		// SessionID can be the empty string for a non-logged in user / new session
		GetCart(ctx context.Context, sessionID string) ([]cart.Item, error)

		// Deletes the entire cart for a user/session
		DeleteCart(ctx context.Context, sessionID string) error

		// Removes an item from the user/session's cart
		RemoveItem(ctx context.Context, sessionID string, itemID string) error

		// Adds an item to the user/session's cart.
		// If there is no user or session, then a session is created and the sessionID is returned.
		AddItem(ctx context.Context, sessionID string, itemID string) (newSessionID string, err error)

		// Update item quantity in the user/session's cart
		// If there is no user or session, then a session is created and the sessionID is returned.
		UpdateItem(ctx context.Context, sessionID string, itemID string, quantity int) (newSessionID string, err error)

		// List socks that match any of the tags specified.  Sort the results by the specified database column.
		// order can be "" in which case the default order is used.
		// pageNum is 1-indexed
		// then return a subset of the results.
		ListItems(ctx context.Context, tags []string, order string, pageNum, pageSize int) ([]catalogue.Sock, error)

		// Gets details about a [Sock]
		GetSock(ctx context.Context, itemID string) (catalogue.Sock, error)

		// Lists all tags
		ListTags(ctx context.Context) ([]string, error)

		// Place an order for the specified items
		NewOrder(ctx context.Context, userID, addressID, cardID, cartID string) (order.Order, error)

		// Get all orders for a customer, sorted by date
		GetOrders(ctx context.Context, userID string) ([]order.Order, error)

		// Get an order by ID
		GetOrder(ctx context.Context, orderID string) (order.Order, error)

		// Log in to an existing user account.  Returns an error if the password
		// doesn't match the registered password
		// Returns the new session ID, which will be the user ID of the logged in user.
		Login(ctx context.Context, sessionID, username, password string) (newSessionID string, u user.User, err error)

		// Register a new user account
		// Returns the new session ID, which will be the user ID of the registered user.
		Register(ctx context.Context, sessionID, username, password, email, first, last string) (newSessionID string, err error)

		// Look up a user by customer ID
		GetUser(ctx context.Context, userID string) (user.User, error)

		// Look up an address by address ID
		GetAddress(ctx context.Context, addressID string) (user.Address, error)

		// Adds a new address for a customer
		PostAddress(ctx context.Context, userID string, address user.Address) (string, error)

		// Look up a card by card id.
		GetCard(ctx context.Context, cardID string) (user.Card, error)

		// Adds a new card for a customer
		PostCard(ctx context.Context, userID string, card user.Card) (string, error)
	}
)

type frontend struct {
	user      user.UserService
	catalogue catalogue.CatalogueService
	cart      cart.CartService
	order     order.OrderService
}

// Instantiates the Frontend service, which makes calls to the user, catalogue, cart, and order services
func NewFrontend(ctx context.Context, user user.UserService, catalogue catalogue.CatalogueService, cart cart.CartService, order order.OrderService) (Frontend, error) {
	f := &frontend{
		user:      user,
		catalogue: catalogue,
		cart:      cart,
		order:     order,
	}
	return f, nil
}

// AddItem implements Frontend.
func (f *frontend) AddItem(ctx context.Context, sessionID string, itemID string) (string, error) {
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	sock, err := f.catalogue.Get(ctx, itemID)
	if err != nil {
		return sessionID, err
	}

	_, err = f.cart.AddItem(ctx, sessionID, cart.Item{ID: sock.ID, Quantity: 1, UnitPrice: sock.Price})
	return sessionID, err
}

// RemoteItem implements Frontend.
func (f *frontend) RemoveItem(ctx context.Context, sessionID string, itemID string) error {
	if sessionID == "" {
		return nil
	}

	return f.cart.RemoveItem(ctx, sessionID, itemID)
}

// GetCart implements Frontend.
func (f *frontend) GetCart(ctx context.Context, sessionID string) ([]cart.Item, error) {
	if sessionID == "" {
		return nil, nil
	}

	return f.cart.GetCart(ctx, sessionID)
}

// DeleteCart implements Frontend.
func (f *frontend) DeleteCart(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}

	return f.cart.DeleteCart(ctx, sessionID)
}

// GetUser implements Frontend.
func (f *frontend) GetUser(ctx context.Context, userID string) (user.User, error) {
	if userID == "" {
		return user.User{}, fmt.Errorf("no userID specified")
	}

	users, err := f.user.GetUsers(ctx, userID)
	if err != nil {
		return user.User{}, err
	} else if len(users) == 0 {
		return user.User{}, fmt.Errorf("invalid userID %v", userID)
	} else {
		return users[0], nil
	}
}

// GetAddresses implements Frontend.
func (f *frontend) GetAddress(ctx context.Context, addressID string) (user.Address, error) {
	if addressID == "" {
		return user.Address{}, fmt.Errorf("no addressID specified")
	}
	addrs, err := f.user.GetAddresses(ctx, addressID)
	if err != nil {
		return user.Address{}, err
	} else if len(addrs) == 0 {
		return user.Address{}, fmt.Errorf("invalid addressID %v", addressID)
	} else {
		return addrs[0], nil
	}
}

// GetCards implements Frontend.
func (f *frontend) GetCard(ctx context.Context, cardID string) (user.Card, error) {
	if cardID == "" {
		return user.Card{}, fmt.Errorf("no cardID specified")
	}
	cards, err := f.user.GetCards(ctx, cardID)
	if err != nil {
		return user.Card{}, err
	} else if len(cards) == 0 {
		return user.Card{}, fmt.Errorf("invalid cardID %v", cardID)
	} else {
		return cards[0], nil
	}
}

// GetOrder implements Frontend.
func (f *frontend) GetOrder(ctx context.Context, orderID string) (order.Order, error) {
	return f.order.GetOrder(ctx, orderID)
}

// GetOrders implements Frontend.
func (f *frontend) GetOrders(ctx context.Context, userID string) ([]order.Order, error) {
	if userID == "" {
		return nil, fmt.Errorf("no userID specified")
	}
	return f.order.GetOrders(ctx, userID)
}

// GetSock implements Frontend.
func (f *frontend) GetSock(ctx context.Context, itemID string) (catalogue.Sock, error) {
	return f.catalogue.Get(ctx, itemID)
}

// ListItems implements Frontend.
func (f *frontend) ListItems(ctx context.Context, tags []string, order string, pageNum int, pageSize int) ([]catalogue.Sock, error) {
	return f.catalogue.List(ctx, tags, order, pageNum, pageSize)
}

// ListTags implements Frontend.
func (f *frontend) ListTags(ctx context.Context) ([]string, error) {
	return f.catalogue.Tags(ctx)
}

// Login implements Frontend.  Merges the session into the user, and returns the user ID
func (f *frontend) Login(ctx context.Context, sessionID string, username string, password string) (string, user.User, error) {
	u, err := f.user.Login(ctx, username, password)
	if err != nil {
		return sessionID, user.User{}, err
	}

	if sessionID != "" {
		if err := f.cart.MergeCarts(ctx, u.UserID, sessionID); err != nil {
			return u.UserID, u, err
		}
	}

	return u.UserID, u, nil
}

// NewOrder implements Frontend.
func (f *frontend) NewOrder(ctx context.Context, userID string, addressID string, cardID string, cartID string) (order.Order, error) {
	return f.order.NewOrder(ctx, userID, addressID, cardID, cartID)
}

// PostAddress implements Frontend.
func (f *frontend) PostAddress(ctx context.Context, userID string, address user.Address) (string, error) {
	return f.user.PostAddress(ctx, userID, address)
}

// PostCard implements Frontend.
func (f *frontend) PostCard(ctx context.Context, userID string, card user.Card) (string, error) {
	return f.user.PostCard(ctx, userID, card)
}

// Register implements Frontend.
func (f *frontend) Register(ctx context.Context, sessionID string, username string, password string, email string, first string, last string) (string, error) {
	userID, err := f.user.Register(ctx, username, password, email, first, last)
	if err != nil {
		return sessionID, err
	}

	if sessionID != "" {
		return userID, f.cart.MergeCarts(ctx, userID, sessionID)
	} else {
		return userID, nil
	}
}

// UpdateItem implements Frontend.
func (f *frontend) UpdateItem(ctx context.Context, sessionID string, itemID string, quantity int) (string, error) {
	item, err := f.catalogue.Get(ctx, itemID)
	if err != nil {
		return sessionID, err
	}

	return sessionID, f.cart.UpdateItem(ctx, sessionID, cart.Item{ID: item.ID, Quantity: quantity, UnitPrice: item.Price})
}
