package dnav2

import (
	"fmt"
	"sort"
	"sync"

	"skyra-v1/src/skyra"
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
	mu             sync.Mutex
	slice          skyra.Slice // nucleus slice for the nodes (nil-tolerant until activation lands)
	byID           map[string]*MemoryCell
	terms          map[string]*MemoryCell // lookup key → term node ("has this label a node yet")
	causalByNode   map[string]map[string]*MemoryCell
	causalByPair   map[SemanticPair]map[string]*MemoryCell
	causalByTriple map[SemanticTriple]map[string]*MemoryCell
}

func NewNodeRegistry(slice skyra.Slice) *NodeRegistry {
	return &NodeRegistry{
		slice:          slice,
		byID:           map[string]*MemoryCell{},
		terms:          map[string]*MemoryCell{},
		causalByNode:   map[string]map[string]*MemoryCell{},
		causalByPair:   map[SemanticPair]map[string]*MemoryCell{},
		causalByTriple: map[SemanticTriple]map[string]*MemoryCell{},
	}
}

// SemanticPair is the direction-preserving endpoint portion of a semantic
// edge signature. It is a retrieval key, not a stored causal relationship.
type SemanticPair struct {
	FromID string
	ToID   string
}

// SemanticTriple is the complete directed semantic edge signature used for
// exact candidate verification.
type SemanticTriple struct {
	FromID string
	Label  string
	ToID   string
}

func addSummary[K comparable](index map[K]map[string]*MemoryCell, key K, summary *MemoryCell) {
	if index[key] == nil {
		index[key] = map[string]*MemoryCell{}
	}
	index[key][summary.ID()] = summary
}

// IndexCausalNode publishes the reverse retrieval projections for one
// immutable birth subgraph. The semantic node/edge content remains the source
// of meaning; these maps only find candidate summaries efficiently.
func (ei *NodeRegistry) IndexCausalNode(summary *MemoryCell, nodes []*MemoryCell, edges []EdgeRecord) {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	for _, node := range nodes {
		addSummary(ei.causalByNode, node.ID(), summary)
		node.AddChildTo(PortSummaries, summary.ID(), summary)
	}
	for _, edge := range edges {
		if edge.Kind != EdgeStructural {
			continue
		}
		addSummary(ei.causalByPair, SemanticPair{FromID: edge.FromID, ToID: edge.ToID}, summary)
		addSummary(ei.causalByTriple, SemanticTriple{FromID: edge.FromID, Label: edge.Label, ToID: edge.ToID}, summary)
	}
}

// SummariesForNode returns the reverse-index candidates containing nodeID.
func (ei *NodeRegistry) SummariesForNode(nodeID string) []*MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	return summarySet(ei.causalByNode[nodeID])
}

// SummariesForPair returns candidates containing the ordered endpoints.
func (ei *NodeRegistry) SummariesForPair(fromID, toID string) []*MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	return summarySet(ei.causalByPair[SemanticPair{FromID: fromID, ToID: toID}])
}

// SummariesForTriple returns candidates containing the complete directed edge.
func (ei *NodeRegistry) SummariesForTriple(fromID, label, toID string) []*MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	return summarySet(ei.causalByTriple[SemanticTriple{FromID: fromID, Label: label, ToID: toID}])
}

// CandidateOccurrences counts how many query features index each summary.
// The count is suitable for initial amplitude allocation only; callers must
// still verify the complete directed subgraph before treating a candidate as a
// semantic match.
func (ei *NodeRegistry) CandidateOccurrences(nodeIDs []string, triples []SemanticTriple) map[string]int {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	out := map[string]int{}
	for _, nodeID := range nodeIDs {
		for summaryID := range ei.causalByNode[nodeID] {
			out[summaryID]++
		}
	}
	for _, triple := range triples {
		for summaryID := range ei.causalByTriple[triple] {
			out[summaryID]++
		}
	}
	return out
}

