package world

import (
	"fmt"
	"sync"

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

// Observe is the world's entry point — how the program drives a wave into the
// world. A being acts outward by calling its environment's Observe with itself
// as origin (the wave owner); Observe delegates the wave-rooting to Emit. This
// is the live pattern: the run goes through Observe; Emit is the helper.
func (w *World) Observe(origin skyra.Space, impulse any) {
	w.Emit(origin, impulse)
}

// Emit roots a wave into the world from the owner's location (the helper Observe
// delegates to). The seed of the world.observe facilitator.
func (w *World) Emit(owner skyra.Space, impulse any) {
	if owner == nil || w.kernel == nil {
		return
	}
	point := w.pointOf(owner)
	if point == nil {
		return
	}
	impulse = withSource(impulse, point.source())
	w.kernel.Realize(owner, point, impulse)
}

func (w *World) Express(origin skyra.Space, impulse any) any { return nil }

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
	}
}

type PointResult struct {
	PointID   string
	Primitive string
	Source    *skyra.ImpulseSource
	OK        bool
	Content   any
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

type spaceActivation struct {
	base float64
}

func newSpaceActivation() skyra.ActivationRecord {
	return &spaceActivation{base: spaceBaseline}
}

func (a *spaceActivation) Activate(visited bool, traversalCount int) float64 {
	return a.base
}
