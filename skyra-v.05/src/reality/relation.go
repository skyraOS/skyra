package reality

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"skyra-v1/src/inference"
)

// Slice — a being's private cognitive state on the Relation.
// Each being gets its own slice. Own lock. Own goroutine.
// No being can access another being's slice.
type Slice struct {
	mu              sync.Mutex
	Thread             map[string]map[string][]Reality // thread ID → reality ID → entries
	TrainOfThought     map[string]Reality // cell ID → Reality. Presence = visited. Non-nil = picked up.
	TrainOfThoughtKeys []string           // insertion order for stable iteration
	Understandings     []Reality // the being's cognitive output — inference from propagate AND compressed products.
	PreviousActions    []Reality // what the being already did. Written by speak/motor.
	Signal             float64
	Load               int    // accumulated content size
	Forming            *Skyra // the cell being built during this traversal
	Goal               string // the destination — persistent, changed explicitly
	Intent             IntentGraph // the predicted future — refined at every compression boundary
	Surroundings       map[string]float64 // what the being can see right now — keyed by Reality ID, value is activation score. Falls off below threshold.
	LastSeen           Reality
	Depth              int
	Trace              []string
}

func (s *Slice) Lock()   { s.mu.Lock() }
func (s *Slice) Unlock() { s.mu.Unlock() }

// ThreadAppend adds an entry to the nested thread map. Caller must hold s.mu.
func (s *Slice) ThreadAppend(threadID, realityID string, entry Reality) {
	if s.Thread[threadID] == nil {
		s.Thread[threadID] = make(map[string][]Reality)
	}
	s.Thread[threadID][realityID] = append(s.Thread[threadID][realityID], entry)
}

// ThreadEntries returns all entries across all realities for a given thread. Caller must hold s.mu.
func (s *Slice) ThreadEntries(threadID string) []Reality {
	realities, ok := s.Thread[threadID]
	if !ok {
		return nil
	}
	var all []Reality
	for _, entries := range realities {
		all = append(all, entries...)
	}
	return all
}

// ThreadAll returns all entries across all threads and realities. Caller must hold s.mu.
func (s *Slice) ThreadAll() []Reality {
	var all []Reality
	for _, realities := range s.Thread {
		for _, entries := range realities {
			all = append(all, entries...)
		}
	}
	return all
}

type Relation struct {
	mu sync.RWMutex // protects shared state: World, Fields
	id        string
	Origin    string
	ThreadID  string
	Impulse   string
	World     map[string]Reality             // shared — beings, places, tools. The world is the Relation.
	Providers map[string]Reality             // shared — inference endpoints
	Fields    map[string]map[string]float64  // shared — binding fields
	Threads         map[string]*Thread           // active threads — keyed by thread ID. A being can listen to many.
	ActionHistories map[string]*ActionHistory   // action binding — keyed by "action:" + beingID
	Slices          map[string]*Slice            // private — keyed by being ID, tenant-isolated
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
	for _, sl := range r.Slices {
		sl.mu.Lock()
		sl.TrainOfThought = make(map[string]Reality)
		sl.TrainOfThoughtKeys = nil
		sl.mu.Unlock()
	}

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
	for _, id := range s.TrainOfThoughtKeys {
		t := s.TrainOfThought[id]
		if t == nil {
			continue
		}
		if sk, ok := t.(*Skyra); ok {
			parts = append(parts, sk.Content)
			cell.Relationships[sk.id] = sk
		}
	}
	cell.Content = strings.Join(parts, "\n")
	cell.FamiliarityWeight = s.Signal
	cell.IdentityWeight = s.Signal
	host := s.LastSeen
	s.mu.Unlock()

	vec, err := inference.Embed(cell.Content)
	if err == nil {
		cell.Vector = vec
	}

	if h, ok := host.(*Skyra); ok {
		h.Relationships[cell.id] = cell
	}

	r.mu.RLock()
	for _, thread := range r.Threads {
		if thread.BeingID == beingID {
			thread.Bind(cell.id)
		}
	}
	r.mu.RUnlock()
}

