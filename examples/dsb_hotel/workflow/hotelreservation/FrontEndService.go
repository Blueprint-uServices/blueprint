package hotelreservation

import (
	"context"
	"errors"
)

// FrontEndService implements the front end server from the hotel reservation application
type FrontEndService interface {
	// Returns a list of hotels that fit the search criteria provided by the user.
	SearchHandler(ctx context.Context, customerName string, inDate string, outDate string, lat float64, lon float64, locale string) ([]HotelProfile, error)
	// Returns a list of recommended hotels based on the provided location (`lat`, `lon`) and the criteria for ranking hotels (`require`)
	RecommendHandler(ctx context.Context, lat float64, lon float64, require string, locale string) ([]HotelProfile, error)
	// Logs in a user based on the username and password provided
	UserHandler(ctx context.Context, username string, password string) (string, error)
	// Makes a reservation at the user-requested hotel for the provided dates
	ReservationHandler(ctx context.Context, inDate string, outDate string, hotelId string, customerName string, username string, password string, roomNumber int64) (string, error)
}

// Implementation of the FrontEndService
type FrontEndServiceImpl struct {
	searchService         SearchService
	profileService        ProfileService
	recommendationService RecommendationService
	userService           UserService
	reservationService    ReservationService
}

// Creates and Returns a new FrontEndService object
func NewFrontEndServiceImpl(ctx context.Context, searchService SearchService, profileService ProfileService, recommendationService RecommendationService, userService UserService, reservationService ReservationService) (FrontEndService, error) {
	return &FrontEndServiceImpl{searchService: searchService, profileService: profileService, recommendationService: recommendationService, userService: userService, reservationService: reservationService}, nil
}

func (f *FrontEndServiceImpl) SearchHandler(ctx context.Context, customerName string, inDate string, outDate string, lat float64, lon float64, locale string) ([]HotelProfile, error) {
	nearby_hotels, err := f.searchService.Nearby(ctx, lat, lon, inDate, outDate)
	if err != nil {
		return []HotelProfile{}, err
	}
	if locale == "" {
		locale = "en"
	}
	available_hotels, err := f.reservationService.CheckAvailability(ctx, "", nearby_hotels, inDate, outDate, 1)
	if err != nil {
		return []HotelProfile{}, nil
	}
	profiles, err := f.profileService.GetProfiles(ctx, available_hotels, locale)
	if err != nil {
		return []HotelProfile{}, err
	}
	return profiles, nil
}

func (f *FrontEndServiceImpl) RecommendHandler(ctx context.Context, lat float64, lon float64, require string, locale string) ([]HotelProfile, error) {
	if require != "dis" && require != "rate" && require != "price" {
		return []HotelProfile{}, errors.New("Invalid require param " + require)
	}
	recommended_hotels, err := f.recommendationService.GetRecommendations(ctx, require, lat, lon)
	if err != nil {
		return []HotelProfile{}, err
	}
	if locale == "" {
		locale = "en"
	}
	profiles, err := f.profileService.GetProfiles(ctx, recommended_hotels, locale)
	if err != nil {
		return []HotelProfile{}, err
	}
	return profiles, nil
}

func (f *FrontEndServiceImpl) UserHandler(ctx context.Context, username string, password string) (string, error) {
	if username == "" || password == "" {
		return "", errors.New("Please specify username and password")
	}
	exists, err := f.userService.CheckUser(ctx, username, password)
	if err != nil || !exists {
		return "Invalid Credentials", errors.New("Invalid Credentials")
	}
	return "Login successful", nil
}

func (f *FrontEndServiceImpl) ReservationHandler(ctx context.Context, inDate string, outDate string, hotelId string, customerName string, username string, password string, roomNumber int64) (string, error) {
	if inDate == "" || outDate == "" {
		return "", errors.New("Please specify inDate/outDate params")
	}
	if !checkDataFormat(inDate) || !checkDataFormat(outDate) {
		return "", errors.New("Please check inDate/outDate format (YYYY-MM-DD)")
	}
	if hotelId == "" {
		return "", errors.New("Please specify hotelId params")
	}
	if customerName == "" {
		return "", errors.New("Please specify customer name")
	}
	if username == "" || password == "" {
		return "", errors.New("Please specify username or password")
	}

	exists, err := f.userService.CheckUser(ctx, username, password)
	if err != nil {
		return "", err
	}
	if !exists {
		return "Invalid credentials", errors.New("Invalid credentials")
	}
	resevered_hotels, err := f.reservationService.MakeReservation(ctx, customerName, []string{hotelId}, inDate, outDate, roomNumber)
	if err != nil {
		return "", err
	}
	if len(resevered_hotels) == 0 {
		return "Failed to make reservation", errors.New("Failed to make reservation")
	}
	return "Reservation successful", nil
}

func checkDataFormat(date string) bool {
	if len(date) != 10 {
		return false
	}
	for i := 0; i < 10; i++ {
		if i == 4 || i == 7 {
			if date[i] != '-' {
				return false
			}
		} else {
			if date[i] < '0' || date[i] > '9' {
				return false
			}
		}
	}
	return true
}
