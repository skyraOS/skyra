package dnav2

import (
	"fmt"
	"strings"
	"sync"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// CausalNode persistence — the immutable side of the minimal Aitia loop
// (minimal-aitia-set.md). A CausalNode is ONE base node (carrying the clean
// sentence) with its subgraph behind it: member term/entity nodes and separate
// edge RECORDS (adjacency is edge-record-local, never a mutable child map —
// the immutability model). Nodes are reused across CausalNodes; edges are always
// fresh and contextual (invariant 4).

// Kinds minted by persistence (structural vocabulary, like KindGist — the
// raw-string rule of 8a applies to affordance-stamped kinds, not these).
const (
	KindCausalNode = dna.Kind("compound") // summary node; Content = CausalNodeContent
	KindTerm     = dna.Kind("term")     // subgraph member; Content = TermContent
)

// CausalNodeContent is the summary node's payload: the clean sentence.
type CausalNodeContent struct {
	Sentence  string
	Timestamp int
}

// TermContent is a member node's payload.
type TermContent struct {
	Label string
}

// Edge kinds on EdgeRecord records.
const (
	EdgeStructural = "structural" // semantic: within-subgraph sentence relationship (E of G)
	EdgeMember     = "member"     // member → summary ("the subgraph points to it")
	EdgeEfficient  = "efficient"  // weighted: CausalNode → CausalNode (Phase 4)
	EdgeFinal      = "final"      // weighted: CausalNode → CausalNode (Phase 4)
)

// EdgeRecord is one immutable edge record. Its mutable learned state lives in
// the SignedStore, keyed by ID — identity, endpoints, and kind never change.
type EdgeRecord struct {
	ID         string
	ContextID string
	FromID     string
	ToID       string
	Label      string
	Kind       string
}

// CausalNode is one persisted query result.
type CausalNode struct {
	Cell    *dna.MemoryCell
	Parents []*dna.MemoryCell
	Edges   []EdgeRecord
}

// AitiaSet is the active explanatory set (minimal-aitia-set.md): one present
// CausalNode + one Efficient and one Final causal edge (both "" until Phase 4 —
// item 5's summary-node-only lean licenses the absent-causal shape). Immutable
// after birth; the ACTIVE set reference lives on the Slice, never here.
type AitiaSet struct {
	ID              string
	PresentID       string // present CausalNode's base-cell ID
	EfficientEdgeID string // "" until the fan-out lands (Phase 4)
	FinalEdgeID     string // ""
	Timestamp       int
}

// CausalNodes — the process-side plumbing: CausalNode cells, their birth
// subgraphs' edge records, and Aitia sets, on the nucleus. (Not a design
// object — the design object is the CausalNode; this is its plural.)
// Resolution is exact-match (the token-web discipline; fuzzy/alias parked in
// the ledger) and delegates material nodes to the knowledge graph's
// NodeRegistry: entities are born from OBSERVATION (sensory/ingest), terms
// from a query's ASSERTION, and a claimed referent whose label matches an
// observed entity resolves to that entity node — one michael, whether
// perceived or mentioned. Stable-id references resolve to existing
// CausalNodes first (composition — invariant 14).
type CausalNodes struct {
	mu    sync.Mutex
	slice skyra.Slice
	nodes *NodeRegistry               // the knowledge graph (material nodes + label lookup)
	byID  map[string]*dna.MemoryCell // CausalNode cells
	edges map[string][]EdgeRecord    // causalNodeID → its birth subgraph's edge records
	sets  map[string]AitiaSet
	seq   int
}

func NewCausalNodes(slice skyra.Slice, nodes *NodeRegistry) *CausalNodes {
	return &CausalNodes{
		slice: slice,
		nodes: nodes,
		byID:  map[string]*dna.MemoryCell{},
		edges: map[string][]EdgeRecord{},
		sets:  map[string]AitiaSet{},
	}
}

// MintSet births an immutable AitiaSet around a present CausalNode. The caller
// owns activation (writes the set ID to the Slice).
func (ci *CausalNodes) MintSet(presentID string, ts int) AitiaSet {
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
func (ci *CausalNodes) Set(id string) (AitiaSet, bool) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	s, ok := ci.sets[id]
	return s, ok
}

// Persist mints the immutable CausalNode for one parsed emission: a fresh base
// cell, resolved-or-created parent nodes, and fresh semantic edge records
// (formal edges between parents + parent→CausalNode member edges — the birth
// frame). Returns the CausalNode; the caller owns activation (Slice) and any
// grading.
func (ci *CausalNodes) Persist(p ParsedSubgraph, ts int) *CausalNode {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	ci.seq++
	baseID := fmt.Sprintf("causal:%d:%d", ts, ci.seq)
	base := dna.NewMemoryCell(baseID, KindCausalNode, CausalNodeContent{Sentence: p.Sentence, Timestamp: ts}, ci.slice)
	ci.byID[baseID] = base

	nodeByOrdinal := map[int]*dna.MemoryCell{}
	var parents []*dna.MemoryCell
	for _, n := range p.Nodes {
		cell := ci.resolveLocked(n)
		nodeByOrdinal[n.Ordinal] = cell
		parents = append(parents, cell)
	}

	var edges []EdgeRecord
	for i, e := range p.Edges {
		edges = append(edges, EdgeRecord{
			ID:        fmt.Sprintf("%s/e%d", baseID, i),
			ContextID: baseID,
			FromID:    nodeByOrdinal[e.From].ID(),
			ToID:      nodeByOrdinal[e.To].ID(),
			Label:     e.Label,
			Kind:      EdgeStructural,
		})
	}
	for i, m := range parents {
		edges = append(edges, EdgeRecord{
			ID:        fmt.Sprintf("%s/m%d", baseID, i),
			ContextID: baseID,
			FromID:    m.ID(),
			ToID:      baseID,
			Kind:      EdgeMember,
		})
	}
	ci.edges[baseID] = edges

	return &CausalNode{Cell: base, Parents: parents, Edges: edges}
}

// resolveLocked resolves one parsed node to its immutable cell (caller holds
// ci.mu) — create-if-absent, reuse-if-present. CausalNode references resolve
// here (composition); material nodes delegate to the knowledge graph's
// registry.
//
// Stable id:  CausalNode (composition) → entity (accepting an "entity:"
//             prefix) → term by id → create a term carrying that id.
// Label:      observed entity by exact label → term lookup → create a term.
//
// A mentioned-but-never-observed referent stays a TERM; entities are only ever
// born from observation. (Merging a term into a later-observed entity is the
// deferred merge op — token-web doc.)
func (ci *CausalNodes) resolveLocked(n ParsedNode) *dna.MemoryCell {
	if n.StableID != "" {
		if c, ok := ci.byID[n.StableID]; ok {
			return c // a previous CausalNode as referent — composition, free
		}
		id := strings.TrimPrefix(n.StableID, "entity:")
		if e := ci.nodes.Get(id); e != nil {
			return e
		}
		return ci.nodes.EnsureTerm(n.StableID, n.StableID, n.Label)
	}
	if e := ci.nodes.Get(n.Label); e != nil {
		return e // the observed michael IS the mentioned michael
	}
	return ci.nodes.EnsureTerm(n.Label, "term:"+n.Label, n.Label)
}

// Get returns a CausalNode's base cell by id (nil if none).
func (ci *CausalNodes) Get(id string) *dna.MemoryCell {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return ci.byID[id]
}

// EdgesOf returns a CausalNode's birth-frame edge records (copy).
func (ci *CausalNodes) EdgesOf(causalNodeID string) []EdgeRecord {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	out := make([]EdgeRecord, len(ci.edges[causalNodeID]))
	copy(out, ci.edges[causalNodeID])
	return out
}
