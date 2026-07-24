package world3d

import (
	"testing"
	"time"

	"skyra-v1/src/skyra/spacetime"
	"skyra-v1/src/skyra/spacetime/relation"
	"skyra-v1/src/skyra/distributed-agent/world"
)

// stubAgent is a minimal Space that captures return impulses.
type stubAgent struct {
	id    string
	slice skyra.Slice
	got   chan skyra.Impulse
}

func newStubAgent(id string) *stubAgent {
	return &stubAgent{id: id, slice: world.NewSlice("stub:"+id, id), got: make(chan skyra.Impulse, 64)}
}

func (s *stubAgent) ID() string                            { return s.id }
func (s *stubAgent) Slice() skyra.Slice                    { return s.slice }
func (s *stubAgent) NewActivation() skyra.ActivationRecord { return newSpaceActivation() }
func (s *stubAgent) Content() any                          { return nil }
func (s *stubAgent) Agents() map[string]skyra.Space        { return nil }
func (s *stubAgent) Children() map[string]skyra.Space      { return nil }
func (s *stubAgent) Express(o skyra.Space, i any) any      { return nil }
func (s *stubAgent) Observe(o skyra.Space, impulse any) {
	if imp, ok := impulse.(skyra.Impulse); ok && imp.Return {
		s.got <- imp
	}
}

func (s *stubAgent) waitFor(t *testing.T, prim string) skyra.Impulse {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case imp := <-s.got:
			if imp.Primitive == prim {
				return imp
			}
		case <-deadline:
			t.Fatalf("no %q return impulse arrived", prim)
		}
	}
}

func testWorld() (*World, *relation.Relation) {
	kernel := relation.NewRelation()
	w := New(kernel, 8, Cell{F: 2, X: 4, Y: 4}, Cell{F: 4, X: 2, Y: 2})
	return w, kernel
}

// A move is a WALK: the agent advances one cell per tick and the return wave fires on
// arrival, carrying the destination the world says it reached.
func TestMoveWalksOverTicks(t *testing.T) {
	w, kernel := testWorld()
	agent := newStubAgent("walker")
	start := Cell{F: 0, X: 1, Y: 1}
	w.Place(agent, start)

	dest := Cell{F: 0, X: 5, Y: 1}
	kernel.Realize(agent, w, skyra.Impulse{Primitive: PrimitiveMove, TargetID: "coord:" + dest.ID()})

	steps := 0
	for i := 0; i < 100; i++ {
		w.Tick()
		steps++
		if pos, _ := w.PositionOf(agent.ID()); pos == dest {
			break
		}
	}
	if pos, _ := w.PositionOf(agent.ID()); pos != dest {
		t.Fatalf("agent at %s, want %s", pos.ID(), dest.ID())
	}
	if steps != 4 {
		t.Fatalf("walked %d ticks, want 4 (BFS distance)", steps)
	}
	imp := agent.waitFor(t, PrimitiveMove)
	res, ok := imp.Payload.(MovementResult)
	if !ok || res.Cell != dest.ID() {
		t.Fatalf("bad arrival payload: %#v", imp.Payload)
	}
}

// The perceive wave traverses the kernel and returns agents in radius, landmark
// positions, and the wave's own footprint (cell → distance) for anyone watching.
func TestPerceiveWaveFindsNeighbor(t *testing.T) {
	w, kernel := testWorld()
	a, b := newStubAgent("a"), newStubAgent("b")
	w.Place(a, Cell{F: 0, X: 2, Y: 2})
	w.Place(b, Cell{F: 0, X: 4, Y: 2}) // distance 2, well inside radius 8

	kernel.Realize(a, w, skyra.Impulse{Primitive: PrimitivePerceive})

	imp := a.waitFor(t, PrimitivePerceive)
	res, ok := imp.Payload.(PerceptionResult)
	if !ok {
		t.Fatalf("bad payload: %#v", imp.Payload)
	}
	if res.Self != (Cell{F: 0, X: 2, Y: 2}).ID() {
		t.Fatalf("self = %s", res.Self)
	}
	if got := res.Agents["b"]; got != (Cell{F: 0, X: 4, Y: 2}).ID() {
		t.Fatalf("agent b at %q in perception, want 0,4,2 (agents: %v)", got, res.Agents)
	}
	if res.Shrine != w.ShrineCell().ID() || res.Beacon != w.BeaconCell().ID() {
		t.Fatalf("landmarks wrong: %s %s", res.Shrine, res.Beacon)
	}
	if len(res.Distances) < 100 {
		t.Fatalf("wave footprint too small: %d cells", len(res.Distances))
	}
}

