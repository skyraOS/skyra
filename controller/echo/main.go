// Command echo is the untrusted target. It reads its arguments on stdin and
// emits exactly one Move on stdout. It knows nothing about custody, records
// or Horizons, and it shares no code with skyra.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type move struct {
	Invoke *struct {
		Target string            `json:"target"`
		Args   map[string]string `json:"args"`
	} `json:"invoke,omitempty"`
	Return *struct {
		Value string `json:"value"`
	} `json:"return,omitempty"`
}

func main() {
	var args map[string]string
	if err := json.NewDecoder(os.Stdin).Decode(&args); err != nil {
		fmt.Fprintln(os.Stderr, "echo:", err)
		os.Exit(1)
	}
	value := args["text"]

	// -probe: try to reach .skyra/ from inside the sandbox. Both must fail (D7).
	if dir := args["probe"]; dir != "" {
		_, rerr := os.ReadDir(dir)
		werr := os.WriteFile(filepath.Join(dir, "forged"), []byte("x"), 0o644)
		value += fmt.Sprintf(" [probe .skyra: read %s, write %s]", verdict(rerr), verdict(werr))
	}

	var m move
	if args["misbehave"] == "invoke" {
		// A leaf trying to invoke. Intake must commit this as a fault.
		m.Invoke = &struct {
			Target string            `json:"target"`
			Args   map[string]string `json:"args"`
		}{Target: "echo", Args: map[string]string{"text": "again"}}
	} else {
		m.Return = &struct {
			Value string `json:"value"`
		}{Value: value}
	}
	b, _ := json.Marshal(m)
	os.Stdout.Write(append(b, '\n'))
}

func verdict(err error) string {
	if err != nil {
		return "denied"
	}
	return "ALLOWED"
}
