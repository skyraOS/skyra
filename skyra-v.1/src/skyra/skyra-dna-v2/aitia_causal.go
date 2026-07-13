package dnav2

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"skyra-v1/src/skyra"
)

// CausalNode persistence — the immutable side of the minimal Aitia loop
// (minimal-aitia-set.md). A CausalNode is ONE base node (carrying the clean
// sentence) with its subgraph behind it: member term/entity nodes and separate
// edge RECORDS (adjacency is edge-record-local, never a mutable child map —
// the immutability model). Nodes are reused across CausalNodes; edges are always
// fresh and contextual (invariant 4).

// Kinds minted by persistence (structural vocabulary, like KindFrame — the
// raw-string rule of 8a applies to affordance-stamped kinds, not these).
const (
	KindCausalNode   = Kind("causal-node")   // summary node; Content = CausalNodeContent
	KindTerm         = Kind("term")          // semantic node born from assertion
	KindSemanticEdge = Kind("semantic-edge") // immutable birth-subgraph edge
	KindCausalEdge   = Kind("causal-edge")   // weighted contextual edge
)

// CausalNodeContent is the summary node's payload: the clean sentence.
type CausalNodeContent struct {
	Sentence   string
	Timestamp  int
	Provenance Provenance
}

// TermContent is a member node's payload.
type TermContent struct {
	Label      string
	Provenance Provenance
}

// Edge kinds on EdgeRecord records.
const (
	EdgeStructural = "structural" // semantic: within-subgraph sentence relationship (E of G)
	EdgeInherited  = "inherited"  // weighted: composed parent → new CausalNode
	EdgeEfficient  = "efficient"  // weighted: prior CausalNode → present CausalNode
	EdgeFinal      = "final"      // weighted: present CausalNode → future CausalNode
)

// Typed child ports are substrate vocabulary. Observe strategies select these
// surfaces; the generic kernel never switches on their names.
const (
	PortMaterial  = "material"
	PortFormal    = "formal"
	PortFrom      = "from"
	PortTo        = "to"
	PortSummaries = "summaries"
	PortInherited = "inherited"
	PortEfficient = "efficient"
	PortFinal     = "final"
	PortEvidence  = "evidence"
)

// EdgeRecord is one immutable edge record. Its mutable learned state lives in
// the SignedStore, keyed by ID — identity, endpoints, and kind never change.
type EdgeRecord struct {
	ID        string
	ContextID string
	FromID    string
	ToID      string
	Label     string
	Kind      string
}

// EdgeCell makes an immutable edge occurrence traversable. Topology lives in
// Record and is mirrored by the from/to child ports; learned causal activation
// remains in the graph's SignedStore under the edge ID.
type EdgeCell struct {
	*MemoryCell
	Record EdgeRecord
	signed *SignedStore
}

// Strength returns the causal edge's signed learned state. Semantic edges are
// deliberately unweighted and always return ok=false.
func (e *EdgeCell) Strength(tick int) (strength float64, ok bool) {
	if e == nil || e.Record.Kind == EdgeStructural || e.signed == nil {
		return 0, false
	}
	return e.signed.Strength(e.Record.ID, tick)
}

// CausalNode is one persisted query result.
type CausalNode struct {
	Cell    *MemoryCell
	Parents []*MemoryCell
	Edges   []EdgeRecord
}

// SubgraphQuery is a resolved query-by-subgraph request. IDs refer to durable
// semantic nodes; Edges preserve complete directed semantic signatures.
type SubgraphQuery struct {
	NodeIDs []string
	Edges   []SemanticTriple
}

// CausalCandidate separates retrieval weight from semantic verification.
// Occurrences may seed amplitude; Exact alone says the complete requested
// directed structure is present.
type CausalCandidate struct {
	CausalNodeID string
	Occurrences  int
	Exact        bool
}

// AitiaSet is the active explanatory set (minimal-aitia-set.md): one present
// CausalNode + one Efficient and one Final causal edge (both "" until Phase 4 —
// item 5's summary-node-only lean licenses the absent-causal shape). Immutable
// after birth; the ACTIVE set reference lives on the Slice, never here.
type AitiaSet struct {
	ID              string
	PresentID       string // present CausalNode's base-cell ID
	EfficientEdgeID string // set by MintFullSet; "" only on legacy pre-expansion sets
	FinalEdgeID     string // set by MintFullSet; "" only on legacy pre-expansion sets
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
	mu          sync.Mutex
	slice       skyra.Slice
	nodes       *NodeRegistry // knowledge graph + semantic reverse indexes
	signed      *SignedStore
	byID        map[string]*MemoryCell  // CausalNode cells
	edges       map[string][]EdgeRecord // CausalNodeID → owned edge records
	edgeByID    map[string]*EdgeCell
	materials   map[string][]string          // CausalNodeID → immutable birth-node IDs
	assessments map[string][]CauseAssessment // observationID → grader outputs
	sets        map[string]AitiaSet
	seq         int
}

