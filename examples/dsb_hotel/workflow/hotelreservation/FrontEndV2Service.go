package hotelreservation

import (
	"context"
	"errors"
)

// FrontEndV2Service implements the new front end server from the hotel reservation application
type FrontEndV2Service interface {
	// Returns a list of hotels that fit the search criteria provided by the user.
	SearchHandler(ctx context.Context, customerName string, inDate string, outDate string, lat float64, lon float64, locale string) ([]HotelProfile, error)
	// Returns a list of recommended hotels based on the provided location (`lat`, `lon`) and the criteria for ranking hotels (`require`)
	RecommendHandler(ctx context.Context, lat float64, lon float64, require string, locale string) ([]HotelProfile, error)
	// Logs in a user based on the username and password provided
	UserHandler(ctx context.Context, username string, password string) (string, error)
	// Makes a reservation at the user-requested hotel for the provided dates
	ReservationHandler(ctx context.Context, inDate string, outDate string, hotelId string, customerName string, username string, password string, roomNumber int64) (string, error)
	// Returns a list of reviews for a given hotel
	// Must be a valid pre-registered user.
	ReviewHandler(ctx context.Context, username string, password string, hotelId string) ([]Review, error)
	// Returns a list of restaurants nearby a given hotel
	// Must be a valid pre-registered user.
	RestaurantHandler(ctx context.Context, username string, password string, hotelId string) ([]Restaurant, error)
	// Returns a list of museums nearby a given hotel
	// Must be a valid pre-registered user.
	MuseumHandler(ctx context.Context, username string, password string, hotelId string) ([]Museum, error)
	// Returns a list of cinemas nearby a given hotel
	// Must be a valid pre-registered user.
	CinemaHandler(ctx context.Context, username string, password string, hotelId string) ([]Cinema, error)
}

type FrontEndV2ServiceImpl struct {
	profileService     ProfileService
	attractionsService AttractionsService
	reviewService      ReviewService
	frontendService    *FrontEndServiceImpl
}

func NewFrontEndV2ServiceImpl(ctx context.Context, searchService SearchService, profileService ProfileService, recommendationService RecommendationService, userService UserService, reservationService ReservationService, attractionsService AttractionsService, reviewService ReviewService) (FrontEndV2Service, error) {
	feimpl := &FrontEndServiceImpl{searchService: searchService, profileService: profileService, userService: userService, recommendationService: recommendationService, reservationService: reservationService}

	return &FrontEndV2ServiceImpl{profileService: profileService, attractionsService: attractionsService, reviewService: reviewService, frontendService: feimpl}, nil
}

func (f *FrontEndV2ServiceImpl) SearchHandler(ctx context.Context, customerName string, inDate string, outDate string, lat float64, lon float64, locale string) ([]HotelProfile, error) {
	return f.frontendService.SearchHandler(ctx, customerName, inDate, outDate, lat, lon, locale)
}

func (f *FrontEndV2ServiceImpl) RecommendHandler(ctx context.Context, lat float64, lon float64, require string, locale string) ([]HotelProfile, error) {
	return f.frontendService.RecommendHandler(ctx, lat, lon, require, locale)
}

func (f *FrontEndV2ServiceImpl) UserHandler(ctx context.Context, username string, password string) (string, error) {
	return f.frontendService.UserHandler(ctx, username, password)
}

func (f *FrontEndV2ServiceImpl) ReservationHandler(ctx context.Context, inDate string, outDate string, hotelId string, customerName string, username string, password string, roomNumber int64) (string, error) {
	return f.frontendService.ReservationHandler(ctx, inDate, outDate, hotelId, customerName, username, password, roomNumber)
}

func (f *FrontEndV2ServiceImpl) ReviewHandler(ctx context.Context, username string, password string, hotelId string) ([]Review, error) {
	_, err := f.UserHandler(ctx, username, password)
	if err != nil {
		return []Review{}, err
	}

	if hotelId == "" {
		return []Review{}, errors.New("Please specify hotelId params")
	}

	return f.reviewService.GetReviews(ctx, hotelId)
}

func (f *FrontEndV2ServiceImpl) RestaurantHandler(ctx context.Context, username string, password string, hotelId string) ([]Restaurant, error) {
	_, err := f.UserHandler(ctx, username, password)
	if err != nil {
		return []Restaurant{}, err
	}

	if hotelId == "" {
		return []Restaurant{}, errors.New("Please specify hotelId params")
	}

	profiles, err := f.profileService.GetProfiles(ctx, []string{hotelId}, "en")
	if err != nil {
		return []Restaurant{}, err
	}
	if len(profiles) == 0 || profiles[0].ID == "" {
		return []Restaurant{}, errors.New("hotel not found")
	}
	profile := profiles[0]
	lat := profile.Address.Lat
	lon := profile.Address.Lon

	return f.attractionsService.NearbyRest(ctx, lat, lon)
}

func (f *FrontEndV2ServiceImpl) MuseumHandler(ctx context.Context, username string, password string, hotelId string) ([]Museum, error) {
	_, err := f.UserHandler(ctx, username, password)
	if err != nil {
		return []Museum{}, err
	}

	if hotelId == "" {
		return []Museum{}, errors.New("Please specify hotelId params")
	}

	profiles, err := f.profileService.GetProfiles(ctx, []string{hotelId}, "en")
	if err != nil {
		return []Museum{}, err
	}
	if len(profiles) == 0 || profiles[0].ID == "" {
		return []Museum{}, errors.New("hotel not found")
	}
	profile := profiles[0]
	lat := profile.Address.Lat
	lon := profile.Address.Lon

	return f.attractionsService.NearbyMus(ctx, lat, lon)
}

func (f *FrontEndV2ServiceImpl) CinemaHandler(ctx context.Context, username string, password string, hotelId string) ([]Cinema, error) {
	_, err := f.UserHandler(ctx, username, password)
	if err != nil {
		return []Cinema{}, err
	}

	if hotelId == "" {
		return []Cinema{}, errors.New("Please specify hotelId params")
	}

	profiles, err := f.profileService.GetProfiles(ctx, []string{hotelId}, "en")
	if err != nil {
		return []Cinema{}, err
	}
	if len(profiles) == 0 || profiles[0].ID == "" {
		return []Cinema{}, errors.New("hotel not found")
	}
	profile := profiles[0]
	lat := profile.Address.Lat
	lon := profile.Address.Lon

	return f.attractionsService.NearbyCinema(ctx, lat, lon)
}
