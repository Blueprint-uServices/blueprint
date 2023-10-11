package users

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNoCustomerInResponse = errors.New("Response has no matching customer")
	ErrMissingField         = "Error missing %v"

	entitymap = map[string]string{
		"customer": "customers",
		"address":  "addresses",
		"card":     "cards",
	}
)

type (
	User struct {
		FirstName string    `json:"firstName" bson:"firstName"`
		LastName  string    `json:"lastName" bson:"lastName"`
		Email     string    `json:"-" bson:"email"`
		Username  string    `json:"username" bson:"username"`
		Password  string    `json:"-" bson:"password,omitempty"`
		Addresses []Address `json:"-,omitempty" bson:"-"`
		Cards     []Card    `json:"-,omitempty" bson:"-"`
		UserID    string    `json:"id" bson:"-"`
		Links     Links     `json:"_links"`
		Salt      string    `json:"-" bson:"salt"`
	}

	Address struct {
		Street   string `json:"street" bson:"street,omitempty"`
		Number   string `json:"number" bson:"number,omitempty"`
		Country  string `json:"country" bson:"country,omitempty"`
		City     string `json:"city" bson:"city,omitempty"`
		PostCode string `json:"postcode" bson:"postcode,omitempty"`
		ID       string `json:"id" bson:"-"`
		Links    Links  `json:"_links"`
	}

	Card struct {
		LongNum string `json:"longNum" bson:"longNum"`
		Expires string `json:"expires" bson:"expires"`
		CCV     string `json:"ccv" bson:"ccv"`
		ID      string `json:"id" bson:"-"`
		Links   Links  `json:"_links" bson:"-"`
	}

	Links map[string]Href

	Href struct {
		string `json:"href"`
	}
)

func NewUser() User {
	u := User{Addresses: make([]Address, 0), Cards: make([]Card, 0)}
	u.NewSalt()
	return u
}

func (u *User) Validate() error {
	if u.FirstName == "" {
		return fmt.Errorf(ErrMissingField, "FirstName")
	}
	if u.LastName == "" {
		return fmt.Errorf(ErrMissingField, "LastName")
	}
	if u.Username == "" {
		return fmt.Errorf(ErrMissingField, "Username")
	}
	if u.Password == "" {
		return fmt.Errorf(ErrMissingField, "Password")
	}
	return nil
}

func (u *User) MaskCCs() {
	for k, c := range u.Cards {
		c.MaskCC()
		u.Cards[k] = c
	}
}

func (u *User) AddLinks() {
	u.Links.AddCustomer(u.UserID)
}

func (u *User) NewSalt() {
	h := sha1.New()
	io.WriteString(h, strconv.Itoa(int(time.Now().UnixNano())))
	u.Salt = fmt.Sprintf("%x", h.Sum(nil))
}

func (a *Address) AddLinks() {
	a.Links.AddAddress(a.ID)
}

func (c *Card) MaskCC() {
	l := len(c.LongNum) - 4
	c.LongNum = fmt.Sprintf("%v%v", strings.Repeat("*", l), c.LongNum[l:])
}

func (c *Card) AddLinks() {
	c.Links.AddCard(c.ID)
}

func (l *Links) AddLink(ent string, id string) {
	nl := make(Links)
	link := fmt.Sprintf("userservice/%v/%v", entitymap[ent], id)
	nl[ent] = Href{link}
	nl["self"] = Href{link}
	*l = nl

}

func (l *Links) AddAttrLink(attr string, corent string, id string) {
	link := fmt.Sprintf("userservice/%v/%v/%v", entitymap[corent], id, entitymap[attr])
	nl := *l
	nl[entitymap[attr]] = Href{link}
	*l = nl
}

func (l *Links) AddCustomer(id string) {
	l.AddLink("customer", id)
	l.AddAttrLink("address", "customer", id)
	l.AddAttrLink("card", "customer", id)
}

func (l *Links) AddAddress(id string) {
	l.AddLink("address", id)
}

func (l *Links) AddCard(id string) {
	l.AddLink("card", id)
}
