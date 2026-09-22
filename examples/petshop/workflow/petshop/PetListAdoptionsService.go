package petshop

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type PetListAdoptionsService interface {
	ListAdoptions(ctx context.Context, userID string, limit int) ([]AdoptionDetails, error)
}

type PetListAdoptionsServiceImpl struct {
	transactions backend.NoSQLCollection
	petSearch    PetSearchService
	logger       backend.Logger
}

func NewPetListAdoptionsServiceImpl(ctx context.Context, db backend.NoSQLDatabase, petSearch PetSearchService) (PetListAdoptionsService, error) {
	transactions, err := db.GetCollection(ctx, "petshop", "adoptions")
	if err != nil {
		return nil, err
	}
	return &PetListAdoptionsServiceImpl{transactions: transactions, petSearch: petSearch, logger: backend.GetLogger()}, nil
}

func (s *PetListAdoptionsServiceImpl) ListAdoptions(ctx context.Context, userID string, limit int) ([]AdoptionDetails, error) {
	filter := bson.D{}
	if userID != "" {
		filter = bson.D{{Key: "userid", Value: userID}}
	}
	cursor, err := s.transactions.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	var adoptions []Adoption
	if err := cursor.All(ctx, &adoptions); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 25 {
		limit = 25
	}
	if len(adoptions) > limit {
		adoptions = adoptions[len(adoptions)-limit:]
	}
	out := make([]AdoptionDetails, 0, len(adoptions))
	for i := len(adoptions) - 1; i >= 0; i-- {
		pet, err := s.petSearch.GetPet(ctx, adoptions[i].PetType, adoptions[i].PetID)
		if err != nil {
			ctx, _ = s.logger.Warn(ctx, "unable to enrich adoption %s: %v", adoptions[i].TransactionID, err)
			continue
		}
		out = append(out, AdoptionDetails{Adoption: adoptions[i], Pet: pet})
	}
	return out, nil
}
