package world

import (
	"sync"
	"testing"
	"time"

	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
)

type fakeTime struct {
	origin    skyra.Space
	traversed skyra.Space
	impulse   any
	calls     int
}

func (f *fakeTime) ID() string { return "fake-time" }

func (f *fakeTime) Realize(origin, traversed skyra.Space, impulse any) {
	f.origin = origin
	f.traversed = traversed
	f.impulse = impulse
	f.calls++
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
	if origin == f {
		f.observes = append(f.observes, impulse)
		return
	}
	if imp, ok := impulse.(skyra.Impulse); ok && imp.Return {
		f.observes = append(f.observes, impulse)
	}
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

func TestWorldObserveRootsActionAtActorLocation(t *testing.T) {
	time := &fakeTime{}
	w := NewSized(time, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 3, 4)

	w.Observe(actor, "engage")

	if time.calls != 1 {
		t.Fatalf("Realize calls = %d, want 1", time.calls)
	}
	if time.origin != actor {
		t.Fatalf("origin = %v, want actor", time.origin)
	}
	if time.traversed == actor {
		t.Fatalf("traversed actor interior, want world point")
	}
	if time.traversed.ID() != "3,4" {
		t.Fatalf("traversed = %q, want 3,4", time.traversed.ID())
	}
	if time.traversed.Slice() != w.Slice() {
		t.Fatalf("traversed slice is not world slice")
	}
	if time.impulse != "engage" {
		t.Fatalf("impulse = %v, want engage", time.impulse)
	}
}

func TestWorldObserveAddsImpulseSourceAtActorLocation(t *testing.T) {
	time := &fakeTime{}
	w := NewSized(time, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 3, 4)

	w.Observe(actor, skyra.Impulse{Primitive: PrimitivePerceive})

	imp, ok := time.impulse.(skyra.Impulse)
	if !ok {
		t.Fatalf("impulse type = %T, want skyra.Impulse", time.impulse)
	}
	if imp.Source == nil {
		t.Fatalf("source is nil")
	}
	// The sender self-identifies: Source.SpaceID is the actor (owner), not "world".
	if imp.Source.SpaceID != "skyra" || imp.Source.PointID != "3,4" {
		t.Fatalf("source = %+v, want sender skyra @ 3,4", imp.Source)
	}
	if imp.Source.Coord != (skyra.Coord{X: 3, Y: 4}) {
		t.Fatalf("coord = %+v, want 3,4", imp.Source.Coord)
	}
}

func TestWorldObserveNilDoesNotRelay(t *testing.T) {
	time := &fakeTime{}
	w := NewSized(time, 10)

	w.Observe(nil, "walk")

	if time.calls != 0 {
		t.Fatalf("Realize calls = %d, want 0", time.calls)
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

	if got := p.Content(); got != actor {
		t.Fatalf("content = %v, want actor", got)
	}
}

func TestPointExpressPerceive(t *testing.T) {
	w := NewSized(nil, 10)
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
	if result.Content != actor {
		t.Fatalf("content = %v, want actor", result.Content)
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

func TestWorldActionDeliversWorldSliceResultToActor(t *testing.T) {
	kernel := relation.NewRelation()
	w := NewSized(kernel, 10)
	actor := &fakeSpace{id: "skyra", slice: NewSlice("skyra-slice", "skyra")}
	w.Place(actor, 0, 0)

	w.Observe(actor, skyra.Impulse{Primitive: PrimitivePerceive, ResultMap: MapPerception, Amplitude: 1.0})

	deadline := time.After(time.Second)
	for len(actor.observed()) == 0 {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for actor delivery")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	var result map[string]any
	for _, observed := range actor.observed() {
		if imp, ok := observed.(skyra.Impulse); ok && imp.Return {
			if candidate, ok := imp.Payload.(map[string]any); ok {
				result = candidate
				break
			}
		}
	}
	if result == nil {
		t.Fatalf("no return perception result delivered; observes = %+v", actor.observed())
	}
	if len(result) == 0 {
		t.Fatalf("delivery result is empty")
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
}
