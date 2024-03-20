// Package user implements the SockShop user microservice.
//
// The service stores three kinds of information:
//   - user accounts
//   - addresses
//   - credit cards
//
// The sock shop allows customers to check out without creating a user
// account; in this case the customer's address and credit card data
// will be stored without a user accont.
//
// The UserService thus uses three collections for the above information.
// To get the data for a user also means more than one database call.
package user

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type (
	// UserService stores information about user accounts.
	// Having a user account is optional, and not required for placing orders.
	// UserService also stores addresses and credit card details used
	// in orders that aren't associated with a user account.
	UserService interface {
		// Log in to an existing user account.  Returns an error if the password
		// doesn't match the registered password
		Login(ctx context.Context, username, password string) (User, error)

		// Register a new user account.
		// Returns the user ID
		Register(ctx context.Context, username, password, email, first, last string) (string, error)

		// Look up a user by id.  If id is the empty string, returns all users.
		GetUsers(ctx context.Context, id string) ([]User, error)

		// Insert a (possibly new) user into the DB.  Returns the user's ID
		PostUser(ctx context.Context, user User) (string, error)

		// Look up an address by id.  If id is the empty string, returns all addresses.
		GetAddresses(ctx context.Context, id string) ([]Address, error)

		// Insert a (possibly new) address into the DB.  Returns the address ID
		PostAddress(ctx context.Context, userid string, address Address) (string, error)

		// Look up a card by id.  If id is the empty string, returns all cards.
		GetCards(ctx context.Context, cardid string) ([]Card, error)

		// Insert a (possibly new) card into the DB.  Returns the card ID
		PostCard(ctx context.Context, userid string, card Card) (string, error)

		// Deletes an entity with ID id from the DB.
		//
		// entity can be one of "customers", "addresses", or "cards".
		// ID should be the id of the entity to delete
		Delete(ctx context.Context, entity string, id string) error
	}

	// A user with an account.  Accounts are optional for ordering.
	User struct {
		FirstName string    `json:"firstName" bson:"firstName"`
		LastName  string    `json:"lastName" bson:"lastName"`
		Email     string    `json:"-" bson:"email"`
		Username  string    `json:"username" bson:"username"`
		Password  string    `json:"-" bson:"password,omitempty"`
		Addresses []Address `json:"addresses" bson:"-"`
		Cards     []Card    `json:"cards" bson:"-"`
		UserID    string    `json:"id" bson:"-"`
		Salt      string    `json:"-" bson:"salt"`
	}

	// A street address
	Address struct {
		Street   string
		Number   string
		Country  string
		City     string
		PostCode string
		ID       string
	}

	// A credit card
	Card struct {
		LongNum string
		Expires string
		CCV     string
		ID      string
	}
)

// An implementation of the UserService that stores information in a NoSQLDatabase.
// It uses three collections within the database: users, addresses, and cards.
// Addresses and cards are stored separately from user information, because having
// a user account is optional when placing an order.
type userServiceImpl struct {
	UserService
	users *userStore
}

// Creates a UserService implementation that stores user, address, and credit card
// information in a NoSQLDatabase.
//
// Returns an error if unable to get the users, addresses, or cards collection from the DB
func NewUserServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (UserService, error) {
	users, err := newUserStore(ctx, db)
	return &userServiceImpl{users: users}, err
}

func (s *userServiceImpl) Login(ctx context.Context, username, password string) (User, error) {
	// Load the user from the DB
	u, err := s.users.getUserByName(ctx, username)
	if err != nil {
		return newUser(), err
	}

	// Check the password
	if u.Password != calculatePassHash(password, u.Salt) {
		return newUser(), errors.New("Unauthorized")
	}

	// Fetch user's card and address data, mask out CC numbers
	err = s.users.getUserAttributes(ctx, &u)
	u.maskCCs()
	return u, err
}

