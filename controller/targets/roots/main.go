// Command roots is an untrusted leaf. It replies with the paths its Horizon
// lets it see, as the runtime told it: the workspace (where a relative read
// resolves), every read grant, and the scratch space. It opens nothing; what
// it reports is information, and the sandbox is what enforces it. It shares
// no code with skyra.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// result is the observation. A grant the Horizon does not hold is absent.
type result struct {
	Workspace string   `json:"workspace,omitempty"` // a relative read path resolves here
	Read      []string `json:"read"`                // every directory or file it may read
	Scratch   string   `json:"scratch,omitempty"`   // read and write; read it as "scratch:PATH"
}

func main() {
	var in map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "roots:", err)
		os.Exit(1)
	}
	out, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": roots(os.Getenv("SKYRA_READ"), os.Getenv("SKYRA_SCRATCH"))})
	fmt.Println(string(out))
}

func roots(read, scratch string) result {
	r := result{Read: []string{}, Scratch: scratch}
	for _, p := range strings.Split(read, ":") {
		if p != "" {
			r.Read = append(r.Read, p)
		}
	}
	if len(r.Read) > 0 {
		r.Workspace = r.Read[0]
	}
	return r
}
