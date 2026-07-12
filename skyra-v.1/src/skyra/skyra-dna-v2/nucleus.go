package dnav2

import (
	"fmt"
	"sort"
	"sync"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// NodeRegistry — the master registry of MATERIAL NODES, held on the nucleus
// (the DNA). It owns both provenance kinds and the label lookup between them:
// Entities (born from OBSERVATION — sensory/ingest) and Terms (born from a
// query's ASSERTION). Entities are v1 MemoryCells (KindEntity) with
// interaction/state records as children. This is durable working memory: it
// rides EVERY worker's working set (handed by reference at transcription) and
// is traversed every wave. Writes go through the handed reference straight
// back to this master registry (shared by reference).
type NodeRegistry struct {
	mu    sync.Mutex
	slice skyra.Slice // nucleus slice for the nodes (nil-tolerant until activation lands)
	byID  map[string]*dna.MemoryCell
	terms map[string]*dna.MemoryCell // lookup key → term node ("has this label a node yet")
}

func NewNodeRegistry(slice skyra.Slice) *NodeRegistry {
	return &NodeRegistry{
		slice: slice,
		byID:  map[string]*dna.MemoryCell{},
		terms: map[string]*dna.MemoryCell{},
	}
}

// Get returns the entity node for an id WITHOUT creating (nil if unknown) —
// the non-creating half of resolution: entities are born from OBSERVATION
// (sensory/ingest), never from a query's assertion.
func (ei *NodeRegistry) Get(id string) *dna.MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	return ei.byID[id]
}

// EnsureEntity get-or-creates the entity node for a resolved id. (Entity
// resolution is exact-match for now — the identity-registry discipline; fuzzy
// later.)
func (ei *NodeRegistry) EnsureEntity(id string) *dna.MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	if e, ok := ei.byID[id]; ok {
		return e
	}
	e := dna.NewMemoryCell(id, dna.KindEntity, dna.Entity{ID: id}, ei.slice)
	ei.byID[id] = e
	return e
}

// Term returns the term node for a lookup key WITHOUT creating (nil if none).
func (ei *NodeRegistry) Term(key string) *dna.MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	return ei.terms[key]
}

// EnsureTerm get-or-creates a term node — a material node born from ASSERTION
// (a mentioned-but-never-observed referent). Exact-match on the lookup key;
// merging a term into a later-observed entity is the deferred merge op.
func (ei *NodeRegistry) EnsureTerm(key, cellID, label string) *dna.MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	if t, ok := ei.terms[key]; ok {
		return t
	}
	t := dna.NewMemoryCell(cellID, KindTerm, TermContent{Label: label}, ei.slice)
	ei.terms[key] = t
	return t
}

