package socialnetwork

import (
	"context"
	"regexp"
	"strings"
	"sync"
)

// The TextService interface
type TextService interface {
	// Parses the raw `text` to return an edited text with the urls replaced with shortened urls, usermention objects to be stored with the post, and the url objects to be stored with the post.
	ComposeText(ctx context.Context, reqID int64, text string) (string, []UserMention, []URL, error)
}

// Implementation of [TextService]
type TextServiceImpl struct {
	urlShortenService  UrlShortenService
	userMentionService UserMentionService
}

// Creates a [TextService] instance for parsing texts in created posts.
func NewTextServiceImpl(ctx context.Context, urlShortenService UrlShortenService, userMentionService UserMentionService) (TextService, error) {
	return &TextServiceImpl{urlShortenService: urlShortenService, userMentionService: userMentionService}, nil
}

// Implements TextService interface
func (t *TextServiceImpl) ComposeText(ctx context.Context, reqID int64, text string) (string, []UserMention, []URL, error) {
	r := regexp.MustCompile(`@[a-zA-Z0-9-_]+`)
	matches := r.FindAllString(text, -1)
	var usernames []string
	for _, m := range matches {
		usernames = append(usernames, m[1:])
	}
	url_re := regexp.MustCompile(`(http://|https://)([a-zA-Z0-9_!~*'().&=+$%-]+)`)
	url_strings := url_re.FindAllString(text, -1)

	var err1, err2 error
	var urls []URL
	var user_mentions []UserMention
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		urls, err1 = t.urlShortenService.ComposeUrls(ctx, reqID, url_strings)
	}()
	go func() {
		defer wg.Done()
		user_mentions, err2 = t.userMentionService.ComposeUserMentions(ctx, reqID, usernames)
	}()
	wg.Wait()
	if err1 != nil {
		return text, user_mentions, urls, err1
	}
	if err2 != nil {
		return text, user_mentions, urls, err2
	}

	updated_text := text
	if len(urls) != 0 {
		for idx, url_string := range url_strings {
			updated_text = strings.ReplaceAll(updated_text, url_string, urls[idx].ShortenedUrl)
		}
	}

	return updated_text, user_mentions, urls, nil
}
