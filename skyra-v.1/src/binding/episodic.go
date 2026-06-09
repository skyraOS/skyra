package binding

import (
	"sync"

	"skyra-v1/src/reality"
)

// Episodic — the hippocampus. Episodic binding.
// Retrieval via inverted index: cellID → episodeIDs.
// Scoring by overlap count — no iteration over episode members needed.
// Segmentation driven by Motivation (new goal = new episode).
// Does NOT hold a conversation log — holds sparse binding patterns.
type Episodic struct {
	mu      sync.Mutex
	id      string
	BeingID string

	Episodes map[string]*Episode  // episodeID → episode
	Index    map[string][]string  // cellID → episodeIDs containing this cell
	Current  string               // active episode ID
}

// Episode — a conjunctive binding pattern. Knows its members.
// TotalMembers stored for scoring without iterating the member set.
type Episode struct {
	id           string
	Members      map[string]reality.Reality // cellID → cell. The conjunction at deposit time.
	TotalMembers int
}

func NewEpisodic(beingID string) *Episodic {
	return &Episodic{
		id:       reality.NewID(),
		BeingID:  beingID,
		Episodes: make(map[string]*Episode),
		Index:    make(map[string][]string),
		Current:  reality.NewID(),
	}
}

func (e *Episodic) ID() string {
	return e.id
}

func (e *Episodic) Create(r *reality.Relation) reality.Reality {
	return &Episodic{
		id:       reality.NewID(),
		Episodes: make(map[string]*Episode),
		Index:    make(map[string][]string),
		Current:  reality.NewID(),
	}
}

// Score — inverted index lookup. Returns episodeID → overlap ratio.
// O(active_cells × avg_episodes_per_cell). No episode member iteration.
func (e *Episodic) Score(activeCells []string) map[string]float64 {
	hits := make(map[string]int)
	for _, cellID := range activeCells {
		for _, epID := range e.Index[cellID] {
			hits[epID]++
		}
	}

	scores := make(map[string]float64)
	for epID, count := range hits {
		ep := e.Episodes[epID]
		if ep != nil && ep.TotalMembers > 0 {
			scores[epID] = float64(count) / float64(ep.TotalMembers)
		}
	}
	return scores
}

// Deposit — form a new episode from the currently active cells.
// Adds the episode to the Episodes map and indexes each member.
func (e *Episodic) Deposit(activeCells map[string]reality.Reality) string {
	ep := &Episode{
		id:           reality.NewID(),
		Members:      make(map[string]reality.Reality, len(activeCells)),
		TotalMembers: len(activeCells),
	}
	for id, cell := range activeCells {
		ep.Members[id] = cell
	}

	e.mu.Lock()
	e.Episodes[ep.id] = ep
	for cellID := range activeCells {
		e.Index[cellID] = append(e.Index[cellID], ep.id)
	}
	e.Current = ep.id
	e.mu.Unlock()

	return ep.id
}

// Observe — read the attention map, score episodes, imprint fields.
// Episodes above threshold get their binding field written to the Relation.
func (e *Episodic) Observe(r *reality.Relation) {
	sl := r.GetSlice(e.BeingID)
	sl.Lock()
	activeCells := make([]string, 0, len(sl.Surroundings))
	for cellID := range sl.Surroundings {
		activeCells = append(activeCells, cellID)
	}
	sl.Unlock()

	if len(activeCells) == 0 {
		return
	}

	e.mu.Lock()
	scores := e.Score(activeCells)
	e.mu.Unlock()

	// imprint episodic binding field onto the Relation's Fields
	// episodes above threshold amplify their members' activation
	const threshold = 0.3
	r.RLock()
	fields := r.Fields
	r.RUnlock()

	if fields == nil {
		return
	}

	for epID, score := range scores {
		if score < threshold {
			continue
		}
		ep := e.Episodes[epID]
		if ep == nil {
			continue
		}
		for cellID := range ep.Members {
			if fields[cellID] == nil {
				fields[cellID] = make(map[string]float64)
			}
			fields[cellID]["episodic:overlap"] = score
		}
	}
}

// Express — episode boundary detection.
// When Motivation signals a new goal, the current episode closes
// and a new one begins on the next deposit.
func (e *Episodic) Express(r *reality.Relation) *reality.Relation {
	// TODO: read Motivation signal from Relation
	// if goal changed → close current episode, start new one
	// closed episodes become pure retrieval targets
	return r
}

func (e *Episodic) Realize(r *reality.Relation) *reality.Relation {
	e.Observe(r)
	r = e.Express(r)
	return r
}

func (e *Episodic) Competence(r *reality.Relation) {
}

func (e *Episodic) Activation(r *reality.Relation) float64 {
	return 1.0
}

func (e *Episodic) Reinforce() {
}

func (e *Episodic) Decay() {
}
