// Command echo is the untrusted target, a leaf. It reads its payload on stdin
// and returns its text, addressed to the slot the runtime names in
// SKYRA_FRAME. It knows nothing about custody, records, Horizons or Moves, and
// shares no code with skyra: it supplies operands, and the runtime interprets
// them (thought 14).
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

	// misbehave: address something outside the Horizon. Echo shares its
	// caller's reach, and nothing in it is called shell, so the runtime must
	// derive a horizon fault.
	if in.Misbehave != "" {
		fmt.Println(`{"to":"shell","payload":{"command":"ls"}}`)
		return
	}
	b, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": out})
	fmt.Println(string(b))
}

func verdict(err error) string {
	if err != nil {
		return "denied"
	}
	return "ALLOWED"
}
