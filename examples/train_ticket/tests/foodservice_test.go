package tests

import (
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
	"github.com/stretchr/testify/require"
)

func genTestFoodData() []common.Food {
	res := []common.Food{}
	for i := 0; i < 10; i++ {
		f := common.Food{
			Name:  fmt.Sprintf("Food%d", i),
			Price: float64(100*i + 100),
		}
		res = append(res, f)
	}
	return res
}

func requireFood(t *testing.T, expected common.Food, actual common.Food) {
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Price, actual.Price)
}
