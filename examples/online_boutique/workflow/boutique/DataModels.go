package boutique

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"
)

const (
	nanosMin = -999999999
	nanosMax = +999999999
	nanosMod = 1000000000
)

var (
	ErrInvalidValue        = errors.New("one of the specified money values is invalid")
	ErrMismatchingCurrency = errors.New("mismatching currency codes")
)

// Money represents an amount of money along with the currency type.
type Money struct {
	// The 3-letter currency code defined in ISO 4217.
	CurrencyCode string `json:"currencyCode"`

	// The whole units of the amount.
	// For example if `currencyCode` is `"USD"`, then 1 unit is one US dollar.
	Units int64 `json:"units"`

	// Number of nano (10^-9) units of the amount.
	// The value must be between -999,999,999 and +999,999,999 inclusive.
	// If `units` is positive, `nanos` must be positive or zero.
	// If `units` is zero, `nanos` can be positive, zero, or negative.
	// If `units` is negative, `nanos` must be negative or zero.
	// For example $-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.
	Nanos int32 `json:"nanos"`
}

// IsValid checks if specified value has a valid units/nanos signs and ranges.
func IsValid(m Money) bool {
	return signMatches(m) && validNanos(m.Nanos)
}

func signMatches(m Money) bool {
	return m.Nanos == 0 || m.Units == 0 || (m.Nanos < 0) == (m.Units < 0)
}

func validNanos(nanos int32) bool { return nanosMin <= nanos && nanos <= nanosMax }

// IsZero returns true if the specified money value is equal to zero.
func IsZero(m Money) bool { return m.Units == 0 && m.Nanos == 0 }

// IsPositive returns true if the specified money value is valid and is
// positive.
func IsPositive(m Money) bool {
	return IsValid(m) && m.Units > 0 || (m.Units == 0 && m.Nanos > 0)
}

// IsNegative returns true if the specified money value is valid and is
// negative.
func IsNegative(m Money) bool {
	return IsValid(m) && m.Units < 0 || (m.Units == 0 && m.Nanos < 0)
}

// AreSameCurrency returns true if values l and r have a currency code and
// they are the same values.
func AreSameCurrency(l, r Money) bool {
	return l.CurrencyCode == r.CurrencyCode && l.CurrencyCode != ""
}

// AreEquals returns true if values l and r are the equal, including the
// currency. This does not check validity of the provided values.
func AreEquals(l, r Money) bool {
	return l.CurrencyCode == r.CurrencyCode &&
		l.Units == r.Units && l.Nanos == r.Nanos
}

// Negate returns the same amount with the sign negated.
func Negate(m Money) Money {
	return Money{
		Units:        -m.Units,
		Nanos:        -m.Nanos,
		CurrencyCode: m.CurrencyCode}
}

// Must panics if the given error is not nil. This can be used with other
// functions like: "m := Must(Sum(a,b))".
func Must(v Money, err error) Money {
	if err != nil {
		panic(err)
	}
	return v
}

// Sum adds two values. Returns an error if one of the values are invalid or
// currency codes are not matching (unless currency code is unspecified for
// both).
func Sum(l, r Money) (Money, error) {
	if !IsValid(l) || !IsValid(r) {
		return Money{}, ErrInvalidValue
	} else if l.CurrencyCode != r.CurrencyCode {
		return Money{}, ErrMismatchingCurrency
	}
	units := l.Units + r.Units
	nanos := l.Nanos + r.Nanos

	if (units == 0 && nanos == 0) || (units > 0 && nanos >= 0) || (units < 0 && nanos <= 0) {
		// same sign <units, nanos>
		units += int64(nanos / nanosMod)
		nanos = nanos % nanosMod
	} else {
		// different sign. nanos guaranteed to not to go over the limit
		if units > 0 {
			units--
			nanos += nanosMod
		} else {
			units++
			nanos -= nanosMod
		}
	}

	return Money{
		Units:        units,
		Nanos:        nanos,
		CurrencyCode: l.CurrencyCode}, nil
}

// MultiplySlow is a slow multiplication operation done through adding the value
// to itself n-1 times.
func MultiplySlow(m Money, n uint32) Money {
	out := m
	for n > 1 {
		out = Must(Sum(out, m))
		n--
	}
	return out
}

type Address struct {
	StreetAddress string
	City          string
	State         string
	Country       string
	ZipCode       int32
}

type CartItem struct {
	ProductID string
	Quantity  int32
}

// Order represents a user order.
type Order struct {
	OrderID            string
	ShippingTrackingID string
	ShippingCost       Money
	ShippingAddress    Address
	Items              []OrderItem
}

type OrderItem struct {
	Item CartItem
	Cost Money
}

// Quote represents a currency value.
type quote struct {
	Dollars uint32
	Cents   uint32
}

// String representation of the Quote.
func (q quote) String() string {
	return fmt.Sprintf("$%d.%d", q.Dollars, q.Cents)
}

// createQuoteFromCount takes a number of items and returns a quote.
func createQuoteFromCount(count int) quote {
	return createQuoteFromFloat(8.99)
}

// createQuoteFromFloat takes a price represented as a float and creates a quote.
func createQuoteFromFloat(value float64) quote {
	units, fraction := math.Modf(value)
	return quote{
		uint32(units),
		uint32(math.Trunc(fraction * 100)),
	}
}

// Ad represents an advertisement.
type Ad struct {
	RedirectURL string // URL to redirect to when an ad is clicked.
	Text        string // Short advertisement text to display.
}

type CreditCardInfo struct {
	Number          string
	CVV             int32
	ExpirationYear  int
	ExpirationMonth time.Month
}

// LastFour returns the last four digits of the card number.
func (c CreditCardInfo) LastFour() string {
	num := c.Number
	if len(num) > 4 {
		num = num[len(num)-4:]
	}
	return num
}

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Picture     string `json:"picture"`
	PriceUSD    Money  `json:"priceUsd"`

	// Categories such as "clothing" or "kitchen" that can be used to look up
	// other related products.
	Categories []string `json:"categories"`
}

func newUUID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	var out [36]byte
	hex.Encode(out[0:8], id[0:4])
	out[8] = '-'
	hex.Encode(out[9:13], id[4:6])
	out[13] = '-'
	hex.Encode(out[14:18], id[6:8])
	out[18] = '-'
	hex.Encode(out[19:23], id[8:10])
	out[23] = '-'
	hex.Encode(out[24:36], id[10:16])
	return string(out[:]), nil
}
