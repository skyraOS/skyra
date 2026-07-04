package dnav2

import "sync"

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

	// The master graph (the nucleus's durable working memory).
	entities  *EntityIndex
	groupings *GroupingIndex

	// infer is the injected LLM seam (system, present) → text. Genes that think
	// call it; nil → those genes skip (tests without inference).
	infer InferFunc

	mu    sync.Mutex
	clock int // monotonic proper-time counter for node timestamps
}

// InferFunc is the LLM seam: a system prompt + the present (frame) → the model's
// text. Injected so dnav2 stays free of the inference package.
type InferFunc func(system, present string) (string, error)

func NewDNA(beingID string) *DNA {
	return &DNA{
		beingID:   beingID,
		observe:   map[string]ObserveCompetence{},
		express:   map[string]ExpressCompetence{},
		entities:  NewEntityIndex(nil), // nil nucleus slice until activation lands
		groupings: NewGroupingIndex(nil),
	}
}

func (d *DNA) BeingID() string           { return d.beingID }
func (d *DNA) Entities() *EntityIndex    { return d.entities }
func (d *DNA) Groupings() *GroupingIndex { return d.groupings }
func (d *DNA) Infer() InferFunc          { return d.infer }
func (d *DNA) SetInfer(fn InferFunc)     { d.infer = fn }

// Tick advances the being's proper-time clock (node timestamps / ordering).
func (d *DNA) Tick() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.clock++
	return d.clock
}

// RegisterObserve / RegisterExpress wire a gene at boot. Keyed "namespace:gene".
func (d *DNA) RegisterObserve(key string, fn ObserveCompetence) { d.observe[key] = fn }
func (d *DNA) RegisterExpress(key string, fn ExpressCompetence) { d.express[key] = fn }

// observeGene / expressGene look up a registered gene; nil if unregistered (a nil
// gene is a no-op bookend, so an unwired step is inert, not a panic).
func (d *DNA) observeGene(key string) ObserveCompetence { return d.observe[key] }
func (d *DNA) expressGene(key string) ExpressCompetence { return d.express[key] }
