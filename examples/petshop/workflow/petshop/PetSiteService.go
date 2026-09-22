package petshop

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type PetSiteService interface {
	Home(ctx context.Context, userID string, query PetSearchQuery) (PetShopHome, error)
	Adopt(ctx context.Context, userID, petType, petID string) (Adoption, error)
	AdoptionHistory(ctx context.Context, userID string) ([]AdoptionDetails, error)
	AddFoodToCart(ctx context.Context, userID, foodID string, quantity uint32) (Cart, error)
	CheckoutFood(ctx context.Context, userID string, request CheckoutRequest) (FoodOrder, error)
	RecommendFood(ctx context.Context, userID, prompt, petID string) ([]Recommendation, error)
}

type PetSiteServiceImpl struct {
	petSearch PetSearchService
	adoption  PayForAdoptionService
	history   PetListAdoptionsService
	petFood   PetFoodService
	foodAgent PetFoodAgentService
	logger    backend.Logger
}

func NewPetSiteServiceImpl(ctx context.Context, petSearch PetSearchService, adoption PayForAdoptionService, history PetListAdoptionsService, petFood PetFoodService, foodAgent PetFoodAgentService) (PetSiteService, error) {
	return &PetSiteServiceImpl{petSearch: petSearch, adoption: adoption, history: history, petFood: petFood, foodAgent: foodAgent, logger: backend.GetLogger()}, nil
}

func (s *PetSiteServiceImpl) Home(ctx context.Context, userID string, query PetSearchQuery) (PetShopHome, error) {
	query.UserID = userID
	pets, err := s.petSearch.Search(ctx, query)
	if err != nil {
		return PetShopHome{}, err
	}
	foods, err := s.petFood.ListFoods(ctx, FoodFilters{InStockOnly: true})
	if err != nil {
		return PetShopHome{}, err
	}
	adoptions, err := s.history.ListAdoptions(ctx, userID, 25)
	if err != nil {
		return PetShopHome{}, err
	}
	cart, err := s.petFood.GetCart(ctx, userID)
	if err != nil {
		return PetShopHome{}, err
	}
	return PetShopHome{Pets: pets, Foods: foods, Adoptions: adoptions, Cart: cart, SearchQuery: query}, nil
}

func (s *PetSiteServiceImpl) Adopt(ctx context.Context, userID, petType, petID string) (Adoption, error) {
	return s.adoption.CompleteAdoption(ctx, petID, petType, userID)
}

func (s *PetSiteServiceImpl) AdoptionHistory(ctx context.Context, userID string) ([]AdoptionDetails, error) {
	return s.history.ListAdoptions(ctx, userID, 25)
}

func (s *PetSiteServiceImpl) AddFoodToCart(ctx context.Context, userID, foodID string, quantity uint32) (Cart, error) {
	return s.petFood.AddCartItem(ctx, userID, foodID, quantity)
}

func (s *PetSiteServiceImpl) CheckoutFood(ctx context.Context, userID string, request CheckoutRequest) (FoodOrder, error) {
	return s.petFood.Checkout(ctx, userID, request)
}

func (s *PetSiteServiceImpl) RecommendFood(ctx context.Context, userID, prompt, petID string) ([]Recommendation, error) {
	return s.foodAgent.Recommend(ctx, userID, prompt, petID)
}
