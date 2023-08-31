package irutil

import (
	"regexp"
	"strconv"
)

var r = regexp.MustCompile(`[^a-zA-Z0-9_ ]+`)

func Clean(name string) string {
	cleanName := r.ReplaceAllString(name, "_")
	for len(cleanName) > 0 {
		if _, err := strconv.Atoi(cleanName[0:1]); err != nil {
			return cleanName
		} else {
			cleanName = cleanName[1:]
		}
	}
	return cleanName
}
