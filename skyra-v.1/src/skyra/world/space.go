package world

import (
	"fmt"
	"sync"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

const (
	DefaultSize       = 1000
	spaceBaseline     = 0.10
	spaceThreshold    = 0.001
	PrimitivePerceive = "perceive"
	PrimitiveEngage   = "engage"
	MapPerception     = "perception"
	MapEngagement     = "engagement"
	worldSliceID      = "world-slice"
	worldSliceOwner   = "world"
	defaultWorldID    = "world"
)

type coord struct {
	x int
	y int
}

type World struct {
	id        string
	kernel    skyra.Time
	slice     skyra.Slice
	size      int
	mu        sync.RWMutex
	placement map[coord]skyra.Space
	position  map[string]coord
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
		placement: map[coord]skyra.Space{},
		position:  map[string]coord{},
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
	debug.Being("world", "perceive", fmt.Sprintf("origin=%s payload=%v", origin.ID(), impulse.Payload))
	boundary := &PerceiveCell{
		id:      "world:perceive:" + origin.ID(),
		world:   w,
		owner:   origin,
		start:   start,
		impulse: impulse,
	}
	w.kernel.Realize(origin, boundary, impulse)
}

func (w *World) observeEngage(origin skyra.Space, impulse skyra.Impulse) {
	if origin == nil || w.kernel == nil || impulse.TargetID == "" {
		return
	}
	debug.Being("world", "engage", fmt.Sprintf("origin=%s target=%s payload=%v", origin.ID(), impulse.TargetID, impulse.Payload))
	target := w.agentByID(impulse.TargetID)
	start := w.pointOf(target)
	if target == nil || start == nil {
		return
	}
	impulse.ResultMap = MapEngagement
	boundary := &EngageCell{
		id:      "world:engage:" + origin.ID() + ":" + target.ID(),
		world:   w,
		owner:   origin,
		target:  target,
		start:   start,
		impulse: impulse,
	}
	w.kernel.Realize(origin, boundary, impulse)
}

func (w *World) Express(origin skyra.Space, impulse any) any { return nil }

func (w *World) Agents() map[string]skyra.Space {
	w.mu.RLock()
	defer w.mu.RUnlock()

	agents := make(map[string]skyra.Space, len(w.placement))
	for _, space := range w.placement {
		agents[space.ID()] = space
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

func (w *World) Place(space skyra.Space, x, y int) {
	if space == nil {
		return
	}
	c := w.wrapCoord(x, y)
	w.mu.Lock()
	defer w.mu.Unlock()

	if old, ok := w.position[space.ID()]; ok {
		delete(w.placement, old)
	}
	if occupant, ok := w.placement[c]; ok {
		delete(w.position, occupant.ID())
	}
	w.placement[c] = space
	w.position[space.ID()] = c
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

func (w *World) agentByID(id string) skyra.Space {
	w.mu.RLock()
	defer w.mu.RUnlock()
	c, ok := w.position[id]
	if !ok {
		return nil
	}
	return w.placement[c]
}

func (w *World) PointOf(space skyra.Space) skyra.Space {
	return w.pointOf(space)
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

func (w *World) occupantAt(x, y int) skyra.Space {
	c := w.wrapCoord(x, y)
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.placement[c]
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

func (p *PerceiveCell) Children() map[string]skyra.Space {
	if p.start == nil {
		return nil
	}
	return map[string]skyra.Space{p.start.ID(): p.start}
}

func (p *PerceiveCell) Express(origin skyra.Space, impulse any) any {
	if p.world.kernel == nil || origin == nil {
		return nil
	}
	// Preserve the original Source and Payload (the utterance + speaker) on the
	// delivery wave. Perception data is already in world.slice.Maps() — it doesn't
	// need to ride the impulse payload and displace the utterance. Not marked
	// Return: the perceive delivery IS the primary wave (not a loop-back).
	p.world.kernel.Realize(origin, origin, skyra.Impulse{
		Primitive: PrimitivePerceive,
		Payload:   p.impulse.Payload,
		Source:    p.impulse.Source,
	})
	return nil
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
	// Traversal intake: gather neighbors. origin is the wave owner (unused here).
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

func (p *Point) Agents() map[string]skyra.Space {
	occupant := p.Content()
	space, ok := occupant.(skyra.Space)
	if !ok || space == nil {
		return nil
	}
	return map[string]skyra.Space{space.ID(): space}
}

func (p *Point) Content() any {
	return p.world.occupantAt(p.x, p.y)
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
