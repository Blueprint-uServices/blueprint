package socialnetwork

// The format of a user stored in the user database
type User struct {
	UserID    int64
	FirstName string
	LastName  string
	Username  string
	PwdHashed string
	Salt      string
}

// The format of a media stored as part of a post.
type Media struct {
	MediaID   int64
	MediaType string
}

// The format of a url stored in the url-shorten database
type URL struct {
	ShortenedUrl string
	ExpandedUrl  string
}

// The format of a usermention stored as part of a post
type UserMention struct {
	UserID   int64
	Username string
}

// The format of a creator stored as part of a post
type Creator struct {
	UserID   int64
	Username string
}

// The type of the post.
type PostType int64

// Enums aren't supported atm. So just use integers instead.
const (
	POST int64 = iota
	REPOST
	REPLY
	DM
)

type Post struct {
	PostID       int64
	Creator      Creator
	ReqID        int64
	Text         string
	UserMentions []UserMention
	Medias       []Media
	Urls         []URL
	Timestamp    int64
	PostType     int64
}
