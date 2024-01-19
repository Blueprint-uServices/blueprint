package stringutil

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Returns s with the first letter converted to uppercase.
func Capitalize(s string) string {
	r := []rune(s)
	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
}

var r = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

// Returns name with only alphanumeric characters and all other
// symbols converted to underscores.
//
// CleanName is primarily used by plugins to convert user-defined
// service names into names that are valid as e.g. environment variables,
// command line arguments, etc.
func CleanName(name string) string {
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

// If s ends with the provided suffix, removes the suffix from s and
// replaces it with replacement.  If s does not end with the provided
// suffix, then simply returns s . replacement
func ReplaceSuffix(s string, suffix string, replacement string) string {
	if before, hasSuffix := strings.CutSuffix(s, suffix); hasSuffix {
		return before + replacement
	}
	return s + "." + replacement
}
