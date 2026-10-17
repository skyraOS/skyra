package provider

import (
	"strings"
	"testing"
)

func TestBridgeArguments(t *testing.T) {
	args, err := BridgeArguments("s", "-prompt", "", "http://127.0.0.1:1234/mcp", strings.Repeat("x", 32), []string{"echo", "repo_write"})
	if err != nil { t.Fatal(err) }
	joined := strings.Join(args, "|")
	for _, want := range []string{"--tools||", "--strict-mcp-config", "mcp__skyra__echo,mcp__skyra__repo_write", "--resume|s", "--|-prompt"} {
		if !strings.Contains(joined, want) { t.Errorf("missing %s", want) }
	}
	for _, endpoint := range []string{"http://evil.invalid:1234/mcp", "https://127.0.0.1:1234/mcp", "http://127.0.0.1/mcp", "http://user@127.0.0.1:1234/mcp", "http://127.0.0.1:1234/mcp?secret=x"} {
		if _, err := BridgeArguments("", "", "", endpoint, strings.Repeat("x",32), nil); err == nil { t.Errorf("accepted %s", endpoint) }
	}
	if _, err := BridgeArguments("", "", "", "http://127.0.0.1:1234/mcp", strings.Repeat("x",32), []string{"*"}); err == nil { t.Fatal("wildcard accepted") }
}

func TestPartialToolEnforcement(t *testing.T) {
	final := `{"type":"result","subtype":"success","session_id":"s"}`
	for _, tc := range []struct { name string; ok bool }{{"mcp__skyra__echo",true},{"Bash",false},{"mcp__ambient__echo",false}} {
		stream := `{"type":"stream_event","event":{"content_block":{"type":"tool_use","name":"`+tc.name+`"}}}`+"\n"+final
		err := ReadWithTools(strings.NewReader(stream), []string{"echo"}, func(Event) error { return nil })
		if (err == nil) != tc.ok { t.Errorf("%s: %v",tc.name,err) }
	}
	stream := `{"type":"system","session_id":"other"}`+"\n"+final
	if ReadWithTools(strings.NewReader(stream), nil, func(Event) error { return nil }) == nil { t.Fatal("session changed") }
}
