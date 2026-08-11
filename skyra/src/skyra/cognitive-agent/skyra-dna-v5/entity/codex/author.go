package codex

import (
	"fmt"
	"regexp"
	"strings"

	"skyra-v1/src/inference"
)

// Author writes a new expectation for a need.
//
// This is the ONE place the design requires something be handed to it rather
// than growing it (interfaces.md, Bootstrap):
//
//	The runtime must know how to reach at least one entity that can author an
//	expectation. The root's frame must name it.
//
// NAME THE NEED, NOT THE COMPLAINT. The name is what gets tokenized for the
// index, so a new possibility must be named for what it DOES — "find a red
// utensil" — never for the rejection that prompted it. The negation is the cite.
//
// FLAGGED: this call is questions.md §2 made concrete. Its output appears in no
// citation, so provenance is exact for observations and absent for priors. The
// authored expression is the largest unattributed term in any chain that uses it.
func Author(need string) (name string, expression string, err error) {
	const system = `You author ONE expectation for a runtime that runs it later.

An expectation is a standing instruction — what a possibility is FOR. It is
re-run many times against different inputs and NEVER edited, so write it for the
general case, not for one example.

Reply with exactly two lines and nothing else:

NAME: three-or-fewer-lowercase-hyphenated-words
EXPECTATION: one paragraph of plain instruction, in the second person

The NAME is what it DOES, never the problem that prompted it. "count-file-lines"
not "fix-the-count-bug".

The EXPECTATION must tell whoever reads it what to produce and what fields to put
in its answer. Assume the reader is capable and has the input in front of them.`

	raw, err := inference.CallCodex(system, "The need:\n\n"+need)
	if err != nil {
		return "", "", err
	}

	name = firstGroup(reName, raw)
	expression = firstGroup(reExpect, raw)
	if name == "" || expression == "" {
		return "", "", fmt.Errorf("author: could not read NAME/EXPECTATION out of %q", trunc(raw, 300))
	}
	return slug(name), expression, nil
}

var (
	reName   = regexp.MustCompile(`(?im)^\s*NAME:\s*(.+)$`)
	reExpect = regexp.MustCompile(`(?ims)^\s*EXPECTATION:\s*(.+)$`)
)

func firstGroup(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case ('a' <= r && r <= 'z') || ('0' <= r && r <= '9'):
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	if out == "" {
		return "unnamed"
	}
	return out
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
