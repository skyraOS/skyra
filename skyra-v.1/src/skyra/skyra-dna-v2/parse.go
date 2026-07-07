package dnav2

import (
	"regexp"
	"strings"
)

var (
	frameEditRe = regexp.MustCompile(`(?is)<edit\s+id="([^"]*)"\s*>(.*?)</edit>`)
	frameNewRe  = regexp.MustCompile(`(?is)<new>(.*?)</new>`)
)

// ParseFrameOp parses midlevel's framing operator (active-grouping.md):
//
//	<edit id="ID">sharpened gist</edit>  → ("edit", "ID", gist, true)
//	<new>gist</new>                      → ("new",  "",   gist, true)
//
// Edit is tried first so a stray "<new>" inside an edit's gist can't win.
func ParseFrameOp(resp string) (op, id, gist string, ok bool) {
	if m := frameEditRe.FindStringSubmatch(resp); m != nil {
		return "edit", strings.TrimSpace(m[1]), strings.TrimSpace(m[2]), true
	}
	if m := frameNewRe.FindStringSubmatch(resp); m != nil {
		return "new", "", strings.TrimSpace(m[1]), true
	}
	return "", "", "", false
}
