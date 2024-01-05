package news

type News struct {
	Title   string `bson:"Title"`
	Content string `bson:"Content"`
}
