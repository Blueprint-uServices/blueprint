package socialnetwork

import (
	"context"
	"errors"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type UserIDService interface {
	GetUserId(ctx context.Context, reqID int64, username string) (int64, error)
}

type UserIDServiceImpl struct {
	userCache backend.Cache
	userDB    backend.NoSQLDatabase
}

func NewUserIDServiceImpl(ctx context.Context, userCache backend.Cache, userDB backend.NoSQLDatabase) (UserIDService, error) {
	return &UserIDServiceImpl{userCache: userCache, userDB: userDB}, nil
}

func (u *UserIDServiceImpl) GetUserId(ctx context.Context, reqID int64, username string) (int64, error) {
	user_id := int64(-1)
	err := u.userCache.Get(ctx, username+":UserID", &user_id)
	if err != nil {
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
