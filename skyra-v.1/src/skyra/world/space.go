package world

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

const (
	DefaultSize           = 1000
	DefaultPerceiveRadius = 20
	DefaultEngageRadius   = 20
	DefaultMoveRadius     = 20
	spaceBaseline         = 0.10
	spaceThreshold        = 0.001
	PrimitivePerceive     = "perceive"
	PrimitiveEngage       = "engage"
	PrimitiveMove         = "move"
	PrimitiveBump         = "bump"
	PrimitiveWait         = "wait"
	MapPerception         = "perception"
	MapEngagement         = "engagement"
	MapMovement           = "movement"
	worldSliceID          = "world-slice"
	worldSliceOwner       = "world"
	defaultWorldID        = "world"
)

type coord struct {
	x int
	y int
}

type World struct {
	id     string
	kernel skyra.Time
	slice  skyra.Slice
	size   int
	mu     sync.RWMutex
	// placement is occupancy: the set of agents standing on each cell. Agents can
	// overlap, so a cell holds a set — never a single slot a later arrival could
	// silently overwrite (an overwritten agent becomes a ghost: present in
	// position, invisible to delivery and perception).
	placement map[coord]map[string]skyra.Space
	position  map[string]coord
	// byID is the identity registry: stable agent id → the agent. Identity is
	// never derived from occupancy (a cell lookup answers "who is HERE", not "who
	// IS this") — resolving ids through placement is what unpersoned agents when
	// overlap landed. Written at Place time.
	byID map[string]skyra.Space
	walls     map[coord]bool
	paths     map[coord]bool
}

func New(kernel skyra.Time) *World {
	return NewSized(kernel, DefaultSize)
}

func NewSized(kernel skyra.Time, size int) *World {
	if size <= 0 {
		size = DefaultSize
	}
	worldSlice := NewSliceWithThreshold(worldSliceID, worldSliceOwner, spaceThreshold)
	worldSlice.Maps()[MapPerception] = map[string]any{}
	worldSlice.Maps()[MapEngagement] = map[string]any{}
	return &World{
		id:        defaultWorldID,
		kernel:    kernel,
		slice:     worldSlice,
		size:      size,
		placement: map[coord]map[string]skyra.Space{},
		position:  map[string]coord{},
		byID:      map[string]skyra.Space{},
		walls:     map[coord]bool{},
		paths:     map[coord]bool{},
	}
}

func (w *World) ID() string                            { return w.id }
func (w *World) Slice() skyra.Slice                    { return w.slice }
func (w *World) NewActivation() skyra.ActivationRecord { return newSpaceActivation() }

// Observe is the world's primitive boundary. It reads the impulse envelope and
// dispatches on the requested world primitive. The origin value has no special
// meaning here, and Observe does not root another wave.
func (w *World) Observe(origin skyra.Space, impulse any) {
	imp, ok := asImpulse(impulse)
	if !ok {
		return
	}
	switch imp.Primitive {
	case PrimitivePerceive:
		w.observePerceive(origin, imp)
	case PrimitiveEngage:
		w.observeEngage(origin, imp)
	case PrimitiveMove:
		w.observeMove(origin, imp)
	case PrimitiveWait:
		w.observeWait(origin, imp)
	}
}

func (w *World) observePerceive(origin skyra.Space, impulse skyra.Impulse) {
	if origin == nil || w.kernel == nil {
		return
	}
	start := w.pointOf(origin)
	if start == nil {
		return
	}
	// ResultMap routes Point.Express results into the "perception" slot so
	// PerceiveCell.Express can collect them. Without this they'd land under
	// "perceive" (the primitive name fallback) and be invisible to the collector.
	impulse.ResultMap = MapPerception
	impulse.Amplitude = DefaultPerceiveRadius
	debug.Being("world", "perceive", fmt.Sprintf("origin=%s pos=%d,%d radius=%d", origin.ID(), start.x, start.y, DefaultPerceiveRadius))
	boundary := &PerceiveCell{
		id:      "world:perceive:" + origin.ID(),
		world:   w,
		owner:   origin,
		start:   start,
		radius:  DefaultPerceiveRadius,
		impulse: impulse,
	}
	w.kernel.Realize(origin, boundary, impulse)
}

