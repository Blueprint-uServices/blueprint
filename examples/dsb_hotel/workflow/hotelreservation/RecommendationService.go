package hotelreservation

import (
	"context"
	"math"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/hailocab/go-geoindex"
	"go.mongodb.org/mongo-driver/bson"
)

// RecommendationService implements Recommendation Service from the hotel reservation application
type RecommendationService interface {
	// Returns the recommended hotels based on the desired location (`lat`, `lon`) and the metric (`require`) for ranking recommendations
	GetRecommendations(ctx context.Context, require string, lat float64, lon float64) ([]string, error)
}

// Implements RecommendationService
type RecommendationServiceImpl struct {
	recommendDB backend.NoSQLDatabase
	hotels      map[string]Hotel
}

func initRecommendationDB(ctx context.Context, db backend.NoSQLDatabase) error {
	c, err := db.GetCollection(ctx, "recommendation-db", "recommendation")
	if err != nil {
		return err
	}
	err = c.InsertOne(ctx, &Hotel{"1", 37.7867, -122.4112, 109.00, 150.00})
	if err != nil {
		return err
	}
	err = c.InsertOne(ctx, &Hotel{"2", 37.7854, -122.4005, 139.00, 120.00})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &Hotel{"3", 37.7834, -122.4071, 109.00, 190.00})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &Hotel{"4", 37.7936, -122.3930, 129.00, 160.00})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &Hotel{"5", 37.7831, -122.4181, 119.00, 140.00})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &Hotel{"6", 37.7863, -122.4015, 149.00, 200.00})
	if err != nil {
		return err
	}

	// add up to 80 hotels
	for i := 7; i <= 80; i++ {
		hotel_id := strconv.Itoa(i)
		lat := 37.7835 + float64(i)/500.0*3
		lon := -122.41 + float64(i)/500.0*4

		rate := 135.00
		rate_inc := 179.00
		if i%3 == 0 {
			if i%5 == 0 {
				rate = 109.00
				rate_inc = 123.17
			} else if i%5 == 1 {
				rate = 120.00
				rate_inc = 140.00
			} else if i%5 == 2 {
				rate = 124.00
				rate_inc = 144.00
			} else if i%5 == 3 {
				rate = 132.00
				rate_inc = 158.00
			} else if i%5 == 4 {
				rate = 232.00
				rate_inc = 258.00
			}
		}

		err = c.InsertOne(ctx, &Hotel{hotel_id, lat, lon, rate, rate_inc})
		if err != nil {
			return err
		}
	}
	return nil
}

// Creates and Returns a new RecommendationService object
func NewRecommendationServiceImpl(ctx context.Context, recommendDB backend.NoSQLDatabase) (RecommendationService, error) {
	service := &RecommendationServiceImpl{recommendDB: recommendDB, hotels: make(map[string]Hotel)}
	err := initRecommendationDB(ctx, recommendDB)
	if err != nil {
		return nil, err
	}
	err = service.LoadRecommendations(context.Background())
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (r *RecommendationServiceImpl) LoadRecommendations(ctx context.Context) error {
	collection, err := r.recommendDB.GetCollection(ctx, "recommendation-db", "recommendation")
	if err != nil {
		return err
	}

	filter := bson.D{}
	res, err := collection.FindMany(ctx, filter)
	if err != nil {
		return err
	}
	var hotels []Hotel
	res.All(ctx, &hotels)
	for _, hotel := range hotels {
		r.hotels[hotel.HId] = hotel
	}

	return nil
}

func (r *RecommendationServiceImpl) GetRecommendations(ctx context.Context, require string, lat float64, lon float64) ([]string, error) {

	var hotelIds []string
	if require == "dis" {
		p1 := &geoindex.GeoPoint{Pid: "", Plat: lat, Plon: lon}
		min := math.MaxFloat64
		dist := make(map[string]float64)
		for _, hotel := range r.hotels {
			hotel_pt := &geoindex.GeoPoint{Pid: "", Plat: hotel.HLat, Plon: hotel.HLon}
			tmp := float64(geoindex.Distance(p1, hotel_pt)) / 1000
			if tmp < min {
				min = tmp
			}
			dist[hotel.HId] = tmp
		}
		for _, hotel := range r.hotels {
			distance := dist[hotel.HId]
			if distance == min {
				hotelIds = append(hotelIds, hotel.HId)
			}
		}
	} else if require == "rate" {
		max := 0.0
		rates := make(map[string]float64)
		for _, hotel := range r.hotels {
			if hotel.HRate > max {
				max = hotel.HRate
			}
			rates[hotel.HId] = hotel.HRate
		}
		for _, hotel := range r.hotels {
			rate := rates[hotel.HId]
			if rate == max {
				hotelIds = append(hotelIds, hotel.HId)
			}
		}
	} else if require == "price" {
		min := math.MaxFloat64
		prices := make(map[string]float64)
		for _, hotel := range r.hotels {
			if hotel.HPrice < min {
				min = hotel.HPrice
			}
			prices[hotel.HId] = hotel.HPrice
		}
		for hid, price := range prices {
			if min == price {
				hotelIds = append(hotelIds, hid)
			}
		}
	}

	return hotelIds, nil
}
