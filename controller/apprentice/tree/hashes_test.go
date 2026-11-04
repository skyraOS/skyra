// The hashes: their format pinned by known answers, computed apart from the
// code (Python's hashlib over the same canonical JSON), and their domains
// kept apart.
package tree

import (
	"regexp"
	"strings"
	"testing"
)

var (
	hb      = Binary("sha256:" + strings.Repeat("1", 64))
	hpassed = []ReturnBase{{"sha256:" + strings.Repeat("2", 64)}}
)

func TestHashesKnownAnswers(t *testing.T) {
	in := must(CreateInvoke(hb, hpassed))
	ret := must(CreateReturn(in, shapeS))
	for _, tc := range []struct{ name, got, want string }{
		{"input", in.id, "sha256:1cf89a85d7d5563b2afc55257ee86c7ba0433ad322c7b50e7ac433a8b9fbf704"},
		{"input, nothing passed", must(CreateInvoke(hb, nil)).id, "sha256:9c48a2bc40e7f22e73734b5fe8aee1f70c6f01a672b61a4a5a54d67f494bf1db"},
		{"return", ret.id, "sha256:38d3293cf25a18a2da0b13f8b9c9b98a96df3689e3f14d8a406dfe8ec4559453"},
		{"refinement, default", string(RefinementOf(ret, DefaultGuard())), "sha256:5a462ec1478a3c3157c5a88546e02d592d562ccf0fc1c0ead40cf5e52c0d1b74"},
		// <, > and & are written as they are, as JSON.stringify writes them.
		{"refinement, written", string(RefinementOf(ret, WrittenGuard("a < b && x"))), "sha256:6f3395965ed805f02b01a99f2a083cd7edf3dddda49be232246b91cd39c3bacd"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: %s; want %s", tc.name, tc.got, tc.want)
		}
	}
}

var hashForm = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func TestHashesTellPartsApart(t *testing.T) {
	two := []ReturnBase{{"sha256:" + strings.Repeat("2", 64)}, {"sha256:" + strings.Repeat("3", 64)}}
	in := func(b Binary, p []ReturnBase) InputBase { return must(CreateInvoke(b, p)) }
	ret := func(i InputBase, s Schema) ReturnBase { return must(CreateReturn(i, s)) }
	r := ret(in(hb, hpassed), shapeS)
	ids := map[string]string{
		"input":                      in(hb, hpassed).id,
		"input, another binary":      in(binX, hpassed).id,
		"input, nothing passed":      in(hb, nil).id,
		"input, two passed":          in(hb, two).id,
		"input, two passed, swapped": in(hb, []ReturnBase{two[1], two[0]}).id,
		"return":                     r.id,
		"return, another input":      ret(in(binX, hpassed), shapeS).id,
		"return, another shape":      ret(in(hb, hpassed), shapeA).id,
		"default guard":              string(RefinementOf(r, DefaultGuard())),
		"written guard":              string(RefinementOf(r, WrittenGuard("x"))),
		"another return":             string(RefinementOf(ret(in(hb, nil), shapeS), DefaultGuard())),
		// A written guard whose code is "default" is not the default guard.
		"written \"default\"": string(RefinementOf(r, WrittenGuard("default"))),
	}
	seen := map[string]string{}
	for name, id := range ids {
		if !hashForm.MatchString(id) {
			t.Errorf("%s: %q is not sha256: and 64 hex digits", name, id)
		}
		if other, ok := seen[id]; ok {
			t.Errorf("%s and %s share an id", name, other)
		}
		seen[id] = name
	}
	if in(hb, hpassed) != in(hb, hpassed) || r != ret(in(hb, hpassed), shapeS) {
		t.Error("the hashes are not deterministic")
	}
}

// The edges refuse what is not well formed: a binary without a known scheme
// and digest, a shape that is empty.
func TestEdgesRefuse(t *testing.T) {
	if _, ok := CreateInvoke("md5:abc", nil); ok {
		t.Error("createInvoke: a binary without a known scheme")
	}
	if _, ok := CreateReturn(must(CreateInvoke(hb, nil)), ""); ok {
		t.Error("createReturn: an empty shape")
	}
}

// Same parts, three domains: no two kinds of id share one, the domain tags
// keep them apart.
func TestHashDomains(t *testing.T) {
	a, b, c := sum([]any{inputTag, "p"}), sum([]any{returnTag, "p"}), sum([]any{refinementTag, "p"})
	if a == b || b == c || a == c {
		t.Error("the domain tags do not separate the hashes")
	}
}
