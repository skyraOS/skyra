// Command echo is the untrusted target, a leaf. It reads its payload on stdin
// and prints its text. It knows nothing about custody, records, Horizons or
// Moves, and shares no code with skyra: a tool just prints its data, and the
// runtime interprets it (thought 14).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	var in struct {
		Text      string `json:"text"`
		Probe     string `json:"probe"`
		Misbehave string `json:"misbehave"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "echo:", err)
		os.Exit(1)
	}
	out := in.Text

	// probe: try to reach .skyra/ from inside the sandbox. Both must fail (D7).
	if in.Probe != "" {
		_, rerr := os.ReadDir(in.Probe)
		werr := os.WriteFile(filepath.Join(in.Probe, "forged"), []byte("x"), 0o644)
		out += fmt.Sprintf(" [probe .skyra: read %s, write %s]", verdict(rerr), verdict(werr))
	}

	// misbehave: address a capability. A leaf reaches nothing, so the runtime
	// must derive a horizon fault.
	if in.Misbehave != "" {
		out = `{"to":"echo","payload":{"text":"again"}}`
	}
	fmt.Println(out)
}

func verdict(err error) string {
	if err != nil {
		return "denied"
	}
	return "ALLOWED"
}
