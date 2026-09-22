// Package ctu is the execution core — the HOT path (ACT).
//
// FRESH BUILD. The CTU is nearly-stateless. Move is its single ENTRY (ingress calls it; local THINK
// calls it). It CARRIES OS-specific driver mechanisms but has NO ambient authority — it fires nothing
// without a horizon handed to it (CHERI-shaped). It holds only VIEWS/HANDLES: the carried horizon and
// cut-IDs as fds. The authoritative state (table, socket table, mint) is the kernel's.
//
// AUTHORITY: skyra/specs/2026-08-24/skyra-dna-v11/research/move-and-the-kernel-barrier.md,
// think-act-paths.md, capability-table-and-horizon.md, bootstrap/ctu.md.
package ctu

import (
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v11/causal"
)

// CTU is the execution endpoint. Transfer is the door; the Transition core runs behind it.
type CTU interface {
	// Transfer is the CTU's single ENTRY — a MOVE crosses the barrier in. Called by ingress (or local
	// THINK). It does NOT drive: it LANDS + ENQUEUES + wakes the loop, then returns. Dispatches on Header:
	//   Request  → VERIFY the carried horizon offline (per capability: sig + not-expired), MATERIALIZE the
	//              Context, insert the move into the WorkList at Ready. the HOT path needs no kernel.
	//   Response → land m.Returned on its fd — Cache[m.ID.Possibility].Registers[m.ID.Writes] — POP the
	//              Issued move from Work, flip dependents Waiting→Ready.
	// Then it signals the loop. It never mints and never widens a horizon. (move-and-the-kernel-barrier)
	Transfer(m c.Move) c.Outcome

	// Run is the driver loop. It PARKS when the Ready bucket is empty (the thread blocks) and WAKES on a
	// Transfer — it NEVER exits on quiescence. Each pass hands the WorkList to the one Transition core.
	Run()
}

// Core is the CTU's single execution core — STATIC, exactly ONE instance (the pipeline hardware, the T).
// It holds the WHOLE of the CTU's microarchitectural state AND runs Transition over it. This is the
// ground of "nearly-stateless": every field is a reconstructible CACHE or transient in-flight work, and
// NONE is authority (the capability table, socket table, and mint are the kernel's). Register files live
// inside the Contexts (Cache); the reorder buffer is Work; the i-cache is Bodies. The CTU has-a Core; its
// Run loop drives Core.Transition, and it is NOT a pluggable seam (hence a struct, not an interface).
type Core struct {
	Self   c.Target            // this CTU's identity — egress source and return-route destination
	Cache  c.Cache             // resident possibilities (context map) — Materialize writes, Remember drains
	Bodies c.BodyCache         // compiled-body i-cache — a miss is the page fault to the kernel
	Work   c.WorkList          // the global move work-list (the ROB), bucketed by state
	Interp Interpret           // horizon block-selection seam: provisioned target or Fault
	Local  map[c.Target]Driver // installed mechanisms: presence here distinguishes LOCAL from CUT
	Out    Egress              // the ONE external handle — issue a cut / emit a response across the barrier
	// + a wake signal for Run (park/unpark) — impl detail, not a cross-boundary type
}

