package graph

import (
	"testing"

	"skyra-v1/src/skyra/spacetime/kernel"
)

// inert is a placeholder behavior. No real Behavior exists yet — what one has to do on
// descent is still open — so these tests exercise structure and boundaries, not traversal.
type inert struct{ id string }

func (b inert) ID() string { return b.id }
func (b inert) Perturb(h *Handle, origin kernel.Space, impulse any) kernel.Space {
	return nil
}
func (b inert) Observe(h *Handle, origin kernel.Space, impulse any, realized kernel.Space) kernel.Space {
	return h
}

// buildCognitive is the worked example. Note that no Node literal appears: every node
// arrives through Ground / AddExpectation / AddAction, carrying the observation that grounds it.
func buildCognitive(t *testing.T) *Namespace {
	t.Helper()
	ns := NewNamespace("skyra", inert{"cognitive/v1"})
	ns.Ground("Maintain Agent", Act{Primitive: "perceive_self"})

	respond, err := ns.AddExpectation("skyra:/", "respond", "Respond",
		Act{Primitive: "perceive_reply"},
		Observation{ID: "obs:1", Support: 0.6, Confidence: 0.8})
	if err != nil {
		t.Fatalf("expect: %v", err)
	}
	if _, err := ns.AddAction(respond.ID, "generate", Act{Primitive: "generate_response"},
		Observation{ID: "obs:2", Support: 0.5, Confidence: 0.9}); err != nil {
		t.Fatalf("perform: %v", err)
	}
	return ns
}

// An expectation cannot be built without its perceive. Not checked afterward — there is no
// call that produces one.
func TestEveryExpectationArrivesWithItsPerceive(t *testing.T) {
	ns := buildCognitive(t)

	for _, id := range []NodeID{"skyra:/", "skyra:/respond"} {
		node, ok := ns.Node(id)
		if !ok {
			t.Fatalf("missing %q", id)
		}
		found := false
		for _, e := range node.Entries {
			child, _ := ns.Node(e.Target)
			if child != nil && child.IsFile() && child.Act.Perceive {
				found = true
			}
		}
		if !found {
			t.Errorf("%q has no perceive beneath it; it cannot be wrong", id)
		}
	}
}

// Every edge carries the observation that IS it. There is no entry pointing at nothing.
func TestEveryEdgeIsAnObservation(t *testing.T) {
	ns := buildCognitive(t)
	for id, node := range ns.nodes {
		for _, e := range node.Entries {
			if e.Via == "" {
				t.Errorf("%q → %q has no observation; the structure IS the evidence", id, e.Target)
				continue
			}
			if _, ok := ns.Observation(e.Via); !ok {
				t.Errorf("%q → %q cites observation %q, which was never written", id, e.Target, e.Via)
			}
		}
	}
}

// Kind is derived from composition. Nothing is tagged.
func TestCompositionDecidesKind(t *testing.T) {
	ns := buildCognitive(t)
	respond, _ := ns.Node("skyra:/respond")
	generate, _ := ns.Node("skyra:/respond/generate")

	if !respond.IsDirectory() {
		t.Error("Respond has entries and should be a directory / expectation")
	}
	if !generate.IsFile() || generate.IsDirectory() {
		t.Error("an action is a file and must not be traversable into")
	}
}

// A FILE NEVER BECOMES A DIRECTORY. questions.md Q2 — composition-derived typing would
// otherwise let an entry under an action silently promote it.
func TestActionsCannotBeGrownIntoExpectations(t *testing.T) {
	ns := buildCognitive(t)

	_, err := ns.AddExpectation("skyra:/respond/generate", "sub", "Sub",
		Act{Primitive: "perceive_x"}, Observation{ID: "obs:9"})
	if err == nil {
		t.Fatal("attaching beneath an action succeeded; it silently became an expectation")
	}
	if _, err := ns.AddAction("skyra:/respond/generate", "sub2", Act{Primitive: "x"}, Observation{ID: "obs:10"}); err == nil {
		t.Fatal("attaching an action beneath an action succeeded")
	}
}

