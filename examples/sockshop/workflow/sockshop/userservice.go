package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type (
	Address struct {
		Id       string
		Street   string
		Number   string
		Country  string
		City     string
		Postcode string
		Href     string
		Link     map[string]string
	}

	Card struct {
		Id         string
		CustomerId string
		LongNum    string
		Expires    string
		Cvv        string
		Href       string
		Link       map[string]string
	}

	Customer struct {
		Id         string
		Username   string
		Email      string
		Password   string
		FirstName  string
		LastName   string
		CardIds    []string //* keeping only the IDs
		AddressIds []string
		Addresses  []Address //* keeping the entire thing
		Cards      []Card
	}

	UserService interface {
		Login(ctx context.Context, username, password string) (Customer, error)
		AddCustomer(ctx context.Context, customer Customer, cards []Card, addresses []Address) (string, error)
		GetCustomerById(ctx context.Context, customerId string) (Customer, error)
		GetCustomers(ctx context.Context) ([]Customer, error)
		DeleteCustomer(ctx context.Context, customerId string) (string, error)
		GetCards(ctx context.Context) ([]Card, error)
		GetCard(ctx context.Context, cardId string) (Card, error)
		GetAddress(ctx context.Context, addressId string) (Address, error)
		DeleteCard(ctx context.Context, cardId string) (string, error)
		AddCard(ctx context.Context, card Card) (string, error)
		GetCardsForCustomer(ctx context.Context, customerId string) ([]Card, error)
		GetAddresses(ctx context.Context) ([]Address, error)
		GetAddressesForCustomer(ctx context.Context, customerId string) ([]Address, error)
		DeleteAddress(ctx context.Context, customerId, addressId string) (string, error)
		AddAddress(ctx context.Context, customerId string, address Address) (string, error)
	}
)

type UserServiceImpl struct {
	customers  backend.NoSQLCollection
	cards      backend.NoSQLCollection
	addresses  backend.NoSQLCollection
	linkDomain string
}

func CreateUserService(ctx context.Context, db backend.NoSQLDatabase) (UserService, error) {
	us := &UserServiceImpl{linkDomain: "userService"}
	var err error
	if us.customers, err = db.GetCollection(ctx, "user", "customers"); err != nil {
		return nil, err
	}
	if us.cards, err = db.GetCollection(ctx, "user", "cards"); err != nil {
		return nil, err
	}
	if us.addresses, err = db.GetCollection(ctx, "user", "addresses"); err != nil {
		return nil, err
	}
	return us, nil
}

func (*UserServiceImpl) prepareSaltedHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hash)
}

func (*UserServiceImpl) maskCard(cardNum string) string {
	if len(cardNum) < 10 {
		panic("Length of cardNum is less than 10")
	}

	return cardNum[:4] + strings.Repeat("*", len(cardNum)-8) + cardNum[len(cardNum)-4:]
}

func queryOne(ctx context.Context, collection backend.NoSQLCollection, filter bson.D, receiver any) error {
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return err
	}
	return cursor.Decode(ctx, receiver)
}

func queryMany(ctx context.Context, collection backend.NoSQLCollection, filter bson.D, receiver any) error {
	cursor, err := collection.FindMany(ctx, filter)
	if err != nil {
		return err
	}
	return cursor.All(ctx, receiver)
}

