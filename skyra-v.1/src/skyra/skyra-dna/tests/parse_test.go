package tests

import (
	"testing"

	dna "skyra-v1/src/skyra/skyra-dna"
)

// Parsing the comparator response is a Space's concern (dna), not inference's.
func TestParseAlignment(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want float64
		ok   bool
	}{
		{"well-formed", "<comparison><alignment> 0.5 </alignment></comparison>", 0.5, true},
		{"no spaces", "<alignment>1</alignment>", 1, true},
		{"bare, no closing tag", "reason\n<alignment> -0.3\nmore text", -0.3, true},
		{"negative", "<alignment>-1</alignment>", -1, true},
		{"clamp over", "<alignment> 1.4 </alignment>", 1, true},
		{"clamp under", "<alignment> -2.0 </alignment>", -1, true},
		{"case-insensitive tag", "<ALIGNMENT> 0.2", 0.2, true},
		{"no alignment", "the setpoint is met", 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := dna.ParseAlignment(c.in)
			if ok != c.ok || (ok && got != c.want) {
				t.Fatalf("ParseAlignment(%q) = (%v, %v), want (%v, %v)", c.in, got, ok, c.want, c.ok)
			}
		})
	}
}
