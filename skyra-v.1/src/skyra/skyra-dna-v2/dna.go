package dnav2

import (
	"sync"

	"skyra-v1/src/skyra"
)

// DNA — the one canonical base cell (the genome). Holds ALL durable indexes (the
// long-term graph) AND all competences (namespaced genes, keyed "namespace:gene"
// e.g. "other:grade"). The v1 dispatcher cells collapse into this: sensory /
// midlevel / other / self / motor are competence NAMESPACES, not separate objects
// — expression contexts of the one genome.
//
// Competences are registered once at boot (read-only during a wave → concurrent
// transcription is safe). Durable indexes are guarded by the DNA's own fine-grained
// locks (the substrate model), NOT held across inference.
type DNA struct {
	beingID string

	observe map[string]ObserveCompetence // "namespace:gene" → gene
	express map[string]ExpressCompetence

	// nucleus is the being's one slice — the substrate the index nodes live on and
	// the home of the active state (activeGroupingID/activeGistID via Slice.Maps()).
	// It carries no kernel waves; it is index substrate only.
	nucleus skyra.Slice

	// The knowledge graph's material nodes (the nucleus's durable working memory).
	nodes     *NodeRegistry
	groupings *GroupingIndex

	// activation holds edge-owned learning (A/S/F/N) keyed by edge identity — the
	// economy the traversal reads and the bump maintains.
	activation *ActivationStore

	// The minimal-Aitia loop's stores (cause-and-effect, Phase 0/1): immutable
	// CausalNodes + Aitia sets, and the signed edge EMAs. Beside the legacy
	// grouping/activation machinery, not a mutation of it.
	causal *CausalNodes
	signed *SignedStore

	// infer is the injected LLM seam (system, present) → text. Genes that think
	// call it; nil → those genes skip (tests without inference).
	infer InferFunc
	// emit delivers motor's concrete action to the world (injected after the world
	// exists). nil → motor collapses but does not emit.
	emit func(skyra.Impulse)

	mu         sync.Mutex
	clock      int    // monotonic proper-time counter for node timestamps
	lastAction string // latest emitted action, for CLI/validate display
}

// InferFunc is the LLM seam: a system prompt + the present (frame) → the model's
// text. Injected so dnav2 stays free of the inference package.
type InferFunc func(system, present string) (string, error)

func NewDNA(beingID string, nucleus skyra.Slice) *DNA {
	nodes := NewNodeRegistry(nucleus)
	return &DNA{
		beingID:    beingID,
		observe:    map[string]ObserveCompetence{},
		express:    map[string]ExpressCompetence{},
		nucleus:    nucleus,
		nodes:      nodes,
		groupings:  NewGroupingIndex(nucleus),
		activation: NewActivationStore(),
		causal:     NewCausalNodes(nucleus, nodes), // mentioned michael resolves to observed michael
		signed:     NewSignedStore(),
	}
}

func (d *DNA) BeingID() string           { return d.beingID }
func (d *DNA) Nucleus() skyra.Slice      { return d.nucleus }
func (d *DNA) Nodes() *NodeRegistry       { return d.nodes }
func (d *DNA) Groupings() *GroupingIndex { return d.groupings }
func (d *DNA) Infer() InferFunc          { return d.infer }
func (d *DNA) SetInfer(fn InferFunc)     { d.infer = fn }
func (d *DNA) Activation() *ActivationStore { return d.activation }
func (d *DNA) CausalNodes() *CausalNodes    { return d.causal }
func (d *DNA) Signed() *SignedStore         { return d.signed }

// ActiveSet returns the being's active AitiaSet (ok=false before the first
// query commits). V0 anchors on the being's own ID (ledger item 7).
func (d *DNA) ActiveSet() (AitiaSet, bool) {
	id := activeSetGet(d.nucleus, d.beingID)
	if id == "" {
		return AitiaSet{}, false
	}
	return d.causal.Set(id)
}

// ActivateSet commits a set as the being's active situation — atomic
// replacement, never mutation of the prior set.
func (d *DNA) ActivateSet(setID string) {
	activeSetSet(d.nucleus, d.beingID, setID)
}

// CrossBirth crosses freshly-created edges (node birth) so the new nodes are immediately
// selectable — a born edge was just active, so F must start > 0 or the edge deadlocks at
// strength 0 and can never be re-selected to earn its alignment.
func (d *DNA) CrossBirth(keys ...string) {
	d.mu.Lock()
	tick := d.clock
	d.mu.Unlock()
	for _, k := range keys {
		d.activation.Cross(k, tick)
	}
}

// Tick advances the being's proper-time clock (node timestamps / ordering).
func (d *DNA) Tick() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.clock++
	return d.clock
}

// Clock reads the proper-time clock without advancing it (decay-read time).
func (d *DNA) Clock() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.clock
}

func (d *DNA) Emit() func(skyra.Impulse)    { return d.emit }
func (d *DNA) SetEmit(fn func(skyra.Impulse)) { d.emit = fn }

func (d *DNA) SetLastAction(s string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lastAction = s
}

func (d *DNA) LastAction() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.lastAction
}

// RegisterObserve / RegisterExpress wire a gene at boot. Keyed "namespace:gene".
func (d *DNA) RegisterObserve(key string, fn ObserveCompetence) { d.observe[key] = fn }
func (d *DNA) RegisterExpress(key string, fn ExpressCompetence) { d.express[key] = fn }

// observeGene / expressGene look up a registered gene; nil if unregistered (a nil
// gene is a no-op bookend, so an unwired step is inert, not a panic).
func (d *DNA) observeGene(key string) ObserveCompetence { return d.observe[key] }
func (d *DNA) expressGene(key string) ExpressCompetence { return d.express[key] }
