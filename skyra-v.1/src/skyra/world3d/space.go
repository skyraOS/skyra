package world3d

import (
	"fmt"
	"strings"
	"sync"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/world"
)

// The globe world: the flat world's kernel contract on a stitched sphere.
//
// Same architecture as world/space.go — the world is a skyra.Space, primitives arrive as
// impulses through Observe, perception is a WAVE the kernel traverses (boundary cell →
// cells-in-radius → collect → return wave), and results only ever come back as return
// impulses into the origin's Observe. Two deliberate differences:
//
//   - MOVEMENT IS A WALK, NOT A TELEPORT. A move registers a path and the agent advances
//     one cell per world tick; the return wave fires on ARRIVAL. Distance costs time,
//     which makes `within` deadlines physical and makes the globe watchable.
//
//   - THE WORLD CARRIES TASK FURNITURE. A shrine, a sigil, a beacon: the beacon lights
//     only for an agent standing on it AND carrying the sigil, and a failed activation
//     reports the FIRST missing prerequisite — the same one-hint-at-a-time shape as the
//     hidden build recipe, so the same learning curve has somewhere to live out here.
//     The world REPORTS; it never grades (it has never seen a prediction).
const (
	DefaultFaceSize       = 12
	DefaultPerceiveRadius = 8
	spaceBaseline         = 0.10
	spaceThreshold        = 0.001
	PrimitivePerceive     = "perceive"
	PrimitiveEngage       = "engage"
	PrimitiveMove         = "move"
	PrimitiveTake         = "take"
	PrimitiveActivate     = "activate"
	PrimitiveCatch        = "catch"
	MapPerception         = "perception"
	globeSliceID          = "globe-slice"
	globeSliceOwner       = "globe"
	defaultGlobeID        = "globe"
)

// Event is the world narrating itself for an observer (the frontend). Not part of
// cognition — no being ever sees these.
type Event struct {
	Kind  string         `json:"kind"` // step | arrive | perceive | take | activate | engage
	Agent string         `json:"agent,omitempty"`
	Cell  string         `json:"cell,omitempty"`
	Extra map[string]any `json:"extra,omitempty"`
}

type walkState struct {
	origin  skyra.Space
	path    []Cell
	at      int
	impulse skyra.Impulse
}

type World struct {
	id     string
	kernel skyra.Time
	slice  skyra.Slice
	globe  *Globe

	mu        sync.RWMutex
	placement map[Cell]map[string]skyra.Space
	position  map[string]Cell
	byID      map[string]skyra.Space
	walks     map[string]*walkState

	shrine   Cell
	beacon   Cell
	carrying map[string]bool
	lit      bool
	tick     int

	// the snitch: a registered agent some game is currently hunting. caughtBy is the
	// game's whole outcome — the world records WHO, and only the first catch writes it.
	snitchID string
	caughtBy string

	onEvent func(Event)
}

func New(kernel skyra.Time, faceSize int, shrine, beacon Cell) *World {
	if faceSize <= 0 {
		faceSize = DefaultFaceSize
	}
	slice := world.NewSliceWithThreshold(globeSliceID, globeSliceOwner, spaceThreshold)
	slice.Maps()[MapPerception] = map[string]any{}
	return &World{
		id:        defaultGlobeID,
		kernel:    kernel,
		slice:     slice,
		globe:     NewGlobe(faceSize),
		placement: map[Cell]map[string]skyra.Space{},
		position:  map[string]Cell{},
		byID:      map[string]skyra.Space{},
		walks:     map[string]*walkState{},
		shrine:    shrine,
		beacon:    beacon,
		carrying:  map[string]bool{},
	}
}

func (w *World) ID() string                            { return w.id }
func (w *World) Slice() skyra.Slice                    { return w.slice }
func (w *World) NewActivation() skyra.ActivationRecord { return newSpaceActivation() }
func (w *World) Express(origin skyra.Space, impulse any) any { return nil }
func (w *World) Content() any                          { return nil }
func (w *World) Globe() *Globe                         { return w.globe }
func (w *World) ShrineCell() Cell                      { return w.shrine }
func (w *World) BeaconCell() Cell                      { return w.beacon }

// SetOnEvent installs the narration hook. Called before the world starts ticking.
func (w *World) SetOnEvent(fn func(Event)) { w.onEvent = fn }

func (w *World) emit(e Event) {
	if w.onEvent != nil {
		w.onEvent(e)
	}
}