func (s *userServiceImpl) Register(ctx context.Context, username, password, email, first, last string) (string, error) {
	// Create the public user info
	u := newUser()
	u.Username = username
	u.Password = calculatePassHash(password, u.Salt)
	u.Email = email
	u.FirstName = first
	u.LastName = last
	u.Addresses = []Address{}
	u.Cards = []Card{}

	// Save the user in the DB
	err := s.users.createUser(ctx, &u)
	return u.UserID, err
}

func (s *userServiceImpl) GetUsers(ctx context.Context, userid string) ([]User, error) {
	if userid == "" {
		return s.users.getUsers(ctx)
	} else {
		user, err := s.users.getUser(ctx, userid)
		return []User{user}, err
	}
}

func (s *userServiceImpl) PostUser(ctx context.Context, u User) (string, error) {
	u.newSalt()
	u.Password = calculatePassHash(u.Password, u.Salt)
	err := s.users.createUser(ctx, &u)
	return u.UserID, err
}

func (s *userServiceImpl) GetAddresses(ctx context.Context, addressid string) ([]Address, error) {
	if addressid == "" {
		return s.users.addresses.getAllAddresses(ctx)
	} else {
		address, err := s.users.addresses.getAddress(ctx, addressid)
		return []Address{address}, err
	}
}

func (s *userServiceImpl) PostAddress(ctx context.Context, userid string, address Address) (string, error) {
	err := s.users.createAddress(ctx, userid, &address)
	return address.ID, err
}

func (s *userServiceImpl) GetCards(ctx context.Context, cardid string) ([]Card, error) {
	if cardid == "" {
		return s.users.cards.getAllCards(ctx)
	} else {
		card, err := s.users.cards.getCard(ctx, cardid)
		return []Card{card}, err
	}
}

func (s *userServiceImpl) PostCard(ctx context.Context, userid string, card Card) (string, error) {
	err := s.users.createCard(ctx, userid, &card)
	return card.ID, err
}

func (s *userServiceImpl) Delete(ctx context.Context, entity string, id string) error {
	return s.users.delete(ctx, entity, id)
}

// Creates a new, empty user, with a salt.
func newUser() User {
	u := User{Addresses: make([]Address, 0), Cards: make([]Card, 0)}
	u.newSalt()
	return u
}

var (
	errMissingField = "Error missing %v"
)

func (u *User) validate() error {
	if u.FirstName == "" {
		return fmt.Errorf(errMissingField, "FirstName")
	}
	if u.LastName == "" {
		return fmt.Errorf(errMissingField, "LastName")
	}
	if u.Username == "" {
		return fmt.Errorf(errMissingField, "Username")
	}
	if u.Password == "" {
		return fmt.Errorf(errMissingField, "Password")
	}
	return nil
}

// Replace all CC numbers with asterisks, for returning to the user for display
func (u *User) maskCCs() {
	for i := range u.Cards {
		u.Cards[i].maskCC()
	}
}

func (u *User) newSalt() {
	h := sha1.New()
	io.WriteString(h, strconv.Itoa(int(time.Now().UnixNano())))
	u.Salt = fmt.Sprintf("%x", h.Sum(nil))
}

func (u *User) addressIDs() []string {
	ids := []string{}
	for _, address := range u.Addresses {
		ids = append(ids, address.ID)
	}
	return ids
}

func (u *User) cardIDs() []string {
	ids := []string{}
	for _, card := range u.Cards {
		ids = append(ids, card.ID)
	}
	return ids
}

// Replaces the CC number with asterisks for returning to the user for display
func (c *Card) maskCC() {
	l := len(c.LongNum) - 4
	c.LongNum = fmt.Sprintf("%v%v", strings.Repeat("*", l), c.LongNum[l:])
}

func calculatePassHash(pass, salt string) string {
	h := sha1.New()
	io.WriteString(h, salt)
	io.WriteString(h, pass)
	return fmt.Sprintf("%x", h.Sum(nil))
}
