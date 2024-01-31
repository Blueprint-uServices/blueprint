package fooddelivery

type FoodDeliveryOrder struct {
	ID                 string
	StationFoodStoreID string
	FoodList           []string
	TripID             string
	SeatNum            int64
	CreatedTime        string
	DeliveryTime       string
	DeliveryFee        float64
}

type DeliveryInfo struct {
	DeliveryTime string
	OrderID      string
}

type TripOrderInfo struct {
	OrderID string
	TripID  string
}

type SeatInfo struct {
	OrderID string
	SeatNum int64
}
