package provider

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSubscriptionEnvironment(t *testing.T) {
	for _, key := range []string{"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "CLAUDE_CODE_OAUTH_TOKEN", "ANTHROPIC_BASE_URL", "CLAUDE_CODE_USE_BEDROCK", "CLAUDE_CODE_USE_VERTEX", "CLAUDE_CODE_USE_FOUNDRY"} {
		if _, err := Environment([]string{key + "=secret"}); err == nil {
			t.Fatalf("accepted %s", key)
		}
	}
	got, err := Environment([]string{"HOME=/local", "PATH=/bin", "ANTHROPIC_API_KEY="})
	if err != nil || !reflect.DeepEqual(got, []string{"HOME=/local", "PATH=/bin"}) {
		t.Fatalf("%v %v", got, err)
	}
}

func TestArguments(t *testing.T) {
	args := Arguments("session-1", "-hostile prompt", "")
	joined := strings.Join(args, "|")
	for _, want := range []string{"-p", "--tools||", "--strict-mcp-config", "--resume|session-1", "--|-hostile prompt"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q: %v", want, args)
		}
	}
	if strings.Contains(joined, "skip-permissions") {
		t.Fatal("permissions bypass")
	}
}

func TestRead(t *testing.T) {
	success := `{"type":"result","subtype":"success","session_id":"s","result":"done"}`
	cases := []struct {
		name, stream string
		ok           bool
	}{
		{"success", `{"type":"system","subtype":"init","session_id":"s"}` + "\n" + success + "\n", true},
		{"EOF", `{"type":"system"}` + "\n", false},
		{"malformed", "{", false},
		{"native", `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash"}]}}`, false},
		{"failed", `{"type":"result","subtype":"error_during_execution","is_error":true}`, false},
		{"no session", `{"type":"result","subtype":"success"}`, false},
		{"extra", success + "\n" + success, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ReadWithTools(strings.NewReader(tc.stream), nil, func(e Event) error {
				if !strings.HasPrefix(string(e.Raw), "{") {
					t.Fatal("missing raw record")
				}
				return nil
			})
			if (err == nil) != tc.ok {
				t.Fatalf("error = %v", err)
			}
		})
	}
	sentinel := errors.New("record failed")
	if err := ReadWithTools(strings.NewReader(success), nil, func(Event) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
}