// Transition steps the global WorkList, keyed by STATE: drain Ready → dispatch locally or issue a cut (via
// co.Out) → move to Issued; a landing pops Issued and flips dependents Waiting→Ready. Returns at
// quiescence (no move Ready this pass). A move to a target NOT in the context's horizon FAULTs (page fault
// → kernel). Never blocks, never recurses. Body: execute lane.
func (co *Core) Transition(moves c.WorkList) c.Outcome {
	ensureWork(moves)
	completed := false

	// A pass drains the run queue, including dependents made Ready by an earlier local completion. That
	// is the dataflow wake cascade: there is no recursive Drive call and no separate scheduler structure.
	for len(moves[c.Ready]) > 0 {
		ready := make([]c.MoveID, 0, len(moves[c.Ready]))
		for id := range moves[c.Ready] {
			ready = append(ready, id)
		}

		for _, id := range ready {
			move, ok := moves[c.Ready][id]
			if !ok { // an earlier completion in this pass may already have changed the bucket
				continue
			}
			ctx := co.Cache[id.Possibility]
			if ctx == nil {
				return c.Failure
			}

			eligible, out := co.interpret(ctx, &move)
			if out != c.Success {
				return out
			}
			if len(eligible.Targets) == 0 {
				return c.Fault
			}
			target := eligible.Targets[0]

			if driver, local := co.Local[target]; local {
				inputs, ready := operands(ctx, move.Template.Reads)
				if !ready {
					move.State = c.Waiting
					delete(moves[c.Ready], id)
					moves[c.Waiting][id] = move
					continue
				}

				observation, performed := driver.Perform(move.Template.Action, inputs)
				if performed != c.Success {
					return performed
				}
				land(ctx, id.Writes, move.Template.Contract, &observation)
				delete(moves[c.Ready], id)
				flipDependents(moves, ctx)
				completed = setResult(co.Bodies, ctx) || completed
				continue
			}

			if co.Out == nil {
				return c.Failure
			}
			outbound := move
			outbound.Template.Target = target
			outbound.Header = c.Request
			outbound.Returned = nil
			outbound.Seed = cutSeed(co.Self, ctx, move, target)
			if issued := co.Out.Egress(co.Self, outbound); issued != c.Success {
				return issued
			}

			delete(moves[c.Ready], id)
			if move.Template.Return {
				move.State = c.Issued
				moves[c.Issued][id] = move
			} else {
				flipDependents(moves, ctx)
			}
		}
	}

	if completed {
		return c.Success
	}
	return c.NotYet
}

func (co *Core) interpret(ctx *c.Context, move *c.Move) (EligibleSet, c.Outcome) {
	if co.Interp != nil {
		return co.Interp.Interpret(ctx, move)
	}
	return HorizonInterpreter{}.Interpret(ctx, move)
}

func ensureWork(work c.WorkList) {
	for _, state := range []c.MoveState{c.Waiting, c.Ready, c.Issued} {
		if work[state] == nil {
			work[state] = make(map[c.MoveID]c.Move)
		}
	}
}

func operands(ctx *c.Context, reads []c.RegKey) ([]c.Observation, bool) {
	in := make([]c.Observation, 0, len(reads))
	for _, key := range reads {
		register, ok := ctx.Registers[key]
		if !ok || register.Value == nil {
			return nil, false
		}
		in = append(in, *register.Value)
	}
	return in, true
}

func land(ctx *c.Context, key c.RegKey, contract c.Contract, value *c.Observation) {
	register := ctx.Registers[key]
	if register.Contract == (c.Contract{}) {
		register.Contract = contract
	}
	register.Value = value
	ctx.Registers[key] = register
}

func flipDependents(work c.WorkList, ctx *c.Context) {
	for id, move := range work[c.Waiting] {
		if id.Possibility != ctx.ID {
			continue
		}
		if _, ready := operands(ctx, move.Template.Reads); !ready {
			continue
		}
		delete(work[c.Waiting], id)
		move.State = c.Ready
		work[c.Ready][id] = move
	}
}

func setResult(bodies c.BodyCache, ctx *c.Context) bool {
	body, ok := bodies[ctx.Body]
	if !ok {
		return false
	}
	register, ok := ctx.Registers[body.Return]
	if !ok || register.Value == nil {
		return false
	}
	ctx.Result = register.Value
	return true
}

func cutSeed(self c.Target, parent *c.Context, move c.Move, target c.Target) *c.Context {
	registers := make(map[c.RegKey]c.Register, len(move.Template.Reads))
	for _, key := range move.Template.Reads {
		registers[key] = parent.Registers[key]
	}
	return &c.Context{
		ID:        move.ID.Possibility,
		Registers: registers,
		Horizon:   bound(parent.Horizon, target, move.Template.Action),
		ReturnTo:  &c.Route{To: self, Fd: move.ID},
	}
}

