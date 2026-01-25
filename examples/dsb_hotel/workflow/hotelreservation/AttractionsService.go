package hotelreservation

import "context"

// AttractionsService implements the Attractions Service from the hotel reservation application
type AttractionsService interface {
	NearbyRest(ctx context.Context, lat float64, lon float64) ([]string, error)
	NearbyMus(ctx context.Context, lat float64, lon float64) ([]string, error)
	NearbyCinema(ctx context.Context, lat float64, lon float64) ([]string, error)
}
