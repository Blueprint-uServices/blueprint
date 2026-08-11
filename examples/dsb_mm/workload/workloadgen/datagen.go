package workloadgen

import (
	"fmt"
	"math/rand"
)

var movie_titles []string
var textRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 ")

func gen_text(length int) string {
	text := make([]rune, length)
	for i := range text {
		text[i] = textRunes[rand.Intn(len(textRunes))]
	}
	return string(text)
}

func GenReviewHandler() (string, string, string, string, int) {
	title_id := rand.Intn(len(movie_titles))
	title := movie_titles[title_id]
	text := gen_text(rand.Intn(33) + 32)
	userid := rand.Intn(1000) + 1
	username := fmt.Sprintf("username_%d", userid)
	password := fmt.Sprintf("password_%d", userid)
	rating := rand.Intn(10) + 1

	return title, text, username, password, rating
}
