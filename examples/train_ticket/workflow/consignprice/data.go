package consignprice

type ConsignPrice struct {
	ID            string
	Index         int64
	InitialWeight float64
	InitialPrice  float64
	WithinPrice   float64
	BeyondPrice   float64
}