func (usi *UserServiceImpl) Login(ctx context.Context, username, password string) (Customer, error) {
	var user Customer
	query := bson.D{{"username", username}}
	if err := queryOne(ctx, usi.customers, query, &user); err != nil {
		return Customer{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return Customer{}, errors.New("incorrect password")
	}

	var responsePayload Customer

	if user.CardIds != nil {
		var cards []Card
		query := bson.D{{"id", bson.D{{"$in", user.CardIds}}}}
		if err := queryMany(ctx, usi.cards, query, &cards); err != nil {
			return Customer{}, err
		}

		var updatedCards []Card
		for _, c := range cards {
			c.Href = "GetCardsForCustomer/" + usi.linkDomain + "/" + user.Id
			c.LongNum = usi.maskCard(c.LongNum)
			updatedCards = append(updatedCards, c)
		}

		responsePayload.Cards = updatedCards
	}

	if user.AddressIds != nil {
		var addrs []Address
		query := bson.D{{"id", bson.D{{"$in", user.AddressIds}}}}
		if err := queryMany(ctx, usi.addresses, query, &addrs); err != nil {
			return Customer{}, err
		}

		var updatedAddresses []Address
		for _, a := range addrs {
			a.Href = "GetAddressesForCustomer/" + usi.linkDomain + "/" + user.Id
			updatedAddresses = append(updatedAddresses, a)
		}

		responsePayload.Addresses = updatedAddresses
	}

	return responsePayload, nil
}

func (usi *UserServiceImpl) GetCustomers(ctx context.Context) ([]Customer, error) {
	var customers []Customer
	query := bson.D{{}}
	if err := queryMany(ctx, usi.customers, query, &customers); err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, errors.New("no customers found!")
	}

	var updatedCustomers []Customer
	for _, cus := range customers {

		if len(cus.CardIds) != 0 {
			var cards []Card
			query := bson.D{{"id", bson.D{{"$in", cus.CardIds}}}}
			if err := queryMany(ctx, usi.cards, query, &cards); err != nil {
				return nil, err
			}
			cus.CardIds = nil
			cus.Cards = cards
		}

		if len(cus.AddressIds) != 0 {
			var addrs []Address
			query := bson.D{{"id", bson.D{{"$in", cus.AddressIds}}}}
			if err := queryMany(ctx, usi.addresses, query, &addrs); err != nil {
				return nil, err
			}
			cus.AddressIds = nil
			cus.Addresses = addrs
		}

		updatedCustomers = append(updatedCustomers, cus)
	}

	return updatedCustomers, nil
}

func (usi *UserServiceImpl) AddCustomer(ctx context.Context, customer Customer, cards []Card, addresses []Address) (string, error) {
	collection := usi.db.GetDatabase("user").GetCollection("customers")

	query := fmt.Sprintf(`{"Username": %s }`, customer.Username)

	result, err := collection.FindOne(query)

	if err != nil {
		return "", err
	}

	var user Customer
	result.Decode(&user)

	if user.Id != "" {
		return "", errors.New("User already exists")
	}

	if cards != nil {
		cardCollection := usi.db.GetDatabase("user").GetCollection("cards")

		var cardsToInsert []interface{}
		for _, c := range cards {

			cardsToInsert = append(cardsToInsert, c)
		}

		err := cardCollection.InsertMany(cardsToInsert)
		if err != nil {
			return "", err
		}
	}

	if addresses != nil {
		addrCollection := usi.db.GetDatabase("user").GetCollection("addresses")

		var addrsToInsert []interface{}
		for _, a := range addresses {

			addrsToInsert = append(addrsToInsert, a)
		}

		err := addrCollection.InsertMany(addrsToInsert)
		if err != nil {
			return "", err
		}
	}

	customer.Id = uuid.New().String()
	customer.Password = usi.prepareSaltedHash(customer.Password)
	customer.Addresses = addresses
	customer.Cards = cards
	err = collection.InsertOne(customer)

	if err != nil {
		return "", err
	}

	return customer.Id, nil
}

func (usi *UserServiceImpl) DeleteCustomer(ctx context.Context, customerId string) (string, error) {
	collection := usi.db.GetDatabase("user").GetCollection("customers")

	query := fmt.Sprintf(`{"Id": %s }`, customerId)

	err := collection.DeleteOne(query)

	if err != nil {
		return "", err
	}

	return "Customer removed successfully", nil
}