// ResetEpisode clears the task furniture (sigil carried, beacon lit, snitch catch) and
// abandons any in-flight walks. Placement survives — the agents are still standing where
// they stand; it is the WORLD'S state that renews, the way the build sandbox is remade.
func (w *World) ResetEpisode() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.carrying = map[string]bool{}
	w.lit = false
	w.caughtBy = ""
	w.walks = map[string]*walkState{}
}

// SetSnitch registers which agent is the snitch. Place it like any agent first.
func (w *World) SetSnitch(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.snitchID = id
	w.caughtBy = ""
}

// SnitchAt reports the snitch's cell — a LIVE fact, staleness is the hunter's problem.
func (w *World) SnitchAt() (Cell, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.snitchID == "" {
		return Cell{}, false
	}
	c, ok := w.position[w.snitchID]
	return c, ok
}

// CaughtBy reports who holds the snitch; "" while it flies.
func (w *World) CaughtBy() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.caughtBy
}

// Catch attempts to grab the snitch: within REACH (your cell or one step) and only while
// it is still free. First catch wins — the write is guarded by the same lock that reads,
// so two simultaneous swipes cannot both succeed. Failures report the first missing
// prerequisite, the world's usual one-hint-at-a-time.
func (w *World) Catch(agentID string) (ok bool, missing string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.snitchID == "" {
		return false, "snitch"
	}
	if w.caughtBy != "" {
		return false, "snitch-free"
	}
	pos, hasPos := w.position[agentID]
	spos, hasSnitch := w.position[w.snitchID]
	if !hasPos || !hasSnitch {
		return false, "reach"
	}
	if pos != spos && !withinReach(w.globe, pos, spos) {
		return false, "reach"
	}
	w.caughtBy = agentID
	delete(w.walks, w.snitchID) // a caught snitch stops mid-flight
	return true, ""
}

// withinReach is the 8-cell ring: orthogonal neighbors, plus diagonals — recognized
// graph-theoretically as cells two steps away via TWO distinct paths (a straight
// two-step has one; a diagonal corner has two), which keeps the definition exact across
// stitched face edges where "diagonal" has no coordinate meaning.
func withinReach(g *Globe, a, b Cell) bool {
	twoStep := 0
	for _, n := range g.Neighbors(a) {
		if n == b {
			return true
		}
		for _, nn := range g.Neighbors(n) {
			if nn == b {
				twoStep++
				break
			}
		}
	}
	return twoStep >= 2
}

// Positions snapshots every placed agent — the frontend's initial state.
func (w *World) Positions() map[string]string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make(map[string]string, len(w.position))
	for id, c := range w.position {
		out[id] = c.ID()
	}
	return out
}

func (w *World) Agents() map[string]skyra.Space {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make(map[string]skyra.Space, len(w.byID))
	for id, s := range w.byID {
		out[id] = s
	}
	return out
}

// Children exposes the occupied cells as Points — the traversal graph mirrors the flat
// world: only where something stands is there anything to walk to from the world root.
func (w *World) Children() map[string]skyra.Space {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make(map[string]skyra.Space, len(w.placement))
	for c := range w.placement {
		p := w.newPoint(c, 0)
		out[p.ID()] = p
	}
	return out
}

// Place stands an agent on a cell. Occupancy is a set — agents overlap, nobody is
// overwritten, nobody ghosts (the flat world learned this one the hard way).
func (w *World) Place(space skyra.Space, c Cell) {
	if space == nil || !w.globe.Valid(c) {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.placeLocked(space, c)
}

func (w *World) placeLocked(space skyra.Space, c Cell) {
	if old, ok := w.position[space.ID()]; ok {
		if cell := w.placement[old]; cell != nil {
			delete(cell, space.ID())
			if len(cell) == 0 {
				delete(w.placement, old)
			}
		}
	}
	if w.placement[c] == nil {
		w.placement[c] = map[string]skyra.Space{}
	}
	w.placement[c][space.ID()] = space
	w.position[space.ID()] = c
	w.byID[space.ID()] = space
}

// PositionOf reports where an agent stands. A world instrument: authoritative, free.
func (w *World) PositionOf(agentID string) (Cell, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	c, ok := w.position[agentID]
	return c, ok
}

// Carrying reports whether an agent holds the sigil.
func (w *World) Carrying(agentID string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.carrying[agentID]
}

// Lit reports the beacon — a LIVE fact, like build status: checked, never stored stale.
func (w *World) Lit() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lit
}

// Take attempts to pick up the sigil. Reports what happened and, on failure, the FIRST
// missing prerequisite — a hint the world gives, never a judgment.
func (w *World) Take(agentID string) (ok bool, missing string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	pos, placed := w.position[agentID]
	if !placed || pos != w.shrine {
		return false, "shrine-presence"
	}
	w.carrying[agentID] = true
	return true, ""
}

