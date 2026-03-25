package graph

import (
	"regexp"
	"strings"
)

var reHTML = regexp.MustCompile(`(?s)<[^>]*>`)

// StripHTML removes tags and collapses whitespace (simple, for mail body preview).
func StripHTML(s string) string {
	s = reHTML.ReplaceAllString(s, " ")
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}
