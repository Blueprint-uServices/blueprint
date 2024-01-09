package hotelreservation

import (
	"context"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// ProfileService implements Profile Service from the hotel reservation application
type ProfileService interface {
	// Returns the profiles of hotels based on the `hotelIds` provided
	GetProfiles(ctx context.Context, hotelIds []string, locale string) ([]HotelProfile, error)
}

// Implementation of Profile Service
type ProfileServiceImpl struct {
	profileCache backend.Cache
	profileDB    backend.NoSQLDatabase
}

func initProfileDB(ctx context.Context, db backend.NoSQLDatabase) error {
	c, err := db.GetCollection(ctx, "profile-db", "hotels")
	if err != nil {
		return err
	}
	err = c.InsertOne(ctx, &HotelProfile{
		"1",
		"Clift Hotel",
		"(415) 775-4700",
		"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
		Address{
			"495",
			"Geary St",
			"San Francisco",
			"CA",
			"United States",
			"94102",
			37.7867,
			-122.4112}})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &HotelProfile{
		"2",
		"W San Francisco",
		"(415) 777-5300",
		"Less than a block from the Yerba Buena Center for the Arts, this trendy hotel is a 12-minute walk from Union Square.",
		Address{
			"181",
			"3rd St",
			"San Francisco",
			"CA",
			"United States",
			"94103",
			37.7854,
			-122.4005}})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &HotelProfile{
		"3",
		"Hotel Zetta",
		"(415) 543-8555",
		"A 3-minute walk from the Powell Street cable-car turnaround and BART rail station, this hip hotel 9 minutes from Union Square combines high-tech lodging with artsy touches.",
		Address{
			"55",
			"5th St",
			"San Francisco",
			"CA",
			"United States",
			"94103",
			37.7834,
			-122.4071}})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &HotelProfile{
		"4",
		"Hotel Vitale",
		"(415) 278-3700",
		"This waterfront hotel with Bay Bridge views is 3 blocks from the Financial District and a 4-minute walk from the Ferry Building.",
		Address{
			"8",
			"Mission St",
			"San Francisco",
			"CA",
			"United States",
			"94105",
			37.7936,
			-122.3930}})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &HotelProfile{
		"5",
		"Phoenix Hotel",
		"(415) 776-1380",
		"Located in the Tenderloin neighborhood, a 10-minute walk from a BART rail station, this retro motor lodge has hosted many rock musicians and other celebrities since the 1950s. It’s a 4-minute walk from the historic Great American Music Hall nightclub.",
		Address{
			"601",
			"Eddy St",
			"San Francisco",
			"CA",
			"United States",
			"94109",
			37.7831,
			-122.4181}})
	if err != nil {
		return err
	}

	err = c.InsertOne(ctx, &HotelProfile{
		"6",
		"St. Regis San Francisco",
		"(415) 284-4000",
		"St. Regis Museum Tower is a 42-story, 484 ft skyscraper in the South of Market district of San Francisco, California, adjacent to Yerba Buena Gardens, Moscone Center, PacBell Building and the San Francisco Museum of Modern Art.",
		Address{
			"125",
			"3rd St",
			"San Francisco",
			"CA",
			"United States",
			"94109",
			37.7863,
			-122.4015}})
	if err != nil {
		return err
	}

	// add up to 80 hotels
	for i := 7; i <= 80; i++ {
		hotel_id := strconv.Itoa(i)
		phone_num := "(415) 284-40" + hotel_id
		lat := 37.7835 + float64(i)/500.0*3
		lon := -122.41 + float64(i)/500.0*4
		err = c.InsertOne(ctx, &HotelProfile{
			hotel_id,
			"St. Regis San Francisco",
			phone_num,
			"St. Regis Museum Tower is a 42-story, 484 ft skyscraper in the South of Market district of San Francisco, California, adjacent to Yerba Buena Gardens, Moscone Center, PacBell Building and the San Francisco Museum of Modern Art.",
			Address{
				"125",
				"3rd St",
				"San Francisco",
				"CA",
				"United States",
				"94109",
				lat,
				lon}})
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates and Returns a new Profile Service object
func NewProfileServiceImpl(ctx context.Context, profileCache backend.Cache, profileDB backend.NoSQLDatabase) (ProfileService, error) {
	err := initProfileDB(ctx, profileDB)
	if err != nil {
		return nil, err
	}
	return &ProfileServiceImpl{profileCache: profileCache, profileDB: profileDB}, nil
}

func (p *ProfileServiceImpl) GetProfiles(ctx context.Context, hotelIds []string, locale string) ([]HotelProfile, error) {
	var profiles []HotelProfile

	for _, hid := range hotelIds {
		var profile HotelProfile
		exists, err := p.profileCache.Get(ctx, hid, &profile)
		if err != nil {
			return profiles, err
		}
		if !exists {
			// Check Database
			collection, err := p.profileDB.GetCollection(ctx, "profile-db", "hotels")
			if err != nil {
				return []HotelProfile{}, err
			}
			query := bson.D{{"id", hid}}
			res, err := collection.FindOne(ctx, query)
			if err != nil {
				return profiles, err
			}
			res.One(ctx, &profile)
			err = p.profileCache.Put(ctx, hid, profile)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}
