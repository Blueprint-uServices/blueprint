package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/petshop/workflow/petshop"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
)

var petFoodRegistry = registry.NewServiceRegistry[petshop.PetFoodService]("pet_food_service")

func init() {
	petFoodRegistry.Register("local", func(ctx context.Context) (petshop.PetFoodService, error) {
		db, cache, _, err := localBackends(ctx)
		if err != nil {
			return nil, err
		}
		return petshop.NewPetFoodServiceImpl(ctx, db, cache)
	})
}

func TestPetFoodService(t *testing.T) {
	ctx := context.Background()
	service, err := petFoodRegistry.Get(ctx)
	require.NoError(t, err)
	foods, err := service.ListFoods(ctx, petshop.FoodFilters{PetType: "puppy", InStockOnly: true})
	require.NoError(t, err)
	require.Len(t, foods, 2)
	cart, err := service.AddCartItem(ctx, "food-user", "F12345678", 2)
	require.NoError(t, err)
	require.Len(t, cart.Items, 1)
	order, err := service.Checkout(ctx, "food-user", petshop.CheckoutRequest{PaymentMethod: "CreditCard"})
	require.NoError(t, err)
	require.Equal(t, "confirmed", order.Status)
	require.Equal(t, 25.98, order.Subtotal)
}
