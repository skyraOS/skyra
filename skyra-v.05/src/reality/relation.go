package reality

import (
	"fmt"
	"strings"
	"sync"

	"skyra-v1/src/inference"
)

type Relation struct {
	mu        sync.Mutex // TODO: lock should be per-agent segment, not global. Thoughts/Load/Signal are private per traversal. Fields/Visited are shared and need coordination.
	id        string
	Origin    string
	ThreadID  string
	Impulse   string
	Providers map[string]Reality
	Fields    map[string]map[string]float64 // memoryID → {bindingReality: activation}
	Depth     int
	Trace     []string
	Visited   map[string]bool
	LastSeen  Reality
	Forming   *Skyra // the cell being built during this traversal — accumulates relationships as it descends
	Thoughts  []Reality // accumulated thought cells — each is a Reality with content, vector, and relationships
	Goal      string // the destination — persistent across traversals, changed explicitly by the being
	Intent    string // the route — how to reach the goal right now, refined at every compression boundary
	Signal    float64
	Load      int // accumulated content size in tokens/chars — checked on ascent
	MaxLoad   int
	MaxDepth  int
	Cancelled bool
}

func (r *Relation) ID() string {
	return r.id
}

func (r *Relation) Create(rel *Relation) Reality {
	return nil
}

func (r *Relation) Realize(rel *Relation) *Relation {
	return rel
}

func (r *Relation) Competence(rel *Relation) {
}

func (r *Relation) Observe(rel *Relation) {
	r.mu.Lock()
	r.Depth++
	r.Trace = append(r.Trace, rel.ID())
	r.mu.Unlock()
}

func (r *Relation) Express(rel *Relation) *Relation {
	r.mu.Lock()
	spent := r.Signal < 0.1
	r.mu.Unlock()
	if spent {
		r.Deposit()
	}
	return rel
}

func (r *Relation) Activation(rel *Relation) float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Signal
}

func (r *Relation) Deposit() {
	r.mu.Lock()
	cell := r.Forming
	if cell == nil {
		r.mu.Unlock()
		return
	}
	var parts []string
	for _, t := range r.Thoughts {
		if s, ok := t.(*Skyra); ok {
			parts = append(parts, s.Content)
			cell.Relationships[s.id] = s
		}
	}
	cell.Content = strings.Join(parts, "\n")
	cell.Weight = r.Signal
	host := r.LastSeen
	r.mu.Unlock()

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

	return &Relation{
		id:        NewID(),
		Origin:    origin,
		Impulse:   raw,
		Forming:   forming,
		Providers: make(map[string]Reality),
		Fields:    make(map[string]map[string]float64),
		Visited:   make(map[string]bool),
		Signal:    1.0,
		MaxLoad:   75000, // context window limit — thought fires when Load approaches this
		MaxDepth:  50,
	}, nil
}
