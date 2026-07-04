package dnav2

import (
	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// Orchestrator — membrane as the workflow engine. Holds the durable references
// (DNA, board, kernel, the per-namespace active-state slices, its own throwaway
// tree slice). Per WAVE it is a `membraneRoot` (below): create the worker for this
// step as a child, let the kernel run it (same wave), then read the worker's
// reported condition and dispatch the next step(s) by re-invoking itself.
//
// See specs/2026-07-03/slice-based-dispatch.md ("The orchestrator").
type Orchestrator struct {
	dna    *DNA
	board  *Board
	kernel skyra.Time

	membraneSlice skyra.Slice            // per-step kernel tree-workspace (throwaway)
	slices        map[string]skyra.Slice // namespace → active-state slice (other/self/midlevel/motor)

	// flush is the trace-export hook, called when an impulse completes (deferred to
	// an external store; nil = no-op).
	flush func(impulseID string, trace []Span)
}

func NewOrchestrator(dna *DNA, board *Board, kernel skyra.Time, membraneSlice skyra.Slice, slices map[string]skyra.Slice) *Orchestrator {
	return &Orchestrator{dna: dna, board: board, kernel: kernel, membraneSlice: membraneSlice, slices: slices}
}

func (o *Orchestrator) SetFlush(fn func(string, []Span)) { o.flush = fn }

// Fire enters the pipeline for an external impulse: registers its board row and
// starts the first membrane wave at sensory:intake.
func (o *Orchestrator) Fire(impulseID, prev string, impulse any) {
	o.board.Row(impulseID, prev)
	o.run(StepSensoryIntake, impulseID, impulse)
}

// run spawns one membrane-bracketed wave for `step` (async — a fresh wave that
// terminates the previous one). The membraneRoot both owns the wave and is
// traversed; the worker is its child.
func (o *Orchestrator) run(step Step, impulseID string, impulse any) {
	m := &membraneRoot{orch: o, step: step, impulseID: impulseID, impulse: impulse}
	go o.kernel.Realize(m, m, impulse)
}

// membraneRoot — the per-wave orchestration transcript. Carries the step handoff
// (which worker to run) + the impulseID; the impulse itself stays pure.
type membraneRoot struct {
	orch      *Orchestrator
	step      Step
	impulseID string
	impulse   any

	worker *EphemeralRoot // created in Observe, the kernel's child this wave
}

func (m *membraneRoot) ID() string                            { return "membrane" }
func (m *membraneRoot) Slice() skyra.Slice                    { return m.orch.membraneSlice }
func (m *membraneRoot) NewActivation() skyra.ActivationRecord { return &rootActivation{} }
func (m *membraneRoot) Content() any                          { return nil }
func (m *membraneRoot) Agents() map[string]skyra.Space        { return nil }

// Children exposes the worker so the kernel descends into it (same wave).
func (m *membraneRoot) Children() map[string]skyra.Space {
	if m.worker == nil {
		return nil
	}
	return map[string]skyra.Space{m.worker.ID(): m.worker}
}

// Observe — transcribe the worker for this step (the gene expresses on the
// namespace's active-state slice). The kernel then descends into it.
func (m *membraneRoot) Observe(origin skyra.Space, impulse any) {
	ns := namespaceOf(m.step)
	slice := m.orch.slices[ns]
	key := string(m.step)
	m.worker = m.orch.dna.Transcribe(key, key, key, slice, m.impulseID, nil)
	// hand the working set — the entity index rides EVERY worker (the mRNA carries
	// what it needs). Step-specific: the predict genes get the impulse's active
	// grouping (produced by midlevel:frame, stashed on the row).
	m.worker.Entities = m.orch.dna.Entities()
	switch m.step {
	case StepOtherPredict, StepSelfPredict, StepMotorCollapse:
		if g, ok := m.orch.board.Row(m.impulseID, "").Grouping().(*dna.MemoryCell); ok {
			m.worker.Grouping = g
		}
	}
}

// Express — the worker has run and set its condition; read it, record the span, and
// dispatch the next step(s) per the workflow definition + the board.
func (m *membraneRoot) Express(origin skyra.Space, impulse any) any {
	cond := ""
	if m.worker != nil {
		cond = m.worker.Condition
	}
	row := m.orch.board.Row(m.impulseID, "")
	row.Append(Span{Step: string(m.step), Condition: cond})

	// midlevel:frame produced the active grouping — stash it on the row so the
	// fan-out's predict workers receive it as their working set.
	if m.step == StepMidlevelFrame && m.worker != nil && m.worker.Grouping != nil {
		row.SetGrouping(m.worker.Grouping)
	}

	route, ok := routeFor(m.step, cond)
	if !ok {
		return nil // inert edge — nothing dispatched
	}
	o := m.orch

	switch {
	case route.Reframe:
		// the one-reframe rail: only the first reframe per impulse is allowed
		if row.TryReframe() {
			o.run(StepMidlevelFrame, m.impulseID, m.impulse)
		}

	case route.Done:
		o.finish(row, m.impulseID)

	case route.Wait:
		// a predict branch completed — mark the join and try to fire motor
		if agent := agentOf(m.step); agent != "" {
			row.Mark("G", agent) // MVP: one grouping "G"
		}
		if row.TryFireMotor() {
			o.run(StepMotorCollapse, m.impulseID, m.impulse)
		}

	case len(route.Next) == 1:
		o.run(route.Next[0], m.impulseID, m.impulse)

	case len(route.Next) > 1:
		// fan-out: declare the join set, then spawn one membrane wave per branch
		agents := make([]string, 0, len(route.Next))
		for _, s := range route.Next {
			if a := agentOf(s); a != "" {
				agents = append(agents, a)
			}
		}
		row.Declare("G", agents...) // MVP: one grouping "G"
		for _, s := range route.Next {
			o.run(s, m.impulseID, m.impulse)
		}
	}
	return nil
}

// finish flips the impulse row to done and flushes its trace.
func (o *Orchestrator) finish(row *Row, impulseID string) {
	trace := row.Done()
	if o.flush != nil {
		o.flush(impulseID, trace)
	}
}