func (w *World) observeEngage(origin skyra.Space, impulse skyra.Impulse) {
	if origin == nil || w.kernel == nil || impulse.TargetID == "" {
		return
	}
	target := w.agentByID(impulse.TargetID)
	targetPos := w.pointOf(target)
	if target == nil || targetPos == nil {
		return
	}
	originPos := w.pointOf(origin)
	if originPos != nil && chebyshevDist(originPos.x, originPos.y, targetPos.x, targetPos.y) > DefaultEngageRadius {
		debug.Being("world", "engage", fmt.Sprintf("blocked: origin=%s target=%s out of range", origin.ID(), target.ID()))
		go w.kernel.Realize(origin, origin, skyra.Impulse{
			Primitive: PrimitiveEngage,
			Payload:   EngageBlockedResult{TargetID: target.ID(), Reason: "out of range"},
			Source:    &skyra.ImpulseSource{SpaceID: w.id},
			Return:    true,
		})
		return
	}
	debug.Being("world", "engage", fmt.Sprintf("origin=%s target=%s payload=%v", origin.ID(), impulse.TargetID, impulse.Payload))
	impulse.ResultMap = MapEngagement
	boundary := &EngageCell{
		id:      "world:engage:" + origin.ID() + ":" + target.ID(),
		world:   w,
		owner:   origin,
		target:  target,
		start:   targetPos,
		impulse: impulse,
	}
	w.kernel.Realize(origin, boundary, impulse)
}

// observeMove resolves the target to a coord and places the origin there.
// Target resolution order:
//  1. Parse as "x,y" → absolute coord
//  2. Look up as agent ID → move to that agent's current position
//  3. (future) Named location index
func (w *World) observeMove(origin skyra.Space, impulse skyra.Impulse) {
	if origin == nil || w.kernel == nil {
		return
	}
	target := impulse.TargetID
	x, y, ok := w.resolveDestination(target)
	if !ok {
		debug.Being("world", "move", fmt.Sprintf("origin=%s target=%q unresolvable", origin.ID(), target))
		return
	}
	w.mu.RLock()
	isWall := w.walls[w.wrapCoord(x, y)]
	w.mu.RUnlock()
	if isWall {
		debug.Being("world", "move", fmt.Sprintf("blocked: origin=%s wall at %d,%d", origin.ID(), x, y))
		go w.kernel.Realize(origin, origin, skyra.Impulse{
			Primitive: PrimitiveMove,
			Payload:   MoveBlockedResult{DestX: x, DestY: y, Reason: "wall"},
			Source:    &skyra.ImpulseSource{SpaceID: w.id},
			Return:    true,
		})
		return
	}
	debug.Being("world", "move", fmt.Sprintf("origin=%s target=%q → %d,%d", origin.ID(), target, x, y))
	boundary := &MoveCell{
		id:      "world:move:" + origin.ID(),
		world:   w,
		owner:   origin,
		destX:   x,
		destY:   y,
		impulse: impulse,
	}
	w.kernel.Realize(origin, boundary, impulse)
}

// resolveDestination maps a target string to a world coord.
// Tries coord parse first, then agent lookup.
func (w *World) resolveDestination(target string) (x, y int, ok bool) {
	// 1. "x,y" literal coord
	if parts := strings.SplitN(target, ",", 2); len(parts) == 2 {
		px, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		py, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 == nil && err2 == nil {
			c := w.wrapCoord(px, py)
			return c.x, c.y, true
		}
	}
	// 2. Agent ID → use their current position
	if agent := w.agentByID(target); agent != nil {
		if p := w.pointOf(agent); p != nil {
			return p.x, p.y, true
		}
	}
	return 0, 0, false
}

// MoveCell is the move primitive's boundary. It places the owner at the
// resolved destination and fires a return wave with the new position.
// No kernel traversal needed — destination is already resolved.
type MoveCell struct {
	id      string
	world   *World
	owner   skyra.Space
	destX   int
	destY   int
	impulse skyra.Impulse
}

func (m *MoveCell) ID() string                            { return m.id }
func (m *MoveCell) Slice() skyra.Slice                    { return m.world.slice }
func (m *MoveCell) NewActivation() skyra.ActivationRecord { return newSpaceActivation() }
func (m *MoveCell) Content() any                          { return nil }
func (m *MoveCell) Agents() map[string]skyra.Space        { return nil }
func (m *MoveCell) Children() map[string]skyra.Space      { return nil }
func (m *MoveCell) Observe(origin skyra.Space, impulse any) {}

