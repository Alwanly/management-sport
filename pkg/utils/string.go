package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// Slugify converts a string to a URL-safe slug
// Example: "Manchester United FC" -> "manchester-united-fc"
func Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Keep only letters and numbers, remove everything else
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, s)

	// Replace multiple spaces with single hyphen
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "-")

	// Replace multiple hyphens with single hyphen
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")

	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")

	return s
}