func summarySet(in map[string]*MemoryCell) []*MemoryCell {
	ids := make([]string, 0, len(in))
	for id := range in {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]*MemoryCell, 0, len(ids))
	for _, id := range ids {
		out = append(out, in[id])
	}
	return out
}

// Get returns the entity node for an id WITHOUT creating (nil if unknown) —
// the non-creating half of resolution: entities are born from OBSERVATION
// (sensory/ingest), never from a query's assertion.
func (ei *NodeRegistry) Get(id string) *MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	return ei.byID[id]
}

// EnsureEntity get-or-creates the entity node for a resolved id. (Entity
// resolution is exact-match for now — the identity-registry discipline; fuzzy
// later.)
func (ei *NodeRegistry) EnsureEntity(id string) *MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	if e, ok := ei.byID[id]; ok {
		return e
	}
	e := NewMemoryCell(id, KindEntity, Entity{ID: id, Provenance: ProvenanceObserved}, ei.slice)
	ei.byID[id] = e
	return e
}

// Term returns the term node for a lookup key WITHOUT creating (nil if none).
func (ei *NodeRegistry) Term(key string) *MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	return ei.terms[key]
}

// EnsureTerm get-or-creates a term node — a material node born from ASSERTION
// (a mentioned-but-never-observed referent). Exact-match on the lookup key;
// merging a term into a later-observed entity is the deferred merge op.
func (ei *NodeRegistry) EnsureTerm(key, cellID, label string) *MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	if t, ok := ei.terms[key]; ok {
		return t
	}
	t := NewMemoryCell(cellID, KindTerm, TermContent{Label: label, Provenance: ProvenanceAsserted}, ei.slice)
	ei.terms[key] = t
	return t
}

// AppendInteraction records an utterance/action on an entity as a KindResponse
// child (the write-time discipline; the reverse index depends on it).
func (ei *NodeRegistry) AppendInteraction(entity *MemoryCell, rec InteractionRecord) {
	if entity == nil {
		return
	}
	ei.mu.Lock()
	defer ei.mu.Unlock()
	id := fmt.Sprintf("%s:interaction:%d:%d", entity.ID(), rec.Timestamp, len(entity.Children()))
	cell := NewMemoryCell(id, KindResponse, rec, ei.slice)
	entity.AddChild(entity.ID()+"→"+id, cell)
}

// Latest returns the newest interaction node (KindResponse) across all entities, with
// its owning entity — the interaction that just arrived, for midlevel to consolidate
// onto the active episode.
func (ei *NodeRegistry) Latest() (node, entity *MemoryCell, ok bool) {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	bestTS := -1
	for _, e := range ei.byID {
		for _, child := range e.Children() {
			mc, isCell := child.(*MemoryCell)
			if !isCell {
				continue
			}
			if rec, isRec := mc.Content().(InteractionRecord); isRec && rec.Timestamp > bestTS {
				bestTS, node, entity = rec.Timestamp, mc, e
			}
		}
	}
	return node, entity, node != nil
}

// LinkEpisode wires entity → episode (the episodes a participant is in), idempotent on
// the episode id — cool entities point to episodes, not raw interactions.
func (ei *NodeRegistry) LinkEpisode(entity, episode *MemoryCell) {
	if entity == nil || episode == nil {
		return
	}
	ei.mu.Lock()
	defer ei.mu.Unlock()
	entity.AddChild(episode.ID(), episode)
}

// Snapshot returns the current entity nodes under the lock (for frame assembly —
// read live, the active state).
func (ei *NodeRegistry) Snapshot() []*MemoryCell {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	out := make([]*MemoryCell, 0, len(ei.byID))
	for _, e := range ei.byID {
		out = append(out, e)
	}
	return out
}

