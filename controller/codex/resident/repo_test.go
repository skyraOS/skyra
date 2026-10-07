package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestRepositoryToolsReachAndExactArguments(t *testing.T) {
	for name, contract := range map[string]string{"execve": "contract.json", "repo.write": "write-contract.json", "repo.check": "check-contract.json"} {
		b, err := os.ReadFile("../../hosts/repo/" + contract)
		if err != nil {
			t.Fatal(err)
		}
		var declared struct {
			Input *schema `json:"input"`
		}
		if err := json.Unmarshal(b, &declared); err != nil {
			t.Fatal(err)
		}
		args := `{"target":"git","args":["commit","-m","message with spaces; $(literal)"]}`
		if name == "repo.write" {
			args = `{"path":"src/note.go","text":"package note\n"}`
		}
		if name == "repo.check" {
			args = `{"action":"go.test","package":"./controller/blender"}`
		}
		c := turnWithCalls(strings.ReplaceAll(name, ".", "_") + " " + args)
		c.Tools[name] = declared.Input
		to, out := emitted(t, c)
		if to != name || string(out["payload"]) != args {
			// Emission canonicalizes object keys; compare decoded values.
			var want, got any
			json.Unmarshal([]byte(args), &want)
			json.Unmarshal(out["payload"], &got)
			if to != name || !reflect.DeepEqual(want, got) {
				t.Fatalf("tool %s routed to %s: %s", name, to, out["payload"])
			}
		}
	}
}

func TestRepositoryFailuresAreUnsuccessfulProviderResults(t *testing.T) {
	for _, tc := range []struct{ target, body string }{
		{"repo.check", `{"action":"go.test","exit_code":1,"output":"FAIL","error":"check_failed"}`},
		{"repo.write", `{"error":"outside_repo"}`},
		{"execve", `{"exit_code":-1,"error":"refused: target must be git"}`},
	} {
		r := toolResult(json.RawMessage(`"call"`), step{To: tc.target, State: "return", Value: json.RawMessage(tc.body)})["result"].(map[string]any)
		if r["success"] != false {
			t.Fatalf("%s failure claimed success", tc.target)
		}
	}
}