func NewCausalNodes(slice skyra.Slice, nodes *NodeRegistry, stores ...*SignedStore) *CausalNodes {
	signed := NewSignedStore()
	if len(stores) > 0 && stores[0] != nil {
		signed = stores[0]
	}
	return &CausalNodes{
		slice:       slice,
		nodes:       nodes,
		signed:      signed,
		byID:        map[string]*MemoryCell{},
		edges:       map[string][]EdgeRecord{},
		edgeByID:    map[string]*EdgeCell{},
		materials:   map[string][]string{},
		assessments: map[string][]CauseAssessment{},
		sets:        map[string]AitiaSet{},
	}
}

// AddAssessments records independent active-cause grader outputs. Phase 4
// consumes these to distribute learning; recording one does not mutate an EMA.
func (ci *CausalNodes) AddAssessments(observationID string, assessments ...CauseAssessment) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.assessments[observationID] = append(ci.assessments[observationID], assessments...)
}

func (ci *CausalNodes) AssessmentsOf(observationID string) []CauseAssessment {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	out := make([]CauseAssessment, len(ci.assessments[observationID]))
	copy(out, ci.assessments[observationID])
	return out
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

// MintFullSet atomically publishes one complete explanatory context. All three
// CausalNodes must already exist; the two contextual belief edges are always
// fresh and therefore inherit neither signed strength nor edge-keyed budget.
func (ci *CausalNodes) MintFullSet(presentID, efficientID, finalID string, ts int) (AitiaSet, error) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	present, efficient, final := ci.byID[presentID], ci.byID[efficientID], ci.byID[finalID]
	if present == nil || efficient == nil || final == nil {
		return AitiaSet{}, fmt.Errorf("full AitiaSet requires known present/efficient/final nodes")
	}
	ci.seq++
	setID := fmt.Sprintf("aitia:%d:%d", ts, ci.seq)
	efficientRecord := EdgeRecord{
		ID: fmt.Sprintf("%s/efficient", setID), ContextID: setID,
		FromID: efficientID, ToID: presentID, Kind: EdgeEfficient,
	}
	finalRecord := EdgeRecord{
		ID: fmt.Sprintf("%s/final", setID), ContextID: setID,
		FromID: presentID, ToID: finalID, Kind: EdgeFinal,
	}
	efficientEdge := ci.publishEdgeLocked(efficientRecord, efficient, present)
	finalEdge := ci.publishEdgeLocked(finalRecord, present, final)
	present.AddChildTo(PortEfficient, efficientRecord.ID, efficientEdge)
	present.AddChildTo(PortFinal, finalRecord.ID, finalEdge)
	ci.edges[presentID] = append(ci.edges[presentID], efficientRecord, finalRecord)
	set := AitiaSet{
		ID: setID, PresentID: presentID, EfficientEdgeID: efficientRecord.ID,
		FinalEdgeID: finalRecord.ID, Timestamp: ts,
	}
	ci.sets[set.ID] = set
	return set, nil
}

// Set returns an AitiaSet by ID.
func (ci *CausalNodes) Set(id string) (AitiaSet, bool) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	s, ok := ci.sets[id]
	return s, ok
}

// Persist mints the immutable CausalNode for one parsed emission: a fresh base
// cell, resolved-or-created material nodes, traversable semantic edge cells,
// reverse retrieval projections, and fresh inherited edges for CausalNode
// members. The caller owns active-set activation and later grading.
func (ci *CausalNodes) Persist(p ParsedSubgraph, ts int) *CausalNode {
	return ci.persist(p, ts, ProvenanceComposed)
}

// PersistObserved births the same CausalNode shape with observation provenance.
// Observation is epistemic content, not a distinct node kind.
func (ci *CausalNodes) PersistObserved(p ParsedSubgraph, ts int) *CausalNode {
	return ci.persist(p, ts, ProvenanceObserved)
}

