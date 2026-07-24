package world

import (
	"sync"
	"testing"

	"skyra-v1/src/skyra/spacetime"
	"skyra-v1/src/skyra/spacetime/relation"
)

type fakeKernel struct {
	calls     int
	origin    skyra.Space
	traversed skyra.Space
	impulse   any
}

func (f *fakeKernel) ID() string { return "fake-kernel" }

func (f *fakeKernel) Realize(origin, traversed skyra.Space, impulse any) {
	f.calls++
	f.origin = origin
	f.traversed = traversed
	f.impulse = impulse
}

type fakeSpace struct {
	mu       sync.Mutex
	id       string
	slice    skyra.Slice
	observes []any
}

func (f *fakeSpace) ID() string                            { return f.id }
func (f *fakeSpace) Slice() skyra.Slice                    { return f.slice }
func (f *fakeSpace) NewActivation() skyra.ActivationRecord { return nil }
func (f *fakeSpace) Observe(origin skyra.Space, impulse any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.observes = append(f.observes, impulse)
}

// observed returns a snapshot of recorded impulses. Return-perception waves
// append from a spawned goroutine, so test reads must be synchronized.
func (f *fakeSpace) observed() []any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]any(nil), f.observes...)
}
func (f *fakeSpace) Express(origin skyra.Space, impulse any) any {
	return nil
}
func (f *fakeSpace) Children() map[string]skyra.Space { return nil }
func (f *fakeSpace) Content() any                     { return nil }
func (f *fakeSpace) Agents() map[string]skyra.Space   { return nil }

func TestWorldObserveCreatesPerceiveBoundary(t *testing.T) {
	kernel := &fakeKernel{}
	w := NewSized(kernel, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 3, 4)

	w.Observe(actor, skyra.Impulse{Primitive: PrimitivePerceive})

	if kernel.calls != 1 {
		t.Fatalf("Realize calls = %d, want 1", kernel.calls)
	}
	if kernel.origin != actor {
		t.Fatalf("origin = %v, want actor", kernel.origin)
	}
	if _, ok := kernel.traversed.(*PerceiveCell); !ok {
		t.Fatalf("traversed = %T, want *PerceiveCell", kernel.traversed)
	}
}

