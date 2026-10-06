package prompt

import "testing"

func TestBaseIsTheHeaderLine(t *testing.T) {
	if got := Base("think"); got != "You are inside Skyra, a cognitive architecture, fulfilling: think" {
		t.Fatalf("prompt %q", got)
	}
	if got := Base(""); got != "You are inside Skyra, a cognitive architecture." {
		t.Fatalf("headerless prompt %q", got)
	}
}
