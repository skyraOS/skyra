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
	hpassed = []BaseID{BaseID("sha256:" + strings.Repeat("2", 64))}
)

func TestHashesKnownAnswers(t *testing.T) {
	base := BaseOf(hb, hpassed, shapeS)
	for _, tc := range []struct{ name, got, want string }{
		{"base", string(base), "sha256:b5291f06ad90b483de6f76ab49fd6c0faf834ec69d4c0ca808044d3c11851947"},
		{"refinement, default", string(RefinementOf(base, DefaultGuard())), "sha256:b81bc5db5a45a4594e7ade06b7680fed28ebef9e078a171ca5e0b806c00c4bc1"},
		// <, > and & are written as they are, as JSON.stringify writes them.
		{"refinement, written", string(RefinementOf(base, WrittenGuard("a < b && x"))), "sha256:a9de825d32121773a98112c793177dcfc49fecf3ab964ecb621867b39e18a295"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: %s; want %s", tc.name, tc.got, tc.want)
		}
	}
}

var hashForm = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func TestHashesTellPartsApart(t *testing.T) {
	two := []BaseID{"sha256:" + BaseID(strings.Repeat("2", 64)), "sha256:" + BaseID(strings.Repeat("3", 64))}
	ids := map[string]string{
		"base":                BaseOf(hb, hpassed, shapeS).str(),
		"another binary":      BaseOf(binX, hpassed, shapeS).str(),
		"nothing passed":      BaseOf(hb, nil, shapeS).str(),
		"two passed":          BaseOf(hb, two, shapeS).str(),
		"two passed, swapped": BaseOf(hb, []BaseID{two[1], two[0]}, shapeS).str(),
		"another shape":       BaseOf(hb, hpassed, shapeA).str(),
		"default guard":       RefinementOf(BaseOf(hb, hpassed, shapeS), DefaultGuard()).str(),
		"written guard":       RefinementOf(BaseOf(hb, hpassed, shapeS), WrittenGuard("x")).str(),
		"another base":        RefinementOf(BaseOf(hb, nil, shapeS), DefaultGuard()).str(),
		// A written guard whose code is "default" is not the default guard.
		"written \"default\"": RefinementOf(BaseOf(hb, hpassed, shapeS), WrittenGuard("default")).str(),
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
	if BaseOf(hb, hpassed, shapeS) != BaseOf(hb, hpassed, shapeS) {
		t.Error("BaseOf is not deterministic")
	}
}

// Same parts, two domains: a base type and a refinement type never share an
// id, the domain tag keeps them apart.
func TestHashDomains(t *testing.T) {
	if sum([]any{baseTag, "p"}) == sum([]any{refinementTag, "p"}) {
		t.Error("the domain tags do not separate the hashes")
	}
}

func (b BaseID) str() string       { return string(b) }
func (r RefinementID) str() string { return string(r) }