func bound(horizon c.Horizon, target c.Target, action c.Name) c.Horizon {
	bounded := c.Horizon{}
	for _, cap := range horizon.Capabilities {
		if cap.Target == target && cap.Action == action {
			bounded.Capabilities = append(bounded.Capabilities, cap)
		}
	}
	return bounded
}

// Egress is the CTU's consumer-side port on the kernel barrier — the only OUTWARD call Transition makes.
// Declared HERE, not imported from kernel, so the CTU stays decoupled; kernel.Barrier satisfies it.
type Egress interface {
	Egress(from c.Target, m c.Move) c.Outcome
}

// Materialize is the LOADER: BodyID + Horizon → a fresh Context. It RESOLVES the Body from the CTU's
// content-addressed body cache (an i-cache), then builds the register file. On the hot path the CTU does
// this locally (no kernel); on the cold path the kernel does it. (cpu-execution-model; closes the cached-Body seam)
type Materialize interface {
	Materialize(id c.BodyID, h c.Horizon) *c.Context
}

// Verify is the offline capability check — the INGRESS permission check. No kernel, no caveat
// evaluation: signature ok, not expired, and the capability authorizes (t, action). Fine-grained +
// timestamped makes it a lookup. This is what keeps the hot path kernel-free.
type Verify interface {
	Verify(cap c.Capability, t c.Target, action c.Name) bool
}

// Interpret is block selection over the CONTEXT's HORIZON (its provisioned view). A name that resolves
// in the horizon → dispatch; a name that MISSES because it is unprovisioned → FAULT (page fault). A
// name that resolves to a driver on THIS ctu runs it; the ctu never enumerates its own drivers.
type Interpret interface {
	Interpret(ctx *c.Context, mv *c.Move) (EligibleSet, c.Outcome)
}

type EligibleSet struct{ Targets []c.Target }

// Drivers are the CTU's OS-specific mechanisms — inert, present, unknown to the CTU. It ADVERTISES them
// to the kernel at handshake (enumeration); it never fires one except via a horizon. (bootstrap/ctu.md)
type Drivers interface {
	// Advertise exposes each driver WITH the fine-grained capabilities it defines (its actions), for the
	// kernel to READ and register into the sysblock. PERMISSIONS LIVE ON THE DRIVERS: the driver declares
	// what (Target, Action) capabilities exist; the kernel provisions; ingress enforces. The CTU does not
	// use this inventory itself.
	Advertise() []DriverDecl
}

// Driver is an installed OS-specific mechanism. Decl is the inert kernel-facing advertisement; Perform
// is reachable only after Interpret has established that the resident Context's horizon authorizes it.
type Driver interface {
	Decl() DriverDecl
	Perform(action c.Name, in []c.Observation) (c.Observation, c.Outcome)
}

// DriverDecl is a driver's declaration: the target it IS, and the fine-grained actions it defines. Each
// (Target, Action) is one capability the kernel can provision. This is where the OS-specific permission
// semantics live — keeping the kernel generic and OS-independent. (permissions live on the drivers.)
type DriverDecl struct {
	Target  c.Target
	Actions []c.Name
}

// Machine is the concrete CTU endpoint. The mutex serializes the one Core and the condition variable is
// its interrupt line: Transfer signals it after enqueue/landing; Run sleeps on it when no move is Ready.
type Machine struct {
	Core     *Core
	Loader   Materialize
	Verifier Verify

	mu   sync.Mutex
	wake *sync.Cond
	gen  uint64 // incremented for every landing/enqueue so completion-only responses also wake Run
}

var (
	_ CTU         = (*Machine)(nil)
	_ Materialize = CachedMaterializer{}
	_ Interpret   = HorizonInterpreter{}
	_ Drivers     = DriverSet(nil)
)

// New returns a hot-path endpoint over core. A nil loader selects the local cached-body materializer;
// verification remains explicit because signature policy and key material belong outside the CTU.
func New(core *Core, loader Materialize, verifier Verify) *Machine {
	if core == nil {
		core = &Core{}
	}
	initializeCore(core)
	if loader == nil {
		loader = CachedMaterializer{Bodies: core.Bodies}
	}
	machine := &Machine{Core: core, Loader: loader, Verifier: verifier}
	machine.wake = sync.NewCond(&machine.mu)
	return machine
}