func (ci *CausalNodes) persist(p ParsedSubgraph, ts int, provenance Provenance) *CausalNode {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	ci.seq++
	baseID := fmt.Sprintf("causal:%d:%d", ts, ci.seq)
	base := NewMemoryCell(baseID, KindCausalNode, CausalNodeContent{
		Sentence: p.Sentence, Timestamp: ts, Provenance: provenance,
	}, ci.slice)
	ci.byID[baseID] = base

	nodeByOrdinal := map[int]*MemoryCell{}
	var parents []*MemoryCell
	for _, n := range p.Nodes {
		cell := ci.resolveLocked(n)
		nodeByOrdinal[n.Ordinal] = cell
		parents = append(parents, cell)
	}

	var edges []EdgeRecord
	for i, e := range p.Edges {
		record := EdgeRecord{
			ID:        fmt.Sprintf("%s/e%d", baseID, i),
			ContextID: baseID,
			FromID:    nodeByOrdinal[e.From].ID(),
			ToID:      nodeByOrdinal[e.To].ID(),
			Label:     e.Label,
			Kind:      EdgeStructural,
		}
		edge := ci.publishEdgeLocked(record, nodeByOrdinal[e.From], nodeByOrdinal[e.To])
		base.AddChildTo(PortFormal, record.ID, edge)
		edges = append(edges, record)
	}
	for i, parent := range parents {
		base.AddChildTo(PortMaterial, parent.ID(), parent)
		if parent.Kind() != KindCausalNode {
			continue
		}
		record := EdgeRecord{
			ID:        fmt.Sprintf("%s/i%d", baseID, i),
			ContextID: baseID,
			FromID:    parent.ID(),
			ToID:      baseID,
			Kind:      EdgeInherited,
		}
		edge := ci.publishEdgeLocked(record, parent, base)
		base.AddChildTo(PortInherited, record.ID, edge)
		edges = append(edges, record)
	}
	ci.edges[baseID] = edges
	materialIDs := make([]string, 0, len(parents))
	for _, parent := range parents {
		materialIDs = append(materialIDs, parent.ID())
	}
	ci.materials[baseID] = materialIDs
	ci.nodes.IndexCausalNode(base, parents, edges)

	return &CausalNode{Cell: base, Parents: parents, Edges: edges}
}

// publishEdgeLocked constructs endpoint fields and child ports together before
// publishing the edge into the graph-owned registry. Caller holds ci.mu.
func (ci *CausalNodes) publishEdgeLocked(record EdgeRecord, from, to *MemoryCell) *EdgeCell {
	if from == nil || to == nil || record.FromID != from.ID() || record.ToID != to.ID() {
		panic("dnav2: edge endpoints disagree with immutable topology")
	}
	kind := KindCausalEdge
	if record.Kind == EdgeStructural {
		kind = KindSemanticEdge
	}
	cell := NewMemoryCell(record.ID, kind, record, ci.slice)
	cell.AddChildTo(PortFrom, from.ID(), from)
	cell.AddChildTo(PortTo, to.ID(), to)
	edge := &EdgeCell{MemoryCell: cell, Record: record, signed: ci.signed}
	ci.edgeByID[record.ID] = edge
	return edge
}

// MintCausalEdge creates one fresh contextual weighted relationship and makes
// it reachable from the perspective-bearing endpoint. It never reuses an edge
// or learned activation merely because its endpoint nodes were reused.
func (ci *CausalNodes) MintCausalEdge(kind, contextID, fromID, toID string) (*EdgeCell, error) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	if kind != EdgeInherited && kind != EdgeEfficient && kind != EdgeFinal {
		return nil, fmt.Errorf("causal edge kind %q is not weighted", kind)
	}
	from, to := ci.byID[fromID], ci.byID[toID]
	if from == nil || to == nil {
		return nil, fmt.Errorf("causal edge endpoints must be CausalNodes: %q -> %q", fromID, toID)
	}
	ci.seq++
	record := EdgeRecord{
		ID:        fmt.Sprintf("edge:%s:%d", kind, ci.seq),
		ContextID: contextID,
		FromID:    fromID,
		ToID:      toID,
		Kind:      kind,
	}
	edge := ci.publishEdgeLocked(record, from, to)
	switch kind {
	case EdgeInherited:
		to.AddChildTo(PortInherited, record.ID, edge)
		ci.edges[toID] = append(ci.edges[toID], record)
	case EdgeEfficient:
		to.AddChildTo(PortEfficient, record.ID, edge)
		ci.edges[toID] = append(ci.edges[toID], record)
	case EdgeFinal:
		from.AddChildTo(PortFinal, record.ID, edge)
		ci.edges[fromID] = append(ci.edges[fromID], record)
	}
	return edge, nil
}

