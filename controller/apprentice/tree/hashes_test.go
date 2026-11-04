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
	hc      = contract(hb, Schema(`{"type":"object"}`))
	hpassed = []ReturnBase{{"sha256:" + strings.Repeat("2", 64)}}
)

func TestHashesKnownAnswers(t *testing.T) {
	in := CreateInvoke(hc, hpassed)
	ret := must(CreateReturn(in, shapeS))
	for _, tc := range []struct{ name, got, want string }{
		{"input", in.id, "sha256:eeed6b565a9469dc163b3757e1f49b23f795b1db9aed94b5609f92696f56ee3a"},
		{"input, nothing passed", CreateInvoke(hc, nil).id, "sha256:900dc53cbdfa9e7bf549dd8e1c0c56ef4c83543017c7e1b68bdf69869d23d11a"},
		{"return", ret.id, "sha256:283d0f465c4bf6e0017b53a7f678da10ee450cfd9c561770a6fd9d1b0e0bf3ad"},
		{"refinement, default", string(RefinementOf(ret, DefaultGuard())), "sha256:3f707334c26b9649a258f0017772d9a3661e11c0463fd54a83e7a1a32119be15"},
		// <, > and & are written as they are, as JSON.stringify writes them.
		{"refinement, written", string(RefinementOf(ret, WrittenGuard("a < b && x"))), "sha256:99f1853cb7d1c1b4220225864806be5f439df490327b0e034ee5b0fed78909df"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: %s; want %s", tc.name, tc.got, tc.want)
		}
	}
}

var hashForm = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func TestHashesTellPartsApart(t *testing.T) {
	two := []ReturnBase{{"sha256:" + strings.Repeat("2", 64)}, {"sha256:" + strings.Repeat("3", 64)}}
	in := func(c Contract, p []ReturnBase) InputBase { return CreateInvoke(c, p) }
	ret := func(i InputBase, s Schema) ReturnBase { return must(CreateReturn(i, s)) }
	r := ret(in(hc, hpassed), shapeS)
	ids := map[string]string{
		"input":                      in(hc, hpassed).id,
		"input, another binary":      in(x, hpassed).id,
		"input, another shape":       in(contract(hb, shapeB), hpassed).id,
		"input, nothing passed":      in(hc, nil).id,
		"input, two passed":          in(hc, two).id,
		"input, two passed, swapped": in(hc, []ReturnBase{two[1], two[0]}).id,
		"return":                     r.id,
		"return, another input":      ret(in(x, hpassed), shapeS).id,
		"return, another shape":      ret(in(hc, hpassed), shapeA).id,
		"default guard":              string(RefinementOf(r, DefaultGuard())),
		"written guard":              string(RefinementOf(r, WrittenGuard("x"))),
		"another return":             string(RefinementOf(ret(in(hc, nil), shapeS), DefaultGuard())),
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
	if in(hc, hpassed) != in(hc, hpassed) || r != ret(in(hc, hpassed), shapeS) {
		t.Error("the hashes are not deterministic")
	}
}

// Up an edge refuses a shape that is empty. Down one never refuses: a
// contract is well formed by construction.
func TestEdgesRefuse(t *testing.T) {
	if _, ok := CreateReturn(CreateInvoke(hc, nil), ""); ok {
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
