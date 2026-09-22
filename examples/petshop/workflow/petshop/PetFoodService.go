package petshop

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrFoodNotFound = errors.New("food not found")
	ErrCartNotFound = errors.New("cart item not found")
)

type PetFoodService interface {
	ListFoods(ctx context.Context, filters FoodFilters) ([]Food, error)
	GetFood(ctx context.Context, foodID string) (Food, error)
	CreateFood(ctx context.Context, request CreateFoodRequest) (Food, error)
	UpdateFood(ctx context.Context, foodID string, request UpdateFoodRequest) (Food, error)
	DeleteFood(ctx context.Context, foodID string) error
	GetCart(ctx context.Context, userID string) (Cart, error)
	AddCartItem(ctx context.Context, userID, foodID string, quantity uint32) (Cart, error)
	UpdateCartItem(ctx context.Context, userID, foodID string, quantity uint32) (Cart, error)
	RemoveCartItem(ctx context.Context, userID, foodID string) error
	ClearCart(ctx context.Context, userID string) error
	Checkout(ctx context.Context, userID string, request CheckoutRequest) (FoodOrder, error)
}

type PetFoodServiceImpl struct {
	foods  backend.NoSQLCollection
	orders backend.NoSQLCollection
	carts  backend.Cache
	logger backend.Logger
	mu     sync.Mutex
}

func NewPetFoodServiceImpl(ctx context.Context, db backend.NoSQLDatabase, carts backend.Cache) (PetFoodService, error) {
	foods, err := db.GetCollection(ctx, "petshop", "foods")
	if err != nil {
		return nil, err
	}
	orders, err := db.GetCollection(ctx, "petshop", "food-orders")
	if err != nil {
		return nil, err
	}
	for _, food := range seedFoods() {
		if _, err := foods.Upsert(ctx, bson.D{{Key: "id", Value: food.ID}}, food); err != nil {
			return nil, err
		}
	}
	return &PetFoodServiceImpl{foods: foods, orders: orders, carts: carts, logger: backend.GetLogger()}, nil
}

func (s *PetFoodServiceImpl) ListFoods(ctx context.Context, filters FoodFilters) ([]Food, error) {
	cursor, err := s.foods.FindMany(ctx, bson.D{{Key: "is_active", Value: true}})
	if err != nil {
		return nil, err
	}
	var foods []Food
	if err := cursor.All(ctx, &foods); err != nil {
		return nil, err
	}
	term := strings.ToLower(filters.SearchTerm)
	out := foods[:0]
	for _, food := range foods {
		if filters.PetType != "" && food.PetType != strings.ToLower(filters.PetType) || filters.FoodType != "" && food.FoodType != strings.ToLower(filters.FoodType) {
			continue
		}
		if filters.MinPrice > 0 && food.Price < filters.MinPrice || filters.MaxPrice > 0 && food.Price > filters.MaxPrice {
			continue
		}
		if filters.InStockOnly && (food.AvailabilityStatus != "in_stock" || food.StockQuantity == 0) {
			continue
		}
		if term != "" && !strings.Contains(strings.ToLower(food.Name), term) && !strings.Contains(strings.ToLower(food.Description), term) {
			continue
		}
		out = append(out, food)
	}
	return out, nil
}

func (s *PetFoodServiceImpl) GetFood(ctx context.Context, foodID string) (Food, error) {
	cursor, err := s.foods.FindOne(ctx, bson.D{{Key: "id", Value: foodID}, {Key: "is_active", Value: true}})
	if err != nil {
		return Food{}, err
	}
	var food Food
	found, err := cursor.One(ctx, &food)
	if err != nil {
		return Food{}, err
	}
	if !found {
		return Food{}, ErrFoodNotFound
	}
	return food, nil
}

func (s *PetFoodServiceImpl) CreateFood(ctx context.Context, request CreateFoodRequest) (Food, error) {
	if err := validateFoodRequest(request); err != nil {
		return Food{}, err
	}
	id, err := randomID()
	if err != nil {
		return Food{}, err
	}
	now := time.Now().UTC()
	food := Food{ID: "F" + strings.ToUpper(id[:8]), PetType: strings.ToLower(request.PetType), Name: request.Name, FoodType: strings.ToLower(request.FoodType), Description: request.Description, Price: request.Price, Image: request.Image, Ingredients: request.Ingredients, FeedingGuidelines: request.FeedingGuidelines, AvailabilityStatus: "in_stock", StockQuantity: request.StockQuantity, CreatedAt: now, UpdatedAt: now, IsActive: true}
	if request.StockQuantity == 0 {
		food.AvailabilityStatus = "out_of_stock"
	}
	return food, s.foods.InsertOne(ctx, food)
}

