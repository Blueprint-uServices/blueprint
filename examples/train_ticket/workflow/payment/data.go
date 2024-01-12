package payment

type Payment struct {
	ID      string
	OrderID string
	UserID  string
	Price   string
	Type    string
}

type Money struct {
	ID     string
	UserID string
	Price  string
	Type   string
}
