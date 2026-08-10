package media

type CastInfo struct {
	CastInfoId int
	Name       string
	Gender     bool
	Intro      string
}

func (c CastInfo) remote() {}

type MovieID struct {
	MovID string
	Title string
}

func (m MovieID) remote() {}

type Cast struct {
	CastID     int
	Character  string
	CastInfoID int
}

func (c Cast) remote() {}

type MovieInfo struct {
	MovieID      string
	Title        string
	Casts        []Cast
	PlotID       int
	ThumbnailIDs []string
	PhotoIDs     []string
	VideoIDs     []string
	AvgRating    float64
	NumRating    int
}

func (m MovieInfo) remote() {}

type Review struct {
	ReviewID  int
	UserID    int
	ReqID     int
	Text      string
	MovieID   string
	Rating    int
	Timestamp string
}

func (r Review) remote() {}

type Page struct {
	MovieInfo MovieInfo
	Reviews   []Review
	CastInfo  []CastInfo
	Plot      string
}

func (p Page) remote() {}

type User struct {
	UserID    int
	FirstName string
	LastName  string
	Username  string
	PwdHashed string
	Salt      string
}

func (u User) remote() {}
