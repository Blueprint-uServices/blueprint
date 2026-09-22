package petshop

import "time"

type Pet struct {
	PetID        string `json:"petid" bson:"petid"`
	Availability string `json:"availability" bson:"availability"`
	CutenessRate string `json:"cuteness_rate" bson:"cuteness_rate"`
	Image        string `json:"image,omitempty" bson:"image"`
	PetColor     string `json:"petcolor" bson:"petcolor"`
	PetType      string `json:"pettype" bson:"pettype"`
	Price        string `json:"price" bson:"price"`
	PetURL       string `json:"peturl" bson:"peturl,omitempty"`
}

type PetSearchQuery struct {
	PetType  string
	PetColor string
	PetID    string
	UserID   string
}

type Adoption struct {
	TransactionID string    `json:"transactionid" bson:"transactionid"`
	PetID         string    `json:"petid" bson:"petid"`
	PetType       string    `json:"pettype" bson:"pettype"`
	UserID        string    `json:"userid" bson:"userid"`
	AdoptionDate  time.Time `json:"adoptiondate" bson:"adoptiondate"`
}

type AdoptionDetails struct {
	Adoption
	Pet Pet `json:"pet" bson:"pet"`
}

type FoodFilters struct {
	PetType     string
	FoodType    string
	MinPrice    float64
	MaxPrice    float64
	SearchTerm  string
	InStockOnly bool
}

type Food struct {
	ID                 string    `json:"id" bson:"id"`
	PetType            string    `json:"pet_type" bson:"pet_type"`
	Name               string    `json:"name" bson:"name"`
	FoodType           string    `json:"food_type" bson:"food_type"`
	Description        string    `json:"description" bson:"description"`
	Price              float64   `json:"price" bson:"price"`
	Image              string    `json:"image" bson:"image"`
	Ingredients        []string  `json:"ingredients" bson:"ingredients"`
	FeedingGuidelines  string    `json:"feeding_guidelines" bson:"feeding_guidelines"`
	AvailabilityStatus string    `json:"availability_status" bson:"availability_status"`
	StockQuantity      uint32    `json:"stock_quantity" bson:"stock_quantity"`
	CreatedAt          time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" bson:"updated_at"`
	IsActive           bool      `json:"is_active" bson:"is_active"`
}

type CreateFoodRequest struct {
	PetType           string
	Name              string
	FoodType          string
	Description       string
	Price             float64
	Image             string
	Ingredients       []string
	FeedingGuidelines string
	StockQuantity     uint32
}

type UpdateFoodRequest struct {
	Name               *string
	Description        *string
	Price              *float64
	Image              *string
	Ingredients        []string
	FeedingGuidelines  *string
	AvailabilityStatus *string
	StockQuantity      *uint32
}

type CartItem struct {
	FoodID    string    `json:"food_id"`
	Quantity  uint32    `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	AddedAt   time.Time `json:"added_at"`
}

type Cart struct {
	UserID    string     `json:"user_id"`
	Items     []CartItem `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Address struct {
	Name, Street, City, State, ZipCode, Country string
}

type CheckoutRequest struct {
	PaymentMethod   string
	ShippingAddress Address
	BillingAddress  Address
}

type FoodOrder struct {
	OrderID           string
	UserID            string
	Items             []CartItem
	Subtotal          float64
	Tax               float64
	Shipping          float64
	TotalAmount       float64
	PaymentMethod     string
	Status            string
	CreatedAt         time.Time
	EstimatedDelivery time.Time
}

type Recommendation struct {
	Food   Food
	Reason string
}

type PetShopHome struct {
	Pets        []Pet
	Foods       []Food
	Adoptions   []AdoptionDetails
	Cart        Cart
	SearchQuery PetSearchQuery
}
