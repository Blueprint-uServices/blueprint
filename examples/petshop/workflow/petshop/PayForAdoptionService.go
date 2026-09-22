package petshop

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrPetUnavailable  = errors.New("pet is not available for adoption")
	ErrInvalidAdoption = errors.New("pet id, pet type, and user id are required")
)

type PayForAdoptionService interface {
	CompleteAdoption(ctx context.Context, petID, petType, userID string) (Adoption, error)
	CleanupAdoptions(ctx context.Context, userID string) error
}

type PayForAdoptionServiceImpl struct {
	transactions backend.NoSQLCollection
	petSearch    PetSearchService
	historyQueue backend.Queue
	logger       backend.Logger
}

func NewPayForAdoptionServiceImpl(ctx context.Context, db backend.NoSQLDatabase, petSearch PetSearchService, historyQueue backend.Queue) (PayForAdoptionService, error) {
	transactions, err := db.GetCollection(ctx, "petshop", "adoptions")
	if err != nil {
		return nil, err
	}
	return &PayForAdoptionServiceImpl{transactions: transactions, petSearch: petSearch, historyQueue: historyQueue, logger: backend.GetLogger()}, nil
}

func (s *PayForAdoptionServiceImpl) CompleteAdoption(ctx context.Context, petID, petType, userID string) (Adoption, error) {
	petID, petType, userID = strings.TrimSpace(petID), strings.ToLower(strings.TrimSpace(petType)), strings.TrimSpace(userID)
	if petID == "" || petType == "" || userID == "" {
		return Adoption{}, ErrInvalidAdoption
	}
	pet, err := s.petSearch.GetPet(ctx, petType, petID)
	if err != nil {
		return Adoption{}, err
	}
	if pet.Availability != "yes" {
		return Adoption{}, ErrPetUnavailable
	}
	transactionID, err := randomID()
	if err != nil {
		return Adoption{}, err
	}
	adoption := Adoption{TransactionID: transactionID, PetID: petID, PetType: petType, UserID: userID, AdoptionDate: time.Now().UTC()}
	if err := s.transactions.InsertOne(ctx, adoption); err != nil {
		return Adoption{}, err
	}
	if err := s.petSearch.SetAvailability(ctx, petType, petID, "no"); err != nil {
		_ = s.transactions.DeleteOne(ctx, bson.D{{Key: "transactionid", Value: transactionID}})
		return Adoption{}, err
	}
	if s.historyQueue != nil {
		if pushed, err := s.historyQueue.Push(ctx, adoption); err != nil || !pushed {
			ctx, _ = s.logger.Warn(ctx, "adoption history event was not queued: transaction=%s err=%v", transactionID, err)
		}
	}
	ctx, _ = s.logger.Info(ctx, "adoption completed: transaction=%s pet=%s user=%s", transactionID, petID, userID)
	return adoption, nil
}

func (s *PayForAdoptionServiceImpl) CleanupAdoptions(ctx context.Context, userID string) error {
	filter := bson.D{}
	if userID != "" {
		filter = bson.D{{Key: "userid", Value: userID}}
	}
	cursor, err := s.transactions.FindMany(ctx, filter)
	if err != nil {
		return err
	}
	var adoptions []Adoption
	if err := cursor.All(ctx, &adoptions); err != nil {
		return err
	}
	for _, adoption := range adoptions {
		if err := s.petSearch.SetAvailability(ctx, adoption.PetType, adoption.PetID, "yes"); err != nil {
			return err
		}
	}
	return s.transactions.DeleteMany(ctx, filter)
}

func randomID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(bytes[:]), nil
}
