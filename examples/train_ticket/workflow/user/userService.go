// Package user provides an implementation of the UserService
// UserService uses a backend.NoSQLDatabase to store user data
package user

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// UserService manages the users in the application
type UserService interface {
	// Finds a user given `username`
	FindByUsername(ctx context.Context, username string) (User, error)
	// Finds a user ID given `userID`
	FindByUserID(ctx context.Context, userID string) (User, error)
	// Deletes a user with ID `userID`
	DeleteUser(ctx context.Context, userID string) error
	// Gets all users
	GetAllUsers(ctx context.Context) ([]User, error)
	// Saves a new user
	SaveUser(ctx context.Context, user User) error
	// Updates an existing user
	UpdateUser(ctx context.Context, user User) (bool, error)
}

// Implementation of UserService
type UserServiceImpl struct {
	userDB backend.NoSQLDatabase
}

// Creates and returns a UserService object
func NewUserServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*UserServiceImpl, error) {
	return &UserServiceImpl{userDB: db}, nil
}

func (u *UserServiceImpl) FindByUserID(ctx context.Context, userID string) (User, error) {
	coll, err := u.userDB.GetCollection(ctx, "user", "user")
	if err != nil {
		return User{}, err
	}
	query := bson.D{{"userid", userID}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return User{}, err
	}
	var user User
	exists, err := res.One(ctx, &user)
	if err != nil {
		return User{}, err
	}
	if !exists {
		return User{}, errors.New("User with UserID " + userID + " does not exist")
	}
	return user, nil
}

func (u *UserServiceImpl) FindByUsername(ctx context.Context, username string) (User, error) {
	coll, err := u.userDB.GetCollection(ctx, "user", "user")
	if err != nil {
		return User{}, err
	}
	query := bson.D{{"username", username}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return User{}, err
	}
	var user User
	exists, err := res.One(ctx, &user)
	if err != nil {
		return User{}, err
	}
	if !exists {
		return User{}, errors.New("User with UserID " + username + " does not exist")
	}
	return user, nil
}

func (u *UserServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	coll, err := u.userDB.GetCollection(ctx, "user", "user")
	if err != nil {
		return err
	}
	query := bson.D{{"userid", userID}}
	err = coll.DeleteOne(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserServiceImpl) GetAllUsers(ctx context.Context) ([]User, error) {
	coll, err := u.userDB.GetCollection(ctx, "user", "user")
	if err != nil {
		return []User{}, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return []User{}, err
	}
	var users []User
	err = res.All(ctx, &users)
	if err != nil {
		return []User{}, err
	}
	return users, nil
}

func (u *UserServiceImpl) SaveUser(ctx context.Context, user User) error {
	coll, err := u.userDB.GetCollection(ctx, "user", "user")
	if err != nil {
		return err
	}
	return coll.InsertOne(ctx, user)
}

func (u *UserServiceImpl) UpdateUser(ctx context.Context, user User) (bool, error) {
	coll, err := u.userDB.GetCollection(ctx, "user", "user")
	if err != nil {
		return false, err
	}
	query := bson.D{{"userid", user.UserID}}
	return coll.Upsert(ctx, query, user)
}