// EpisodeStore — the episode spine, on the nucleus. An episode is an episode
// SHELL (KindEpisode, content Episode); its framing lives in KindFrame CHILDREN —
// a sequence (edit appends a new frame, priors retained and decaying) — and the
// frame owns the records/predictions made under it (the frame-owned spine,
// perceive-internalization.md §3). Handed to the predict workers as their working
// set (the active episode for the impulse).
type EpisodeStore struct {
	mu        sync.Mutex
	slice     skyra.Slice
	byID      map[string]*MemoryCell
	frameByID map[string]*MemoryCell
	// the current episode + its current frame (active-grouping.md) live ON THE
	// SLICE as IDs (mapActiveEpisodeID/mapActiveFrameID) — active state's
	// canonical home; resolved to nodes through these maps on read.
}

func NewEpisodeStore(slice skyra.Slice) *EpisodeStore {
	return &EpisodeStore{slice: slice, byID: map[string]*MemoryCell{}, frameByID: map[string]*MemoryCell{}}
}

// Create mints an episode with its first frame child, registers both, and makes
// them the ACTIVE episode/frame. (Cold-start "create"; warm edit appends instead.)
func (store *EpisodeStore) Create(episodeID, frame string, ts int) *MemoryCell {
	store.mu.Lock()
	g := NewMemoryCell(episodeID, KindEpisode, Episode{Created: ts}, store.slice)
	store.byID[episodeID] = g
	frameCell := store.mintFrameLocked(g, frame, ts)
	store.mu.Unlock()
	sliceSetID(store.slice, mapActiveEpisodeID, episodeID)
	sliceSetID(store.slice, mapActiveFrameID, frameCell.ID())
	return g
}

// Edit appends a NEW frame to an existing episode (the prior frame stays and
// decays — the sequence) and makes the episode + fresh frame active. Returns the
// episode, nil if the id is unknown (caller falls through to Create).
func (store *EpisodeStore) Edit(episodeID, frame string, ts int) *MemoryCell {
	store.mu.Lock()
	g := store.byID[episodeID]
	if g == nil {
		store.mu.Unlock()
		return nil
	}
	frameCell := store.mintFrameLocked(g, frame, ts)
	store.mu.Unlock()
	sliceSetID(store.slice, mapActiveEpisodeID, episodeID)
	sliceSetID(store.slice, mapActiveFrameID, frameCell.ID())
	return g
}

// mintFrameLocked appends a frame child to g (caller holds store.mu).
func (store *EpisodeStore) mintFrameLocked(g *MemoryCell, frame string, ts int) *MemoryCell {
	id := fmt.Sprintf("%s:frame:%d", g.ID(), ts)
	cell := NewMemoryCell(id, KindFrame, Frame{Text: frame, Timestamp: ts}, store.slice)
	store.frameByID[id] = cell
	g.AddChild(g.ID()+"→"+id, cell)
	return cell
}

