package consign

type Consign struct {
	Id         string
	OrderId    string
	AccountId  string
	HandleDate string
	TargetDate string
	From       string
	To         string
	Consignee  string
	Phone      string
	Weight     float64
	Within     bool
	Price      float64
}
