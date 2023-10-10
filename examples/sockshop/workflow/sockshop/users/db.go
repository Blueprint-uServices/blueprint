package users

/*
Blueprint additional comment:

This code is directly translated from the UserService implementation on GitHub,
which includes several layers of database indirection.

All subsequent comments in this file are from the original code
*/

import (
	"errors"
)

// Database represents a simple interface so we can switch to a new system easily
// this is just basic and specific to this microservice
type Database interface {
	Init() error
	GetUserByName(string) (User, error)
	GetUser(string) (User, error)
	GetUsers() ([]User, error)
	CreateUser(*User) error
	GetUserAttributes(*User) error
	GetAddress(string) (Address, error)
	GetAddresses() ([]Address, error)
	CreateAddress(*Address, string) error
	GetCard(string) (Card, error)
	GetCards() ([]Card, error)
	Delete(string, string) error
	CreateCard(*Card, string) error
}

var (
	//ErrNoDatabaseFound error returnes when database interface does not exists in DBTypes
	ErrNoDatabaseFound = "No database with name %v registered"
	//ErrNoDatabaseSelected is returned when no database was designated in the flag or env
	ErrNoDatabaseSelected = errors.New("No DB selected")
)

// CreateUser invokes DefaultDb method
func CreateUser(db Database, u *User) error {
	return db.CreateUser
	return DefaultDb.CreateUser(u)
}

// GetUserByName invokes DefaultDb method
func GetUserByName(n string) (users.User, error) {
	u, err := DefaultDb.GetUserByName(n)
	if err == nil {
		u.AddLinks()
	}
	return u, err
}

// GetUser invokes DefaultDb method
func GetUser(n string) (users.User, error) {
	u, err := DefaultDb.GetUser(n)
	if err == nil {
		u.AddLinks()
	}
	return u, err
}

// GetUsers invokes DefaultDb method
func GetUsers() ([]users.User, error) {
	us, err := DefaultDb.GetUsers()
	for k, _ := range us {
		us[k].AddLinks()
	}
	return us, err
}

// GetUserAttributes invokes DefaultDb method
func GetUserAttributes(u *users.User) error {
	err := DefaultDb.GetUserAttributes(u)
	if err != nil {
		return err
	}
	for k, _ := range u.Addresses {
		u.Addresses[k].AddLinks()
	}
	for k, _ := range u.Cards {
		u.Cards[k].AddLinks()
	}
	return nil
}

// CreateAddress invokes DefaultDb method
func CreateAddress(a *users.Address, userid string) error {
	return DefaultDb.CreateAddress(a, userid)
}

// GetAddress invokes DefaultDb method
func GetAddress(n string) (users.Address, error) {
	a, err := DefaultDb.GetAddress(n)
	if err == nil {
		a.AddLinks()
	}
	return a, err
}

// GetAddresses invokes DefaultDb method
func GetAddresses() ([]users.Address, error) {
	as, err := DefaultDb.GetAddresses()
	for k, _ := range as {
		as[k].AddLinks()
	}
	return as, err
}

// CreateCard invokes DefaultDb method
func CreateCard(c *users.Card, userid string) error {
	return DefaultDb.CreateCard(c, userid)
}

// GetCard invokes DefaultDb method
func GetCard(n string) (users.Card, error) {
	return DefaultDb.GetCard(n)
}

// GetCards invokes DefaultDb method
func GetCards() ([]users.Card, error) {
	cs, err := DefaultDb.GetCards()
	for k, _ := range cs {
		cs[k].AddLinks()
	}
	return cs, err
}

// Delete invokes DefaultDb method
func Delete(entity, id string) error {
	return DefaultDb.Delete(entity, id)
}

// Ping invokes DefaultDB method
func Ping() error {
	return DefaultDb.Ping()
}
