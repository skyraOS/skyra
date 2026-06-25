package relation

import "skyra-v1/src/skyra"

const (
	defaultAmplitude = 1.0 // fallback initial wave amplitude when impulse does not specify one
	amplitudeDecay   = 0.9 // attenuation per depth step
)

type Relation struct {
	id        string
	Amplitude float64 // initial amplitude, set at spawn; per-node amplitude attenuates from it
}

func NewRelation() *Relation {
	return &Relation{
		Amplitude: defaultAmplitude,
	}
}

func (r *Relation) ID() string {
	return r.id
}

// Realize runs one wave. Its only result path is return perception; callers
// do not receive a return value.
func (r *Relation) Realize(origin, traversed skyra.Space, impulse any) {
	r.realize(origin, traversed, impulse)
}

// realize is one wave: a locked traversal, then a return perception wave rooted
// at the origin. The traversal lock is released before return perception starts.
func (r *Relation) realize(origin, traversed skyra.Space, impulse any) {
	sl := traversed.Slice()
	if sl == nil {
		return
	}
	r.traverse(origin, traversed, impulse, sl, amplitudeOf(impulse, r.Amplitude))
	// Perception Invariant SUSPENDED for the sensorimotor-boundary build: the kernel
	// returns void (no return-perception wave). The forward pipeline carries results
	// via explicit cell-initiated spawns (go kernel.Realize) instead. Restore the line
	// below once the forward path works. See specs/2026-06-16/skyra-dna/sensorimotor.md
	// §5 and design-docs/laws/invariants/perception-invariant.md (suspension banner).
	//
	//	result := r.traverse(origin, traversed, impulse, sl, amplitudeOf(impulse, r.Amplitude))
	//	r.returnToOrigin(origin, result, amplitudeOf(impulse, r.Amplitude), impulse)
}

// traverse walks `traversed` under its slice lock and returns the wave's
// result. Pure transport: in-plane DFS only. Beings are rooted directly
// by the caller; no crossing.
func (r *Relation) traverse(origin, traversed skyra.Space, impulse any, sl skyra.Slice, amplitude float64) any {
	sl.Lock()
	defer sl.Unlock()

	sl.SetCount(sl.Count() + 1) // proper time — this frame's clock

	// double buffer: last traversal's tree becomes Prior; Current rebuilds.
	tree := sl.Tree()
	tree.Prior = tree.Current
	tree.Current = nil
	sl.SetResult(sl.Maps())

	// visited-once (the rail): nodes on the current path. A traversal never
	// re-descends a node already on its path, so it always terminates.
	onPath := map[string]bool{}

	traversed.Observe(origin, impulse)
	rootNode := r.rootNode(traversed, sl, amplitude)
	r.scoreChildren(rootNode, sl)
	onPath[rootNode.CellID] = true
	stack := []*skyra.Node{rootNode}

	for len(stack) > 0 {
		top := stack[len(stack)-1]

		if top.NextChild < len(top.Survivors) {
			childNode := top.Survivors[top.NextChild]
			top.NextChild++

			if onPath[childNode.CellID] {
				continue // already on the current path — the rail
			}

			onPath[childNode.CellID] = true
			childNode.Cell.Observe(origin, impulse)
			r.scoreChildren(childNode, sl)
			stack = append(stack, childNode)
		} else {
			stack = stack[:len(stack)-1]
			onPath[top.CellID] = false

			out := top.Cell.Express(origin, impulse)
			if out != nil {
				r.recordResult(sl, top.CellID, impulse, out)
			}

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, top)
			} else {
				tree.Current = []*skyra.Node{top} // the finished tree
			}
		}
	}

	return sl.Result()
}

// returnToOrigin turns a completed wave result into an inward perception wave.
// Return waves are currently terminal to keep the runtime finite until the
// scheduler/time controls exist.
func (r *Relation) returnToOrigin(origin skyra.Space, result any, amplitude float64, completed any) {
	if origin == nil || isReturnImpulse(completed) {
		return
	}
	r.realize(origin, origin, skyra.Impulse{
		Primitive: "perceive",
		Payload:   snapshotResult(result),
		Amplitude: amplitude,
		Return:    true,
	})
}

// rootNode scores the root once — it has no parent to carry its score.
func (r *Relation) rootNode(cell skyra.Space, sl skyra.Slice, amplitude float64) *skyra.Node {
	rec := recordFor(cell.ID(), cell, sl)
	node := &skyra.Node{
		CellID:    cell.ID(),
		Cell:      cell,
		Record:    rec,
		Amplitude: amplitude,
	}
	if rec != nil {
		node.Score = rec.Activate(true, sl.Count()) * amplitude
	}
	return node
}