func TestWorldObserveCreatesEngageBoundary(t *testing.T) {
	kernel := &fakeKernel{}
	w := NewSized(kernel, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	target := &fakeSpace{id: "michael", slice: NewSlice("michael-slice", "michael")}
	w.Place(actor, 3, 4)
	w.Place(target, 5, 6)

	w.Observe(actor, skyra.Impulse{Primitive: PrimitiveEngage, TargetID: "michael", Payload: "hello"})

	if kernel.calls != 1 {
		t.Fatalf("Realize calls = %d, want 1", kernel.calls)
	}
	if kernel.origin != actor {
		t.Fatalf("origin = %v, want actor", kernel.origin)
	}
	boundary, ok := kernel.traversed.(*EngageCell)
	if !ok {
		t.Fatalf("traversed = %T, want *EngageCell", kernel.traversed)
	}
	if boundary.start.ID() != "5,6" {
		t.Fatalf("engage start = %q, want target point 5,6", boundary.start.ID())
	}
	imp, ok := kernel.impulse.(skyra.Impulse)
	if !ok {
		t.Fatalf("impulse = %T, want skyra.Impulse", kernel.impulse)
	}
	if imp.ResultMap != MapEngagement {
		t.Fatalf("engage result map = %q, want %q", imp.ResultMap, MapEngagement)
	}
}

func TestWorldObserveIgnoresUnknownPrimitive(t *testing.T) {
	kernel := &fakeKernel{}
	w := NewSized(kernel, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 3, 4)

	w.Observe(actor, skyra.Impulse{Primitive: "unknown"})
	w.Observe(actor, "not an impulse")

	if kernel.calls != 0 {
		t.Fatalf("Realize calls = %d, want 0", kernel.calls)
	}
}

func TestPointObserveBuildsMooreNeighborsOnWorldSlice(t *testing.T) {
	w := NewSized(nil, 10)
	p := w.newPoint(0, 0)

	p.Observe(nil, "look")

	children := p.Children()
	if len(children) != 8 {
		t.Fatalf("children = %d, want 8", len(children))
	}
	for _, id := range []string{"9,9", "0,9", "1,9", "9,0", "1,0", "9,1", "0,1", "1,1"} {
		child, ok := children[id]
		if !ok {
			t.Fatalf("missing neighbor %s in %+v", id, children)
		}
		if child.Slice() != w.Slice() {
			t.Fatalf("neighbor %s does not return world slice", id)
		}
		if child.NewActivation() == nil {
			t.Fatalf("neighbor %s has nil activation; space does not conduct", id)
		}
	}
}

func TestPointContentReadsWorldOccupancy(t *testing.T) {
	w := NewSized(nil, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 2, 3)

	p := w.newPoint(2, 3)

	occupants, ok := p.Content().(map[string]skyra.Space)
	if !ok || occupants["skyra"] != actor {
		t.Fatalf("content = %v, want occupants map containing actor", p.Content())
	}
}

func TestWorldAgentsExposePlacedSpaces(t *testing.T) {
	w := NewSized(nil, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 2, 3)

	if got := w.Agents()["skyra"]; got != actor {
		t.Fatalf("world agent skyra = %v, want actor", got)
	}
	if got := w.newPoint(2, 3).Agents()["skyra"]; got != actor {
		t.Fatalf("point agent skyra = %v, want actor", got)
	}
}

func TestPointExpressPerceive(t *testing.T) {
	w := NewSized(nil, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 0, 0)
	p := w.newPoint(0, 0)
	p.Observe(nil, nil)

	got := p.Express(nil, skyra.Impulse{Primitive: PrimitivePerceive, ResultMap: MapPerception})
	result, ok := got.(PointResult)
	if !ok {
		t.Fatalf("result type = %T, want PointResult", got)
	}
	if !result.OK {
		t.Fatalf("perceive failed: %+v", result)
	}
	if result.PointID != "0,0" {
		t.Fatalf("point id = %q, want 0,0", result.PointID)
	}
	if occ, ok := result.Content.(map[string]skyra.Space); !ok || occ["skyra"] != actor {
		t.Fatalf("content = %v, want occupants map containing actor", result.Content)
	}
	if result.Agents["skyra"] != actor {
		t.Fatalf("agents = %+v, want skyra occupant", result.Agents)
	}
	if result.Source == nil {
		t.Fatalf("source is nil")
	}
	if result.Source.SpaceID != "world" || result.Source.PointID != "0,0" {
		t.Fatalf("source = %+v, want world 0,0", result.Source)
	}
	if result.Source.Coord != (skyra.Coord{X: 0, Y: 0}) {
		t.Fatalf("coord = %+v, want 0,0", result.Source.Coord)
	}
	if len(result.Points) != 8 {
		t.Fatalf("points = %d, want 8", len(result.Points))
	}
	if result.Points["1,0"] == nil {
		t.Fatalf("missing actual point 1,0 in %+v", result.Points)
	}
	if result.Points["1,0"].ID() != "1,0" {
		t.Fatalf("point ID = %q, want 1,0", result.Points["1,0"].ID())
	}
}

func TestPointExpressEngage(t *testing.T) {
	w := NewSized(nil, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 2, 3)
	p := w.newPoint(2, 3)

	got := p.Express(nil, skyra.Impulse{Primitive: PrimitiveEngage, ResultMap: MapEngagement})
	result, ok := got.(PointResult)
	if !ok {
		t.Fatalf("result type = %T, want PointResult", got)
	}
	if !result.OK {
		t.Fatalf("engage failed: %+v", result)
	}
	if occ, ok := result.Content.(map[string]skyra.Space); !ok || occ["skyra"] != actor {
		t.Fatalf("content = %v, want occupants map containing actor", result.Content)
	}
	if result.Agents["skyra"] != actor {
		t.Fatalf("agents = %+v, want skyra occupant", result.Agents)
	}
}

func TestPointExpressRejectsInvalidImpulse(t *testing.T) {
	w := NewSized(nil, 10)
	p := w.newPoint(0, 0)

	got := p.Express(nil, "engage")
	result, ok := got.(PointResult)
	if !ok {
		t.Fatalf("result type = %T, want PointResult", got)
	}
	if result.OK {
		t.Fatalf("invalid impulse succeeded: %+v", result)
	}
}

func TestSliceStoresResult(t *testing.T) {
	sl := NewSlice("slice", "owner")
	perception := map[string]any{"0,0": PointResult{PointID: "0,0", OK: true}}
	sl.Maps()[MapPerception] = perception

	sl.SetResult(sl.Maps())

	got, ok := sl.Result().(map[string]any)
	if !ok {
		t.Fatalf("result type = %T, want map[string]any", sl.Result())
	}
	gotPerception, ok := got[MapPerception].(map[string]any)
	if !ok {
		t.Fatalf("perception map = %T, want map[string]any", got[MapPerception])
	}
	gotPoint, ok := gotPerception["0,0"].(PointResult)
	if !ok {
		t.Fatalf("point result = %T, want PointResult", gotPerception["0,0"])
	}
	wantPoint := perception["0,0"].(PointResult)
	if gotPoint.PointID != wantPoint.PointID || gotPoint.OK != wantPoint.OK {
		t.Fatalf("result = %v, want stored perception result", sl.Result())
	}
}

func TestWorldPerceiveBoundaryRecordsAndDeliversPerception(t *testing.T) {
	kernel := relation.NewRelation()
	w := NewSized(kernel, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 0, 0)

	w.Observe(actor, skyra.Impulse{Primitive: PrimitivePerceive, ResultMap: MapPerception, Amplitude: 1.0})

	result, ok := w.Slice().Result().(map[string]any)
	if !ok {
		t.Fatalf("world slice result = %T, want map[string]any", w.Slice().Result())
	}
	if len(result) == 0 {
		t.Fatalf("world slice result is empty")
	}
	perception, ok := result[MapPerception].(map[string]any)
	if !ok {
		t.Fatalf("perception map = %T, want map[string]any", result[MapPerception])
	}
	point, ok := perception["0,0"].(PointResult)
	if !ok {
		t.Fatalf("root point result = %T, want PointResult", perception["0,0"])
	}
	if point.PointID != "0,0" || point.Primitive != PrimitivePerceive || !point.OK {
		t.Fatalf("root point result = %+v, want successful perceive at 0,0", point)
	}
	if point.Agents["skyra"] != actor {
		t.Fatalf("root point agents = %+v, want skyra", point.Agents)
	}
	empty, ok := perception["1,0"].(PointResult)
	if !ok {
		t.Fatalf("neighbor point result missing; perception = %+v", perception)
	}
	if empty.PointID != "1,0" || len(empty.Agents) != 0 {
		t.Fatalf("empty point result = %+v, want empty 1,0", empty)
	}

	// Perception data lives in world.slice.Maps(), not on the delivered impulse
	// payload. The impulse carries the original payload (utterance) the caller
	// supplied — the being reads who spoke and what they said off Source + Payload.
	var delivered skyra.Impulse
	for _, observed := range actor.observed() {
		imp, ok := observed.(skyra.Impulse)
		if ok && imp.Primitive == PrimitivePerceive {
			delivered = imp
			break
		}
	}
	if delivered.Primitive != PrimitivePerceive {
		t.Fatalf("no perceive delivery found; observed = %+v", actor.observed())
	}
}

func TestWorldEngageBoundaryDeliversImpulseAndReturnsRecipients(t *testing.T) {
	kernel := relation.NewRelation()
	w := NewSized(kernel, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	target := &fakeSpace{id: "michael", slice: NewSlice("michael-slice", "michael")}
	w.Place(actor, 0, 0)
	w.Place(target, 2, 3)

	w.Observe(actor, skyra.Impulse{
		Primitive: PrimitiveEngage,
		TargetID:  "michael",
		Payload:   "hello michael",
		Amplitude: 1.0,
	})

	result, ok := w.Slice().Result().(map[string]any)
	if !ok {
		t.Fatalf("world slice result = %T, want map[string]any", w.Slice().Result())
	}
	engagement, ok := result[MapEngagement].(map[string]any)
	if !ok {
		t.Fatalf("engagement map = %T, want map[string]any", result[MapEngagement])
	}
	point, ok := engagement["2,3"].(PointResult)
	if !ok {
		t.Fatalf("target point result = %T, want PointResult", engagement["2,3"])
	}
	if point.PointID != "2,3" || point.Primitive != PrimitiveEngage || !point.OK {
		t.Fatalf("target point result = %+v, want successful engage at 2,3", point)
	}
	if point.Agents["michael"] != target {
		t.Fatalf("target point agents = %+v, want michael", point.Agents)
	}

	var targetDelivery skyra.Impulse
	for _, observed := range target.observed() {
		imp, ok := observed.(skyra.Impulse)
		if ok && imp.Primitive == PrimitiveEngage && !imp.Return {
			targetDelivery = imp
			break
		}
	}
	if targetDelivery.TargetID != "michael" || targetDelivery.Payload != "hello michael" {
		t.Fatalf("target delivery = %+v, want original engage impulse", targetDelivery)
	}

	var returned skyra.Impulse
	for _, observed := range actor.observed() {
		imp, ok := observed.(skyra.Impulse)
		if ok && imp.Primitive == PrimitiveEngage && imp.Return {
			returned = imp
			break
		}
	}
	report, ok := returned.Payload.(EngagementResult)
	if !ok {
		t.Fatalf("return payload = %T, want EngagementResult; observed = %+v", returned.Payload, actor.observed())
	}
	if report.Delivered["michael"] != target {
		t.Fatalf("delivered = %+v, want michael target", report.Delivered)
	}
}

// Regression for the 2026-07-02 ghost bug (captains-log/observations/
// 03-first-embodied-runs.md): agents overlap on a cell, and the world must
// keep every occupant addressable. Under the old map[coord]Space placement,
// the last arrival overwrote the occupant — the buried agent stayed in
// position but vanished from occupancy, and agentByID (which resolved through
// the cell) returned the burier. Skyra stood on the exit and erased it;
// michael stood on skyra and erased her; every engage "to skyra" after that
// was delivered to michael alone.
func TestOverlapKeepsAllOccupantsAddressable(t *testing.T) {
	w := NewSized(nil, 20)
	exit := &fakeSpace{id: "exit", slice: NewSlice("exit-slice", "exit")}
	skyraB := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	michael := &fakeSpace{id: "michael", slice: NewSlice("michael-slice", "michael")}

	w.Place(exit, 7, 8)
	w.Place(skyraB, 5, 5)
	w.Place(michael, 3, 3)

	// skyra steps onto the exit's cell — overlap, exit reported as occupant.
	occupants, moved := w.Place(skyraB, 7, 8)
	if !moved {
		t.Fatalf("overlap move did not complete")
	}
	if len(occupants) != 1 || occupants[0] != exit {
		t.Fatalf("occupants = %v, want [exit]", occupants)
	}

	// michael joins the same cell — both existing occupants reported.
	occupants, _ = w.Place(michael, 7, 8)
	if len(occupants) != 2 {
		t.Fatalf("occupants = %v, want exit and skyra", occupants)
	}

	// the cell reports every occupant (perception/delivery read path).
	agents := w.newPoint(7, 8).Agents()
	for _, id := range []string{"exit", "skyra", "michael"} {
		if agents[id] == nil {
			t.Fatalf("point agents missing %q: %+v", id, agents)
		}
	}

	// identity never routes through occupancy: everyone still resolves to
	// themselves, no matter who stands on whom.
	if got := w.agentByID("exit"); got != exit {
		t.Fatalf("agentByID(exit) = %v, want exit", got)
	}
	if got := w.agentByID("skyra"); got != skyraB {
		t.Fatalf("agentByID(skyra) = %v, want skyra", got)
	}

	// michael moves away — the others remain on the cell, michael lands cleanly.
	w.Place(michael, 9, 9)
	agents = w.newPoint(7, 8).Agents()
	if agents["michael"] != nil {
		t.Fatalf("michael still on 7,8 after leaving: %+v", agents)
	}
	if agents["exit"] == nil || agents["skyra"] == nil {
		t.Fatalf("departure evicted co-occupants: %+v", agents)
	}
	if got := w.newPoint(9, 9).Agents()["michael"]; got != michael {
		t.Fatalf("michael not at 9,9 after move")
	}
}
