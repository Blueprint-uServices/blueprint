// Package frontend implements the SockShop frontend service, typically deployed via HTTP
package frontend

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/cart"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/catalogue"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/order"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/user"
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

		// List socks that match any of the tags specified.  Sort the results in the specified order,
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

		// Look up an address by customer ID
		GetAddresses(ctx context.Context, userID string) ([]user.Address, error)

		// Adds a new address for a customer
		PostAddress(ctx context.Context, userID string, address user.Address) (string, error)

		// Look up a card by id.  If id is the empty string, returns all cards.
		GetCards(ctx context.Context, userID string) ([]user.Card, error)

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

var ErrNoUserID = fmt.Errorf("no userID specified")

// GetUser implements Frontend.
func (f *frontend) GetUser(ctx context.Context, userID string) (user.User, error) {
	if userID == "" {
		return user.User{}, ErrNoUserID
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
func (f *frontend) GetAddresses(ctx context.Context, userID string) ([]user.Address, error) {
	if userID == "" {
		return nil, ErrNoUserID
	}
	return f.user.GetAddresses(ctx, userID)
}

// GetCards implements Frontend.
func (f *frontend) GetCards(ctx context.Context, userID string) ([]user.Card, error) {
	if userID == "" {
		return nil, ErrNoUserID
	}
	return f.user.GetCards(ctx, userID)
}

// GetOrder implements Frontend.
func (f *frontend) GetOrder(ctx context.Context, orderID string) (order.Order, error) {
	return f.order.GetOrder(ctx, orderID)
}

// GetOrders implements Frontend.
func (f *frontend) GetOrders(ctx context.Context, userID string) ([]order.Order, error) {
	if userID == "" {
		return nil, ErrNoUserID
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
	return f.user.PostAddress(ctx, address)
}

// PostCard implements Frontend.
func (f *frontend) PostCard(ctx context.Context, userID string, card user.Card) (string, error) {
	return f.user.PostCard(ctx, card)
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
