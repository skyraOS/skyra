package dnav2

import (
	"fmt"
	"strings"
	"sync"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// Compound persistence — the immutable side of the minimal Aitia loop
// (minimal-aitia-set.md). A compound is ONE summary node (carrying the clean
// sentence) with its subgraph behind it: member term/entity nodes and separate
// edge RECORDS (adjacency is edge-record-local, never a mutable child map —
// the immutability model). Nodes are reused across compounds; edges are always
// fresh and contextual (invariant 4).

// Kinds minted by persistence (structural vocabulary, like KindGist — the
// raw-string rule of 8a applies to affordance-stamped kinds, not these).
const (
	KindCompound = dna.Kind("compound") // summary node; Content = CompoundContent
	KindTerm     = dna.Kind("term")     // subgraph member; Content = TermContent
)

// CompoundContent is the summary node's payload: the clean sentence.
type CompoundContent struct {
	Sentence  string
	Timestamp int
}

// TermContent is a member node's payload.
type TermContent struct {
	Label string
}

// Edge kinds on CausalEdge records.
const (
	EdgeStructural = "structural" // within-compound sentence relationship (E of G)
	EdgeMember     = "member"     // member → summary ("the subgraph points to it")
	EdgeEfficient  = "efficient"  // compound-summary → compound-summary (Phase 4)
	EdgeFinal      = "final"      // compound-summary → compound-summary (Phase 4)
)

// CausalEdge is one immutable edge record. Its mutable learned state lives in
// the SignedStore, keyed by ID — identity, endpoints, and kind never change.
type CausalEdge struct {
	ID         string
	CompoundID string
	FromID     string
	ToID       string
	Label      string
	Kind       string
}

// Compound is one persisted query result.
type Compound struct {
	Summary *dna.MemoryCell
	Members []*dna.MemoryCell
	Edges   []CausalEdge
}

// AitiaSet is the active explanatory set (minimal-aitia-set.md): one present
// compound + one Efficient and one Final causal edge (both "" until Phase 4 —
// item 5's summary-node-only lean licenses the absent-causal shape). Immutable
// after birth; the ACTIVE set reference lives on the Slice, never here.
type AitiaSet struct {
	ID              string
	PresentID       string // present compound's summary-node ID
	EfficientEdgeID string // "" until the fan-out lands (Phase 4)
	FinalEdgeID     string // ""
	Timestamp       int
}

// CompoundIndex — compounds + the term registry + edge records + Aitia sets, on
// the nucleus. Resolution is exact-match (the token-web discipline; fuzzy/alias
// parked in the ledger) and UNIFIED with the entity graph: entities are born
// from OBSERVATION (sensory/ingest), terms from a query's ASSERTION, and a
// claimed referent whose label matches an observed entity resolves to that
// entity node — one michael, whether perceived or mentioned. Stable-id
// references resolve to existing compounds first (composition: summary nodes
// are legal referents — invariant 14).
type CompoundIndex struct {
	mu       sync.Mutex
	slice    skyra.Slice
	entities *EntityIndex               // the observed graph (nil-tolerant for tests)
	byID     map[string]*dna.MemoryCell // compound summary nodes
	terms    map[string]*dna.MemoryCell // normalized label → term node
	edges    map[string][]CausalEdge    // compoundID → its edge records
	sets     map[string]AitiaSet
	seq      int
}

func NewCompoundIndex(slice skyra.Slice) *CompoundIndex {
	return &CompoundIndex{
		slice: slice,
		byID:  map[string]*dna.MemoryCell{},
		terms: map[string]*dna.MemoryCell{},
		edges: map[string][]CausalEdge{},
		sets:  map[string]AitiaSet{},
	}
}

// WireEntities connects the observed entity graph for resolution (set once at
// DNA construction).
func (ci *CompoundIndex) WireEntities(ei *EntityIndex) { ci.entities = ei }

// MintSet births an immutable AitiaSet around a present compound. The caller
// owns activation (writes the set ID to the Slice).
func (ci *CompoundIndex) MintSet(presentID string, ts int) AitiaSet {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.seq++
	set := AitiaSet{
		ID:        fmt.Sprintf("aitia:%d:%d", ts, ci.seq),
		PresentID: presentID,
		Timestamp: ts,
	}
	ci.sets[set.ID] = set
	return set
}

// Set returns an AitiaSet by ID.
func (ci *CompoundIndex) Set(id string) (AitiaSet, bool) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	s, ok := ci.sets[id]
	return s, ok
}

