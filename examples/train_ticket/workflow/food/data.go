package food

type Food struct {
	Name  string
	Price float64
}

type FoodOrder struct {
	ID          string
	OrderID     string
	FoodType    int64
	StationName string
	StoreName   string
	FoodName    string
	Price       float64
}
