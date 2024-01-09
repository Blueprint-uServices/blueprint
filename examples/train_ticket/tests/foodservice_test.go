package tests

import (
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/food"
	"github.com/stretchr/testify/require"
)

func genTestFoodData() []food.Food {
	res := []food.Food{}
	for i := 0; i < 10; i++ {
		f := food.Food{
			Name:  fmt.Sprintf("Food%d", i),
			Price: float64(100*i + 100),
		}
		res = append(res, f)
	}
	return res
}

func requireFood(t *testing.T, expected food.Food, actual food.Food) {
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Price, actual.Price)
}
