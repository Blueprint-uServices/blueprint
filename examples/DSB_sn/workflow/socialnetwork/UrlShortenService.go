package socialnetwork

import (
	"context"
	"math/rand"
	"time"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

type UrlShortenService interface {
	ComposeUrls(ctx context.Context, reqID int64, urls []string) ([]URL, error)
	GetExtendedUrls(ctx context.Context, reqID int64, shortened_urls []string) ([]string, error)
}

type UrlShortenServiceImpl struct {
	urlShortenDB backend.NoSQLDatabase
	hostname     string
}

func NewUrlShortenServiceImpl(urlShortenDB backend.NoSQLDatabase) (UrlShortenService, error) {
	rand.Seed(time.Now().UnixNano())
	return &UrlShortenServiceImpl{urlShortenDB: urlShortenDB, hostname: "http://short-url/"}, nil
}

func (u *UrlShortenServiceImpl) genRandomStr(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (u *UrlShortenServiceImpl) ComposeUrls(ctx context.Context, reqID int64, urls []string) ([]URL, error) {
	var target_urls []URL
	var target_url_docs []interface{}
	for _, url := range urls {
		shortened_url := u.hostname + u.genRandomStr(10)
		target_url := URL{ShortenedUrl: shortened_url, ExpandedUrl: url}
		target_urls = append(target_urls, target_url)
		target_url_docs = append(target_url_docs, target_url)
	}

	if len(target_urls) > 0 {
		collection, err := u.urlShortenDB.GetCollection(ctx, "url-shorten", "url-shorten")
		if err != nil {
			return []URL{}, err
		}
		err = collection.InsertMany(ctx, target_url_docs)
		if err != nil {
			return []URL{}, err
		}
	}

	return target_urls, nil
}

func (u *UrlShortenServiceImpl) GetExtendedUrls(ctx context.Context, reqID int64, shortened_urls []string) ([]string, error) {
	// Not implemented in Original DSB
	return []string{}, nil
}