type SliceSnapshot struct {
	BeingID              string
	Signal               float64
	Load                 int
	ThreadCount          int
	TrainOfThoughtCount  int
	UnderstandingCount   int
	PreviousActionCount  int
	Goal                 string
	Intent               string
	Surroundings         map[string]float64
}

type ThreadSnapshot struct {
	ID      string
	BeingID string
	Members map[string]float64
}

func (r *Relation) SliceList() []SliceSnapshot {
	var out []SliceSnapshot
	for name, sl := range r.Slices {
		sl.mu.Lock()
		surroundings := make(map[string]float64, len(sl.Surroundings))
		for k, v := range sl.Surroundings {
			surroundings[k] = v
		}
		out = append(out, SliceSnapshot{
			BeingID:             name,
			Signal:              sl.Signal,
			Load:                sl.Load,
			ThreadCount:          len(sl.Thread),
			TrainOfThoughtCount:  len(sl.TrainOfThought),
			UnderstandingCount:   len(sl.Understandings),
			PreviousActionCount:  len(sl.PreviousActions),
			Goal:                sl.Goal,
			Intent:              sl.Intent.Render(),
			Surroundings:        surroundings,
		})
		sl.mu.Unlock()
	}
	return out
}

func (r *Relation) ThreadList() []ThreadSnapshot {
	var out []ThreadSnapshot
	for _, t := range r.Threads {
		t.mu.Lock()
		members := make(map[string]float64, len(t.Members))
		for k, v := range t.Members {
			members[k] = v
		}
		out = append(out, ThreadSnapshot{
			ID:      t.id,
			BeingID: t.BeingID,
			Members: members,
		})
		t.mu.Unlock()
	}
	return out
}

type ActionHistorySnapshot struct {
	ID      string
	BeingID string
	Members map[string]float64
}

func (r *Relation) ActionHistoryList() []ActionHistorySnapshot {
	var out []ActionHistorySnapshot
	for _, ah := range r.ActionHistories {
		ah.mu.Lock()
		members := make(map[string]float64, len(ah.Members))
		for k, v := range ah.Members {
			members[k] = v
		}
		out = append(out, ActionHistorySnapshot{
			ID:      ah.id,
			BeingID: ah.BeingID,
			Members: members,
		})
		ah.mu.Unlock()
	}
	return out
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
		id:                NewID(),
		IdentityWeight:    1.0,
		FamiliarityWeight: 1.0,
		CreatedAt:         time.Now().Unix(),
		Relationships:     make(map[string]Reality),
		Expressors:        make(map[string]Reality),
		ObserveFns:        make(map[string]func(*Relation)),
		ExpressFns:        make(map[string]func(*Relation) *Relation),
	}

	slice := &Slice{
		Thread:          make(map[string]map[string][]Reality),
		TrainOfThought:  make(map[string]Reality),
		Understandings:  []Reality{},
		PreviousActions: []Reality{},
		Signal:          1.0,
		Intent:          NewIntentGraph(),
		Forming:         forming,
		Surroundings:    make(map[string]float64),
	}

	return &Relation{
		id:        NewID(),
		Origin:    origin,
		Impulse:   raw,
		World:     make(map[string]Reality),
		Providers: make(map[string]Reality),
		Fields:    make(map[string]map[string]float64),
		Threads:         make(map[string]*Thread),
		ActionHistories: make(map[string]*ActionHistory),
		Slices:          map[string]*Slice{origin: slice},
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
		id:                NewID(),
		IdentityWeight:    1.0,
		FamiliarityWeight: 1.0,
		CreatedAt:         time.Now().Unix(),
		Relationships:     make(map[string]Reality),
		Expressors:        make(map[string]Reality),
		ObserveFns:        make(map[string]func(*Relation)),
		ExpressFns:        make(map[string]func(*Relation) *Relation),
	}

	s = &Slice{
		Thread:          make(map[string]map[string][]Reality),
		TrainOfThought:  make(map[string]Reality),
		Understandings:  []Reality{},
		PreviousActions: []Reality{},
		Signal:          1.0,
		Intent:          NewIntentGraph(),
		Forming:         forming,
		Surroundings:    make(map[string]float64),
	}

	r.mu.Lock()
	r.Slices[beingID] = s
	r.mu.Unlock()
	return s
}