// scoreChildren activates each child of node exactly once, divides the
// attenuated signal across candidate branches, builds child nodes for the
// survivors, and fills node.Survivors. No cell is ever activated twice within
// one parent — a survivor's Score becomes its own node score when descended.
// Branching conserves amplitude: wider waves split energy instead of
// duplicating it.
func (r *Relation) scoreChildren(node *skyra.Node, sl skyra.Slice) {
	type candidate struct {
		id         string
		cell       skyra.Space
		record     skyra.ActivationRecord
		activation float64
	}
	var candidates []candidate

	for id, child := range node.Cell.Children() {
		rec := recordFor(id, child, sl)
		if rec == nil {
			continue // no record — doesn't participate in activation this frame
		}
		candidates = append(candidates, candidate{
			id:         id,
			cell:       child,
			record:     rec,
			activation: rec.Activate(true, sl.Count()),
		})
	}
	if len(candidates) == 0 {
		return
	}

	childAmp := (node.Amplitude * amplitudeDecay) / float64(len(candidates))
	for _, cand := range candidates {
		score := cand.activation * childAmp
		if score > sl.Threshold() {
			node.Survivors = append(node.Survivors, &skyra.Node{
				CellID:    cand.id,
				Cell:      cand.cell,
				Record:    cand.record,
				Score:     score,
				Amplitude: childAmp,
			})
		}
	}
}

// recordFor returns the activation record under `key`, creating it on first
// sight. The key is the caller's name for this slot, not the cell's own ID: the
// root keys by cell ID, but a child keys by its parent's map key — which lets a
// cell hand the kernel edge-scoped keys (A→B vs C→B) so activation is owned by
// the edge, not the node. The kernel stays oblivious: it just keys a record off
// the identity it was handed. The world's map keys equal its point IDs, so this
// is a no-op there.
func recordFor(key string, cell skyra.Space, sl skyra.Slice) skyra.ActivationRecord {
	records := sl.Records()
	rec := records[key]
	if rec == nil {
		rec = cell.NewActivation()
		if rec != nil {
			records[key] = rec
		}
	}
	return rec
}

func amplitudeOf(impulse any, fallback float64) float64 {
	if impulse, ok := impulse.(skyra.Impulse); ok && impulse.Amplitude > 0 {
		return impulse.Amplitude
	}
	if impulse, ok := impulse.(*skyra.Impulse); ok && impulse != nil && impulse.Amplitude > 0 {
		return impulse.Amplitude
	}
	return fallback
}

func (r *Relation) recordResult(sl skyra.Slice, id string, impulse any, out any) {
	key := resultMapKey(impulse)
	if key == "" {
		return
	}
	results, ok := sl.Maps()[key].(map[string]any)
	if !ok {
		results = map[string]any{}
		sl.Maps()[key] = results
	}
	results[id] = out
}

func resultMapKey(impulse any) string {
	if impulse, ok := impulse.(skyra.Impulse); ok {
		if impulse.ResultMap != "" {
			return impulse.ResultMap
		}
		return impulse.Primitive
	}
	if impulse, ok := impulse.(*skyra.Impulse); ok && impulse != nil {
		if impulse.ResultMap != "" {
			return impulse.ResultMap
		}
		return impulse.Primitive
	}
	return ""
}

func isReturnImpulse(impulse any) bool {
	if impulse, ok := impulse.(skyra.Impulse); ok {
		return impulse.Return
	}
	if impulse, ok := impulse.(*skyra.Impulse); ok && impulse != nil {
		return impulse.Return
	}
	return false
}

// snapshotResult deep-copies a result hand-off so the perceiver receives a
// stable snapshot, not a live reference to the producer's still-mutating Maps.
// It copies the outer result-map and each inner map[string]any; leaf outputs
// are immutable values and are shared.
func snapshotResult(result any) any {
	outer, ok := result.(map[string]any)
	if !ok {
		return result
	}
	cp := make(map[string]any, len(outer))
	for key, val := range outer {
		if inner, ok := val.(map[string]any); ok {
			innerCp := make(map[string]any, len(inner))
			for id, out := range inner {
				innerCp[id] = out
			}
			cp[key] = innerCp
		} else {
			cp[key] = val
		}
	}
	return cp
}