// Transfer implements the CTU door. It mutates only resident cache/work state and wakes Run; execution
// remains exclusively behind the loop (or an explicit Transition call in a single-core test).
func (m *Machine) Transfer(move c.Move) c.Outcome {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch move.Header {
	case c.Request:
		if move.Seed == nil || move.Seed.Body == (c.BodyID{}) || m.Loader == nil || m.Verifier == nil {
			return c.Fault
		}
		if !authorized(m.Verifier, move.Seed.Horizon, move.Template.Target, move.Template.Action) {
			return c.Failure
		}

		ctx := m.Loader.Materialize(move.Seed.Body, move.Seed.Horizon)
		if ctx == nil {
			return c.Fault
		}
		mergeSeed(ctx, move.Seed, move.ID.Possibility)
		if ctx.ID == (c.PossibilityID{}) {
			return c.Failure
		}

		body, ok := m.Core.Bodies[ctx.Body]
		if !ok {
			return c.Fault
		}
		removePossibility(m.Core.Work, ctx.ID)
		m.Core.Cache[ctx.ID] = ctx
		for _, template := range body.Moves {
			id := c.MoveID{Possibility: ctx.ID, Writes: template.Writes}
			state := c.Waiting
			if _, ready := operands(ctx, template.Reads); ready {
				state = c.Ready
			}
			m.Core.Work[state][id] = c.Move{ID: id, Template: template, State: state, Header: c.Request}
		}
		setResult(m.Core.Bodies, ctx)
		m.notify()
		return c.Success

	case c.Response:
		if move.Returned == nil {
			return c.Failure
		}
		ctx := m.Core.Cache[move.ID.Possibility]
		issued, ok := m.Core.Work[c.Issued][move.ID]
		if ctx == nil || !ok {
			return c.Failure
		}
		land(ctx, move.ID.Writes, issued.Template.Contract, move.Returned)
		delete(m.Core.Work[c.Issued], move.ID)
		flipDependents(m.Core.Work, ctx)
		setResult(m.Core.Bodies, ctx)
		m.notify()
		return c.Success

	default:
		return c.Failure
	}
}

// Run is deliberately immortal. Fault leaves the offending move Ready and parks until a later Transfer
// wakes the core after the out-of-scope kernel path has changed provisioning.
func (m *Machine) Run() {
	var seen uint64
	m.mu.Lock()
	for {
		for m.gen == seen && len(m.Core.Work[c.Ready]) == 0 {
			m.wake.Wait()
		}
		seen = m.gen
		out := c.NotYet
		if len(m.Core.Work[c.Ready]) != 0 {
			out = m.Core.Transition(m.Core.Work)
		}
		m.retireCompleted()
		if out == c.Fault {
			for m.gen == seen {
				m.wake.Wait()
			}
			seen = m.gen
		}
	}
}

func (m *Machine) notify() {
	m.gen++
	m.wake.Signal()
}

func (m *Machine) retireCompleted() {
	for id, ctx := range m.Core.Cache {
		if ctx.Result == nil || active(m.Core.Work, id) {
			continue
		}
		if ctx.ReturnTo != nil {
			if m.Core.Out == nil {
				continue
			}
			response := c.Move{
				ID:       ctx.ReturnTo.Fd,
				Template: c.MoveTemplate{Target: ctx.ReturnTo.To},
				Header:   c.Response,
				Returned: ctx.Result,
			}
			if m.Core.Out.Egress(m.Core.Self, response) != c.Success {
				continue
			}
		}
		removePossibility(m.Core.Work, id)
		delete(m.Core.Cache, id)
	}
}

func initializeCore(core *Core) {
	if core.Cache == nil {
		core.Cache = make(c.Cache)
	}
	if core.Bodies == nil {
		core.Bodies = make(c.BodyCache)
	}
	if core.Work == nil {
		core.Work = make(c.WorkList)
	}
	ensureWork(core.Work)
	if core.Local == nil {
		core.Local = make(map[c.Target]Driver)
	}
}

