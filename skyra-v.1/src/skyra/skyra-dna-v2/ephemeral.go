package dnav2

import (
	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// EphemeralRoot — the transcript (mRNA). A per-wave skyra.Space over the DNA:
// created at wave start (the transcription event), discarded when the wave
// completes. It runs ONE selected competence as the kernel's bookends, traverses
// the relevant namespace's active-state slice, and holds ALL the per-wave mutable
// state — never on the DNA, never on the impulse.
//
// A transcript references the DNA (reference + select, never copy the genome).
// Because it is unique per wave and thrown away, per-wave state on it can't be
// clobbered by a concurrent wave — scenario 4 dissolves by construction, the way
// cells do parallelism: keep the DNA read-only, spin off disposable mRNA.
type EphemeralRoot struct {
	id      string
	dna     *DNA
	slice   skyra.Slice       // the namespace's active-state slice (the tree builds here)
	observe ObserveCompetence // the selected gene (nil → no-op bookend)
	express ExpressCompetence

	// per-wave state — the transcript's private scratch, dies with the root:
	ImpulseID string // the correlation ID this wave carries
	Condition string // the parse-condition this competence reports; the orchestrator routes on it
	Handoff   any    // stamped by the creator at transcription (e.g. a verdict to this step)

	// working set — references into the nucleus's master graph, handed by the
	// membrane at transcription (the mRNA carries what it needs). The entity index
	// rides every worker; step-specific handles (the active grouping) are set for
	// the steps that need them.
	Nodes *NodeRegistry
	Grouping *dna.MemoryCell // the active grouping — set by midlevel:frame (out), handed to predicts (in)
}

// Transcribe expresses one gene for one wave: select the observe/express genes,
// point at the DNA + the namespace slice, stamp any cross-step handoff. Returns a
// disposable root ready to hand the kernel.
func (d *DNA) Transcribe(id, observeKey, expressKey string, slice skyra.Slice, impulseID string, handoff any) *EphemeralRoot {
	return &EphemeralRoot{
		id:        id,
		dna:       d,
		slice:     slice,
		observe:   d.observeGene(observeKey),
		express:   d.expressGene(expressKey),
		ImpulseID: impulseID,
		Handoff:   handoff,
	}
}

// DNA is how a competence reaches durable state (the one place it may).
func (e *EphemeralRoot) DNA() *DNA { return e.dna }

// ── skyra.Space ──────────────────────────────────────────────────────────────

func (e *EphemeralRoot) ID() string         { return e.id }
func (e *EphemeralRoot) Slice() skyra.Slice { return e.slice }

// Children — a leaf for now. Pipeline genes (grade/predict/frame/collapse) do their
// work directly in Observe (inference + active-state reads); they don't rely on
// kernel descent. Traversal genes (recall) that need descent come later.
func (e *EphemeralRoot) Children() map[string]skyra.Space { return nil }
func (e *EphemeralRoot) Agents() map[string]skyra.Space   { return nil }
func (e *EphemeralRoot) Content() any                     { return nil }

func (e *EphemeralRoot) NewActivation() skyra.ActivationRecord { return &rootActivation{} }

// Observe / Express — the bookends run the selected gene. A nil gene is inert.
func (e *EphemeralRoot) Observe(origin skyra.Space, impulse any) {
	if e.observe != nil {
		e.observe(e, origin, impulse)
	}
}

func (e *EphemeralRoot) Express(origin skyra.Space, impulse any) any {
	if e.express != nil {
		return e.express(e, origin, impulse)
	}
	return nil
}

type rootActivation struct{}

func (a *rootActivation) Activate(visited bool, traversalCount int) float64 { return 1.0 }