func (m *MoveCell) Express(origin skyra.Space, impulse any) any {
	if m.world.kernel == nil || origin == nil {
		return nil
	}
	// Agents can overlap — Place always succeeds (wall check already happened in
	// observeMove). Everyone already on the cell gets a bump, both directions; the
	// move still completes.
	occupants, _ := m.world.Place(m.owner, m.destX, m.destY)
	for _, occupant := range occupants {
		debug.Being("world", "bump", fmt.Sprintf("%s bumped into %s at %d,%d", m.owner.ID(), occupant.ID(), m.destX, m.destY))
		go m.world.kernel.Realize(origin, origin, skyra.Impulse{
			Primitive: PrimitiveBump,
			Payload:   BumpResult{OtherID: occupant.ID()},
			Source:    &skyra.ImpulseSource{SpaceID: m.world.ID()},
			Return:    true,
		})
		go m.world.kernel.Realize(origin, occupant, skyra.Impulse{
			Primitive: PrimitiveBump,
			Payload:   BumpResult{OtherID: m.owner.ID()},
			Source:    &skyra.ImpulseSource{SpaceID: m.owner.ID()},
		})
	}
	m.world.kernel.Realize(origin, origin, skyra.Impulse{
		Primitive: PrimitiveMove,
		Payload:   MovementResult{X: m.destX, Y: m.destY},
		Source: &skyra.ImpulseSource{
			SpaceID: m.world.ID(),
			Coord:   skyra.Coord{X: m.destX, Y: m.destY},
		},
		Return: true,
	})
	return nil
}

// MovementResult is the return payload after a successful move.
type MovementResult struct {
	X int
	Y int
}

// observeWait arms a timer. After the requested duration the world fires a
// return wave into origin so the cognitive loop picks up where it left off.
func (w *World) observeWait(origin skyra.Space, impulse skyra.Impulse) {
	if origin == nil || w.kernel == nil {
		return
	}
	secs, err := strconv.ParseFloat(strings.TrimSpace(impulse.TargetID), 64)
	if err != nil || secs <= 0 {
		return
	}
	reason, _ := impulse.Payload.(string)
	debug.Being("world", "wait", fmt.Sprintf("origin=%s secs=%.1f reason=%q", origin.ID(), secs, reason))
	go func() {
		time.Sleep(time.Duration(float64(time.Second) * secs))
		w.kernel.Realize(origin, origin, skyra.Impulse{
			Primitive: PrimitiveWait,
			Payload:   WaitResult{Reason: reason},
			Source:    &skyra.ImpulseSource{SpaceID: w.id},
			Return:    true,
		})
	}()
}

func (w *World) Express(origin skyra.Space, impulse any) any { return nil }

func (w *World) Agents() map[string]skyra.Space {
	w.mu.RLock()
	defer w.mu.RUnlock()

	agents := make(map[string]skyra.Space, len(w.byID))
	for id, space := range w.byID {
		agents[id] = space
	}
	return agents
}

func (w *World) Children() map[string]skyra.Space {
	w.mu.RLock()
	defer w.mu.RUnlock()

	children := make(map[string]skyra.Space, len(w.placement))
	for c := range w.placement {
		p := w.newPoint(c.x, c.y)
		children[p.ID()] = p
	}
	return children
}

func (w *World) Content() any { return nil }

// Place moves space to (x,y). Agents can overlap — the move always succeeds.
// Any agents already on the cell are returned so the caller can fire bump
// notifications; the mover joins them (occupancy is a set — nobody is
// overwritten, nobody ghosts).
func (w *World) Place(space skyra.Space, x, y int) (occupants []skyra.Space, moved bool) {
	if space == nil {
		return nil, false
	}
	c := w.wrapCoord(x, y)
	w.mu.Lock()
	defer w.mu.Unlock()

	for id, occ := range w.placement[c] {
		if id != space.ID() {
			occupants = append(occupants, occ)
		}
	}
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
	return occupants, true
}

// BumpResult is the payload on a bump impulse — who the agent collided with.
type BumpResult struct {
	OtherID string
}

// EngageBlockedResult is the return payload when an engage is blocked by range.
type EngageBlockedResult struct {
	TargetID string
	Reason   string
}

// MoveBlockedResult is the return payload when a move is blocked by range.
type MoveBlockedResult struct {
	DestX   int
	DestY   int
	Reason  string
}

// WaitResult is the return payload fired into origin after a wait timer expires.
type WaitResult struct {
	Reason string
}

