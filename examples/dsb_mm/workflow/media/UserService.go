package media

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
)

type UserService interface {
	Login(ctx context.Context, reqID int, username string, password string) (string, error)
	RegisterUserWithId(ctx context.Context, reqID int, firstName string, lastName string, username string, password string, userID int) error
	RegisterUser(ctx context.Context, reqID int, firstName string, lastName string, username string, password string) error
	UploadUserWithUserId(ctx context.Context, reqID int, userID int, username string) error
	UploadUserWithUsername(ctx context.Context, reqID int, username string) error
	GetUserId(ctx context.Context, reqID int, username string) (int, error)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type UserServiceImpl struct {
	machineID            string
	counter              int64
	currentTimestamp     int64
	secret               string
	userCache            backend.Cache
	userDB               backend.NoSQLDatabase
	composeReviewService ComposeReviewService
	mu                   sync.Mutex
}

type LoginObj struct {
	UserID   int
	Password string
	Salt     string
}

type Claims struct {
	Username  string
	UserID    string
	Timestamp int64
	jwt.StandardClaims
}

func NewUserServiceImpl(ctx context.Context, userCache backend.Cache, userDB backend.NoSQLDatabase, composeReviewService ComposeReviewService, secret string) (UserService, error) {
	return &UserServiceImpl{machineID: GetMachineID(), currentTimestamp: -1, secret: secret, userCache: userCache, userDB: userDB, composeReviewService: composeReviewService}, nil
}

func (u *UserServiceImpl) nextUserID() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	timestamp := time.Now().UnixMilli()
	if u.currentTimestamp == timestamp {
		u.counter++
	} else {
		u.currentTimestamp = timestamp
		u.counter = 0
	}
	return int(generatedID(u.machineID, timestamp, u.counter))
}

func (u *UserServiceImpl) GenRandomStr(length int) string {
	chars := make([]rune, length)
	for i := range chars {
		chars[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(chars)
}

func (u *UserServiceImpl) HashPwd(password []byte) string {
	hasher := sha1.New()
	_, _ = hasher.Write(password)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func (u *UserServiceImpl) Login(ctx context.Context, reqID int, username string, password string) (string, error) {
	login, err := u.getLogin(ctx, username)
	if err != nil {
		return "", err
	}
	if u.HashPwd([]byte(password+login.Salt)) != login.Password {
		return "", errors.New("invalid credentials")
	}
	claims := &Claims{
		Username: username, UserID: fmt.Sprintf("%d", login.UserID), Timestamp: time.Now().UnixMilli(),
		StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(6 * time.Minute).Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.secret))
}

func (u *UserServiceImpl) getLogin(ctx context.Context, username string) (LoginObj, error) {
	var login LoginObj
	found, err := u.userCache.Get(ctx, username+":Login", &login)
	if err != nil {
		return login, err
	}
	if found {
		return login, nil
	}
	user, found, err := u.findUser(ctx, username)
	if err != nil {
		return login, err
	}
	if !found {
		return login, errors.New("invalid credentials")
	}
	login = LoginObj{UserID: user.UserID, Password: user.PwdHashed, Salt: user.Salt}
	return login, u.userCache.Put(ctx, username+":Login", login)
}

func (u *UserServiceImpl) findUser(ctx context.Context, username string) (User, bool, error) {
	collection, err := u.userDB.GetCollection(ctx, "user", "user")
	if err != nil {
		return User{}, false, err
	}
	result, err := collection.FindOne(ctx, bson.D{{"username", username}})
	if err != nil {
		return User{}, false, err
	}
	var user User
	found, err := result.One(ctx, &user)
	return user, found, err
}

func (u *UserServiceImpl) GetUserId(ctx context.Context, reqID int, username string) (int, error) {
	var userID int
	found, err := u.userCache.Get(ctx, username+":UserID", &userID)
	if err != nil {
		return 0, err
	}
	if found {
		return userID, nil
	}
	user, found, err := u.findUser(ctx, username)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, errors.New("user not found")
	}
	return user.UserID, u.userCache.Put(ctx, username+":UserID", user.UserID)
}

func (u *UserServiceImpl) UploadUserWithUserId(ctx context.Context, reqID int, userID int, username string) error {
	return u.composeReviewService.UploadUserId(ctx, reqID, userID)
}

func (u *UserServiceImpl) UploadUserWithUsername(ctx context.Context, reqID int, username string) error {
	userID, err := u.GetUserId(ctx, reqID, username)
	if err != nil {
		return err
	}
	return u.composeReviewService.UploadUserId(ctx, reqID, userID)
}

func (u *UserServiceImpl) RegisterUserWithId(ctx context.Context, reqID int, firstName string, lastName string, username string, password string, userID int) error {
	_, found, err := u.findUser(ctx, username)
	if err != nil {
		return err
	}
	if found {
		return errors.New("username already registered")
	}
	salt := u.GenRandomStr(32)
	user := User{UserID: userID, FirstName: firstName, LastName: lastName, Username: username, PwdHashed: u.HashPwd([]byte(password + salt)), Salt: salt}
	collection, err := u.userDB.GetCollection(ctx, "user", "user")
	if err != nil {
		return err
	}
	if err := collection.InsertOne(ctx, user); err != nil {
		return err
	}
	return u.userCache.Put(ctx, username+":UserID", userID)
}

func (u *UserServiceImpl) RegisterUser(ctx context.Context, reqID int, firstName string, lastName string, username string, password string) error {
	return u.RegisterUserWithId(ctx, reqID, firstName, lastName, username, password, u.nextUserID())
}