// Active returns the current episode (nil before the first episode exists) —
// the slice's activeEpisodeID resolved through the index.
func (store *EpisodeStore) Active() *MemoryCell {
	id := sliceGetID(store.slice, mapActiveEpisodeID)
	if id == "" {
		return nil
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.byID[id]
}

// ActiveFrame returns the current frame — where deposits (records, predictions)
// attach. nil before the first episode exists.
func (store *EpisodeStore) ActiveFrame() *MemoryCell {
	id := sliceGetID(store.slice, mapActiveFrameID)
	if id == "" {
		return nil
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.frameByID[id]
}

// SetActive makes g the current episode (writes its ID + its latest frame's ID to
// the slice; nil clears both).
func (store *EpisodeStore) SetActive(g *MemoryCell) {
	episodeID, frameID := "", ""
	if g != nil {
		episodeID = g.ID()
		if frame := latestFrame(g); frame != nil {
			frameID = frame.ID()
		}
	}
	sliceSetID(store.slice, mapActiveEpisodeID, episodeID)
	sliceSetID(store.slice, mapActiveFrameID, frameID)
}

// Get returns an episode by id (nil if none) — the target of an <edit id="..."> op.
func (store *EpisodeStore) Get(id string) *MemoryCell {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.byID[id]
}

// latestFrame returns g's current framing — the max-timestamp KindFrame child (the
// sequence's newest entry; ticks are monotonic so newest == current). nil if none.
func latestFrame(g *MemoryCell) *MemoryCell {
	if g == nil {
		return nil
	}
	var best *MemoryCell
	bestTS := -1
	for _, c := range g.Children() {
		mc, ok := c.(*MemoryCell)
		if !ok || mc.Kind() != KindFrame {
			continue
		}
		if gg, ok := mc.Content().(Frame); ok && gg.Timestamp > bestTS {
			bestTS, best = gg.Timestamp, mc
		}
	}
	return best
}

// frameText extracts an episode's CURRENT framing text (its latest frame child).
func frameText(g *MemoryCell) string {
	if frame := latestFrame(g); frame != nil {
		if gg, ok := frame.Content().(Frame); ok {
			return gg.Text
		}
	}
	return ""
}

// FrameText exposes the latest framing text for runtime adapters and diagnostics.
func FrameText(g *MemoryCell) string { return frameText(g) }

// OpenEpisodes returns every episode's id + current frame — the candidates the
// framing op decides edit-or-new against (a cheap frame-level view).
func (store *EpisodeStore) OpenEpisodes() []struct{ ID, Frame string } {
	store.mu.Lock()
	episodes := make([]*MemoryCell, 0, len(store.byID))
	for _, g := range store.byID {
		episodes = append(episodes, g)
	}
	store.mu.Unlock()
	out := make([]struct{ ID, Frame string }, 0, len(episodes))
	for _, g := range episodes {
		out = append(out, struct{ ID, Frame string }{g.ID(), frameText(g)})
	}
	return out
}

// AddChild appends a node (a prediction) to an episode under the index lock.
func (store *EpisodeStore) AddChild(episode *MemoryCell, edgeKey string, child skyra.Space) {
	store.mu.Lock()
	defer store.mu.Unlock()
	episode.AddChild(edgeKey, child)
}

// Snapshot returns an episode's direct children under the index lock — the safe
// read path when final:predict and efficient:predict fan out concurrently (final writes
// the distribution, efficient reads it). Prediction root nodes are immutable once added,
// so their own children are lock-free to read.
func (store *EpisodeStore) Snapshot(episode *MemoryCell) []skyra.Space {
	store.mu.Lock()
	defer store.mu.Unlock()
	out := make([]skyra.Space, 0, len(episode.Children()))
	for _, c := range episode.Children() {
		out = append(out, c)
	}
	return out
}

// predictionsOf walks an episode's frame children and yields every prediction node
// hanging under them — the frame-owned spine's one-level-deeper read (predictions
// attach to the frame they were made under; a bet made under a PRIOR frame is still
// gradeable — evidence can close a bet conditioned on the old framing).
func predictionsOf(g *MemoryCell) []*MemoryCell {
	var out []*MemoryCell
	for _, c := range g.Children() {
		frame, ok := c.(*MemoryCell)
		if !ok || frame.Kind() != KindFrame {
			continue
		}
		for _, gc := range frame.Children() {
			if mc, ok := gc.(*MemoryCell); ok && mc.Kind() == KindPrediction {
				out = append(out, mc)
			}
		}
	}
	return out
}

// LatestUngradedFinal finds the newest ungraded final-pred distribution across all
// episodes — the prior bet final:grade closes against fresh evidence. Returns the
// distribution root and its branch children (sorted by edge key for stable prompt
// numbering). nil if there is no open bet.
func (store *EpisodeStore) LatestUngradedFinal() (*MemoryCell, []*MemoryCell) {
	store.mu.Lock()
	defer store.mu.Unlock()
	var best *MemoryCell
	bestTS := -1
	for _, g := range store.byID {
		for _, cc := range predictionsOf(g) {
			p, ok := cc.Content().(Prediction)
			if !ok || p.Role != "final" || p.Graded || len(cc.Children()) == 0 {
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
	kids := make([]*MemoryCell, 0, len(keys))
	for _, k := range keys {
		kids = append(kids, best.Children()[k].(*MemoryCell))
	}
	return best, kids
}

// StampFinalGrade records the judge's verdict on a distribution under the index
// lock: the root carries the probability mass on the realized branch (the proper
// score, 0 on a total miss); each branch child carries 1 (realized) / 0. matchedIdx
// is 1-based; 0 marks a total miss.
func (store *EpisodeStore) StampFinalGrade(root *MemoryCell, kids []*MemoryCell, matchedIdx int, score float64, evidence string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if p, ok := root.Content().(Prediction); ok {
		p.Graded, p.Grade, p.Evidence = true, score, evidence
		root.SetContent(p)
	}
	for i, kid := range kids {
		p, ok := kid.Content().(Prediction)
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

// LatestUngradedRealizedEfficient finds the newest ungraded efficient-cause proposal that has been
// REALIZED (carries an action child — the intent was actually emitted). Returns the
// proposal root, its prediction content, and the latest action on it. efficient:grade
// closes it against the reafference.
func (store *EpisodeStore) LatestUngradedRealizedEfficient() (*MemoryCell, Prediction, ActionRecord, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	var best *MemoryCell
	var bestPred Prediction
	bestTS := -1
	for _, g := range store.byID {
		for _, cc := range predictionsOf(g) {
			p, ok := cc.Content().(Prediction)
			if !ok || p.Role != "efficient" || p.Graded || len(cc.Children()) == 0 {
				continue
			}
			if p.Timestamp > bestTS {
				bestTS, best, bestPred = p.Timestamp, cc, p
			}
		}
	}
	if best == nil {
		return nil, Prediction{}, ActionRecord{}, false
	}
	var action ActionRecord
	actionTS := -1
	for _, child := range best.Children() {
		cc, ok := child.(*MemoryCell)
		if !ok || cc.Kind() != KindAction {
			continue
		}
		if rec, ok := cc.Content().(ActionRecord); ok && rec.Timestamp > actionTS {
			actionTS, action = rec.Timestamp, rec
		}
	}
	if actionTS < 0 {
		return nil, Prediction{}, ActionRecord{}, false
	}
	return best, bestPred, action, true
}

// StampEfficientGrade records the execution judge's verdict on a realized proposal under
// the index lock (score 1 / 0.5 / 0; evidence = the world result).// gradePrior's write-back.
func (store *EpisodeStore) StampEfficientGrade(proposal *MemoryCell, score float64, evidence string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if p, ok := proposal.Content().(Prediction); ok {
		p.Graded, p.Grade, p.Evidence = true, score, evidence
		proposal.SetContent(p)
	}
}

// All returns every episode under the index lock (inspection / trace assembly).
func (store *EpisodeStore) All() []*MemoryCell {
	store.mu.Lock()
	defer store.mu.Unlock()
	out := make([]*MemoryCell, 0, len(store.byID))
	for _, g := range store.byID {
		out = append(out, g)
	}
	return out
}

// Latest returns the most-recently-framed episode (by its current frame's
// timestamp), for inspection/display. nil if none.
func (store *EpisodeStore) Latest() *MemoryCell {
	store.mu.Lock()
	episodes := make([]*MemoryCell, 0, len(store.byID))
	for _, g := range store.byID {
		episodes = append(episodes, g)
	}
	store.mu.Unlock()
	var best *MemoryCell
	bestTS := -1
	for _, g := range episodes {
		if frame := latestFrame(g); frame != nil {
			if gg, ok := frame.Content().(Frame); ok && gg.Timestamp > bestTS {
				bestTS, best = gg.Timestamp, g
			}
		}
	}
	return best
}