// Persist mints the immutable compound for one parsed emission: a fresh summary
// node, resolved-or-created member nodes, and fresh contextual edge records
// (structural edges between members + member→summary edges). Returns the
// compound; the caller owns activation (Slice) and any grading.
func (ci *CompoundIndex) Persist(p ParsedCompound, ts int) *Compound {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	ci.seq++
	summaryID := fmt.Sprintf("compound:%d:%d", ts, ci.seq)
	summary := dna.NewMemoryCell(summaryID, KindCompound, CompoundContent{Sentence: p.Sentence, Timestamp: ts}, ci.slice)
	ci.byID[summaryID] = summary

	nodeByOrdinal := map[int]*dna.MemoryCell{}
	var members []*dna.MemoryCell
	for _, n := range p.Nodes {
		cell := ci.resolveLocked(n)
		nodeByOrdinal[n.Ordinal] = cell
		members = append(members, cell)
	}

	var edges []CausalEdge
	for i, e := range p.Edges {
		edges = append(edges, CausalEdge{
			ID:         fmt.Sprintf("%s/e%d", summaryID, i),
			CompoundID: summaryID,
			FromID:     nodeByOrdinal[e.From].ID(),
			ToID:       nodeByOrdinal[e.To].ID(),
			Label:      e.Label,
			Kind:       EdgeStructural,
		})
	}
	for i, m := range members {
		edges = append(edges, CausalEdge{
			ID:         fmt.Sprintf("%s/m%d", summaryID, i),
			CompoundID: summaryID,
			FromID:     m.ID(),
			ToID:       summaryID,
			Kind:       EdgeMember,
		})
	}
	ci.edges[summaryID] = edges

	return &Compound{Summary: summary, Members: members, Edges: edges}
}

// resolveLocked resolves one parsed node to its immutable cell (caller holds
// ci.mu) — create-if-absent, reuse-if-present, unified with the observed graph.
//
// Stable id:  compound (composition) → entity (accepting an "entity:" prefix)
//             → term by id → create a term carrying that id.
// Label:      observed entity by exact label → term registry → create a term.
//
// A mentioned-but-never-observed referent stays a TERM; entities are only ever
// born from observation. (Merging a term into a later-observed entity is the
// deferred merge op — token-web doc.)
func (ci *CompoundIndex) resolveLocked(n ParsedNode) *dna.MemoryCell {
	if n.StableID != "" {
		if c, ok := ci.byID[n.StableID]; ok {
			return c // a previous compound as referent — composition, free
		}
		if ci.entities != nil {
			id := strings.TrimPrefix(n.StableID, "entity:")
			if e := ci.entities.Get(id); e != nil {
				return e
			}
		}
		if t, ok := ci.terms[n.StableID]; ok {
			return t
		}
		cell := dna.NewMemoryCell(n.StableID, KindTerm, TermContent{Label: n.Label}, ci.slice)
		ci.terms[n.StableID] = cell
		return cell
	}
	if ci.entities != nil {
		if e := ci.entities.Get(n.Label); e != nil {
			return e // the observed michael IS the mentioned michael
		}
	}
	if t, ok := ci.terms[n.Label]; ok {
		return t
	}
	cell := dna.NewMemoryCell("term:"+n.Label, KindTerm, TermContent{Label: n.Label}, ci.slice)
	ci.terms[n.Label] = cell
	return cell
}

// Get returns a compound's summary node by id (nil if none).
func (ci *CompoundIndex) Get(id string) *dna.MemoryCell {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return ci.byID[id]
}

// EdgesOf returns a compound's edge records (copy).
func (ci *CompoundIndex) EdgesOf(compoundID string) []CausalEdge {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	out := make([]CausalEdge, len(ci.edges[compoundID]))
	copy(out, ci.edges[compoundID])
	return out
}

// Term returns the shared node for a normalized label (nil if none) — the
// reuse check.
func (ci *CompoundIndex) Term(label string) *dna.MemoryCell {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return ci.terms[label]
}
