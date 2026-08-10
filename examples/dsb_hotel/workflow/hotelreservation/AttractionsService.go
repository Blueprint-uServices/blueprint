package hotelreservation

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	geoindex "github.com/hailocab/go-geoindex"
	"go.mongodb.org/mongo-driver/bson"
)

// AttractionsService implements the Attractions Service from the hotel reservation application
type AttractionsService interface {
	NearbyRest(ctx context.Context, lat float64, lon float64) ([]Restaurant, error)
	NearbyMus(ctx context.Context, lat float64, lon float64) ([]Museum, error)
	NearbyCinema(ctx context.Context, lat float64, lon float64) ([]Cinema, error)
}

// AttractionsServiceImpl finds nearby attractions using indexes populated from
// the attractions database.
type AttractionsServiceImpl struct {
	attractionsDB backend.NoSQLDatabase
	restaurants   *geoindex.ClusteringIndex
	museums       *geoindex.ClusteringIndex
	cinemas       *geoindex.ClusteringIndex
}

const (
	attractionsMaxSearchRadius  = 10
	attractionsMaxSearchResults = 5
)

func initAttractionsDB(ctx context.Context, db backend.NoSQLDatabase) error {
	restaurants, err := db.GetCollection(ctx, "attractions-db", "restaurants")
	if err != nil {
		return err
	}
	for _, restaurant := range []Restaurant{
		{"1", 37.7867, -122.4112, "R1", 3.5, "fusion"},
		{"2", 37.7857, -122.4012, "R2", 3.9, "italian"},
		{"3", 37.7847, -122.3912, "R3", 4.5, "sushi"},
		{"4", 37.7862, -122.4212, "R4", 3.2, "sushi"},
		{"5", 37.7839, -122.4052, "R5", 4.9, "fusion"},
		{"6", 37.7831, -122.3812, "R6", 4.1, "american"},
	} {
		if err := restaurants.InsertOne(ctx, &restaurant); err != nil {
			return err
		}
	}

	museums, err := db.GetCollection(ctx, "attractions-db", "museums")
	if err != nil {
		return err
	}
	for _, museum := range []Museum{
		{"1", 35.7867, -122.4112, "M1", "history"},
		{"2", 36.7867, -122.5112, "M2", "history"},
		{"3", 38.7867, -122.4612, "M3", "nature"},
		{"4", 37.7867, -122.4912, "M4", "nature"},
		{"5", 36.9867, -122.4212, "M5", "nature"},
		{"6", 37.3867, -122.5012, "M6", "technology"},
	} {
		if err := museums.InsertOne(ctx, &museum); err != nil {
			return err
		}
	}

	// The upstream data set does not seed cinemas, but create the collection so
	// deployments can populate it and use NearbyCinema without code changes.
	_, err = db.GetCollection(ctx, "attractions-db", "cinemas")
	return err
}

// NewAttractionsServiceImpl creates an AttractionsService backed by a NoSQL database.
func NewAttractionsServiceImpl(ctx context.Context, attractionsDB backend.NoSQLDatabase) (AttractionsService, error) {
	if err := initAttractionsDB(ctx, attractionsDB); err != nil {
		return nil, err
	}

	service := &AttractionsServiceImpl{attractionsDB: attractionsDB}
	var err error
	service.restaurants, err = service.newGeoIndex(ctx, "restaurants", func() geoindex.Point { return &Restaurant{} })
	if err != nil {
		return nil, err
	}
	service.museums, err = service.newGeoIndex(ctx, "museums", func() geoindex.Point { return &Museum{} })
	if err != nil {
		return nil, err
	}
	service.cinemas, err = service.newGeoIndex(ctx, "cinemas", func() geoindex.Point { return &Cinema{} })
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (a *AttractionsServiceImpl) newGeoIndex(ctx context.Context, collectionName string, newPoint func() geoindex.Point) (*geoindex.ClusteringIndex, error) {
	collection, err := a.attractionsDB.GetCollection(ctx, "attractions-db", collectionName)
	if err != nil {
		return nil, err
	}
	cursor, err := collection.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	index := geoindex.NewClusteringIndex()
	switch newPoint().(type) {
	case *Restaurant:
		var points []Restaurant
		if err := cursor.All(ctx, &points); err != nil {
			return nil, err
		}
		for i := range points {
			index.Add(&points[i])
		}
	case *Museum:
		var points []Museum
		if err := cursor.All(ctx, &points); err != nil {
			return nil, err
		}
		for i := range points {
			index.Add(&points[i])
		}
	case *Cinema:
		var points []Cinema
		if err := cursor.All(ctx, &points); err != nil {
			return nil, err
		}
		for i := range points {
			index.Add(&points[i])
		}
	}
	return index, nil
}

func nearbyAttractions(index *geoindex.ClusteringIndex, lat float64, lon float64) []geoindex.Point {
	center := &geoindex.GeoPoint{Pid: "", Plat: lat, Plon: lon}
	return index.KNearest(
		center,
		attractionsMaxSearchResults,
		geoindex.Km(attractionsMaxSearchRadius),
		func(geoindex.Point) bool { return true },
	)
}

func (a *AttractionsServiceImpl) NearbyRest(ctx context.Context, lat float64, lon float64) ([]Restaurant, error) {
	points := nearbyAttractions(a.restaurants, lat, lon)
	restaurants := make([]Restaurant, 0, len(points))
	for _, point := range points {
		if restaurant, ok := point.(*Restaurant); ok {
			restaurants = append(restaurants, *restaurant)
		}
	}
	return restaurants, nil
}

func (a *AttractionsServiceImpl) NearbyMus(ctx context.Context, lat float64, lon float64) ([]Museum, error) {
	points := nearbyAttractions(a.museums, lat, lon)
	museums := make([]Museum, 0, len(points))
	for _, point := range points {
		if museum, ok := point.(*Museum); ok {
			museums = append(museums, *museum)
		}
	}
	return museums, nil
}

func (a *AttractionsServiceImpl) NearbyCinema(ctx context.Context, lat float64, lon float64) ([]Cinema, error) {
	points := nearbyAttractions(a.cinemas, lat, lon)
	cinemas := make([]Cinema, 0, len(points))
	for _, point := range points {
		if cinema, ok := point.(*Cinema); ok {
			cinemas = append(cinemas, *cinema)
		}
	}
	return cinemas, nil
}
