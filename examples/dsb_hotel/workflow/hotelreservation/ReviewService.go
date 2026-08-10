package hotelreservation

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// ReviewService implements the Review Service from the hotel reservation application
type ReviewService interface {
	GetReviews(ctx context.Context, hotelid string) ([]Review, error)
}

type ReviewServiceImpl struct {
	ReviewCache backend.Cache
	ReviewDB    backend.NoSQLDatabase
}

func initReviewDB(ctx context.Context, db backend.NoSQLDatabase) error {
	c, err := db.GetCollection(ctx, "review-db", "reviews")
	if err != nil {
		return err
	}
	all_reviews := []Review{
		Review{
			"1",
			"1",
			"Person 1",
			3.4,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			Image{
				"some url",
				false}},
		Review{
			"2",
			"1",
			"Person 2",
			4.4,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			Image{
				"some url",
				false}},
		Review{
			"3",
			"1",
			"Person 3",
			4.2,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			Image{
				"some url",
				false}},
		Review{
			"4",
			"1",
			"Person 4",
			3.9,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			Image{
				"some url",
				false}},
		Review{
			"5",
			"2",
			"Person 5",
			4.2,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			Image{
				"some url",
				false}},
		Review{
			"6",
			"2",
			"Person 6",
			3.7,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			Image{
				"some url",
				false}},
	}
	for _, r := range all_reviews {
		err := c.InsertOne(ctx, &r)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewReviewServiceImpl(ctx context.Context, reviewCache backend.Cache, reviewDB backend.NoSQLDatabase) (ReviewService, error) {
	err := initReviewDB(ctx, reviewDB)
	if err != nil {
		return nil, err
	}
	impl := &ReviewServiceImpl{ReviewCache: reviewCache, ReviewDB: reviewDB}
	return impl, nil
}

func (r *ReviewServiceImpl) GetReviews(ctx context.Context, hotelid string) ([]Review, error) {
	var reviews []Review
	exists, err := r.ReviewCache.Get(ctx, hotelid, &reviews)
	if err != nil {
		return reviews, err
	}
	if !exists {
		coll, err := r.ReviewDB.GetCollection(ctx, "review-db", "reviews")
		if err != nil {
			return reviews, err
		}

		cur, err := coll.FindMany(ctx, bson.D{{"hotelid", hotelid}})
		if err != nil {
			return reviews, err
		}
		err = cur.All(ctx, &reviews)
		if err != nil {
			return reviews, err
		}
		err = r.ReviewCache.Put(ctx, hotelid, reviews)
		// Only log the error
		if err != nil {
			backend.GetLogger().Info(ctx, err.Error())
		}
	}
	return reviews, nil
}