func (s *PetFoodServiceImpl) UpdateFood(ctx context.Context, foodID string, request UpdateFoodRequest) (Food, error) {
	food, err := s.GetFood(ctx, foodID)
	if err != nil {
		return Food{}, err
	}
	if request.Name != nil {
		food.Name = *request.Name
	}
	if request.Description != nil {
		food.Description = *request.Description
	}
	if request.Price != nil {
		food.Price = *request.Price
	}
	if request.Image != nil {
		food.Image = *request.Image
	}
	if request.Ingredients != nil {
		food.Ingredients = request.Ingredients
	}
	if request.FeedingGuidelines != nil {
		food.FeedingGuidelines = *request.FeedingGuidelines
	}
	if request.AvailabilityStatus != nil {
		food.AvailabilityStatus = *request.AvailabilityStatus
	}
	if request.StockQuantity != nil {
		food.StockQuantity = *request.StockQuantity
		if food.StockQuantity == 0 {
			food.AvailabilityStatus = "out_of_stock"
		} else if food.AvailabilityStatus == "out_of_stock" {
			food.AvailabilityStatus = "in_stock"
		}
	}
	food.UpdatedAt = time.Now().UTC()
	_, err = s.foods.ReplaceOne(ctx, bson.D{{Key: "id", Value: foodID}}, food)
	return food, err
}

func (s *PetFoodServiceImpl) DeleteFood(ctx context.Context, foodID string) error {
	food, err := s.GetFood(ctx, foodID)
	if err != nil {
		return err
	}
	food.IsActive, food.AvailabilityStatus, food.UpdatedAt = false, "discontinued", time.Now().UTC()
	_, err = s.foods.ReplaceOne(ctx, bson.D{{Key: "id", Value: foodID}}, food)
	return err
}

func (s *PetFoodServiceImpl) GetCart(ctx context.Context, userID string) (Cart, error) {
	var cart Cart
	found, err := s.carts.Get(ctx, userID, &cart)
	if err != nil {
		return Cart{}, err
	}
	if !found {
		now := time.Now().UTC()
		return Cart{UserID: userID, Items: []CartItem{}, CreatedAt: now, UpdatedAt: now}, nil
	}
	return cart, nil
}

func (s *PetFoodServiceImpl) AddCartItem(ctx context.Context, userID, foodID string, quantity uint32) (Cart, error) {
	if quantity == 0 {
		return Cart{}, fmt.Errorf("quantity must be positive")
	}
	food, err := s.GetFood(ctx, foodID)
	if err != nil {
		return Cart{}, err
	}
	if food.AvailabilityStatus != "in_stock" || food.StockQuantity < quantity {
		return Cart{}, fmt.Errorf("food is unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return Cart{}, err
	}
	found := false
	for i := range cart.Items {
		if cart.Items[i].FoodID == foodID {
			if cart.Items[i].Quantity+quantity > food.StockQuantity {
				return Cart{}, fmt.Errorf("insufficient stock for %s", foodID)
			}
			cart.Items[i].Quantity += quantity
			found = true
		}
	}
	if !found {
		cart.Items = append(cart.Items, CartItem{FoodID: foodID, Quantity: quantity, UnitPrice: food.Price, AddedAt: time.Now().UTC()})
	}
	cart.UpdatedAt = time.Now().UTC()
	return cart, s.carts.Put(ctx, userID, cart)
}

func (s *PetFoodServiceImpl) UpdateCartItem(ctx context.Context, userID, foodID string, quantity uint32) (Cart, error) {
	if quantity == 0 {
		return Cart{}, s.RemoveCartItem(ctx, userID, foodID)
	}
	food, err := s.GetFood(ctx, foodID)
	if err != nil {
		return Cart{}, err
	}
	if food.AvailabilityStatus != "in_stock" || quantity > food.StockQuantity {
		return Cart{}, fmt.Errorf("insufficient stock for %s", foodID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return Cart{}, err
	}
	found := false
	for i := range cart.Items {
		if cart.Items[i].FoodID == foodID {
			cart.Items[i].Quantity = quantity
			found = true
		}
	}
	if !found {
		return Cart{}, ErrCartNotFound
	}
	cart.UpdatedAt = time.Now().UTC()
	return cart, s.carts.Put(ctx, userID, cart)
}

func (s *PetFoodServiceImpl) RemoveCartItem(ctx context.Context, userID, foodID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}
	out := cart.Items[:0]
	for _, item := range cart.Items {
		if item.FoodID != foodID {
			out = append(out, item)
		}
	}
	if len(out) == len(cart.Items) {
		return ErrCartNotFound
	}
	cart.Items, cart.UpdatedAt = out, time.Now().UTC()
	return s.carts.Put(ctx, userID, cart)
}

func (s *PetFoodServiceImpl) ClearCart(ctx context.Context, userID string) error {
	return s.carts.Delete(ctx, userID)
}

func (s *PetFoodServiceImpl) Checkout(ctx context.Context, userID string, request CheckoutRequest) (FoodOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return FoodOrder{}, err
	}
	if len(cart.Items) == 0 {
		return FoodOrder{}, fmt.Errorf("cart is empty")
	}
	var subtotal float64
	for _, item := range cart.Items {
		food, err := s.GetFood(ctx, item.FoodID)
		if err != nil {
			return FoodOrder{}, err
		}
		if food.StockQuantity < item.Quantity {
			return FoodOrder{}, fmt.Errorf("insufficient stock for %s", item.FoodID)
		}
		subtotal += item.UnitPrice * float64(item.Quantity)
		stock := food.StockQuantity - item.Quantity
		if _, err := s.UpdateFood(ctx, food.ID, UpdateFoodRequest{StockQuantity: &stock}); err != nil {
			return FoodOrder{}, err
		}
	}
	id, err := randomID()
	if err != nil {
		return FoodOrder{}, err
	}
	now := time.Now().UTC()
	tax := roundMoney(subtotal * .09)
	shipping := 5.99
	order := FoodOrder{OrderID: "ORDER-" + strings.ToUpper(id[:12]), UserID: userID, Items: cart.Items, Subtotal: roundMoney(subtotal), Tax: tax, Shipping: shipping, TotalAmount: roundMoney(subtotal + tax + shipping), PaymentMethod: request.PaymentMethod, Status: "confirmed", CreatedAt: now, EstimatedDelivery: now.Add(96 * time.Hour)}
	if err := s.orders.InsertOne(ctx, order); err != nil {
		return FoodOrder{}, err
	}
	if err := s.carts.Delete(ctx, userID); err != nil {
		return FoodOrder{}, err
	}
	ctx, _ = s.logger.Info(ctx, "food order completed: order=%s user=%s", order.OrderID, userID)
	return order, nil
}

