package protocol

import (
	"os"
	"strings"
	"testing"
)

func TestDescriptorAndObservationRejections(t *testing.T) {
	raw, err := os.ReadFile("../codex/profiles.json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := LoadDescriptor(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{strings.Replace(string(raw), `"protocol": 1`, `"protocol": 2`, 1), strings.Replace(string(raw), `"encoding": "utf8"`, `"encoding": "callback"`, 1), strings.Replace(string(raw), `"key": "configuration"`, `"key": "operation"`, 1), string(raw) + `{}`, strings.Replace(string(raw), `"operation": "health"`, `"operation": "connect"`, 1)} {
		if _, err := LoadDescriptor([]byte(bad)); err == nil {
			t.Fatal("malformed descriptor accepted")
		}
	}
	for _, args := range [][]Observation{Args("missing"), Args("health", "connection", "x", "deadline_ms", "1", "operation", "health"), Args("health", "connection", "x", "deadline_ms", "30001"), Args("health", "connection", "x", "deadline_ms", "1", "extra", "bad"), Args("health", "connection", "x"), Args("start-thread", "parent", "{}", "connection", "x", "text", "x", "tools", `["shell"]`)} {
		if _, _, err := d.Select(args, false); err == nil {
			t.Fatal("bad arguments accepted", args)
		}
	}
	profile, _, err := d.Select(Args("health", "deadline_ms", "1", "connection", "x"), false)
	if err != nil || profile.Operation != "health" {
		t.Fatal("unordered valid arguments rejected", err)
	}
}
