package shell

import (
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
