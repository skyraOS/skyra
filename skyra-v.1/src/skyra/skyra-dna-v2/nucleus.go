package dnav2

import (
	"fmt"
	"sort"
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

// Latest returns the newest interaction node (KindResponse) across all entities, with
// its owning entity — the interaction that just arrived, for midlevel to consolidate
// onto the active episode.
func (ei *EntityIndex) Latest() (node, entity *dna.MemoryCell, ok bool) {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	bestTS := -1
	for _, e := range ei.byID {
		for _, child := range e.Children() {
			mc, isCell := child.(*dna.MemoryCell)
			if !isCell {
				continue
			}
			if rec, isRec := mc.Content().(dna.InteractionRecord); isRec && rec.Timestamp > bestTS {
				bestTS, node, entity = rec.Timestamp, mc, e
			}
		}
	}
	return node, entity, node != nil
}

// LinkGrouping wires entity → grouping (the episodes a participant is in), idempotent on
// the grouping id — cool entities point to episodes, not raw interactions.
func (ei *EntityIndex) LinkGrouping(entity, grouping *dna.MemoryCell) {
	if entity == nil || grouping == nil {
		return
	}
	ei.mu.Lock()
	defer ei.mu.Unlock()
	entity.AddChild(grouping.ID(), grouping)
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
	mu     sync.Mutex
	slice  skyra.Slice
	byID   map[string]*dna.MemoryCell
	active *dna.MemoryCell // the current episode (active-grouping.md); sensory attaches here
}

func NewGroupingIndex(slice skyra.Slice) *GroupingIndex {
	return &GroupingIndex{slice: slice, byID: map[string]*dna.MemoryCell{}}
}

// Create mints a grouping carrying a gist, registers it, and makes it the ACTIVE
// episode. (Cold-start "create"; warm match/edit sharpens the active gist instead.)
func (gi *GroupingIndex) Create(groupID, gist string, ts int) *dna.MemoryCell {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	g := dna.NewMemoryCell(groupID, dna.KindGrouping, dna.Gist{Text: gist, Timestamp: ts}, gi.slice)
	gi.byID[groupID] = g
	gi.active = g
	return g
}

// Active returns the current episode (nil before the first grouping exists).
func (gi *GroupingIndex) Active() *dna.MemoryCell {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	return gi.active
}

// SetActive makes g the current episode.
func (gi *GroupingIndex) SetActive(g *dna.MemoryCell) {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	gi.active = g
}

// Get returns a grouping by id (nil if none) — the target of an <edit id="..."> op.
func (gi *GroupingIndex) Get(id string) *dna.MemoryCell {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	return gi.byID[id]
}

// OpenEpisodes returns every grouping's id + gist — the candidates the framing op
// decides edit-or-new against (a cheap gist-level view, not the full node sets).
func (gi *GroupingIndex) OpenEpisodes() []struct{ ID, Gist string } {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	out := make([]struct{ ID, Gist string }, 0, len(gi.byID))
	for id, g := range gi.byID {
		gist := ""
		if gg, ok := g.Content().(dna.Gist); ok {
			gist = gg.Text
		}
		out = append(out, struct{ ID, Gist string }{id, gist})
	}
	return out
}

// AddChild appends a node (a prediction) to a grouping under the index lock.
func (gi *GroupingIndex) AddChild(grouping *dna.MemoryCell, edgeKey string, child skyra.Space) {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	grouping.AddChild(edgeKey, child)
}

// Snapshot returns a grouping's direct children under the index lock — the safe
// read path when other:predict and self:predict fan out concurrently (other writes
// the distribution, self reads it). Prediction root nodes are immutable once added,
// so their own children are lock-free to read.
func (gi *GroupingIndex) Snapshot(grouping *dna.MemoryCell) []skyra.Space {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	out := make([]skyra.Space, 0, len(grouping.Children()))
	for _, c := range grouping.Children() {
		out = append(out, c)
	}
	return out
}

// LatestUngradedOther finds the newest ungraded other-pred distribution across all
// groupings — the prior bet other:grade closes against fresh evidence. Returns the
// distribution root and its branch children (sorted by edge key for stable prompt
// numbering). nil if there is no open bet. Ported from v1 latestUngradedDistribution.
func (gi *GroupingIndex) LatestUngradedOther() (*dna.MemoryCell, []*dna.MemoryCell) {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	var best *dna.MemoryCell
	bestTS := -1
	for _, g := range gi.byID {
		for _, child := range g.Children() {
			cc, ok := child.(*dna.MemoryCell)
			if !ok {
				continue
			}
			p, ok := cc.Content().(dna.Prediction)
			if !ok || p.Role != "other" || p.Graded || len(cc.Children()) == 0 {
				continue
			}
			if p.Timestamp > bestTS {
				bestTS, best = p.Timestamp, cc
			}
		}
	}
	if best == nil {
		return nil, nil
	}
	keys := make([]string, 0, len(best.Children()))
	for k := range best.Children() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	kids := make([]*dna.MemoryCell, 0, len(keys))
	for _, k := range keys {
		kids = append(kids, best.Children()[k].(*dna.MemoryCell))
	}
	return best, kids
}

// StampOtherGrade records the judge's verdict on a distribution under the index
// lock: the root carries the probability mass on the realized branch (the proper
// score, 0 on a total miss); each branch child carries 1 (realized) / 0. matchedIdx
// is 1-based; 0 marks a total miss. Ported from v1 gradePrior's write-back.
func (gi *GroupingIndex) StampOtherGrade(root *dna.MemoryCell, kids []*dna.MemoryCell, matchedIdx int, score float64, evidence string) {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	if p, ok := root.Content().(dna.Prediction); ok {
		p.Graded, p.Grade, p.Evidence = true, score, evidence
		root.SetContent(p)
	}
	for i, kid := range kids {
		p, ok := kid.Content().(dna.Prediction)
		if !ok {
			continue
		}
		p.Graded, p.Evidence = true, evidence
		if i == matchedIdx-1 {
			p.Grade = 1
		}
		kid.SetContent(p)
	}
}

// LatestUngradedRealizedSelf finds the newest ungraded self proposal that has been
// REALIZED (carries an action child — the intent was actually emitted). Returns the
// proposal root, its prediction content, and the latest action on it. self:grade
// closes it against the reafference. Ported from v1 latestUngradedRealized.
func (gi *GroupingIndex) LatestUngradedRealizedSelf() (*dna.MemoryCell, dna.Prediction, dna.ActionRecord, bool) {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	var best *dna.MemoryCell
	var bestPred dna.Prediction
	bestTS := -1
	for _, g := range gi.byID {
		for _, child := range g.Children() {
			cc, ok := child.(*dna.MemoryCell)
			if !ok {
				continue
			}
			p, ok := cc.Content().(dna.Prediction)
			if !ok || p.Role != "self" || p.Graded || len(cc.Children()) == 0 {
				continue
			}
			if p.Timestamp > bestTS {
				bestTS, best, bestPred = p.Timestamp, cc, p
			}
		}
	}
	if best == nil {
		return nil, dna.Prediction{}, dna.ActionRecord{}, false
	}
	var action dna.ActionRecord
	actionTS := -1
	for _, child := range best.Children() {
		cc, ok := child.(*dna.MemoryCell)
		if !ok || cc.Kind() != dna.KindAction {
			continue
		}
		if rec, ok := cc.Content().(dna.ActionRecord); ok && rec.Timestamp > actionTS {
			actionTS, action = rec.Timestamp, rec
		}
	}
	if actionTS < 0 {
		return nil, dna.Prediction{}, dna.ActionRecord{}, false
	}
	return best, bestPred, action, true
}

// StampSelfGrade records the execution judge's verdict on a realized proposal under
// the index lock (score 1 / 0.5 / 0; evidence = the world result). Ported from v1
// gradePrior's write-back.
func (gi *GroupingIndex) StampSelfGrade(proposal *dna.MemoryCell, score float64, evidence string) {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	if p, ok := proposal.Content().(dna.Prediction); ok {
		p.Graded, p.Grade, p.Evidence = true, score, evidence
		proposal.SetContent(p)
	}
}

// All returns every grouping under the index lock (inspection / trace assembly).
func (gi *GroupingIndex) All() []*dna.MemoryCell {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	out := make([]*dna.MemoryCell, 0, len(gi.byID))
	for _, g := range gi.byID {
		out = append(out, g)
	}
	return out
}

// Latest returns the most-recently-created grouping (by gist timestamp), for
// inspection/display. nil if none.
func (gi *GroupingIndex) Latest() *dna.MemoryCell {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	var best *dna.MemoryCell
	bestTS := -1
	for _, g := range gi.byID {
		if gist, ok := g.Content().(dna.Gist); ok && gist.Timestamp > bestTS {
			bestTS, best = gist.Timestamp, g
		}
	}
	return best
}
