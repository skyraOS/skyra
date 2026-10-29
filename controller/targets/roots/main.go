// Command roots is an untrusted leaf. It replies with the paths its Boundary
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

// result is the observation. A grant the Boundary does not hold is absent.
type result struct {
	Workspace   string          `json:"workspace,omitempty"` // a relative read path resolves here
	Read        []string        `json:"read"`                // every directory or file it may read
	Permissions map[string]bool `json:"permissions,omitempty"`
	Error       string          `json:"error,omitempty"`
	Scratch     string          `json:"scratch,omitempty"` // read and write; read it as "scratch:PATH"
}

func main() {
	var in map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "roots:", err)
		os.Exit(1)
	}
	r := roots(os.Getenv("SKYRA_WORKSPACE"), os.Getenv("SKYRA_READ"), os.Getenv("SKYRA_SCRATCH"))
	r.Permissions = map[string]bool{"scratch_write": r.Scratch != "", "repo_write": false, "git_push_allowed": false, "shell": false, "delete": false}
	if changes, _ := in["changes"].(bool); changes {
		r.Error = "changes_unavailable: conversation history requires the repository host"
	}
	out, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": r})
	fmt.Println(string(out))
}

func roots(workspace, read, scratch string) result {
	r := result{Workspace: workspace, Read: []string{}, Scratch: scratch}
	for _, p := range strings.Split(read, ":") {
		if p != "" {
			r.Read = append(r.Read, p)
		}
	}
	return r
}
