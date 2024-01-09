package socialnetwork

import (
	"context"
	"log"
	"strings"

	"github.com/Blueprint-uServices/blueprint/runtime/core/backend"
)

// The UserMentionService interface
type UserMentionService interface {
	// Composes UserMention objects to be stored in a post by converting raw usernames.
	// Returns an error if any name in the `usernames` array is not registered.
	ComposeUserMentions(ctx context.Context, reqID int64, usernames []string) ([]UserMention, error)
}

// Implementation of [UserMentionService]
type UserMentionServiceImpl struct {
	userCache backend.Cache
	userDB    backend.NoSQLDatabase
}

// Creates a [UserMentionService] instance that is responsible for converting usernames to usermention objects
func NewUserMentionServiceImpl(ctx context.Context, userCache backend.Cache, userDB backend.NoSQLDatabase) (UserMentionService, error) {
	return &UserMentionServiceImpl{userCache: userCache, userDB: userDB}, nil
}

// Implements the UserMentionService interface
func (u *UserMentionServiceImpl) ComposeUserMentions(ctx context.Context, reqID int64, usernames []string) ([]UserMention, error) {
	usernames_not_cached := make(map[string]bool)
	rev_lookup := make(map[string]string)
	var keys []string
	for _, name := range usernames {
		usernames_not_cached[name] = true
		keys = append(keys, name+":UserID")
		rev_lookup[name+":UserID"] = name
	}
	values := make([]int64, len(keys))
	var retvals []interface{}
	for idx, _ := range values {
		retvals = append(retvals, &values[idx])
	}
	u.userCache.Mget(ctx, keys, retvals)
	var user_mentions []UserMention
	for idx, key := range keys {
		if values[idx] != 0 {
			user_mention := UserMention{UserID: values[idx], Username: rev_lookup[key]}
			user_mentions = append(user_mentions, user_mention)
			delete(usernames_not_cached, rev_lookup[key])
		}
	}
	if len(usernames_not_cached) != 0 {
		log.Println("Looking for user IDs in the database")
		var names []string
		for name := range usernames_not_cached {
			names = append(names, `"`+name+`"`)
		}
		collection, err := u.userDB.GetCollection(ctx, "user", "user")
		if err != nil {
			return user_mentions, err
		}
		in_str := strings.Join(names, ",")
		query := `{"Username": {"$in": [` + in_str + `]}}`
		query_d, err := parseNoSQLDBQuery(query)
		if err != nil {
			log.Println(err)
			return []UserMention{}, err
		}
		vals, err := collection.FindMany(ctx, query_d)
		if err != nil {
			return []UserMention{}, err
		}
		var new_user_mentions []UserMention
		err = vals.All(ctx, &new_user_mentions)
		if err != nil {
			return user_mentions, err
		}
		user_mentions = append(user_mentions, new_user_mentions...)
	}
	return user_mentions, nil
}