// AppendInteraction records an utterance/action on an entity as a KindResponse
// child (the write-time discipline; the reverse index depends on it).
func (ei *NodeRegistry) AppendInteraction(entity *dna.MemoryCell, rec dna.InteractionRecord) {
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
func (ei *NodeRegistry) Latest() (node, entity *dna.MemoryCell, ok bool) {
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
func (ei *NodeRegistry) LinkGrouping(entity, grouping *dna.MemoryCell) {
	if entity == nil || grouping == nil {
		return
	}
	ei.mu.Lock()
	defer ei.mu.Unlock()
	entity.AddChild(grouping.ID(), grouping)
}

// Snapshot returns the current entity nodes under the lock (for frame assembly —
// read live, the active state).
func (ei *NodeRegistry) Snapshot() []*dna.MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	out := make([]*dna.MemoryCell, 0, len(ei.byID))
	for _, e := range ei.byID {
		out = append(out, e)
	}
	return out
}

// GroupingIndex — the grouping spine, on the nucleus. A grouping is an episode
// SHELL (KindGrouping, content Grouping); its framing lives in KindGist CHILDREN —
// a sequence (edit appends a new gist, priors retained and decaying) — and the
// gist owns the records/predictions made under it (the gist-owned spine,
// perceive-internalization.md §3). Handed to the predict workers as their working
// set (the active grouping for the impulse).
type GroupingIndex struct {
	mu       sync.Mutex
	slice    skyra.Slice
	byID     map[string]*dna.MemoryCell
	gistByID map[string]*dna.MemoryCell
	// the current episode + its current gist (active-grouping.md) live ON THE
	// SLICE as IDs (mapActiveGroupingID/mapActiveGistID) — active state's
	// canonical home; resolved to nodes through these maps on read.
}

func NewGroupingIndex(slice skyra.Slice) *GroupingIndex {
	return &GroupingIndex{slice: slice, byID: map[string]*dna.MemoryCell{}, gistByID: map[string]*dna.MemoryCell{}}
}

// Create mints a grouping with its first gist child, registers both, and makes
// them the ACTIVE episode/gist. (Cold-start "create"; warm edit appends instead.)
func (gi *GroupingIndex) Create(groupID, gist string, ts int) *dna.MemoryCell {
	gi.mu.Lock()
	g := dna.NewMemoryCell(groupID, dna.KindGrouping, dna.Grouping{Created: ts}, gi.slice)
	gi.byID[groupID] = g
	gistCell := gi.mintGistLocked(g, gist, ts)
	gi.mu.Unlock()
	sliceSetID(gi.slice, mapActiveGroupingID, groupID)
	sliceSetID(gi.slice, mapActiveGistID, gistCell.ID())
	return g
}

// Edit appends a NEW gist to an existing grouping (the prior gist stays and
// decays — the sequence) and makes the grouping + fresh gist active. Returns the
// grouping, nil if the id is unknown (caller falls through to Create).
func (gi *GroupingIndex) Edit(groupID, gist string, ts int) *dna.MemoryCell {
	gi.mu.Lock()
	g := gi.byID[groupID]
	if g == nil {
		gi.mu.Unlock()
		return nil
	}
	gistCell := gi.mintGistLocked(g, gist, ts)
	gi.mu.Unlock()
	sliceSetID(gi.slice, mapActiveGroupingID, groupID)
	sliceSetID(gi.slice, mapActiveGistID, gistCell.ID())
	return g
}

// mintGistLocked appends a gist child to g (caller holds gi.mu).
func (gi *GroupingIndex) mintGistLocked(g *dna.MemoryCell, gist string, ts int) *dna.MemoryCell {
	id := fmt.Sprintf("%s:gist:%d", g.ID(), ts)
	cell := dna.NewMemoryCell(id, dna.KindGist, dna.Gist{Text: gist, Timestamp: ts}, gi.slice)
	gi.gistByID[id] = cell
	g.AddChild(g.ID()+"→"+id, cell)
	return cell
}

// Active returns the current episode (nil before the first grouping exists) —
// the slice's activeGroupingID resolved through the index.
func (gi *GroupingIndex) Active() *dna.MemoryCell {
	id := sliceGetID(gi.slice, mapActiveGroupingID)
	if id == "" {
		return nil
	}
	gi.mu.Lock()
	defer gi.mu.Unlock()
	return gi.byID[id]
}

// ActiveGist returns the current gist — where deposits (records, predictions)
// attach. nil before the first grouping exists.
func (gi *GroupingIndex) ActiveGist() *dna.MemoryCell {
	id := sliceGetID(gi.slice, mapActiveGistID)
	if id == "" {
		return nil
	}
	gi.mu.Lock()
	defer gi.mu.Unlock()
	return gi.gistByID[id]
}

// SetActive makes g the current episode (writes its ID + its latest gist's ID to
// the slice; nil clears both).
func (gi *GroupingIndex) SetActive(g *dna.MemoryCell) {
	groupID, gistID := "", ""
	if g != nil {
		groupID = g.ID()
		if gist := latestGist(g); gist != nil {
			gistID = gist.ID()
		}
	}
	sliceSetID(gi.slice, mapActiveGroupingID, groupID)
	sliceSetID(gi.slice, mapActiveGistID, gistID)
}

// Get returns a grouping by id (nil if none) — the target of an <edit id="..."> op.
func (gi *GroupingIndex) Get(id string) *dna.MemoryCell {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	return gi.byID[id]
}

// latestGist returns g's current framing — the max-timestamp KindGist child (the
// sequence's newest entry; ticks are monotonic so newest == current). nil if none.
func latestGist(g *dna.MemoryCell) *dna.MemoryCell {
	if g == nil {
		return nil
	}
	var best *dna.MemoryCell
	bestTS := -1
	for _, c := range g.Children() {
		mc, ok := c.(*dna.MemoryCell)
		if !ok || mc.Kind() != dna.KindGist {
			continue
		}
		if gg, ok := mc.Content().(dna.Gist); ok && gg.Timestamp > bestTS {
			bestTS, best = gg.Timestamp, mc
		}
	}
	return best
}

// gistOf extracts a grouping's CURRENT framing text (its latest gist child).
func gistOf(g *dna.MemoryCell) string {
	if gist := latestGist(g); gist != nil {
		if gg, ok := gist.Content().(dna.Gist); ok {
			return gg.Text
		}
	}
	return ""
}

// OpenEpisodes returns every grouping's id + current gist — the candidates the
// framing op decides edit-or-new against (a cheap gist-level view).
func (gi *GroupingIndex) OpenEpisodes() []struct{ ID, Gist string } {
	gi.mu.Lock()
	groupings := make([]*dna.MemoryCell, 0, len(gi.byID))
	for _, g := range gi.byID {
		groupings = append(groupings, g)
	}
	gi.mu.Unlock()
	out := make([]struct{ ID, Gist string }, 0, len(groupings))
	for _, g := range groupings {
		out = append(out, struct{ ID, Gist string }{g.ID(), gistOf(g)})
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

// predictionsOf walks a grouping's gist children and yields every prediction node
// hanging under them — the gist-owned spine's one-level-deeper read (predictions
// attach to the gist they were made under; a bet made under a PRIOR gist is still
// gradeable — evidence can close a bet conditioned on the old framing).
func predictionsOf(g *dna.MemoryCell) []*dna.MemoryCell {
	var out []*dna.MemoryCell
	for _, c := range g.Children() {
		gist, ok := c.(*dna.MemoryCell)
		if !ok || gist.Kind() != dna.KindGist {
			continue
		}
		for _, gc := range gist.Children() {
			if mc, ok := gc.(*dna.MemoryCell); ok && mc.Kind() == dna.KindPrediction {
				out = append(out, mc)
			}
		}
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
		for _, cc := range predictionsOf(g) {
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
		for _, cc := range predictionsOf(g) {
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

// Latest returns the most-recently-framed grouping (by its current gist's
// timestamp), for inspection/display. nil if none.
func (gi *GroupingIndex) Latest() *dna.MemoryCell {
	gi.mu.Lock()
	groupings := make([]*dna.MemoryCell, 0, len(gi.byID))
	for _, g := range gi.byID {
		groupings = append(groupings, g)
	}
	gi.mu.Unlock()
	var best *dna.MemoryCell
	bestTS := -1
	for _, g := range groupings {
		if gist := latestGist(g); gist != nil {
			if gg, ok := gist.Content().(dna.Gist); ok && gg.Timestamp > bestTS {
				bestTS, best = gg.Timestamp, g
			}
		}
	}
	return best
}