func authorized(verifier Verify, horizon c.Horizon, target c.Target, action c.Name) bool {
	for _, capability := range horizon.Capabilities {
		if capability.Target == target && capability.Action == action && verifier.Verify(capability, target, action) {
			return true
		}
	}
	return false
}

func mergeSeed(ctx, seed *c.Context, fallback c.PossibilityID) {
	ctx.ID = seed.ID
	if ctx.ID == (c.PossibilityID{}) {
		ctx.ID = fallback
	}
	ctx.Body = seed.Body
	ctx.Horizon = seed.Horizon
	ctx.ReturnTo = seed.ReturnTo
	for key, seeded := range seed.Registers {
		base := ctx.Registers[key]
		if seeded.Contract != (c.Contract{}) {
			base.Contract = seeded.Contract
		}
		base.Value = seeded.Value
		ctx.Registers[key] = base
	}
}

func removePossibility(work c.WorkList, possibility c.PossibilityID) {
	for _, state := range []c.MoveState{c.Waiting, c.Ready, c.Issued} {
		for id := range work[state] {
			if id.Possibility == possibility {
				delete(work[state], id)
			}
		}
	}
}

func active(work c.WorkList, possibility c.PossibilityID) bool {
	for _, state := range []c.MoveState{c.Ready, c.Issued} {
		for id := range work[state] {
			if id.Possibility == possibility {
				return true
			}
		}
	}
	return false
}

// CachedMaterializer is the hot-path loader. It resolves the immutable Body locally and allocates a
// fresh register file; Transfer overlays the request's seed values and routing metadata afterward.
type CachedMaterializer struct{ Bodies c.BodyCache }

func (m CachedMaterializer) Materialize(id c.BodyID, horizon c.Horizon) *c.Context {
	body, ok := m.Bodies[id]
	if !ok {
		return nil
	}
	registers := make(map[c.RegKey]c.Register)
	for _, move := range body.Moves {
		registers[move.Writes] = c.Register{Contract: move.Contract}
		for _, read := range move.Reads {
			if _, ok := registers[read]; !ok {
				registers[read] = c.Register{}
			}
		}
	}
	return &c.Context{Body: id, Registers: registers, Horizon: horizon}
}

// HorizonInterpreter performs block selection against the resident provisioned view. Locality is not
// encoded here: Transition separately looks the selected target up in Core.Local.
type HorizonInterpreter struct{}

func (HorizonInterpreter) Interpret(ctx *c.Context, move *c.Move) (EligibleSet, c.Outcome) {
	for _, capability := range ctx.Horizon.Capabilities {
		if capability.Target == move.Template.Target && capability.Action == move.Template.Action {
			return EligibleSet{Targets: []c.Target{move.Template.Target}}, c.Success
		}
	}
	return EligibleSet{}, c.Fault
}

// Function adapters keep tests and hosted integrations small without weakening the named seams.
type VerifyFunc func(c.Capability, c.Target, c.Name) bool

func (f VerifyFunc) Verify(cap c.Capability, target c.Target, action c.Name) bool {
	return f(cap, target, action)
}

type InterpretFunc func(*c.Context, *c.Move) (EligibleSet, c.Outcome)

func (f InterpretFunc) Interpret(ctx *c.Context, move *c.Move) (EligibleSet, c.Outcome) {
	return f(ctx, move)
}

type EgressFunc func(c.Target, c.Move) c.Outcome

func (f EgressFunc) Egress(from c.Target, move c.Move) c.Outcome { return f(from, move) }

// DriverSet is both the installed-mechanism map and the kernel-facing advertisement source.
type DriverSet map[c.Target]Driver

func (drivers DriverSet) Advertise() []DriverDecl {
	decls := make([]DriverDecl, 0, len(drivers))
	for _, driver := range drivers {
		decls = append(decls, driver.Decl())
	}
	return decls
}
