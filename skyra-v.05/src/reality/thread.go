package reality

import "sync"

// Thread — a binding Reality. Owns a local activation field over the graph.
// On descent: imprints thread_alignment and thread_strength onto the Relation.
// On ascent: updates its own weights from what was actually used.
// The hippocampal index — binds things that co-occurred.
type Thread struct {
	mu      sync.Mutex
	id      string
	BeingID string

	// members — node IDs that belong to this thread. Keyed by Reality ID.
	// Value is thread_strength — how important this node is to the thread right now.
	Members map[string]float64

	// traversal count — proper time for this thread
	TraversalCount int
}

func NewThread(id, beingID string) *Thread {
	return &Thread{
		id:      id,
		BeingID: beingID,
		Members: make(map[string]float64),
	}
}

func (t *Thread) ID() string { return t.id }

func (t *Thread) Create(r *Relation) Reality { return nil }

// Realize — the thread's traversal. Observe imprints, Express learns.
func (t *Thread) Realize(r *Relation) *Relation {
	if r.Cancelled {
		return r
	}
	t.Observe(r)
	r = t.Express(r)
	t.TraversalCount++
	return r
}

func (t *Thread) Competence(r *Relation) {}

// Observe — descent. The thread imprints its field onto the Relation.
// thread_alignment: hard gate. 1.0 if member, 0.0 if not.
// thread_strength: soft amplifier. How important this node is right now.
func (t *Thread) Observe(r *Relation) {
	t.mu.Lock()
	members := make(map[string]float64, len(t.Members))
	for k, v := range t.Members {
		members[k] = v
	}
	t.mu.Unlock()

	r.mu.Lock()
	if r.Fields == nil {
		r.mu.Unlock()
		return
	}
	for nodeID, strength := range members {
		if r.Fields[nodeID] == nil {
			r.Fields[nodeID] = make(map[string]float64)
		}
		r.Fields[nodeID]["thread:alignment"] = 1.0
		r.Fields[nodeID]["thread:strength"] = strength
	}
	r.mu.Unlock()
}

// Express — ascent. The thread reads what was actually hit and updates
// its own strength weights. Nodes that were visited get reinforced.
// Nodes that weren't decay.
func (t *Thread) Express(r *Relation) *Relation {
	sl := r.GetSlice(t.BeingID)
	sl.mu.Lock()
	visited := make(map[string]bool, len(sl.Visited))
	for k, v := range sl.Visited {
		visited[k] = v
	}
	sl.mu.Unlock()

	t.mu.Lock()
	for nodeID := range t.Members {
		if visited[nodeID] {
			t.Members[nodeID] = Familiarity(t.Members[nodeID], true)
		} else {
			t.Members[nodeID] = Familiarity(t.Members[nodeID], false)
		}
	}
	t.mu.Unlock()

	return r
}

// Bind — add a node to this thread. Called when a new cell is created
// during a conversation, or when an existing cell is activated.
func (t *Thread) Bind(nodeID string) {
	t.mu.Lock()
	if _, exists := t.Members[nodeID]; !exists {
		t.Members[nodeID] = 1.0
	}
	t.mu.Unlock()
}

// Speak — bump thread_strength on specific nodes. The being said something
// that touched these nodes. They matter more now.
func (t *Thread) Speak(nodeIDs []string) {
	t.mu.Lock()
	for _, id := range nodeIDs {
		if strength, exists := t.Members[id]; exists {
			t.Members[id] = Familiarity(strength, true)
		}
	}
	t.mu.Unlock()
}

func (t *Thread) Activation(r *Relation) float64 {
	return 1.0
}

func (t *Thread) Reinforce() {}
func (t *Thread) Decay()     {}
