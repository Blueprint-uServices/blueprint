package hotelreservation

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// UserService manages the registered users for the application
type UserService interface {
	// Returns true if the provided credentials are for a valid user and match the stored credentials.
	// Returns false otherwise.
	CheckUser(ctx context.Context, username string, password string) (bool, error)
}

// Implementation of the UserService
type UserServiceImpl struct {
	users  map[string]string
	userDB backend.NoSQLDatabase
}

func initUserDB(ctx context.Context, userDB backend.NoSQLDatabase) error {
	c, err := userDB.GetCollection(ctx, "user-db", "user")
	if err != nil {
		return err
	}
	for i := 0; i <= 500; i++ {
		suffix := strconv.Itoa(i)
		user_name := "Cornell_" + suffix
		password := ""
		for j := 0; j < 10; j++ {
			password += suffix
		}

		sum := sha256.Sum256([]byte(password))
		pass := fmt.Sprintf("%x", sum)
		err := c.InsertOne(ctx, &User{user_name, pass})
		if err != nil {
			return err
		}
	}
	return nil
}

// Creates and returns a new UserService object
func NewUserServiceImpl(ctx context.Context, userDB backend.NoSQLDatabase) (UserService, error) {
	u := &UserServiceImpl{userDB: userDB, users: make(map[string]string)}
	err := initUserDB(ctx, userDB)
	if err != nil {
		return nil, err
	}
	err = u.LoadUsers(context.Background())
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (u *UserServiceImpl) LoadUsers(ctx context.Context) error {
	collection, err := u.userDB.GetCollection(ctx, "user-db", "user")
	if err != nil {
		return err
	}

	var users []User
	filter := bson.D{}
	result, err := collection.FindMany(ctx, filter)
	if err != nil {
		return err
	}
	result.All(ctx, &users)

	for _, user := range users {
		u.users[user.Username] = user.Password
	}
	return nil
}

func (u *UserServiceImpl) CheckUser(ctx context.Context, username string, password string) (bool, error) {
	sum := sha256.Sum256([]byte(password))
	pass := fmt.Sprintf("%x", sum)

	result := false
	if true_pass, found := u.users[username]; found {
		result = pass == true_pass
	}
	return result, nil
}
