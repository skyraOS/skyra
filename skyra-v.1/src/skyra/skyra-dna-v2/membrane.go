package dnav2

import (
	"skyra-v1/src/skyra"
)

// Orchestrator — membrane as the workflow engine. Holds the durable references
// (DNA, board, kernel, the per-namespace active-state slices, its own throwaway
// tree slice). Per WAVE it is a `membraneRoot` (below): create the worker for this
// step as a child, let the kernel run it (same wave), then read the worker's
// reported condition and dispatch the next step(s) by re-invoking itself.
//
// See specs/2026-07-03/slice-based-dispatch.md ("The orchestrator").
type Orchestrator struct {
	genome *DNA
	board  *Board
	kernel skyra.Time

	membraneSlice skyra.Slice            // per-step kernel tree-workspace (throwaway)
	slices        map[string]skyra.Slice // namespace → active-state slice (other/self/midlevel/motor)

	// flush is the trace-export hook, called when an impulse completes (deferred to
	// an external store; nil = no-op).
	flush func(impulseID string, trace []Span)

	// traces mints one content-level RunTrace per board run (nil = untraced).
	traces *TraceStore
}

func NewOrchestrator(genome *DNA, board *Board, kernel skyra.Time, membraneSlice skyra.Slice, slices map[string]skyra.Slice) *Orchestrator {
	return &Orchestrator{genome: genome, board: board, kernel: kernel, membraneSlice: membraneSlice, slices: slices}
}

func (o *Orchestrator) SetFlush(fn func(string, []Span)) { o.flush = fn }

// SetTraces enables per-board-run content traces (one JSONL file per impulse).
func (o *Orchestrator) SetTraces(s *TraceStore) { o.traces = s }

// Fire enters the pipeline for an external impulse: registers its board row,
// opens its run trace, and starts the first membrane wave at sensory:intake.
func (o *Orchestrator) Fire(impulseID, prev string, impulse any) {
	row := o.board.Row(impulseID, prev)
	if o.traces != nil && row.Trace() == nil {
		row.SetTrace(o.traces.Open(impulseID))
	}
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
	m.worker = m.orch.genome.Transcribe(key, key, key, slice, m.impulseID, nil)
	// hand the working set — the entity index rides EVERY worker (the mRNA carries
	// what it needs). Step-specific: the predict genes get the impulse's active
	// episode (produced by midlevel:frame, stashed on the row).
	row := m.orch.board.Row(m.impulseID, "")
	m.worker.Nodes = m.orch.genome.Nodes()
	m.worker.ObservationID = row.ObservationID()
	m.worker.Trace = row.Trace()
	switch m.step {
	case StepOtherPredict, StepSelfPredict, StepMotorCollapse:
		if g, ok := row.Episode().(*MemoryCell); ok {
			m.worker.Episode = g
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
	if m.worker != nil && m.worker.ObservationID != "" {
		row.SetObservationID(m.worker.ObservationID)
	}
	row.Append(Span{Step: string(m.step), Condition: cond})
	row.Trace().Span(string(m.step), cond)

	// midlevel:frame produced the active episode — stash it on the row so the
	// fan-out's predict workers receive it as their working set.
	if m.step == StepMidlevelFrame && m.worker != nil && m.worker.Episode != nil {
		row.SetEpisode(m.worker.Episode)
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
			row.Mark("episode", agent) // MVP: one active episode
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
		row.Declare("episode", agents...) // MVP: one active episode
		for _, s := range route.Next {
			o.run(s, m.impulseID, m.impulse)
		}
	}
	return nil
}

// finish flips the impulse row to done and flushes its trace.
func (o *Orchestrator) finish(row *Row, impulseID string) {
	trace := row.Done()
	row.Trace().Close()
	if o.flush != nil {
		o.flush(impulseID, trace)
	}
}