// pointsInRadius returns all grid points within Chebyshev distance radius from
// (cx, cy), excluding the center. Used by PerceiveCell to enumerate its children
// as a flat set rather than relying on recursive BFS.
func (w *World) pointsInRadius(cx, cy, radius int) map[string]skyra.Space {
	out := make(map[string]skyra.Space, (2*radius+1)*(2*radius+1)-1)
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			c := w.wrapCoord(cx+dx, cy+dy)
			p := w.newPoint(c.x, c.y)
			out[p.ID()] = p
		}
	}
	return out
}

// chebyshevDist is the grid distance metric (max of |dx|, |dy|). Radius 1 = 8
// adjacent cells; radius N = all cells within an N-step king's move.
func chebyshevDist(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

func (w *World) pointOf(space skyra.Space) *Point {
	if space == nil {
		return nil
	}
	w.mu.RLock()
	c, ok := w.position[space.ID()]
	w.mu.RUnlock()
	if !ok {
		return nil
	}
	return w.newPoint(c.x, c.y)
}

// agentByID resolves an agent by its stable id, via the identity registry —
// never through occupancy (placement answers "who is at this cell", which under
// overlap is not the same question).
func (w *World) agentByID(id string) skyra.Space {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.byID[id]
}

func (w *World) PointOf(space skyra.Space) skyra.Space {
	return w.pointOf(space)
}

// CoordOf returns the world coordinate of space, and whether it is placed.
func (w *World) CoordOf(space skyra.Space) (skyra.Coord, bool) {
	p := w.pointOf(space)
	if p == nil {
		return skyra.Coord{}, false
	}
	return skyra.Coord{X: p.x, Y: p.y}, true
}

// IsWall reports whether the given world coordinate is a wall.
func (w *World) IsWall(x, y int) bool {
	c := w.wrapCoord(x, y)
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.walls[c]
}

// IsPath reports whether the given world coordinate is a path.
func (w *World) IsPath(x, y int) bool {
	c := w.wrapCoord(x, y)
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.paths[c]
}

func (w *World) neighborsOf(x, y int) map[string]skyra.Space {
	neighbors := make(map[string]skyra.Space, 8)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			c := w.wrapCoord(x+dx, y+dy)
			p := w.newPoint(c.x, c.y)
			neighbors[p.ID()] = p
		}
	}
	return neighbors
}

