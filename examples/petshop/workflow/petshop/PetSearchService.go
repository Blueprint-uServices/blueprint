package petshop

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrPetNotFound    = errors.New("pet not found")
	ErrInvalidPetType = errors.New("pet type not found")
)

type PetSearchService interface {
	Search(ctx context.Context, query PetSearchQuery) ([]Pet, error)
	GetPet(ctx context.Context, petType, petID string) (Pet, error)
	SetAvailability(ctx context.Context, petType, petID, availability string) error
	ResetAvailability(ctx context.Context) error
}

type PetSearchServiceImpl struct {
	collection backend.NoSQLCollection
	logger     backend.Logger
	imageBase  string
}

func NewPetSearchServiceImpl(ctx context.Context, db backend.NoSQLDatabase, imageBaseURL string) (PetSearchService, error) {
	collection, err := db.GetCollection(ctx, "petshop", "pets")
	if err != nil {
		return nil, err
	}
	service := &PetSearchServiceImpl{collection: collection, logger: backend.GetLogger(), imageBase: strings.TrimRight(imageBaseURL, "/")}
	for _, pet := range seedPets() {
		if _, err := collection.Upsert(ctx, bson.D{{Key: "petid", Value: pet.PetID}, {Key: "pettype", Value: pet.PetType}}, pet); err != nil {
			return nil, err
		}
	}
	return service, nil
}

func (s *PetSearchServiceImpl) Search(ctx context.Context, query PetSearchQuery) ([]Pet, error) {
	petType := strings.ToLower(strings.TrimSpace(query.PetType))
	petColor := strings.ToLower(strings.TrimSpace(query.PetColor))
	if petType != "" && petType != "puppy" && petType != "kitten" && petType != "bunny" {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPetType, petType)
	}
	filter := bson.D{}
	if petType != "" {
		filter = append(filter, bson.E{Key: "pettype", Value: petType})
	}
	if petColor != "" && petColor != "all" {
		filter = append(filter, bson.E{Key: "petcolor", Value: petColor})
	}
	if strings.TrimSpace(query.PetID) != "" {
		filter = append(filter, bson.E{Key: "petid", Value: strings.TrimSpace(query.PetID)})
	}
	cursor, err := s.collection.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	var pets []Pet
	if err := cursor.All(ctx, &pets); err != nil {
		return nil, err
	}
	for i := range pets {
		pets[i].PetURL = s.petURL(pets[i])
	}
	ctx, _ = s.logger.Info(ctx, "pet search completed: user=%s type=%s color=%s id=%s results=%d", query.UserID, petType, petColor, query.PetID, len(pets))
	return pets, nil
}

func (s *PetSearchServiceImpl) GetPet(ctx context.Context, petType, petID string) (Pet, error) {
	pets, err := s.Search(ctx, PetSearchQuery{PetType: petType, PetID: petID})
	if err != nil {
		return Pet{}, err
	}
	if len(pets) == 0 {
		return Pet{}, ErrPetNotFound
	}
	return pets[0], nil
}

func (s *PetSearchServiceImpl) SetAvailability(ctx context.Context, petType, petID, availability string) error {
	if availability != "yes" && availability != "no" {
		return fmt.Errorf("availability must be yes or no")
	}
	updated, err := s.collection.UpdateOne(ctx, bson.D{{Key: "petid", Value: petID}, {Key: "pettype", Value: strings.ToLower(petType)}}, bson.D{{Key: "$set", Value: bson.D{{Key: "availability", Value: availability}}}})
	if err != nil {
		return err
	}
	if updated == 0 {
		return ErrPetNotFound
	}
	return nil
}

func (s *PetSearchServiceImpl) ResetAvailability(ctx context.Context) error {
	_, err := s.collection.UpdateMany(ctx, bson.D{}, bson.D{{Key: "$set", Value: bson.D{{Key: "availability", Value: "yes"}}}})
	return err
}

func (s *PetSearchServiceImpl) petURL(pet Pet) string {
	folder := map[string]string{"puppy": "puppies", "kitten": "kittens", "bunny": "bunnies"}[pet.PetType]
	if s.imageBase == "" {
		return fmt.Sprintf("/%s/%s.jpg", folder, pet.Image)
	}
	return fmt.Sprintf("%s/%s/%s.jpg", s.imageBase, folder, pet.Image)
}

