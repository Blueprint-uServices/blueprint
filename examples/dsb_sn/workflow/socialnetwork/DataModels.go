package socialnetwork

type User struct {
	UserID    int64
	FirstName string
	LastName  string
	Username  string
	PwdHashed string
	Salt      string
}

type Media struct {
	MediaID   int64
	MediaType string
}

type URL struct {
	ShortenedUrl string
	ExpandedUrl  string
}

type UserMention struct {
	UserID   int64
	Username string
}

type Creator struct {
	UserID   int64
	Username string
}

type PostType int

const (
	POST PostType = iota
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
	PostType     PostType
}
