// Proposal 58: cuts, as Frames.lean defines them. An argument is a $n or a
// cut of one: runs of its lines ($2[3:7,9]) or a field path ($2.dir). The
// model never writes the value: the instruction says where to take it
// from, so what is learned applies to new values. A cut is part of binding
// an instruction, not a transition or an ask.
package tree

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// SelKind is what a cut takes of its $n.
type SelKind uint8

const (
	SelWhole SelKind = iota
	SelLines
	SelField
)

// A Sel is a selector: all of $n, runs of its lines from A to B inclusive,
// counted from 1, or the value at a field path.
type Sel struct {
	Kind SelKind
	Runs [][2]int
	Path []string
}

// String is the selector as it is written after $n, and as its base is
// keyed.
func (s Sel) String() string {
	switch s.Kind {
	case SelLines:
		rs := make([]string, len(s.Runs))
		for i, r := range s.Runs {
			rs[i] = itoa(r[0]) + ":" + itoa(r[1])
		}
		return "[" + strings.Join(rs, ",") + "]"
	case SelField:
		return "." + strings.Join(s.Path, ".")
	}
	return ""
}

var errSel = errors.New("a cut is [a:b,c] for lines or .key for a field")

// selector reads the selector at the start of s: its width, 0 for none.
// After $n, `[` opens runs of lines and `.` then a letter or `_` a field
// path; anything else is no selector.
func selector(s string) (Sel, int, error) {
	if s == "" {
		return Sel{}, 0, nil
	}
	if s[0] == '[' {
		end := strings.IndexByte(s, ']')
		if end < 0 {
			return Sel{}, 0, errSel
		}
		sel := Sel{Kind: SelLines, Runs: [][2]int{}}
		if body := s[1:end]; body != "" {
			for _, part := range strings.Split(body, ",") {
				a, b, ranged := strings.Cut(part, ":")
				x, okx := digits(a)
				y, oky := x, okx
				if ranged {
					y, oky = digits(b)
				}
				if !okx || !oky {
					return Sel{}, 0, errSel
				}
				sel.Runs = append(sel.Runs, [2]int{x, y})
			}
		}
		return sel, end + 1, nil
	}
	sel, w := Sel{Kind: SelField}, 0
	for w < len(s) && s[w] == '.' {
		k := w + 1
		for k < len(s) && (s[k] == '_' || isLetter(s[k]) || (k > w+1 && s[k] >= '0' && s[k] <= '9')) {
			k++
		}
		if k == w+1 {
			break
		}
		sel.Path = append(sel.Path, s[w+1:k])
		w = k
	}
	if w == 0 {
		return Sel{}, 0, nil
	}
	return sel, w, nil
}

func isLetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

func digits(s string) (int, bool) {
	if s == "" || len(s) > 9 {
		return 0, false
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, true
}

// ReadSel reads the selector at the start of s, and its width: 0 for none.
func ReadSel(s string) (Sel, int, error) { return selector(s) }

// ParseSel reads a whole selector, as String writes it.
func ParseSel(s string) (Sel, error) {
	sel, w, err := selector(s)
	if err != nil {
		return Sel{}, err
	}
	if w != len(s) {
		return Sel{}, errSel
	}
	return sel, nil
}

// textOf is a value's text: a string, or bytes that are UTF-8.
func textOf(v any) (string, bool) {
	switch v := v.(type) {
	case string:
		return v, true
	case []byte:
		if utf8.Valid(v) {
			return string(v), true
		}
	}
	return "", false
}

// lines splits text at newlines, a final newline ending the last line
// rather than starting another.
func lines(s string) []string {
	ls := strings.Split(s, "\n")
	if len(ls) > 1 && ls[len(ls)-1] == "" {
		ls = ls[:len(ls)-1]
	}
	return ls
}

// Cut is a cut's value: the selected lines joined, or the field's value.
// It fails when the selector does not apply to v.
func Cut(v any, s Sel) (any, bool) {
	switch s.Kind {
	case SelWhole:
		return v, true
	case SelLines:
		text, ok := textOf(v)
		if !ok || len(s.Runs) == 0 {
			return nil, false
		}
		ls := lines(text)
		var out []string
		for _, r := range s.Runs {
			if r[0] < 1 || r[0] > r[1] || r[1] > len(ls) {
				return nil, false
			}
			out = append(out, ls[r[0]-1:r[1]]...)
		}
		return strings.Join(out, "\n"), true
	}
	if len(s.Path) == 0 {
		return nil, false
	}
	for _, k := range s.Path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}
		if v, ok = m[k]; !ok {
			return nil, false
		}
	}
	return v, true
}

// CutObs is a cut observation: its value cut, its base keyed on the base it
// was cut from and the selector, never a tool. All of $n is $n itself.
func CutObs(cutBase func(base, sel string) string, s Sel, o Obs) (Obs, bool) {
	if s.Kind == SelWhole {
		return o, true
	}
	v, ok := Cut(o.Value, s)
	if !ok {
		return Obs{}, false
	}
	return Obs{Value: v, Base: cutBase(o.Base, s.String())}, true
}

// cutRefs rewrites each $n in tmpl, and the selector after it, through to.
func cutRefs(tmpl string, to func(n int, s Sel) (string, error)) (string, error) {
	var out strings.Builder
	for i := 0; i < len(tmpl); {
		if n, w := number(tmpl[i:]); w > 0 {
			s, sw, err := selector(tmpl[i+w:])
			if err != nil {
				return "", err
			}
			text, err := to(n, s)
			if err != nil {
				return "", err
			}
			out.WriteString(text)
			i += w + sw
			continue
		}
		out.WriteByte(tmpl[i])
		i++
	}
	return out.String(), nil
}
