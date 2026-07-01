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

func TestParseMotorEmissionEngage(t *testing.T) {
	got, ok := dna.ParseMotorEmission("<engage><michael>walk through the kernel</michael></engage>")
	if !ok || got.Primitive != "engage" || got.Target != "michael" || got.Payload != "walk through the kernel" {
		t.Fatalf("ParseMotorEmission engage = (%+v, %v)", got, ok)
	}
}

func TestParseMotorEmissionPerceive(t *testing.T) {
	got, ok := dna.ParseMotorEmission(`<perceive>"nothing important yet"</perceive>`)
	if !ok || got.Primitive != "perceive" || got.Target != "" || got.Payload != "nothing important yet" {
		t.Fatalf("ParseMotorEmission perceive = (%+v, %v)", got, ok)
	}
}

func TestParseMotorEmissionToleratesPercieveTypo(t *testing.T) {
	got, ok := dna.ParseMotorEmission(`<percieve>"nothing important yet"</percieve>`)
	if !ok || got.Primitive != "perceive" || got.Payload != "nothing important yet" {
		t.Fatalf("ParseMotorEmission percieve = (%+v, %v)", got, ok)
	}
}
