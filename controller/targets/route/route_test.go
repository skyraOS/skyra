package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func route(source, ret, prev string) Decision {
	if prev == "" {
		prev = "null"
	}
	return Run(Input{Source: source, Return: json.RawMessage(ret), Previous: json.RawMessage(prev), Targets: []string{"echo", "log"}})
}

func faults(t *testing.T, d Decision, cause string) {
	t.Helper()
	if d.Fault == "" || d.Interrupt || d.Send != nil || !strings.Contains(d.Fault, cause) {
		t.Fatalf("want a fault decision mentioning %q, got %+v", cause, d)
	}
}

const ret = `{"text":"hi","media":[{"url":"a"}],"a b":1}`

func TestBuiltins(t *testing.T) {
	for _, c := range []struct {
		name, source, prev string
		want               Decision
	}{
		{"interrupt", "def route():\n  return skyra.interrupt()", "", Decision{Interrupt: true}},
		{"fault", "def route():\n  return skyra.fault(\"truncated\")", "", Decision{Fault: "truncated"}},
		{"send, built from get", "def route():\n  return skyra.send(\"log\", {\"line\": skyra.get(\"$.text\"), \"url\": skyra.get(\"$.media[0].url\"), \"n\": skyra.get('$[\"a b\"]')})", "",
			Decision{Send: &Envelope{To: "log", Payload: json.RawMessage(`{"line":"hi","n":1,"url":"a"}`)}}},
		{"previous is None without one", "def route():\n  if skyra.previous() == None:\n    return skyra.interrupt()\n  return skyra.fault(\"x\")", "", Decision{Interrupt: true}},
		{"previous is the earlier return", "def route():\n  if skyra.previous()[\"text\"] == \"old\":\n    return skyra.interrupt()\n  return skyra.fault(\"x\")", `{"text":"old","media":[],"a b":1}`, Decision{Interrupt: true}},
		{"changed", "def route():\n  if skyra.changed(\"$.text\") and not skyra.changed(\"$.media\"):\n    return skyra.interrupt()\n  return skyra.fault(\"x\")", `{"text":"old","media":[{"url":"a"}],"a b":1}`, Decision{Interrupt: true}},
		{"changed without a previous", "def route():\n  if skyra.changed(\"$.text\"):\n    return skyra.interrupt()\n  return skyra.fault(\"x\")", "", Decision{Interrupt: true}},
		{"diff", "def route():\n  if skyra.diff() > 0.3 and skyra.diff() < 0.4:\n    return skyra.interrupt()\n  return skyra.fault(\"x\")", `{"text":"old","media":[{"url":"a"}],"a b":1}`, Decision{Interrupt: true}},
		{"diff without a previous", "def route():\n  if skyra.diff() == 1.0:\n    return skyra.interrupt()\n  return skyra.fault(\"x\")", "", Decision{Interrupt: true}},
	} {
		if got := route(c.source, ret, c.prev); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

// A route cannot loop: while and for are refused before it runs, recursion
// when it is called, and the step cap is the backstop. Each is a fault
// decision, never a crash.
func TestARouteCannotLoop(t *testing.T) {
	faults(t, route("def route():\n  while True:\n    pass\n  return skyra.interrupt()", ret, ""), "while")
	faults(t, route("def route():\n  for m in skyra.get(\"$.media\"):\n    pass\n  return skyra.interrupt()", ret, ""), "for is not allowed")
	faults(t, route("def route():\n  x = [m for m in skyra.get(\"$.media\")]\n  return skyra.interrupt()", ret, ""), "for is not allowed")
	faults(t, route("def f():\n  return f()\ndef route():\n  return f()", ret, ""), "recursively")
	faults(t, route("def route():\n  x = 0\n"+strings.Repeat("  x = x + 1\n", 40000)+"  return skyra.interrupt()", ret, ""), "too many steps")
}

func TestNoAmbientAuthority(t *testing.T) {
	faults(t, route("load(\"x.star\", \"y\")\ndef route():\n  return skyra.interrupt()", ret, ""), "load")
	faults(t, route("def route():\n  return skyra.send(\"shell\", {})", ret, ""), `"shell" is not in the parent's reach`)
}

func TestExactlyOneDecision(t *testing.T) {
	faults(t, route("def route():\n  pass", ret, ""), "not a decision")
	faults(t, route("def route():\n  skyra.interrupt()\n  return skyra.fault(\"x\")", ret, ""), "2 decisions")
	faults(t, route("def route():\n  return [skyra.interrupt(), skyra.interrupt()]", ret, ""), "not a decision")
	faults(t, route("def other():\n  return skyra.interrupt()", ret, ""), "def route():")
	faults(t, route("def route(:", ret, ""), "route.star")
	faults(t, route("def route():\n  return skyra.get(\"$.nothing\")", ret, ""), "no $.nothing")
	faults(t, route("def route():\n  return skyra.fault(\"\")", ret, ""), "a cause is required")
}

// The same input run twice gives byte-identical output, through the binary.
func TestDeterministic(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "route")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("%s: %v", out, err)
	}
	in, _ := json.Marshal(Input{Source: "def route():\n  if skyra.diff() > 0.3:\n    return skyra.send(\"log\", {\"d\": skyra.diff(), \"t\": skyra.get(\"$.text\")})\n  return skyra.interrupt()",
		Return: json.RawMessage(ret), Previous: json.RawMessage(`{"text":"old","media":[{"url":"a"}],"a b":1}`), Targets: []string{"log"}})
	runOnce := func() string {
		cmd := exec.Command(bin)
		cmd.Stdin, cmd.Env = strings.NewReader(string(in)), append(os.Environ(), "SKYRA_FRAME=f1")
		out, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		return string(out)
	}
	a, b := runOnce(), runOnce()
	if a != b || a != `{"payload":{"send":{"to":"log","payload":{"d":0.3333333333333333,"t":"hi"}}},"to":"f1"}`+"\n" {
		t.Fatalf("not byte-identical, or not the decision:\n%s%s", a, b)
	}
}
