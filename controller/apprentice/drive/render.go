package drive

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"skyra/apprentice/tree"
)

// ShowLimit is how many bytes of each value an ask shows; past it the value
// is truncated and the model may expand it. The value is not settled: a
// variable, to be tuned.
var ShowLimit = 2048

// ExpandLimit is the most expansions in one ask. Expanding is part of one
// ask, so it never counts against the refinement bound; this bounds it on
// its own.
var ExpandLimit = 16

// Text is a value as the model reads it: a string as itself, bytes as text
// when they are UTF-8 and as hex otherwise, anything else as JSON with its
// bytes rendered the same way.
func Text(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return bytesText(v)
	}
	b, _ := json.Marshal(readable(v))
	return string(b)
}

func bytesText(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	return "hex:" + hex.EncodeToString(b)
}

func readable(v any) any {
	switch v := v.(type) {
	case []byte:
		return bytesText(v)
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, x := range v {
			out[k] = readable(x)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, x := range v {
			out[i] = readable(x)
		}
		return out
	}
	return v
}

// chunk is up to limit bytes of s from off, cut at a rune boundary, and
// whether s goes on past it.
func chunk(s string, off, limit int) (string, bool) {
	if off >= len(s) {
		return "", false
	}
	end := off + limit
	if end >= len(s) {
		return s[off:], false
	}
	for end > off && !utf8.RuneStart(s[end]) {
		end--
	}
	if end == off { // a limit smaller than one character still shows it
		_, w := utf8.DecodeRuneInString(s[off:])
		end = off + w
	}
	return s[off:end], end < len(s)
}

// Render is the ask as the model reads it: where it stands, why the last
// guard was refused, and each $n in scope, its shape and, up to ShowLimit
// bytes, its value. Which $n are in the base is said, since a guard reading
// one outside it runs once and is not learned.
func Render(a tree.AskModel) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Base: %s\nDepth: %d\n", a.Base, a.Depth)
	if a.Reason != "" {
		fmt.Fprintf(&b, "Your last guard was refused: %s\n", a.Reason)
	}
	var in []string
	for n := range a.Arrival.Table {
		in = append(in, fmt.Sprintf("$%d", n))
	}
	slices.Sort(in)
	fmt.Fprintf(&b, "In the base: %s\n\nScope:\n", strings.Join(in, " "))
	for i, o := range a.Scope {
		n := i + 1
		if o.Tool != "" {
			fmt.Fprintf(&b, "$%d tool %s\n", n, o.Tool)
			continue
		}
		text := Text(o.Value)
		shown, more := chunk(text, 0, ShowLimit)
		fmt.Fprintf(&b, "$%d %s\n%s\n", n, tree.ShapeOf(o.Value).Enc(), shown)
		if more {
			fmt.Fprintf(&b, "…[truncated: %d of %d bytes shown; expand $%d from %d]\n", len(shown), len(text), n, len(shown))
		}
	}
	return b.String()
}

// Expand is the next part of $n's value from off, up to ShowLimit bytes.
func Expand(a tree.AskModel, n, off int) (string, error) {
	if n < 1 || n > len(a.Scope) {
		return "", fmt.Errorf("$%d is not in scope: the scope is $1 to $%d", n, len(a.Scope))
	}
	text := Text(a.Scope[n-1].Value)
	if off < 0 || off > len(text) {
		return "", fmt.Errorf("$%d is %d bytes: offset %d is outside it", n, len(text), off)
	}
	for off > 0 && off < len(text) && !utf8.RuneStart(text[off]) {
		off-- // an offset inside a character reads from its start
	}
	shown, more := chunk(text, off, ShowLimit)
	if more {
		shown += fmt.Sprintf("\n…[%d to %d of %d bytes; expand $%d from %d]", off, off+len(shown), len(text), n, off+len(shown))
	} else {
		shown += fmt.Sprintf("\n…[%d to %d of %d bytes: the end]", off, len(text), len(text))
	}
	return shown, nil
}
