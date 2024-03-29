package socialnetwork

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
)

// The UserService interface
type UserService interface {
	// Logs in a user using the user's `username` and `password`
	// Returns an error if the username is not registered or if the provided password doesn't match the stored password.
	Login(ctx context.Context, reqID int64, username string, password string) (string, error)
	// Registers a new user with a given `userID`.
	// Returns an error if the `userID` is already taken.
	RegisterUserWithId(ctx context.Context, reqID int64, firstName string, lastName string, username string, password string, userID int64) error
	// Registers a new user. A random `userID` is generated by the service.
	RegisterUser(ctx context.Context, reqID int64, firstName string, lastName string, username string, password string) error
	// Creates and returns a creator object to be included in a Post. The Creator object is created using the `userID`.
	ComposeCreatorWithUserId(ctx context.Context, reqID int64, userID int64, username string) (Creator, error)
	// Creates and returns a creator object to be included in a Post. The service looks up the userID of the user with the given `username`.
	// Returns an error if the user with the given `username` is not registered.
	ComposeCreatorWithUsername(ctx context.Context, reqID int64, username string) (Creator, error)
}

// Implementation of [UserService]
type UserServiceImpl struct {
	machine_id         string
	counter            int64
	current_timestamp  int64
	secret             string
	userCache          backend.Cache
	userDB             backend.NoSQLDatabase
	socialGraphService SocialGraphService
}

// The format of a Login object.
type LoginObj struct {
	UserID   int64
	Password string
	Salt     string
}

// The format of a claims object used for creating jwt login tokens.
type claimsT struct {
	Username  string
	UserID    string
	Timestamp int64
	jwt.StandardClaims
}

// Implementation of [UserService]
func NewUserServiceImpl(ctx context.Context, userCache backend.Cache, userDB backend.NoSQLDatabase, socialGraphService SocialGraphService, secret string) (UserService, error) {
	return &UserServiceImpl{counter: 0, current_timestamp: -1, machine_id: GetMachineID(), userCache: userCache, userDB: userDB, socialGraphService: socialGraphService, secret: secret}, nil
}

func (u *UserServiceImpl) getCounter(timestamp int64) int64 {
	if u.current_timestamp == timestamp {
		retVal := u.counter
		u.counter += 1
		return retVal
	} else {
		u.current_timestamp = timestamp
		u.counter = 1
		return 0
	}
}

func (u *UserServiceImpl) genRandomStr(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (u *UserServiceImpl) hashPwd(pwd []byte) string {
	hasher := sha1.New()
	hasher.Write(pwd)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// Implements UserService interface
func (u *UserServiceImpl) Login(ctx context.Context, reqID int64, username string, password string) (string, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	var login LoginObj
	login.UserID = -1
	// Ignore error for now as we don't have a separate thing for testing if key doesn't exist
	u.userCache.Get(ctx, username+":Login", &login)
	if login.UserID == -1 {
		var user User
		collection, err := u.userDB.GetCollection(ctx, "user", "user")
		if err != nil {
			return "", err
		}
		query := bson.D{{"username", username}}
		res, err := collection.FindOne(ctx, query)
		if err != nil {
			return "", err
		}
		res.One(ctx, &user)
		login.Password = user.PwdHashed
		login.Salt = user.Salt
		login.UserID = user.UserID
	}
	var tokenStr string
	hashed_pwd := u.hashPwd([]byte(password + login.Salt))
	if hashed_pwd != login.Password {
		return "", errors.New("Invalid credentials")
	} else {
		expiration_time := time.Now().Add(6 * time.Minute)
		claims := &claimsT{
			Username:       username,
			UserID:         strconv.FormatInt(login.UserID, 10),
			Timestamp:      timestamp,
			StandardClaims: jwt.StandardClaims{ExpiresAt: expiration_time.Unix()},
		}
		var err error
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err = token.SignedString([]byte(u.secret))
		if err != nil {
			return "", errors.New("Failed to create login token")
		}
	}
	err := u.userCache.Put(ctx, username+":Login", login)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// Implements UserService interface
func (u *UserServiceImpl) ComposeCreatorWithUserId(ctx context.Context, reqID int64, userID int64, username string) (Creator, error) {
	return Creator{UserID: userID, Username: username}, nil
}

// Implements UserService interface
func (u *UserServiceImpl) ComposeCreatorWithUsername(ctx context.Context, reqID int64, username string) (Creator, error) {
	user_id := int64(-1)
	u.userCache.Get(ctx, username+":UserID", &user_id)
	if user_id == -1 {
		var user User
		collection, err := u.userDB.GetCollection(ctx, "user", "user")
		if err != nil {
			return Creator{}, err
		}
		query := bson.D{{"username", username}}
		res, err := collection.FindOne(ctx, query)
		if err != nil {
			return Creator{}, err
		}
		exists, err := res.One(ctx, &user)
		if err != nil {
			return Creator{}, err
		}
		if !exists {
			return Creator{}, errors.New("User doesn't exist")
		}
		user_id = user.UserID

		err = u.userCache.Put(ctx, username+":UserID", user_id)
		if err != nil {
			return Creator{}, err
		}
	}
	return Creator{UserID: user_id, Username: username}, nil
}

// Implements UserService interface
func (u *UserServiceImpl) RegisterUserWithId(ctx context.Context, reqID int64, firstName string, lastName string, username string, password string, userID int64) error {
	collection, err := u.userDB.GetCollection(ctx, "user", "user")
	if err != nil {
		return err
	}
	query := bson.D{{"username", username}}
	res, err := collection.FindOne(ctx, query)
	if err != nil {
		return err
	}
	var user User
	exists, err := res.One(ctx, &user)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("Username already registered")
	}
	salt := u.genRandomStr(32)
	hashed_pwd := u.hashPwd([]byte(password + salt))
	user = User{UserID: userID, FirstName: firstName, LastName: lastName, PwdHashed: hashed_pwd, Salt: salt, Username: username}
	err = collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	return u.socialGraphService.InsertUser(ctx, reqID, userID)
}

// Implements UserService interface
func (u *UserServiceImpl) RegisterUser(ctx context.Context, reqID int64, firstName string, lastName string, username string, password string) error {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	idx := u.getCounter(timestamp)
	timestamp_hex := strconv.FormatInt(timestamp, 16)
	if len(timestamp_hex) > 10 {
		timestamp_hex = timestamp_hex[:10]
	} else if len(timestamp_hex) < 10 {
		timestamp_hex = strings.Repeat("0", 10-len(timestamp_hex)) + timestamp_hex
	}
	counter_hex := strconv.FormatInt(idx, 16)
	if len(counter_hex) > 1 {
		counter_hex = counter_hex[:1]
	} else if len(counter_hex) < 1 {
		counter_hex = strings.Repeat("0", 1-len(counter_hex)) + counter_hex
	}
	user_id_str := u.machine_id + timestamp_hex + counter_hex
	user_id, err := strconv.ParseInt(user_id_str, 16, 64)
	if err != nil {
		return err
	}
	user_id = user_id & 0x7FFFFFFFFFFFFFFF
	return u.RegisterUserWithId(ctx, reqID, firstName, lastName, username, password, user_id)
}
