package reality

import (
	"fmt"
	"strings"
	"sync"

	"skyra-v1/src/inference"
)

// Slice — a being's private cognitive state on the Relation.
// Each being gets its own slice. Own lock. Own goroutine.
// No being can access another being's slice.
type Slice struct {
	mu       sync.Mutex
	Thoughts []Reality // accumulated thought cells
	Signal   float64
	Load     int    // accumulated content size
	Forming  *Skyra // the cell being built during this traversal
	Goal     string // the destination — persistent, changed explicitly
	Intent   string // the predicted future — refined at every compression boundary
	LastSeen Reality
	Depth    int
	Trace    []string
	Visited  map[string]bool
}

func (s *Slice) Lock()   { s.mu.Lock() }
func (s *Slice) Unlock() { s.mu.Unlock() }

type Relation struct {
	mu sync.RWMutex // protects shared state: World, Fields
	id        string
	Origin    string
	ThreadID  string
	Impulse   string
	World     map[string]Reality             // shared — beings, places, tools. The world is the Relation.
	Providers map[string]Reality             // shared — inference endpoints
	Fields    map[string]map[string]float64  // shared — binding fields
	Slices    map[string]*Slice              // private — keyed by being ID, tenant-isolated
	MaxLoad   int
	MaxDepth  int
	Cancelled bool
}

func (r *Relation) RLock()   { r.mu.RLock() }
func (r *Relation) RUnlock() { r.mu.RUnlock() }

func (r *Relation) SetImpulse(origin, impulse string) {
	r.mu.Lock()
	r.Origin = origin
	r.Impulse = impulse
	r.mu.Unlock()
}

func (r *Relation) ID() string {
	return r.id
}

func (r *Relation) Create(rel *Relation) Reality {
	return nil
}

// Realize — the world loop. Observe routes the impulse to the
// highest scoring entity. The entity traverses. Express delivers.
func (r *Relation) Realize(rel *Relation) *Relation {
	r.Observe(rel)
	return rel
}

func (r *Relation) Competence(rel *Relation) {
}

// Observe — the world observes the impulse. Scores every entity in the
// World map by activation. The highest scoring targets receive the traversal.
func (r *Relation) Observe(rel *Relation) {
	r.mu.RLock()
	world := r.World
	r.mu.RUnlock()

	var best Reality
	var bestScore float64

	for _, entity := range world {
		score := entity.Activation(rel)
		if score > bestScore {
			bestScore = score
			best = entity
		}
	}

	if best != nil {
		best.Realize(rel)
	}
}

func (r *Relation) Express(rel *Relation) *Relation {
	s := r.GetSlice(r.Origin)
	s.mu.Lock()
	spent := s.Signal < 0.1
	s.mu.Unlock()
	if spent {
		r.Deposit(r.Origin)
	}
	return rel
}

func (r *Relation) Activation(rel *Relation) float64 {
	s := r.GetSlice(r.Origin)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Signal
}

func (r *Relation) Deposit(beingID string) {
	s := r.GetSlice(beingID)
	s.mu.Lock()
	cell := s.Forming
	if cell == nil {
		s.mu.Unlock()
		return
	}
	var parts []string
	for _, t := range s.Thoughts {
		if sk, ok := t.(*Skyra); ok {
			parts = append(parts, sk.Content)
			cell.Relationships[sk.id] = sk
		}
	}
	cell.Content = strings.Join(parts, "\n")
	cell.Weight = s.Signal
	host := s.LastSeen
	s.mu.Unlock()

	vec, err := inference.Embed(cell.Content)
	if err == nil {
		cell.Vector = vec
	}

	if h, ok := host.(*Skyra); ok {
		h.Relationships[cell.id] = cell
	}
}

func (r *Relation) Reinforce() {
}

func (r *Relation) Decay() {
}

func Impress(origin, raw string) (*Relation, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("reality: empty input")
	}

	forming := &Skyra{
		id:            NewID(),
		Weight:        1.0,
		Alpha:         0.1,
		Relationships: make(map[string]Reality),
		Expressors:    make(map[string]Reality),
		ObserveFns:    make(map[string]func(*Relation)),
		ExpressFns:    make(map[string]func(*Relation) *Relation),
	}

	slice := &Slice{
		Thoughts: []Reality{},
		Signal:   1.0,
		Forming:  forming,
		Visited:  make(map[string]bool),
	}

	return &Relation{
		id:        NewID(),
		Origin:    origin,
		Impulse:   raw,
		World:     make(map[string]Reality),
		Providers: make(map[string]Reality),
		Fields:    make(map[string]map[string]float64),
		Slices:    map[string]*Slice{origin: slice},
		MaxLoad:   75000,
		MaxDepth:  50,
	}, nil
}

// GetSlice returns the being's private slice, creating one if it doesn't exist.
func (r *Relation) GetSlice(beingID string) *Slice {
	r.mu.RLock()
	s, ok := r.Slices[beingID]
	r.mu.RUnlock()
	if ok {
		return s
	}

	forming := &Skyra{
		id:            NewID(),
		Weight:        1.0,
		Alpha:         0.1,
		Relationships: make(map[string]Reality),
		Expressors:    make(map[string]Reality),
		ObserveFns:    make(map[string]func(*Relation)),
		ExpressFns:    make(map[string]func(*Relation) *Relation),
	}

	s = &Slice{
		Thoughts: []Reality{},
		Signal:   1.0,
		Forming:  forming,
		Visited:  make(map[string]bool),
	}

	r.mu.Lock()
	r.Slices[beingID] = s
	r.mu.Unlock()
	return s
}
