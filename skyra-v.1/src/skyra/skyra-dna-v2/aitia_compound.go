package dnav2

import (
	"fmt"
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

// CompoundIndex — compounds + the term registry + edge records, on the nucleus.
// Term reuse is normalized-label exact match (the token-web discipline);
// stable-id references resolve to existing compounds first (composition:
// summary nodes are legal referents — invariant 14).
type CompoundIndex struct {
	mu    sync.Mutex
	slice skyra.Slice
	byID  map[string]*dna.MemoryCell // compound summary nodes
	terms map[string]*dna.MemoryCell // normalized label → term node
	edges map[string][]CausalEdge    // compoundID → its edge records
	seq   int
}

func NewCompoundIndex(slice skyra.Slice) *CompoundIndex {
	return &CompoundIndex{
		slice: slice,
		byID:  map[string]*dna.MemoryCell{},
		terms: map[string]*dna.MemoryCell{},
		edges: map[string][]CausalEdge{},
	}
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
// ci.mu). Resolution order: explicit stable id → existing compound (composition)
// → term registry by id → create with the stable id; no stable id → term
// registry by normalized label → create.
func (ci *CompoundIndex) resolveLocked(n ParsedNode) *dna.MemoryCell {
	if n.StableID != "" {
		if c, ok := ci.byID[n.StableID]; ok {
			return c // a previous compound as referent — composition, free
		}
		if t, ok := ci.terms[n.StableID]; ok {
			return t
		}
		cell := dna.NewMemoryCell(n.StableID, KindTerm, TermContent{Label: n.Label}, ci.slice)
		ci.terms[n.StableID] = cell
		return cell
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