func (usi *UserServiceImpl) GetCustomerById(ctx context.Context, customerId string) (Customer, error) {
	collection := usi.db.GetDatabase("user").GetCollection("customers")

	query := fmt.Sprintf(`{"Id": %s }`, customerId)

	result, err := collection.FindOne(query)

	if err != nil {
		return Customer{}, err
	}

	var user Customer
	result.Decode(&user)

	if user.Id == "" || err != nil {
		return Customer{}, errors.New("Could not fetch customer")
	}

	var cards []Card
	var addrs []Address

	cardCollection := usi.db.GetDatabase("user").GetCollection("cards")
	query = fmt.Sprintf(`{"Id": {"$in": [1]%v}`, user.CardIds)

	res, err := cardCollection.FindMany(query)
	if err != nil {
		return Customer{}, err
	}
	res.All(&cards)

	if len(cards) != 0 {
		user.Cards = cards
		user.CardIds = nil
	}

	addrCollection := usi.db.GetDatabase("user").GetCollection("addresses")

	res2, err := addrCollection.FindMany(query)
	if err != nil {
		return Customer{}, err
	}
	res2.All(&addrs)

	if len(addrs) != 0 {
		user.Addresses = addrs
		user.AddressIds = nil
	}

	return user, nil
}

//----------------------------------------------------------------------

func (usi *UserServiceImpl) GetCards(ctx context.Context) ([]Card, error) {
	collection := usi.db.GetDatabase("user").GetCollection("cards")

	var cards []Card

	result, err := collection.FindMany("")

	if err != nil {
		return nil, err
	}

	result.All(&cards)

	if len(cards) == 0 {
		return nil, errors.New("No cards found!")
	}

	updatedCards := []Card{}
	for _, c := range cards {

		c.Link = map[string]string{"Href": "GetCard/" + usi.linkDomain + "/" + c.Id}
		updatedCards = append(updatedCards, c)
	}

	return updatedCards, nil
}

func (usi *UserServiceImpl) GetCard(ctx context.Context, cardId string) (Card, error) {
	collection := usi.db.GetDatabase("user").GetCollection("cards")

	var card Card

	query := fmt.Sprintf(`{"Id": %s }`, cardId)

	result, err := collection.FindOne(query)

	if err != nil {
		return Card{}, err
	}

	err = result.Decode(&card)
	if err != nil {
		return Card{}, err
	}

	//TODO check the app logic about this, cause it doesnt make much sense
	//apparently the original impl did it like this :shrug:
	// card["links"] = map[string]string{
	// 	"href": "GetCardsForCustomer/" + usi.linkDomain + "/" + card["id"].(string)
	// }

	return card, nil
}

func (usi *UserServiceImpl) AddCard(ctx context.Context, card Card) (string, error) {
	collection := usi.db.GetDatabase("user").GetCollection("cards")

	query := fmt.Sprintf(`{"LongNum": %s }`, card.LongNum)

	result, err := collection.FindOne(query)

	if err != nil {
		return "", err
	}

	result.Decode(&card)

	if card.Id != "" {
		return "", errors.New("Card already exists")
	}

	card.Id = uuid.New().String()
	err = collection.InsertOne(card)

	if err != nil {
		return "", err
	}

	return card.Id, nil

}

func (usi *UserServiceImpl) DeleteCard(ctx context.Context, cardId string) (string, error) {
	collection := usi.db.GetDatabase("user").GetCollection("cards")

	query := fmt.Sprintf(`{"Id": %s }`, cardId)

	err := collection.DeleteOne(query)

	if err != nil {
		return "", err
	}

	return "Card removed successfully", nil
}

