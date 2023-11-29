package hotelreservation

import (
	"context"
)

type SearchService interface {
	Nearby(ctx context.Context, lat float64, lon float64, inDate string, outDate string) ([]string, error)
}

type SearchServiceImpl struct {
	geoService  GeoService
	rateService RateService
}

func NewSearchServiceImpl(ctx context.Context, geoService GeoService, rateService RateService) (SearchService, error) {
	return &SearchServiceImpl{geoService: geoService, rateService: rateService}, nil
}

func (s *SearchServiceImpl) Nearby(ctx context.Context, lat float64, lon float64, inDate string, outDate string) ([]string, error) {
	var nearby_hotels []string
	nearby_hotel_ids, err := s.geoService.Nearby(ctx, lat, lon)
	if err != nil {
		return nearby_hotels, err
	}

	rates, err := s.rateService.GetRates(ctx, nearby_hotel_ids, inDate, outDate)
	if err != nil {
		return nearby_hotels, err
	}
	for _, rate := range rates {
		nearby_hotels = append(nearby_hotels, rate.HotelID)
	}
	return nearby_hotels, nil
}