// occupantsAt returns a snapshot of every agent standing on (x,y).
func (w *World) occupantsAt(x, y int) map[string]skyra.Space {
	c := w.wrapCoord(x, y)
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

func (w *World) newPoint(x, y int) *Point {
	c := w.wrapCoord(x, y)
	return &Point{
		id:    fmt.Sprintf("%d,%d", c.x, c.y),
		x:     c.x,
		y:     c.y,
		world: w,
	}
}

func (w *World) wrapCoord(x, y int) coord {
	return coord{
		x: ((x % w.size) + w.size) % w.size,
		y: ((y % w.size) + w.size) % w.size,
	}
}

type PerceiveCell struct {
	id      string
	world   *World
	owner   skyra.Space
	start   *Point
	radius  int
	impulse skyra.Impulse
}

func (p *PerceiveCell) ID() string                            { return p.id }
func (p *PerceiveCell) Slice() skyra.Slice                    { return p.world.slice }
func (p *PerceiveCell) NewActivation() skyra.ActivationRecord { return newSpaceActivation() }
func (p *PerceiveCell) Content() any                          { return nil }
func (p *PerceiveCell) Agents() map[string]skyra.Space        { return nil }

func (p *PerceiveCell) Observe(origin skyra.Space, impulse any) {
	p.world.slice.Maps()[MapPerception] = map[string]any{}
}

// Children returns every Point within radius as a flat map — one hop, no
// recursive BFS. The kernel's branching-amplitude division kills BFS depth
// past ~5 hops with 8 children per node; flat enumeration gives exact radius
// control. Each Point reports what's at its coord and returns no further children.
func (p *PerceiveCell) Children() map[string]skyra.Space {
	if p.start == nil {
		return nil
	}
	return p.world.pointsInRadius(p.start.x, p.start.y, p.radius)
}

func (p *PerceiveCell) Express(origin skyra.Space, impulse any) any {
	if p.world.kernel == nil || origin == nil {
		return nil
	}
	perception, _ := p.world.slice.Maps()[MapPerception].(map[string]any)
	discovered := collectPerceiveResults(perception)
	discovered.SelfCoord = p.start.source().Coord // center of the wave = origin's position

	// Scan walls and paths within radius so sensory can describe the environment.
	p.world.mu.RLock()
	for dy := -p.radius; dy <= p.radius; dy++ {
		for dx := -p.radius; dx <= p.radius; dx++ {
			c := p.world.wrapCoord(p.start.x+dx, p.start.y+dy)
			if p.world.walls[c] {
				discovered.Walls = append(discovered.Walls, skyra.Coord{X: c.x, Y: c.y})
			}
			if p.world.paths[c] {
				discovered.Paths = append(discovered.Paths, skyra.Coord{X: c.x, Y: c.y})
			}
		}
	}
	p.world.mu.RUnlock()

	agentIDs := make([]string, 0, len(discovered.Agents))
	for id := range discovered.Agents {
		agentIDs = append(agentIDs, id)
	}
	debug.Being("world", "perceive", fmt.Sprintf("express: self=%d,%d points=%d agents=%v walls=%d paths=%d",
		discovered.SelfCoord.X, discovered.SelfCoord.Y, len(discovered.Points), agentIDs, len(discovered.Walls), len(discovered.Paths)))
	// notify each discovered agent: "you were perceived by origin" — no utterance.
	// skip the origin itself (self-perception is a no-op).
	for id, agent := range discovered.Agents {
		if origin != nil && id == origin.ID() {
			continue
		}
		debug.Being("world", "perceive", fmt.Sprintf("notify agent=%s perceived-by=%s", id, origin.ID()))
		p.world.kernel.Realize(origin, agent, skyra.Impulse{
			Primitive: PrimitivePerceive,
			Source:    p.impulse.Source,
		})
	}
	// return the perception result to origin — carries self-position + nearby map
	debug.Being("world", "perceive", fmt.Sprintf("return to origin=%s", origin.ID()))
	p.world.kernel.Realize(origin, origin, skyra.Impulse{
		Primitive: PrimitivePerceive,
		Payload:   discovered,
		Source: &skyra.ImpulseSource{
			SpaceID: p.world.ID(),
		},
		Return: true,
	})
	return nil
}

// PerceptionResult is the return payload from a perceive wave: what was found.
// SelfCoord is the origin's position at the time of the wave — the center of
// the perception radius. Sensory reads this to update its self-location.
type PerceptionResult struct {
	SelfCoord skyra.Coord
	Agents    map[string]skyra.Space // agent ID → space, for each agent found in range
	Points    []PointResult          // all points reached (each carries Source.Coord)
	Walls     []skyra.Coord          // wall coords within perceive radius
	Paths     []skyra.Coord          // path coords within perceive radius
}

func collectPerceiveResults(results map[string]any) PerceptionResult {
	out := PerceptionResult{Agents: map[string]skyra.Space{}}
	for _, result := range results {
		point, ok := result.(PointResult)
		if !ok {
			continue
		}
		out.Points = append(out.Points, point)
		for id, agent := range point.Agents {
			if agent != nil {
				out.Agents[id] = agent
			}
		}
	}
	return out
}

type EngageCell struct {
	id      string
	world   *World
	owner   skyra.Space
	target  skyra.Space
	start   *Point
	impulse skyra.Impulse
}

func (e *EngageCell) ID() string                            { return e.id }
func (e *EngageCell) Slice() skyra.Slice                    { return e.world.slice }
func (e *EngageCell) NewActivation() skyra.ActivationRecord { return newSpaceActivation() }
func (e *EngageCell) Content() any                          { return nil }
func (e *EngageCell) Agents() map[string]skyra.Space        { return nil }

func (e *EngageCell) Observe(origin skyra.Space, impulse any) {
	e.world.slice.Maps()[MapEngagement] = map[string]any{}
}

func (e *EngageCell) Children() map[string]skyra.Space {
	if e.start == nil {
		return nil
	}
	return map[string]skyra.Space{e.start.ID(): e.start}
}

func (e *EngageCell) Express(origin skyra.Space, impulse any) any {
	if e.world.kernel == nil || origin == nil {
		return nil
	}
	engagement, _ := e.world.slice.Maps()[MapEngagement].(map[string]any)
	delivered := collectDeliveredAgents(engagement)
	for _, agent := range delivered {
		e.world.kernel.Realize(origin, agent, e.impulse)
	}
	e.world.kernel.Realize(origin, origin, skyra.Impulse{
		Primitive: PrimitiveEngage,
		Payload: EngagementResult{
			Delivered: delivered,
		},
		Source: &skyra.ImpulseSource{
			SpaceID: e.world.ID(),
		},
		Return: true,
	})
	return nil
}

type Point struct {
	id        string
	x, y      int
	world     *World
	neighbors map[string]skyra.Space
}

func (p *Point) ID() string                            { return p.id }
func (p *Point) Slice() skyra.Slice                    { return p.world.slice }
func (p *Point) NewActivation() skyra.ActivationRecord { return newSpaceActivation() }

func (p *Point) Observe(origin skyra.Space, impulse any) {
	imp, ok := impulse.(skyra.Impulse)
	if ok && imp.Primitive == PrimitivePerceive {
		// perceive: PerceiveCell already enumerated all points in radius as direct
		// children — no further expansion from Points needed.
		p.neighbors = nil
		return
	}
	p.neighbors = p.world.neighborsOf(p.x, p.y)
}

func (p *Point) Express(origin skyra.Space, impulse any) any {
	imp, ok := impulse.(skyra.Impulse)
	if !ok {
		if ptr, ok := impulse.(*skyra.Impulse); ok && ptr != nil {
			imp = *ptr
		} else {
			return PointResult{
				PointID: p.ID(),
				Source:  p.source(),
				OK:      false,
				Error:   "world point expected skyra.Impulse",
			}
		}
	}

	switch imp.Primitive {
	case PrimitivePerceive:
		return p.perceive(origin, imp)
	case PrimitiveEngage:
		return p.engage(origin, imp)
	case PrimitiveMove:
		return p.move(origin, imp)
	default:
		return PointResult{
			PointID:   p.ID(),
			Primitive: imp.Primitive,
			Source:    p.source(),
			OK:        false,
			Error:     "unsupported world primitive",
		}
	}
}

func (p *Point) Children() map[string]skyra.Space {
	if p.neighbors == nil {
		return map[string]skyra.Space{}
	}
	return p.neighbors
}

// Agents returns every agent standing on this point — all occupants, since
// agents can overlap. (Deriving this from a single occupant made an overlapped
// agent invisible to perception and delivery.)
func (p *Point) Agents() map[string]skyra.Space {
	return p.world.occupantsAt(p.x, p.y)
}

func (p *Point) Content() any {
	return p.world.occupantsAt(p.x, p.y)
}

func (p *Point) source() *skyra.ImpulseSource {
	return &skyra.ImpulseSource{
		SpaceID: p.world.ID(),
		PointID: p.ID(),
		Coord: skyra.Coord{
			X: p.x,
			Y: p.y,
		},
	}
}

func (p *Point) perceive(origin skyra.Space, impulse skyra.Impulse) PointResult {
	return PointResult{
		PointID:   p.ID(),
		Primitive: impulse.Primitive,
		Source:    p.source(),
		OK:        true,
		Content:   p.Content(),
		Agents:    p.Agents(),
		Points:    p.Children(),
	}
}

func (p *Point) engage(origin skyra.Space, impulse skyra.Impulse) PointResult {
	return PointResult{
		PointID:   p.ID(),
		Primitive: impulse.Primitive,
		Source:    p.source(),
		OK:        true,
		Content:   p.Content(),
		Agents:    p.Agents(),
	}
}

func (p *Point) move(origin skyra.Space, impulse skyra.Impulse) PointResult {
	return PointResult{
		PointID:   p.ID(),
		Primitive: impulse.Primitive,
		Source:    p.source(),
		OK:        true,
	}
}

type EngagementResult struct {
	Delivered map[string]skyra.Space
}

type PointResult struct {
	PointID   string
	Primitive string
	Source    *skyra.ImpulseSource
	OK        bool
	Content   any
	Agents    map[string]skyra.Space
	Points    map[string]skyra.Space
	Error     string
}

func withSource(impulse any, source *skyra.ImpulseSource) any {
	switch imp := impulse.(type) {
	case skyra.Impulse:
		if imp.Source == nil {
			imp.Source = source
		}
		return imp
	case *skyra.Impulse:
		if imp == nil || imp.Source != nil {
			return impulse
		}
		cp := *imp
		cp.Source = source
		return &cp
	default:
		return impulse
	}
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

func collectDeliveredAgents(results map[string]any) map[string]skyra.Space {
	delivered := map[string]skyra.Space{}
	for _, result := range results {
		point, ok := result.(PointResult)
		if !ok {
			continue
		}
		for id, agent := range point.Agents {
			if agent != nil {
				delivered[id] = agent
			}
		}
	}
	return delivered
}

func snapshotMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
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
