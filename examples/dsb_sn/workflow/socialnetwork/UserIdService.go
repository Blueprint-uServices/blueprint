package socialnetwork

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// The UserIDService interface
type UserIDService interface {
	// Returns the userID of the user associated with the `username`.
	// Returns an error if no user exists with the given `username`.
	GetUserId(ctx context.Context, reqID int64, username string) (int64, error)
}

// Implementation of [UserIDService]
type UserIDServiceImpl struct {
	userCache backend.Cache
	userDB    backend.NoSQLDatabase
}

// Creates a [UserIDService] instance for looking up users with usernames.
func NewUserIDServiceImpl(ctx context.Context, userCache backend.Cache, userDB backend.NoSQLDatabase) (UserIDService, error) {
	return &UserIDServiceImpl{userCache: userCache, userDB: userDB}, nil
}

// Implements UserIDService interface
func (u *UserIDServiceImpl) GetUserId(ctx context.Context, reqID int64, username string) (int64, error) {
	user_id := int64(-1)
	exists, err := u.userCache.Get(ctx, username+":UserID", &user_id)
	if err != nil {
		return user_id, err
	}
	if !exists {
		var user User
		collection, err := u.userDB.GetCollection(ctx, "user", "user")
		if err != nil {
			return -1, err
		}
		query := bson.D{{"username", username}}
		res, err := collection.FindOne(ctx, query)
		if err != nil {
			return -1, err
		}
		result, err := res.One(ctx, &user)
		if err != nil {
			return -1, err
		}
		if !result {
			return -1, errors.New("Username " + username + " not found")
		}
		user_id = user.UserID

		err = u.userCache.Put(ctx, username+":UserID", user_id)
		if err != nil {
			return -1, err
		}
	}
	return user_id, nil
}