func seedPets() []Pet {
	return []Pet{
		{PetID: "023", PetType: "bunny", Availability: "yes", CutenessRate: "5", Image: "b4", PetColor: "white", Price: "75"},
		{PetID: "024", PetType: "bunny", Availability: "yes", CutenessRate: "4", Image: "b3", PetColor: "brown", Price: "45"},
		{PetID: "025", PetType: "bunny", Availability: "yes", CutenessRate: "5", Image: "b2", PetColor: "brown", Price: "55"},
		{PetID: "026", PetType: "bunny", Availability: "yes", CutenessRate: "4", Image: "b1", PetColor: "brown", Price: "95"},
		{PetID: "016", PetType: "kitten", Availability: "yes", CutenessRate: "4", Image: "k1", PetColor: "white", Price: "79"},
		{PetID: "017", PetType: "kitten", Availability: "yes", CutenessRate: "5", Image: "k3", PetColor: "white", Price: "89"},
		{PetID: "018", PetType: "kitten", Availability: "yes", CutenessRate: "4", Image: "k6", PetColor: "black", Price: "99"},
		{PetID: "019", PetType: "kitten", Availability: "yes", CutenessRate: "4", Image: "k2", PetColor: "grey", Price: "99"},
		{PetID: "020", PetType: "kitten", Availability: "yes", CutenessRate: "5", Image: "k5", PetColor: "brown", Price: "89"},
		{PetID: "021", PetType: "kitten", Availability: "yes", CutenessRate: "4", Image: "k4", PetColor: "brown", Price: "65"},
		{PetID: "022", PetType: "kitten", Availability: "yes", CutenessRate: "5", Image: "k7", PetColor: "black", Price: "75"},
		{PetID: "001", PetType: "puppy", Availability: "yes", CutenessRate: "5", Image: "p9", PetColor: "black", Price: "89"},
		{PetID: "002", PetType: "puppy", Availability: "yes", CutenessRate: "5", Image: "p10", PetColor: "black", Price: "99"},
		{PetID: "003", PetType: "puppy", Availability: "yes", CutenessRate: "5", Image: "p1", PetColor: "brown", Price: "99"},
		{PetID: "004", PetType: "puppy", Availability: "yes", CutenessRate: "5", Image: "p11", PetColor: "black", Price: "89"},
		{PetID: "005", PetType: "puppy", Availability: "yes", CutenessRate: "3", Image: "p3", PetColor: "brown", Price: "59"},
		{PetID: "006", PetType: "puppy", Availability: "yes", CutenessRate: "3", Image: "p6", PetColor: "brown", Price: "100"},
		{PetID: "007", PetType: "puppy", Availability: "yes", CutenessRate: "4", Image: "p4", PetColor: "brown", Price: "70"},
		{PetID: "008", PetType: "puppy", Availability: "yes", CutenessRate: "5", Image: "p7", PetColor: "brown", Price: "600"},
		{PetID: "009", PetType: "puppy", Availability: "yes", CutenessRate: "5", Image: "p12", PetColor: "black", Price: "60"},
		{PetID: "010", PetType: "puppy", Availability: "yes", CutenessRate: "3", Image: "p8", PetColor: "brown", Price: "80"},
		{PetID: "011", PetType: "puppy", Availability: "yes", CutenessRate: "4", Image: "p14", PetColor: "white", Price: "59"},
		{PetID: "012", PetType: "puppy", Availability: "yes", CutenessRate: "5", Image: "p15", PetColor: "white", Price: "79"},
		{PetID: "013", PetType: "puppy", Availability: "yes", CutenessRate: "4", Image: "p2", PetColor: "white", Price: "99"},
		{PetID: "014", PetType: "puppy", Availability: "yes", CutenessRate: "5", Image: "p13", PetColor: "black", Price: "79"},
		{PetID: "015", PetType: "puppy", Availability: "yes", CutenessRate: "4", Image: "p5", PetColor: "white", Price: "79"},
	}
}
