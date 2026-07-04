package dnav2

import (
	"fmt"
	"sync"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// EntityIndex — the master entity graph, held on the nucleus (the DNA). Entities
// are v1 MemoryCells (KindEntity) with interaction/state records as children. This
// is durable working memory: it rides EVERY worker's working set (handed by
// reference at transcription) and is traversed every wave. Writes go through the
// handed reference straight back to this master index (shared by reference).
type EntityIndex struct {
	mu    sync.Mutex
	slice skyra.Slice // nucleus slice for the nodes (nil-tolerant until activation lands)
	byID  map[string]*dna.MemoryCell
}

func NewEntityIndex(slice skyra.Slice) *EntityIndex {
	return &EntityIndex{slice: slice, byID: map[string]*dna.MemoryCell{}}
}

// Ensure get-or-creates the entity node for a resolved id. (Entity resolution is
// exact-match for now — the identity-registry discipline; fuzzy later.)
func (ei *EntityIndex) Ensure(id string) *dna.MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	if e, ok := ei.byID[id]; ok {
		return e
	}
	e := dna.NewMemoryCell(id, dna.KindEntity, dna.Entity{ID: id}, ei.slice)
	ei.byID[id] = e
	return e
}

// AppendInteraction records an utterance/action on an entity as a KindResponse
// child (the write-time discipline; the reverse index depends on it).
func (ei *EntityIndex) AppendInteraction(entity *dna.MemoryCell, rec dna.InteractionRecord) {
	if entity == nil {
		return
	}
	ei.mu.Lock()
	defer ei.mu.Unlock()
	id := fmt.Sprintf("%s:interaction:%d:%d", entity.ID(), rec.Timestamp, len(entity.Children()))
	cell := dna.NewMemoryCell(id, dna.KindResponse, rec, ei.slice)
	entity.AddChild(entity.ID()+"→"+id, cell)
}

// Snapshot returns the current entity nodes under the lock (for frame assembly —
// read live, the active state).
func (ei *EntityIndex) Snapshot() []*dna.MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	out := make([]*dna.MemoryCell, 0, len(ei.byID))
	for _, e := range ei.byID {
		out = append(out, e)
	}
	return out
}

// GroupingIndex — the grouping spine, on the nucleus. Groupings are KindGrouping
// MemoryCells carrying a Gist; predictions hang off them as children. Handed to
// the predict workers as their working set (the active grouping for the impulse).
type GroupingIndex struct {
	mu    sync.Mutex
	slice skyra.Slice
	byID  map[string]*dna.MemoryCell
}

func NewGroupingIndex(slice skyra.Slice) *GroupingIndex {
	return &GroupingIndex{slice: slice, byID: map[string]*dna.MemoryCell{}}
}

// Create mints a grouping carrying a gist and registers it (cold-start "create";
// warm match/edit later).
func (gi *GroupingIndex) Create(groupID, gist string, ts int) *dna.MemoryCell {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	g := dna.NewMemoryCell(groupID, dna.KindGrouping, dna.Gist{Text: gist, Timestamp: ts}, gi.slice)
	gi.byID[groupID] = g
	return g
}

// AddChild appends a node (a prediction) to a grouping under the index lock.
func (gi *GroupingIndex) AddChild(grouping *dna.MemoryCell, edgeKey string, child skyra.Space) {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	grouping.AddChild(edgeKey, child)
}
