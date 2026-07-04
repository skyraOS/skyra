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

	// The master graph (the nucleus's durable working memory). The entity index is
	// the first ported; the grouping spine + action index follow.
	entities *EntityIndex

	// mu guards misc scratch (stress, etc.); the entity index has its own lock.
	mu      sync.Mutex
	indexes map[string]any
}

func NewDNA(beingID string) *DNA {
	return &DNA{
		beingID:  beingID,
		observe:  map[string]ObserveCompetence{},
		express:  map[string]ExpressCompetence{},
		entities: NewEntityIndex(nil), // nil nucleus slice until activation lands
		indexes:  map[string]any{},
	}
}

func (d *DNA) BeingID() string        { return d.beingID }
func (d *DNA) Entities() *EntityIndex { return d.entities }

// RegisterObserve / RegisterExpress wire a gene at boot. Keyed "namespace:gene".
func (d *DNA) RegisterObserve(key string, fn ObserveCompetence) { d.observe[key] = fn }
func (d *DNA) RegisterExpress(key string, fn ExpressCompetence) { d.express[key] = fn }

// observeGene / expressGene look up a registered gene; nil if unregistered (a nil
// gene is a no-op bookend, so an unwired step is inert, not a panic).
func (d *DNA) observeGene(key string) ObserveCompetence { return d.observe[key] }
func (d *DNA) expressGene(key string) ExpressCompetence { return d.express[key] }