func (usi *UserServiceImpl) GetCardsForCustomer(ctx context.Context, customerId string) ([]Card, error) {

	collection := usi.db.GetDatabase("user").GetCollection("customers")

	query := fmt.Sprintf(`{"Id": %s }`, customerId)

	result, err := collection.FindOne(query)

	if err != nil {
		return nil, err
	}

	var user Customer
	result.Decode(&user)

	if user.Id == "" {
		return nil, errors.New("Could not fetch customer")
	}

	cardIds := user.CardIds

	cardCollection := usi.db.GetDatabase("user").GetCollection("cards")
	query = fmt.Sprintf(`{"Id": {"$in": [1]%v}`, cardIds)
	res, err := cardCollection.FindMany(query)

	if err != nil {
		return nil, err
	}

	var cards []Card

	res.All(&cards)

	updatedCards := []Card{}

	for _, c := range cards {

		c.Link = map[string]string{
			"Href": "GetCardsForCustomer/" + usi.linkDomain + "/" + customerId,
		}

		updatedCards = append(updatedCards, c)
	}

	return updatedCards, nil
}

//----------------------------------------------------------------------

func (usi *UserServiceImpl) GetAddress(ctx context.Context, addressId string) (Address, error) {
	collection := usi.db.GetDatabase("user").GetCollection("addresses")

	var addr Address

	query := fmt.Sprintf(`{"Id": %s }`, addressId)

	result, err := collection.FindOne(query)

	if err != nil {
		return Address{}, err
	}
	err = result.Decode(&addr)

	if err != nil {
		return Address{}, err
	}

	//TODO check the app logic about this, cause it doesnt make much sense
	//apparently the original impl did it like this :shrug:
	// addr["links"] = map[string]string{
	// 	"href": "GetCardsForCustomer/" + usi.linkDomain + "/" + addr["id"].(string) // we actually need customerId here
	// }

	return addr, nil
}

func (usi *UserServiceImpl) GetAddressesForCustomer(ctx context.Context, customerId string) ([]Address, error) {
	collection := usi.db.GetDatabase("user").GetCollection("customers")

	query := fmt.Sprintf(`{"Id": %s }`, customerId)

	result, err := collection.FindOne(query)

	if err != nil {
		return nil, err
	}

	var user Customer
	err = result.Decode(&user)

	if err != nil {
		return nil, err
	}

	addrIds := user.AddressIds

	addressCollection := usi.db.GetDatabase("user").GetCollection("addresses")
	query = fmt.Sprintf(`{"Id": {"$in": [1]%v}`, addrIds)

	res, err := addressCollection.FindMany(query)

	if err != nil {
		return nil, err
	}

	var addrs []Address

	res.All(&addrs)

	updatedAddrs := []Address{}

	for _, a := range addrs {

		a.Link = map[string]string{
			"href": "GetAddressesForCustomer/" + usi.linkDomain + "/" + customerId,
		}

		updatedAddrs = append(updatedAddrs, a)
	}

	return updatedAddrs, nil
}

func (usi *UserServiceImpl) GetAddresses(ctx context.Context) ([]Address, error) {
	collection := usi.db.GetDatabase("user").GetCollection("addresses")

	var addresses []Address
	res, err := collection.FindMany("")

	if err != nil {
		return nil, err
	}

	err = res.All(&addresses)

	if err != nil {
		return nil, err
	}

	updatedAddresses := []Address{}
	for _, a := range addresses {

		a.Link = map[string]string{"Href": "GetAddress/" + usi.linkDomain + "/" + a.Id}
		updatedAddresses = append(updatedAddresses, a)
	}

	return updatedAddresses, nil
}

func (usi *UserServiceImpl) DeleteAddress(ctx context.Context, customerId string, addressId string) (string, error) {
	collection := usi.db.GetDatabase("user").GetCollection("addresses")

	query := fmt.Sprintf(`{"Id": %s }`, addressId)

	err := collection.DeleteOne(query)

	if err != nil {
		return "", err
	}

	return "Address removed successfully", nil
}

func (usi *UserServiceImpl) AddAddress(ctx context.Context, customerId string, address Address) (string, error) {
	collection := usi.db.GetDatabase("user").GetCollection("addresses")

	address.Id = uuid.New().String()
	err := collection.InsertOne(address)

	if err != nil {
		return "", err
	}

	return address.Id, nil
}
