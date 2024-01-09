package socialnetwork

import (
	"context"
	"math/rand"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

// The UrlShortenService interface
type UrlShortenService interface {
	// Converts raw `urls` into shortened urls to be used within the application. Returns the list of shortened urls.
	ComposeUrls(ctx context.Context, reqID int64, urls []string) ([]URL, error)
	// Converts the list of shortened urls into their extended forms.
	GetExtendedUrls(ctx context.Context, reqID int64, shortenedUrls []string) ([]string, error)
}

// Implementation of [UrlShortenService]
type UrlShortenServiceImpl struct {
	urlShortenDB backend.NoSQLDatabase
	hostname     string
}

// Creates a [UrlShortenService] instance for converting raw urls to shortened urls and vice versa.
func NewUrlShortenServiceImpl(ctx context.Context, urlShortenDB backend.NoSQLDatabase) (UrlShortenService, error) {
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

// Implements ComposeUrls interface
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

// Implements UrlShortenService interface.
// Currently not implemented as the original DSB application doesn't implement this function either.
func (u *UrlShortenServiceImpl) GetExtendedUrls(ctx context.Context, reqID int64, shortenedUrls []string) ([]string, error) {
	// Not implemented in Original DSB
	return []string{}, nil
}
