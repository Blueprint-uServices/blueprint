package user

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type cardStore struct {
	c backend.NoSQLCollection
}

// The format of a card stored in the database
type dbCard struct {
	Card `bson:",inline"`
	ID   primitive.ObjectID `bson:"_id"`
}

func newCardStore(ctx context.Context, db backend.NoSQLDatabase) (*cardStore, error) {
	c, err := db.GetCollection(ctx, "userservice", "cards")
	return &cardStore{c: c}, err
}

// Gets card by objects Id
func (s *cardStore) getCard(ctx context.Context, cardid string) (Card, error) {
	// Convert the card ID
	id, err := primitive.ObjectIDFromHex(cardid)
	if err != nil {
		return Card{}, errors.New("Invalid ID Hex")
	}

	// Run the query
	cursor, err := s.c.FindOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return Card{}, err
	}
	card := dbCard{}
	_, err = cursor.One(ctx, &card)

	// Convert from DB card data to Card object
	card.Card.ID = card.ID.Hex()
	return card.Card, err
}

// Gets cards from the card store
func (s *cardStore) getCards(ctx context.Context, cardIds []string) ([]Card, error) {
	if len(cardIds) == 0 {
		return nil, nil
	}

	// Convert the card IDs from hex strings to objects
	ids, err := hexToObjectIds(cardIds)
	if err != nil {
		return nil, err
	}

	// Run the query
	cursor, err := s.c.FindMany(ctx, bson.D{{"_id", bson.D{{"$in", ids}}}})
	if err != nil {
		return nil, err
	}
	dbCards := make([]dbCard, 0, len(cardIds))
	err = cursor.All(ctx, &dbCards)

	// Convert from DB card data to Card objects
	cards := make([]Card, 0, len(dbCards))
	for _, card := range dbCards {
		card.Card.ID = card.ID.Hex()
		cards = append(cards, card.Card)
	}

	return cards, err
}

func (s *cardStore) getAllCards(ctx context.Context) ([]Card, error) {
	// Run the query
	cursor, err := s.c.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	dbCards := make([]dbCard, 0)
	err = cursor.All(ctx, &dbCards)

	// Convert from DB card data to Card objects
	cards := make([]Card, 0, len(dbCards))
	for _, card := range dbCards {
		card.Card.ID = card.ID.Hex()
		cards = append(cards, card.Card)
	}

	return cards, err
}

// Adds a card to the cards DB
func (s *cardStore) createCard(ctx context.Context, card *Card) (primitive.ObjectID, error) {
	// Create and insert to DB
	dbcard := dbCard{Card: *card, ID: primitive.NewObjectID()}
	if _, err := s.c.UpsertID(ctx, dbcard.ID, dbcard); err != nil {
		return dbcard.ID, err
	}

	// Update the provided card
	dbcard.Card.ID = dbcard.ID.Hex()
	*card = dbcard.Card
	return dbcard.ID, nil
}

// Creates or updates the provided cards in the cardStore.
func (s *cardStore) createCards(ctx context.Context, cards []Card) ([]primitive.ObjectID, error) {
	if len(cards) == 0 {
		return []primitive.ObjectID{}, nil
	}
	createdIds := make([]primitive.ObjectID, 0)
	for _, card := range cards {
		toInsert := dbCard{
			Card: card,
			ID:   primitive.NewObjectID(),
		}
		_, err := s.c.UpsertID(ctx, toInsert.ID, toInsert)
		if err != nil {
			return createdIds, err
		}
		createdIds = append(createdIds, toInsert.ID)
	}

	return createdIds, nil
}

func (s *cardStore) removeCard(ctx context.Context, cardid string) error {
	// Convert the card ID
	id, err := primitive.ObjectIDFromHex(cardid)
	if err != nil {
		return errors.New("Invalid ID Hex")
	}
	return s.removeCards(ctx, []primitive.ObjectID{id})
}

// Removes all specified cards from the DB
func (s *cardStore) removeCards(ctx context.Context, ids []primitive.ObjectID) error {
	return s.c.DeleteMany(ctx, bson.D{{"_id", bson.D{{"$in", ids}}}})
}
