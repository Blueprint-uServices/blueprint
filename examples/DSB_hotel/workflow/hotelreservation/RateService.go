package hotelreservation

import (
	"context"
	"log"
	"strconv"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type RateService interface {
	GetRates(ctx context.Context, hotelIDs []string, inDate string, outDate string) ([]RatePlan, error)
}

type RateServiceImpl struct {
	rateCache backend.Cache
	rateDB    backend.NoSQLDatabase
}

func initRateDB(ctx context.Context, db backend.NoSQLDatabase) error {
	c, err := db.GetCollection(ctx, "rate-db", "inventory")
	if err != nil {
		return err
	}
	err = c.InsertOne(ctx, &RatePlan{
		"1",
		"RACK",
		"2015-04-09",
		"2015-04-10",
		RoomType{
			109.00,
			"KNG",
			"King sized bed",
			109.00,
			123.17}})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &RatePlan{
		"2",
		"RACK",
		"2015-04-09",
		"2015-04-10",
		RoomType{
			139.00,
			"QN",
			"Queen sized bed",
			139.00,
			153.09}})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &RatePlan{
		"3",
		"RACK",
		"2015-04-09",
		"2015-04-10",
		RoomType{
			109.00,
			"KNG",
			"King sized bed",
			109.00,
			123.17}})
	if err != nil {
		return err
	}

	// add up to 80 hotels
	for i := 7; i <= 80; i++ {
		if i%3 == 0 {
			hotel_id := strconv.Itoa(i)
			end_date := "2015-04-"
			rate := 109.00
			rate_inc := 123.17
			if i%2 == 0 {
				end_date = end_date + "17"
			} else {
				end_date = end_date + "24"
			}

			if i%5 == 1 {
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

			err = c.InsertOne(ctx, &RatePlan{
				hotel_id,
				"RACK",
				"2015-04-09",
				end_date,
				RoomType{
					rate,
					"KNG",
					"King sized bed",
					rate,
					rate_inc}})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func NewRateServiceImpl(ctx context.Context, rateCache backend.Cache, rateDB backend.NoSQLDatabase) *RateServiceImpl {
	err := initRateDB(ctx, rateDB)
	if err != nil {
		log.Fatal(err)
	}
	return &RateServiceImpl{rateCache: rateCache, rateDB: rateDB}
}

func (r *RateServiceImpl) GetRates(ctx context.Context, hotelIDs []string, inDate string, outDate string) ([]RatePlan, error) {
	var rate_plans []RatePlan

	for _, hotel_id := range hotelIDs {
		var hotel_rate_plans []RatePlan
		err := r.rateCache.Get(ctx, hotel_id, &hotel_rate_plans)
		if err != nil {
			collection, err2 := r.rateDB.GetCollection(ctx, "rate-db", "inventory")
			if err2 != nil {
				return []RatePlan{}, err2
			}
			query := bson.D{{"HotelID", hotel_id}}
			rs, err := collection.FindMany(ctx, query)
			if err != nil {
				return rate_plans, err
			}
			rs.All(ctx, &hotel_rate_plans)
			err = r.rateCache.Put(ctx, hotel_id, hotel_rate_plans)
		}
		rate_plans = append(rate_plans, hotel_rate_plans...)
	}
	// TODO: Sort rate_plans
	return rate_plans, nil
}