// Opening couples this agent's node with THIS agent's behavior, and nothing else.
func TestNamespaceOpenCouplesItsOwnBehavior(t *testing.T) {
	ns := buildCognitive(t)
	root, ok := ns.Root()
	if !ok {
		t.Fatal("no root")
	}
	if root.Label != "Maintain Agent" {
		t.Fatalf("root = %q, want the standing expectation", root.Label)
	}
	h := ns.Open(root)
	if h.Behavior().ID() != "cognitive/v1" {
		t.Fatalf("coupled %q, want this namespace's own behavior", h.Behavior().ID())
	}
	var _ kernel.Space = h
}

// A crossing spawns a real wave on the TARGET's runner. The mount holds a name and
// transport — never the target's behavior.
func TestCrossingNeverCarriesTheTargetsBehavior(t *testing.T) {
	skyra := buildCognitive(t)

	world := NewNamespace("world", inert{"world/v1"})
	world.Ground("the room", Act{Primitive: "perceive_room"})

	// Granted at CONSTRUCTION by whatever places these two together.
	skyra.Grant("room", "world", func(target string, impulse any) Receipt {
		s, err := world.Accept("skyra", impulse) // B decides, alone
		if err != nil {
			return Receipt{Target: target, Refused: true, Reason: err.Error()}
		}
		return Receipt{Target: target, WaveID: s.WaveID, Delivered: true}
	})

	receipt, err := skyra.Cross("room", nil)
	if err != nil {
		t.Fatalf("cross: %v", err)
	}
	if !receipt.Delivered || receipt.WaveID != "world/wave:1" {
		t.Fatalf("receipt = %+v, want a delivered send naming the target's new wave", receipt)
	}
	if world.Live() != 1 {
		t.Fatalf("world live = %d, want 1 admitted execution", world.Live())
	}
	if skyra.Live() != 0 {
		t.Fatalf("skyra live = %d; the sender must not hold the target's execution", skyra.Live())
	}
}

// Refusal is an outcome, not a negotiation.
func TestRefusalIsAnOutcome(t *testing.T) {
	skyra := buildCognitive(t)
	closed := NewNamespace("closed", inert{"closed/v1"}) // never Grounded: nothing to enter

	skyra.Grant("closed", "closed", func(target string, impulse any) Receipt {
		if _, err := closed.Accept("skyra", impulse); err != nil {
			return Receipt{Target: target, Refused: true, Reason: err.Error()}
		}
		return Receipt{Target: target, Delivered: true}
	})

	receipt, err := skyra.Cross("closed", nil)
	if err != nil {
		t.Fatalf("cross returned a transport error: %v", err)
	}
	if !receipt.Refused || receipt.Delivered {
		t.Fatalf("receipt = %+v, want a refusal", receipt)
	}
	if receipt.Reason == "" {
		t.Error("a refusal with no reason cannot become a graded observation")
	}
}

// An unmounted name is not a permission failure — it is simply not a name here.
func TestUnmountedNameIsNotAName(t *testing.T) {
	skyra := buildCognitive(t)
	if _, err := skyra.Cross("mars", nil); err == nil {
		t.Fatal("crossing an unmounted name succeeded")
	}
}

// The two sides keep separate records and may disagree. There is no shared edge.
func TestEachSideRecordsTheBoundaryItself(t *testing.T) {
	skyra := buildCognitive(t)
	world := NewNamespace("world", inert{"world/v1"})
	world.Ground("the room", Act{Primitive: "perceive_room"})

	if _, err := skyra.AddExpectation("skyra:/", "world-reachable", "the room answers me",
		Act{Primitive: "perceive_reply"},
		Observation{ID: "skyra:obs:world-1", Support: 0.9, Confidence: 0.9}); err != nil {
		t.Fatal(err)
	}
	if _, err := world.AddExpectation("world:/", "skyra-behaves", "skyra sends well-formed requests",
		Act{Primitive: "perceive_request"},
		Observation{ID: "world:obs:skyra-1", Support: -0.4, Confidence: 0.7}); err != nil {
		t.Fatal(err)
	}

	mine, _ := skyra.Observation("skyra:obs:world-1")
	theirs, _ := world.Observation("world:obs:skyra-1")
	if mine.Support <= 0 || theirs.Support >= 0 {
		t.Fatal("setup wrong: the two views should disagree")
	}
	if _, ok := world.Observation("skyra:obs:world-1"); ok {
		t.Error("one side's record leaked into the other")
	}
}