// AttachEvidence adds an observation-born CausalNode to an edge's mutable
// provenance projection. It does not alter the edge's immutable endpoints or
// the target CausalNode's semantic birth subgraph.
func (ci *CausalNodes) AttachEvidence(edgeID, observationID string) error {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	edge := ci.edgeByID[edgeID]
	observation := ci.byID[observationID]
	if edge == nil || observation == nil {
		return fmt.Errorf("evidence attachment requires known edge and CausalNode: %q, %q", edgeID, observationID)
	}
	content, ok := observation.Content().(CausalNodeContent)
	if !ok || content.Provenance != ProvenanceObserved {
		return fmt.Errorf("CausalNode %q is not observation-born", observationID)
	}
	edge.AddChildTo(PortEvidence, observationID, observation)
	return nil
}

// resolveLocked resolves one parsed node to its immutable cell (caller holds
// ci.mu) — create-if-absent, reuse-if-present. CausalNode references resolve
// here (composition); material nodes delegate to the knowledge graph's
// registry.
//
// Stable id:  CausalNode (composition) → entity (accepting an "entity:"
//
//	prefix) → term by id → create a term carrying that id.
//
// Label:      observed entity by exact label → term lookup → create a term.
//
// A mentioned-but-never-observed referent stays a TERM; entities are only ever
// born from observation. (Merging a term into a later-observed entity is the
// deferred merge op — token-web doc.)
func (ci *CausalNodes) resolveLocked(n ParsedNode) *MemoryCell {
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
func (ci *CausalNodes) Get(id string) *MemoryCell {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return ci.byID[id]
}

// Edge returns a traversable edge cell by immutable edge identity.
func (ci *CausalNodes) Edge(id string) *EdgeCell {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return ci.edgeByID[id]
}

// Candidates retrieves by reverse-index occurrence and then verifies the
// requested node and directed-edge sets against immutable birth content.
func (ci *CausalNodes) Candidates(query SubgraphQuery) []CausalCandidate {
	scores := ci.nodes.CandidateOccurrences(query.NodeIDs, query.Edges)
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ids := make([]string, 0, len(scores))
	for id := range scores {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]CausalCandidate, 0, len(ids))
	for _, id := range ids {
		out = append(out, CausalCandidate{
			CausalNodeID: id,
			Occurrences:  scores[id],
			Exact:        ci.matchesLocked(id, query),
		})
	}
	return out
}

// QueryOf reconstructs the immutable resolved query represented by one
// CausalNode's semantic birth subgraph.
func (ci *CausalNodes) QueryOf(id string) (SubgraphQuery, bool) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	if ci.byID[id] == nil {
		return SubgraphQuery{}, false
	}
	query := SubgraphQuery{NodeIDs: append([]string(nil), ci.materials[id]...)}
	for _, edge := range ci.edges[id] {
		if edge.Kind == EdgeStructural {
			query.Edges = append(query.Edges, SemanticTriple{FromID: edge.FromID, Label: edge.Label, ToID: edge.ToID})
		}
	}
	return query, true
}

func (ci *CausalNodes) matchesLocked(id string, query SubgraphQuery) bool {
	nodes := map[string]bool{}
	for _, nodeID := range ci.materials[id] {
		nodes[nodeID] = true
	}
	for _, nodeID := range query.NodeIDs {
		if !nodes[nodeID] {
			return false
		}
	}
	edges := map[SemanticTriple]bool{}
	for _, edge := range ci.edges[id] {
		if edge.Kind == EdgeStructural {
			edges[SemanticTriple{FromID: edge.FromID, Label: edge.Label, ToID: edge.ToID}] = true
		}
	}
	for _, edge := range query.Edges {
		if !edges[edge] {
			return false
		}
	}
	return true
}

// EdgesOf returns a CausalNode's birth-frame edge records (copy).
func (ci *CausalNodes) EdgesOf(causalNodeID string) []EdgeRecord {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	out := make([]EdgeRecord, len(ci.edges[causalNodeID]))
	copy(out, ci.edges[causalNodeID])
	return out
}
