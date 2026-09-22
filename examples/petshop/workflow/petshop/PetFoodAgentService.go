package petshop

import (
	"context"
	"fmt"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type PetFoodAgentService interface {
	Recommend(ctx context.Context, userID, prompt, petID string) ([]Recommendation, error)
}

type PetFoodAgentServiceImpl struct {
	petSearch PetSearchService
	petFood   PetFoodService
	logger    backend.Logger
}

func NewPetFoodAgentServiceImpl(ctx context.Context, petSearch PetSearchService, petFood PetFoodService) (PetFoodAgentService, error) {
	return &PetFoodAgentServiceImpl{petSearch: petSearch, petFood: petFood, logger: backend.GetLogger()}, nil
}

func (s *PetFoodAgentServiceImpl) Recommend(ctx context.Context, userID, prompt, petID string) ([]Recommendation, error) {
	var petType string
	if petID != "" {
		pets, err := s.petSearch.Search(ctx, PetSearchQuery{PetID: petID, UserID: userID})
		if err != nil {
			return nil, err
		}
		if len(pets) == 0 {
			return nil, ErrPetNotFound
		}
		petType = pets[0].PetType
	} else {
		lower := strings.ToLower(prompt)
		for _, candidate := range []string{"puppy", "kitten", "bunny"} {
			if strings.Contains(lower, candidate) {
				petType = candidate
				break
			}
		}
	}
	foods, err := s.petFood.ListFoods(ctx, FoodFilters{PetType: petType, InStockOnly: true})
	if err != nil {
		return nil, err
	}
	if len(foods) > 3 {
		foods = foods[:3]
	}
	recommendations := make([]Recommendation, len(foods))
	for i, food := range foods {
		reason := fmt.Sprintf("%s is formulated for %ss and is currently in stock", food.Name, food.PetType)
		if food.FoodType == "treats" {
			reason = fmt.Sprintf("%s is a useful occasional reward for training", food.Name)
		}
		recommendations[i] = Recommendation{Food: food, Reason: reason}
	}
	ctx, _ = s.logger.Info(ctx, "Waggle recommendation completed: user=%s pet=%s results=%d", userID, petID, len(recommendations))
	return recommendations, nil
}
