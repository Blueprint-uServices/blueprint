package hotelreservation

import (
	"context"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	geoindex "github.com/hailocab/go-geoindex"
	"go.mongodb.org/mongo-driver/bson"
)

// GeoService implements the GeoService from HotelReservation
type GeoService interface {
	// Returns list of hotel IDs that are near to the provided coordinates (`lat`, `lon`)
	Nearby(ctx context.Context, lat float64, lon float64) ([]string, error)
}

// Implementation of GeoService
type GeoServiceImpl struct {
	geoDB backend.NoSQLDatabase
	index *geoindex.ClusteringIndex
}

func initGeoDB(ctx context.Context, db backend.NoSQLDatabase) error {
	c, err := db.GetCollection(ctx, "geo-db", "geo")
	if err != nil {
		return err
	}
	err = c.InsertOne(ctx, &Point{"1", 37.7867, -122.4112})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &Point{"2", 37.7854, -122.4005})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &Point{"3", 37.7854, -122.4071})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &Point{"4", 37.7936, -122.3930})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &Point{"5", 37.7831, -122.4181})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &Point{"6", 37.7863, -122.4015})
	if err != nil {
		return err
	}

	// add up to 80 hotels
	for i := 7; i <= 80; i++ {
		hotel_id := strconv.Itoa(i)
		lat := 37.7835 + float64(i)/500.0*3
		lon := -122.41 + float64(i)/500.0*4
		err = c.InsertOne(ctx, &Point{hotel_id, lat, lon})
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates and returns a new GeoService object
func NewGeoServiceImpl(ctx context.Context, geoDB backend.NoSQLDatabase) (GeoService, error) {
	err := initGeoDB(ctx, geoDB)
	if err != nil {
		return nil, err
	}
	service := &GeoServiceImpl{geoDB: geoDB}
	service.newGeoIndex(ctx)
	return service, nil
}

const (
	MAXSEARCHRESULTS = 5
	MAXSEARCHRADIUS  = 10
)

func (g *GeoServiceImpl) newGeoIndex(ctx context.Context) error {
	collection, err := g.geoDB.GetCollection(ctx, "geo-db", "geo")
	if err != nil {
		return err
	}
	var points []Point
	filter := bson.D{}
	res, err := collection.FindMany(ctx, filter)
	if err != nil {
		return err
	}
	res.All(ctx, &points)
	g.index = geoindex.NewClusteringIndex()
	for _, point := range points {
		g.index.Add(point)
	}
	return nil
}

func (g *GeoServiceImpl) getNearbyPoints(lat float64, lon float64) []geoindex.Point {
	center := &Point{Pid: "", Plat: lat, Plon: lon}

	return g.index.KNearest(center, MAXSEARCHRESULTS, geoindex.Km(MAXSEARCHRADIUS), func(p geoindex.Point) bool { return true })
}

func (g *GeoServiceImpl) Nearby(ctx context.Context, lat float64, lon float64) ([]string, error) {
	points := g.getNearbyPoints(lat, lon)
	var hotelIds []string
	for _, p := range points {
		hotelIds = append(hotelIds, p.Id())
	}

	return hotelIds, nil
}