func validateFoodRequest(r CreateFoodRequest) error {
	if r.Name == "" || r.Description == "" || r.Price <= 0 {
		return fmt.Errorf("invalid food request")
	}
	if r.PetType != "puppy" && r.PetType != "kitten" && r.PetType != "bunny" {
		return fmt.Errorf("invalid pet type")
	}
	if r.FoodType != "dry" && r.FoodType != "wet" && r.FoodType != "treats" && r.FoodType != "supplements" {
		return fmt.Errorf("invalid food type")
	}
	return nil
}

func roundMoney(value float64) float64 { return math.Round(value*100) / 100 }

func seedFoods() []Food {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	return []Food{
		{ID: "F12345678", PetType: "puppy", Name: "Beef and Turkey Kibbles", FoodType: "dry", Description: "Nutritious food for growing puppies", Price: 12.99, Image: "petfood/beef-turkey-kibbles.jpg", Ingredients: []string{"beef", "turkey", "rice", "vegetables"}, FeedingGuidelines: "Feed 2-3 times daily", AvailabilityStatus: "in_stock", StockQuantity: 50, CreatedAt: now, UpdatedAt: now, IsActive: true},
		{ID: "F23456789", PetType: "kitten", Name: "Salmon Kitten Feast", FoodType: "wet", Description: "Protein-rich salmon food for kittens", Price: 9.49, Image: "petfood/salmon-kitten-feast.jpg", Ingredients: []string{"salmon", "broth", "vitamins"}, FeedingGuidelines: "Feed twice daily", AvailabilityStatus: "in_stock", StockQuantity: 40, CreatedAt: now, UpdatedAt: now, IsActive: true},
		{ID: "F34567890", PetType: "bunny", Name: "Garden Hay Blend", FoodType: "dry", Description: "High-fiber hay and herb blend for bunnies", Price: 8.25, Image: "petfood/garden-hay-blend.jpg", Ingredients: []string{"timothy hay", "orchard grass", "herbs"}, FeedingGuidelines: "Keep available throughout the day", AvailabilityStatus: "in_stock", StockQuantity: 60, CreatedAt: now, UpdatedAt: now, IsActive: true},
		{ID: "F45678901", PetType: "puppy", Name: "Training Treats", FoodType: "treats", Description: "Small chicken treats for puppy training", Price: 6.75, Image: "petfood/training-treats.jpg", Ingredients: []string{"chicken", "oats"}, FeedingGuidelines: "Use as an occasional reward", AvailabilityStatus: "in_stock", StockQuantity: 100, CreatedAt: now, UpdatedAt: now, IsActive: true},
	}
}
