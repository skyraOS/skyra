package reality

import "sync"

// ActionHistory — a binding Reality for motor memory. Owns a local activation
// field over nodes the being has acted on. Soft amplifier only — no hard gate.
// Things the being hasn't acted on are unaffected. Things it has get boosted
// by how recently and frequently the action occurred.
type ActionHistory struct {
	mu      sync.Mutex
	id      string
	BeingID string

	Members        map[string]float64
	TraversalCount int
}

func NewActionHistory(id, beingID string) *ActionHistory {
	return &ActionHistory{
		id:      id,
		BeingID: beingID,
		Members: make(map[string]float64),
	}
}

func (ah *ActionHistory) ID() string { return ah.id }

func (ah *ActionHistory) Create(r *Relation) Reality { return nil }

func (ah *ActionHistory) Realize(r *Relation) *Relation {
	if r.Cancelled {
		return r
	}
	ah.Observe(r)
	r = ah.Express(r)
	ah.TraversalCount++
	return r
}

func (ah *ActionHistory) Competence(r *Relation) {}

// Observe — descent. Imprint action:recency as a soft amplifier onto the
// Relation's Fields. Non-members get no entry — effectively floor=1.0.
func (ah *ActionHistory) Observe(r *Relation) {
	ah.mu.Lock()
	members := make(map[string]float64, len(ah.Members))
	for k, v := range ah.Members {
		members[k] = v
	}
	ah.mu.Unlock()

	r.mu.Lock()
	if r.Fields == nil {
		r.mu.Unlock()
		return
	}
	for nodeID, strength := range members {
		if r.Fields[nodeID] == nil {
			r.Fields[nodeID] = make(map[string]float64)
		}
		r.Fields[nodeID]["action:recency"] = strength
	}
	r.mu.Unlock()
}

// Express — ascent. Read what was visited, reinforce members that were hit,
// decay members that weren't. Familiarity EMA (α=0.1).
func (ah *ActionHistory) Express(r *Relation) *Relation {
	sl := r.GetSlice(ah.BeingID)
	sl.mu.Lock()
	visited := make(map[string]bool, len(sl.TrainOfThought))
	for k := range sl.TrainOfThought {
		visited[k] = true
	}
	sl.mu.Unlock()

	ah.mu.Lock()
	for nodeID := range ah.Members {
		if visited[nodeID] {
			ah.Members[nodeID] = Familiarity(ah.Members[nodeID], true)
		} else {
			ah.Members[nodeID] = Familiarity(ah.Members[nodeID], false)
		}
	}
	ah.mu.Unlock()

	return r
}

func (ah *ActionHistory) Bind(nodeID string) {
	ah.mu.Lock()
	if _, exists := ah.Members[nodeID]; !exists {
		ah.Members[nodeID] = 1.0
	}
	ah.mu.Unlock()
}

func (ah *ActionHistory) Activation(r *Relation) float64 {
	return 1.0
}

func (ah *ActionHistory) Reinforce() {}
func (ah *ActionHistory) Decay()     {}