// Activate attempts to light the beacon. The hidden recipe, discovered one hint at a
// time: stand on the beacon, carry the sigil.
func (w *World) Activate(agentID string) (ok bool, missing string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	pos, placed := w.position[agentID]
	if !placed || pos != w.beacon {
		return false, "beacon-presence"
	}
	if !w.carrying[agentID] {
		return false, "sigil"
	}
	w.lit = true
	return true, ""
}

// Tick advances world time: every in-flight walk moves one cell; a walk that reaches its
// destination fires the move's RETURN WAVE into its origin. This is the world running
// over the time kernel — distance is duration here.
func (w *World) Tick() {
	w.mu.Lock()
	w.tick++
	type arrival struct {
		origin skyra.Space
		cell   Cell
	}
	var steps []Event
	var arrivals []arrival
	for id, walk := range w.walks {
		walk.at++
		c := walk.path[walk.at]
		w.placeLocked(w.byID[id], c)
		steps = append(steps, Event{Kind: "step", Agent: id, Cell: c.ID()})
		if walk.at == len(walk.path)-1 {
			arrivals = append(arrivals, arrival{origin: walk.origin, cell: c})
			steps = append(steps, Event{Kind: "arrive", Agent: id, Cell: c.ID()})
			delete(w.walks, id)
		}
	}
	w.mu.Unlock()

	for _, e := range steps {
		w.emit(e)
	}
	for _, a := range arrivals {
		w.kernel.Realize(a.origin, a.origin, skyra.Impulse{
			Primitive: PrimitiveMove,
			Payload:   MovementResult{Cell: a.cell.ID()},
			Source:    &skyra.ImpulseSource{SpaceID: w.id},
			Return:    true,
		})
	}
}

// MovementResult is the return payload on arrival.
type MovementResult struct {
	Cell string
}

// MoveBlockedResult is the return payload when a move cannot start.
type MoveBlockedResult struct {
	Target string
	Reason string
}

// TaskResult is the return payload for take/activate — what the world did, and on
// failure the first missing prerequisite.
type TaskResult struct {
	Primitive string
	OK        bool
	Missing   string
}

// PerceptionResult is the return payload from a perceive wave.
type PerceptionResult struct {
	Self      string            // the origin's cell at wave time
	Agents    map[string]string // agent id → cell id, for each agent in radius
	Shrine    string            // landmark: always visible (the towers stand tall)
	Beacon    string
	BeaconLit bool
	Distances map[string]int // cell id → graph distance — the wave's footprint
}

// EngagementResult is the return payload after an engage: who it reached.
type EngagementResult struct {
	TargetID  string
	Delivered bool
}

// Observe is the world's primitive boundary — impulses in, waves out.
func (w *World) Observe(origin skyra.Space, impulse any) {
	imp, ok := asImpulse(impulse)
	if !ok {
		return
	}
	switch imp.Primitive {
	case PrimitivePerceive:
		w.observePerceive(origin, imp)
	case PrimitiveMove:
		w.observeMove(origin, imp)
	case PrimitiveTake:
		w.observeTask(origin, imp, PrimitiveTake)
	case PrimitiveActivate:
		w.observeTask(origin, imp, PrimitiveActivate)
	case PrimitiveCatch:
		w.observeTask(origin, imp, PrimitiveCatch)
	case PrimitiveEngage:
		w.observeEngage(origin, imp)
	}
}

func (w *World) observePerceive(origin skyra.Space, impulse skyra.Impulse) {
	if origin == nil || w.kernel == nil {
		return
	}
	start, placed := w.PositionOf(origin.ID())
	if !placed {
		return
	}
	impulse.ResultMap = MapPerception
	impulse.Amplitude = DefaultPerceiveRadius
	debug.Being("globe", "perceive", fmt.Sprintf("origin=%s at=%s radius=%d", origin.ID(), start.ID(), DefaultPerceiveRadius))
	boundary := &PerceiveCell{
		id:      "globe:perceive:" + origin.ID(),
		world:   w,
		owner:   origin,
		start:   start,
		radius:  DefaultPerceiveRadius,
		impulse: impulse,
	}
	w.kernel.Realize(origin, boundary, impulse)
}

