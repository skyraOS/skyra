package reality

import (
	"crypto/rand"
	"strings"
)

func NewID() string {
	b := make([]byte, 8)
	rand.Read(b)
	const hex = "0123456789abcdef"
	s := make([]byte, 16)
	for i, v := range b {
		s[i*2] = hex[v>>4]
		s[i*2+1] = hex[v&0x0f]
	}
	return string(s)
}

func extractTag(text, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	start := strings.Index(text, open)
	if start == -1 {
		return ""
	}
	start += len(open)
	end := strings.Index(text[start:], close)
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(text[start : start+end])
}