// The hidden recipe reports its FIRST missing prerequisite, one hint at a time:
// activate off-beacon → beacon-presence; on-beacon without sigil → sigil; take
// off-shrine → shrine-presence; the full sequence lights it.
func TestBeaconRecipeHints(t *testing.T) {
	w, _ := testWorld()
	agent := newStubAgent("seeker")
	w.Place(agent, Cell{F: 0, X: 0, Y: 0})

	if ok, missing := w.Activate("seeker"); ok || missing != "beacon-presence" {
		t.Fatalf("off-beacon activate: ok=%v missing=%q", ok, missing)
	}
	w.Place(agent, w.BeaconCell())
	if ok, missing := w.Activate("seeker"); ok || missing != "sigil" {
		t.Fatalf("no-sigil activate: ok=%v missing=%q", ok, missing)
	}
	if ok, missing := w.Take("seeker"); ok || missing != "shrine-presence" {
		t.Fatalf("off-shrine take: ok=%v missing=%q", ok, missing)
	}
	w.Place(agent, w.ShrineCell())
	if ok, _ := w.Take("seeker"); !ok {
		t.Fatal("take at shrine failed")
	}
	w.Place(agent, w.BeaconCell())
	if ok, _ := w.Activate("seeker"); !ok {
		t.Fatal("full-recipe activate failed")
	}
	if !w.Lit() {
		t.Fatal("beacon not lit")
	}
	w.ResetEpisode()
	if w.Lit() || w.Carrying("seeker") {
		t.Fatal("episode reset did not clear furniture")
	}
}

// Catch: reach is your cell or one step; only the FIRST catch writes caughtBy; failures
// hint in order (no snitch registered → snitch, already held → snitch-free, too far →
// reach).
func TestSnitchCatch(t *testing.T) {
	w, _ := testWorld()
	seeker, rival, snitch := newStubAgent("seeker"), newStubAgent("rival"), newStubAgent("snitch")
	w.Place(seeker, Cell{F: 0, X: 1, Y: 1})
	w.Place(rival, Cell{F: 0, X: 2, Y: 1})
	w.Place(snitch, Cell{F: 0, X: 5, Y: 5})

	if ok, missing := w.Catch("seeker"); ok || missing != "snitch" {
		t.Fatalf("no snitch registered: ok=%v missing=%q", ok, missing)
	}
	w.SetSnitch("snitch")
	if ok, missing := w.Catch("seeker"); ok || missing != "reach" {
		t.Fatalf("far catch: ok=%v missing=%q", ok, missing)
	}
	w.Place(seeker, Cell{F: 0, X: 5, Y: 3}) // two straight steps — still out of reach
	if ok, missing := w.Catch("seeker"); ok || missing != "reach" {
		t.Fatalf("two-step catch: ok=%v missing=%q", ok, missing)
	}
	w.Place(seeker, Cell{F: 0, X: 4, Y: 4}) // diagonal — in reach
	if ok, _ := w.Catch("seeker"); !ok {
		t.Fatal("diagonal catch failed")
	}
	w.ResetEpisode()
	w.SetSnitch("snitch")
	w.Place(seeker, Cell{F: 0, X: 5, Y: 4}) // orthogonal — in reach
	if ok, _ := w.Catch("seeker"); !ok {
		t.Fatal("adjacent catch failed")
	}
	if w.CaughtBy() != "seeker" {
		t.Fatalf("caughtBy=%q", w.CaughtBy())
	}
	w.Place(rival, Cell{F: 0, X: 5, Y: 5}) // on the snitch, but too late
	if ok, missing := w.Catch("rival"); ok || missing != "snitch-free" {
		t.Fatalf("second catch: ok=%v missing=%q", ok, missing)
	}
	w.ResetEpisode()
	if w.CaughtBy() != "" {
		t.Fatal("reset did not free the snitch")
	}
}