// observeMove parses the destination ("f,x,y", "coord:f,x,y", or an agent id), computes
// the BFS path, and registers the walk. The return wave fires from Tick on arrival.
func (w *World) observeMove(origin skyra.Space, impulse skyra.Impulse) {
	if origin == nil || w.kernel == nil {
		return
	}
	dest, ok := w.resolveDestination(impulse.TargetID)
	if !ok {
		go w.kernel.Realize(origin, origin, skyra.Impulse{
			Primitive: PrimitiveMove,
			Payload:   MoveBlockedResult{Target: impulse.TargetID, Reason: "unresolvable destination"},
			Source:    &skyra.ImpulseSource{SpaceID: w.id},
			Return:    true,
		})
		return
	}
	start, placed := w.PositionOf(origin.ID())
	if !placed {
		return
	}
	if start == dest {
		go w.kernel.Realize(origin, origin, skyra.Impulse{
			Primitive: PrimitiveMove,
			Payload:   MovementResult{Cell: dest.ID()},
			Source:    &skyra.ImpulseSource{SpaceID: w.id},
			Return:    true,
		})
		return
	}
	path := w.globe.Path(start, dest)
	debug.Being("globe", "move", fmt.Sprintf("origin=%s %s → %s (%d steps)", origin.ID(), start.ID(), dest.ID(), len(path)-1))
	w.mu.Lock()
	w.walks[origin.ID()] = &walkState{origin: origin, path: path, at: 0, impulse: impulse}
	w.mu.Unlock()
	w.emit(Event{Kind: "walk", Agent: origin.ID(), Cell: dest.ID(), Extra: map[string]any{"steps": len(path) - 1}})
}

