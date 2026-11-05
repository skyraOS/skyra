package shell

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// The view is the tree as Step formed it: the root, then its act child,
// which invokes echo and returns; both end done.
func TestTheViewIsTheTreeForming(t *testing.T) {
	dir := t.TempDir()
	ask := Ask{Goal: "say hello", Kind: "delegate", Actor: "model", Reach: []string{"echo"}}
	kernel(t, dir, "frame-1", ask, run)
	runs, err := View(dir)
	if err != nil || len(runs) != 1 {
		t.Fatalf("%d runs: %v", len(runs), err)
	}
	r := runs[0]
	if r.Goal != "say hello" || r.Kind != "delegate" || len(r.Frames) == 0 {
		t.Fatalf("%+v", r)
	}
	var events []string
	for _, f := range r.Frames {
		events = append(events, f.Event+" "+f.Agent)
	}
	if want := "started ,replied 0,replied 0.0,returned 0.0,replied 0.0,replied 0"; strings.Join(events, ",") != want {
		t.Fatalf("events: %s", strings.Join(events, ","))
	}
	if n := r.Frames[0].Nodes; len(n) != 1 || n[0].ID != "0" || n[0].Phase != "thinking" || n[0].Kind != "delegate" {
		t.Fatalf("after started: %+v", n)
	}
	if n := r.Frames[1].Nodes; len(n) != 2 || n[1].Parent != "0" || n[1].Kind != "act" || strings.Join(n[1].Holds, ",") != "echo" {
		t.Fatalf("after the delegate: %+v", n)
	}
	for _, n := range r.Frames[len(r.Frames)-1].Nodes {
		if n.Phase != "done" {
			t.Fatalf("at the end: %+v", n)
		}
	}
}

// The motion stream is a readable projection of accepted effects. It keeps
// spec contents, reference direction, tool inputs and actual returned values.
func TestViewActivitiesFollowAcceptedWork(t *testing.T) {
	dir := t.TempDir()
	kernel(t, dir, "f", Ask{Goal: "say hello", Kind: "delegate", Actor: "model", Reach: []string{"echo"}}, run)
	runs, err := View(dir)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, frame := range runs[0].Frames {
		for _, a := range frame.Activities {
			seen[a.Kind] = true
			if a.Kind == "pass" && (a.From != "0" || a.To != "0.0" || a.Label != "spec" || !strings.Contains(a.Preview, "say hello")) {
				t.Fatalf("spec transfer: %+v", a)
			}
			if a.Kind == "return" && (a.From != "0.0" || a.To != "0" || !strings.Contains(a.Preview, "hello")) {
				t.Fatalf("return transfer: %+v", a)
			}
			if a.Kind == "tool" && (a.Label != "echo" || !strings.Contains(a.Preview, "hello")) {
				t.Fatalf("tool invocation: %+v", a)
			}
		}
	}
	for _, kind := range []string{"cut", "pass", "return", "tool", "received"} {
		if !seen[kind] {
			t.Errorf("no %s activity", kind)
		}
	}
}

func TestRefusedInstructionsDoNotAnimateTransfers(t *testing.T) {
	dir := t.TempDir()
	bad := func(to string, payload json.RawMessage) Step {
		if to == "model" {
			return Step{To: to, State: "return", Value: json.RawMessage(`{"text":"[{\"goal\":{\"of\":0,\"lines\":[1,1]}},{\"delegate\":{\"observations\":[{\"earlier\":0}],\"contracts\":[99]}}]"}`)}
		}
		return run(to, payload)
	}
	kernel(t, dir, "f", Ask{Goal: "say hello", Kind: "delegate", Actor: "model", Reach: []string{"echo"}}, bad)
	runs, err := View(dir)
	if err != nil {
		t.Fatal(err)
	}
	fault := false
	for _, f := range runs[0].Frames {
		for _, a := range f.Activities {
			if a.Kind != "fault" {
				t.Fatalf("refused call animated %s", a.Kind)
			}
			fault = true
		}
	}
	if !fault {
		t.Fatal("missing fault activity")
	}
}

func TestSiblingJoinWaitsForAllSuccessfulReturns(t *testing.T) {
	for _, failChild := range []bool{false, true} {
		t.Run(fmt.Sprint(failChild), func(t *testing.T) {
			dir := t.TempDir()
			seeded := false
			answer := func(to string, payload json.RawMessage) Step {
				if to == "model" {
					var p struct{ Text string }
					_ = json.Unmarshal(payload, &p)
					var instructions string
					if strings.Contains(p.Text, "You are node 0 of") && !seeded {
						seeded = true
						instructions = `[{"goal":{"of":0,"lines":[1,1]}},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`
					} else if failChild && strings.Contains(p.Text, "You are node 0.1 of") {
						instructions = `not an instruction`
					}
					if instructions != "" {
						v, _ := json.Marshal(map[string]string{"text": instructions})
						return Step{To: to, State: "return", Value: v}
					}
				}
				return run(to, payload)
			}
			kernel(t, dir, "f", Ask{Goal: "say hello", Kind: "delegate", Actor: "model", Reach: []string{"echo"}}, answer)
			runs, err := View(dir)
			if err != nil {
				t.Fatal(err)
			}
			joins, complete := 0, 0
			for _, f := range runs[0].Frames {
				for _, a := range f.Activities {
					if a.Kind == "complete" {
						complete++
					}
					if a.Kind != "join" {
						continue
					}
					joins++
					if a.To != "0" || a.Count != 2 || strings.Join(a.Sources, ",") != "0.0,0.1" {
						t.Fatalf("wrong join: %+v", a)
					}
					for _, n := range f.Nodes {
						if n.Parent == "0" && n.Phase != "done" {
							t.Fatalf("join before child finished: %+v", n)
						}
					}
				}
			}
			expected := 1
			if failChild {
				expected = 0
			}
			if joins != expected || complete != 1 {
				t.Fatalf("joins=%d (want %d), successful root completions=%d (want 1)", joins, expected, complete)
			}
		})
	}
}

func TestRootExitDoesNotCelebrateCompletion(t *testing.T) {
	dir := t.TempDir()
	answer := func(to string, p json.RawMessage) Step {
		return Step{To: to, State: "return", Value: json.RawMessage(`{"text":"[{\"exit\":\"cannot continue\"}]"}`)}
	}
	kernel(t, dir, "f", Ask{Goal: "say hello", Kind: "act", Actor: "model", Reach: []string{"echo"}}, answer)
	runs, err := View(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range runs[0].Frames {
		for _, a := range f.Activities {
			if a.Kind == "complete" {
				t.Fatal("root exit celebrated as success")
			}
		}
	}
}