func (w *World) resolveDestination(target string) (Cell, bool) {
	t := strings.TrimPrefix(strings.TrimSpace(target), "coord:")
	if c, ok := ParseCell(t); ok && w.globe.Valid(c) {
		return c, true
	}
	switch t {
	case "shrine:/", "shrine":
		return w.shrine, true
	case "beacon:/", "beacon":
		return w.beacon, true
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	if c, ok := w.position[t]; ok {
		return c, true
	}
	return Cell{}, false
}

// observeTask runs take/activate instantly (the agent acts where it stands) and fires
// the outcome back as a return wave.
func (w *World) observeTask(origin skyra.Space, impulse skyra.Impulse, prim string) {
	if origin == nil || w.kernel == nil {
		return
	}
	var ok bool
	var missing string
	switch prim {
	case PrimitiveTake:
		ok, missing = w.Take(origin.ID())
	case PrimitiveActivate:
		ok, missing = w.Activate(origin.ID())
	case PrimitiveCatch:
		ok, missing = w.Catch(origin.ID())
	}
	cell, _ := w.PositionOf(origin.ID())
	debug.Being("globe", prim, fmt.Sprintf("origin=%s ok=%v missing=%q", origin.ID(), ok, missing))
	w.emit(Event{Kind: prim, Agent: origin.ID(), Cell: cell.ID(), Extra: map[string]any{"ok": ok, "missing": missing}})
	go w.kernel.Realize(origin, origin, skyra.Impulse{
		Primitive: prim,
		Payload:   TaskResult{Primitive: prim, OK: ok, Missing: missing},
		Source:    &skyra.ImpulseSource{SpaceID: w.id},
		Return:    true,
	})
}

// observeEngage delivers an utterance to another agent, and reports delivery back.
// No range gate on the globe (for now): it is a small world and the demo wants talk.
func (w *World) observeEngage(origin skyra.Space, impulse skyra.Impulse) {
	if origin == nil || w.kernel == nil || impulse.TargetID == "" {
		return
	}
	w.mu.RLock()
	target := w.byID[impulse.TargetID]
	w.mu.RUnlock()
	delivered := target != nil && target.ID() != origin.ID()
	if delivered {
		w.kernel.Realize(origin, target, impulse)
	}
	w.emit(Event{Kind: "engage", Agent: origin.ID(), Cell: impulse.TargetID, Extra: map[string]any{"delivered": delivered}})
	go w.kernel.Realize(origin, origin, skyra.Impulse{
		Primitive: PrimitiveEngage,
		Payload:   EngagementResult{TargetID: impulse.TargetID, Delivered: delivered},
		Source:    &skyra.ImpulseSource{SpaceID: w.id},
		Return:    true,
	})
}

// PerceiveCell is the perceive wave's boundary: its children are every cell within
// graph-distance radius, enumerated flat (one hop — the branching-amplitude division
// makes deep BFS through the kernel decay too fast, same as the flat world).
type PerceiveCell struct {
	id      string
	world   *World
	owner   skyra.Space
	start   Cell
	radius  int
	impulse skyra.Impulse
	dist    map[Cell]int
}

func (p *PerceiveCell) ID() string                            { return p.id }
func (p *PerceiveCell) Slice() skyra.Slice                    { return p.world.slice }
func (p *PerceiveCell) NewActivation() skyra.ActivationRecord { return newSpaceActivation() }
func (p *PerceiveCell) Content() any                          { return nil }
func (p *PerceiveCell) Agents() map[string]skyra.Space        { return nil }

func (p *PerceiveCell) Observe(origin skyra.Space, impulse any) {
	p.world.slice.Maps()[MapPerception] = map[string]any{}
	p.dist = p.world.globe.CellsWithin(p.start, p.radius)
}

func (p *PerceiveCell) Children() map[string]skyra.Space {
	out := make(map[string]skyra.Space, len(p.dist))
	for c, d := range p.dist {
		if c == p.start {
			continue
		}
		pt := p.world.newPoint(c, d)
		out[pt.ID()] = pt
	}
	return out
}

func (p *PerceiveCell) Express(origin skyra.Space, impulse any) any {
	if p.world.kernel == nil || origin == nil {
		return nil
	}
	perception, _ := p.world.slice.Maps()[MapPerception].(map[string]any)
	result := PerceptionResult{
		Self:      p.start.ID(),
		Agents:    map[string]string{},
		Shrine:    p.world.shrine.ID(),
		Beacon:    p.world.beacon.ID(),
		BeaconLit: p.world.Lit(),
		Distances: map[string]int{p.start.ID(): 0},
	}
	for _, r := range perception {
		pr, ok := r.(PointResult)
		if !ok {
			continue
		}
		result.Distances[pr.CellID] = pr.Dist
		for id := range pr.Agents {
			result.Agents[id] = pr.CellID
		}
	}
	// agents standing with the perceiver share its cell
	for id := range p.world.occupantsAt(p.start) {
		if id != origin.ID() {
			result.Agents[id] = p.start.ID()
		}
	}
	debug.Being("globe", "perceive", fmt.Sprintf("express: self=%s cells=%d agents=%d", result.Self, len(result.Distances), len(result.Agents)))
	p.world.emit(Event{Kind: "perceive", Agent: origin.ID(), Cell: p.start.ID(), Extra: map[string]any{"distances": distancesForEvent(result.Distances)}})
	p.world.kernel.Realize(origin, origin, skyra.Impulse{
		Primitive: PrimitivePerceive,
		Payload:   result,
		Source:    &skyra.ImpulseSource{SpaceID: p.world.id},
		Return:    true,
	})
	return nil
}

func distancesForEvent(in map[string]int) map[string]int {
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// Point is one globe cell inside a wave — it reports what stands on it and descends no
// further (the boundary enumerated the whole radius already).
type Point struct {
	id    string
	cell  Cell
	dist  int
	world *World
}

func (w *World) newPoint(c Cell, dist int) *Point {
	return &Point{id: c.ID(), cell: c, dist: dist, world: w}
}

func (p *Point) ID() string                              { return p.id }
func (p *Point) Slice() skyra.Slice                      { return p.world.slice }
func (p *Point) NewActivation() skyra.ActivationRecord   { return newSpaceActivation() }
func (p *Point) Observe(origin skyra.Space, impulse any) {}
func (p *Point) Children() map[string]skyra.Space        { return map[string]skyra.Space{} }
func (p *Point) Content() any                            { return p.world.occupantsAt(p.cell) }

func (p *Point) Agents() map[string]skyra.Space {
	return p.world.occupantsAt(p.cell)
}

func (p *Point) Express(origin skyra.Space, impulse any) any {
	return PointResult{
		CellID: p.id,
		Dist:   p.dist,
		Agents: p.Agents(),
	}
}

// PointResult is what one cell reports into the wave.
type PointResult struct {
	CellID string
	Dist   int
	Agents map[string]skyra.Space
}

func (w *World) occupantsAt(c Cell) map[string]skyra.Space {
	w.mu.RLock()
	defer w.mu.RUnlock()
	cell := w.placement[c]
	if len(cell) == 0 {
		return nil
	}
	out := make(map[string]skyra.Space, len(cell))
	for id, occ := range cell {
		out[id] = occ
	}
	return out
}

func asImpulse(impulse any) (skyra.Impulse, bool) {
	switch imp := impulse.(type) {
	case skyra.Impulse:
		return imp, true
	case *skyra.Impulse:
		if imp != nil {
			return *imp, true
		}
	}
	return skyra.Impulse{}, false
}

type spaceActivation struct {
	base float64
}

func newSpaceActivation() skyra.ActivationRecord {
	return &spaceActivation{base: spaceBaseline}
}

func (a *spaceActivation) Activate(visited bool, traversalCount int) float64 {
	return a.base
}
