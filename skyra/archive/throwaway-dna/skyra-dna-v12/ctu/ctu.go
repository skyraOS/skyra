// Package ctu is the execution core — the HOT path (ACT).
//
// CARRIED from v11 (proven: ctu_test green) as the SEAM surface — NOT the v12 comb focus. The CTU is
// nearly-stateless: it holds VIEWS/HANDLES + reconstructible CACHES, never authority. It fires nothing
// without a horizon handed to it, and — per 08-27 — it NEVER constructs one (bound() moved to the kernel).
//
// AUTHORITY: specs/2026-08-24/skyra-dna-v11/research/move-and-the-kernel-barrier.md, think-act-paths.md;
// specs/2026-08-27/horizon-and-attention.md.
package ctu

import c "skyra-v1/src/skyra/skyra-dna-v12/causal"

// CTU is the execution ENDPOINT — one network identity, the SHARED memory hierarchy, and N execution
// CORES. It is the "socket": the barrier port and drivers attach HERE, and it holds the shared caches +
// work pool the cores draw from. Multi-core is not a bolt-on — the dataflow/ROB model is inherently
// PARALLEL (independent Ready moves can run at once), so Cores are parallel pipelines over one shared
// frame. At N=1 this is exactly v11's proven single core. There is NO horizon-construction here (08-27):
// Out carries cuts, the kernel Bounds horizons; every field is a reconstructible cache/handle, not authority.
type CTU struct {
	Self     c.Target            // endpoint identity (per-CTU, NOT per-core) — egress source, return dest
	Cache    c.Cache             // SHARED resident possibilities (the frames) — cores read/write registers here
	Bodies   c.BodyCache         // SHARED i-cache — read-only, content-addressed, naturally shareable
	Work     c.WorkList          // SHARED READY table — cores claim-and-issue from here (the ready frontier)
	InFlight c.WorkList          // SHARED ISSUED table — dispatched-and-pending moves; the local in-flight latch.
	//                              a move claim-and-issued LEAVES Work and ENTERS here; its Landing POPS it.
	//                              (register-lifecycle-and-transition.md: the three tables — Work/InFlight/Registers.
	//                              for a REMOTE move this is the kernel socket table's local shadow, keyed by fd.)
	Interp   Interpret           // SHARED horizon block-selection
	Local    map[c.Target]Driver // SHARED installed mechanisms
	Out      Egress              // SHARED barrier port
	Verifier Verify              // SHARED offline capability check at ingress — the MMU (on-chip; see below)
	Traps    map[Kind]Handler    // the IDT — the vector table the KERNEL fills at boot; the CTU only VECTORS
	//                              through it, never fills it (08-27 kernel-handler-block.md).
	Cores []Core // the execution units (N≥1)
}

// Transfer is the endpoint DOOR — a FORWARD activation (a Move: identity + authority + TargetFd/Action +
// ReturnFd) crosses in. VERIFY + materialize the Context + enqueue its INITIALLY-READY moves into shared Work
// + wake the cores; no driving. Run starts the core loops (one per Core), each parking on empty Work and
// waking on Transfer.
//
//	Forward  → VERIFY the carried horizon (mv.Horizon) offline; if the Body is in Bodies (i-cache HIT) build
//	           the Context + enqueue its INITIALLY-READY child moves (the leaves) into Work — the rest enter
//	           via flip as their inputs land (lazy) — else RAISE the Materialize vector (Traps[Materialize]).
//	Landing  → the RESPONSE type (c.Landing = {fd, observation}) lands on its register
//	           (Cache[Fd.Possibility].Registers[Fd.Writes] = LANDED), pops the in-flight entry, wakes deps.
//	           (a Landing is NOT a Move; Transfer taking a forward Move vs a separate Landing entry is OPEN.)
//
// Bodies: execute lane.
func (ct *CTU) Transfer(mv c.Move) c.Outcome { return c.NotYet }
func (ct *CTU) Run()                         {}

// Land is the RETURN-leg door — the LAND half of the reactor. BOTH a local async Driver.Perform completion
// AND a remote Landing (via kernel Barrier ingress) arrive here, identically (register-lifecycle-and-
// transition.md: local and remote are the same to the engine). It writes the value into the parent's slot
// (Cache[l.Fd.Possibility].Registers[l.Fd.Writes]), POPs the in-flight entry (InFlight), flips dependents
// into Work, and wakes the cores. Separated from Transfer (the forward Move door) so the two wire types stay
// two entry points, not one Move-with-a-bit.
func (ct *CTU) Land(l c.Landing) c.Outcome { return c.NotYet }

// Core is ONE execution pipeline — lightweight. It carries its identity and (later) its transient in-flight
// state; the memory it works over (Cache/Bodies/Work) is the CTU's, SHARED. At N=1 this is v11's single
// core; N>1 is the SAME structure plus the concurrency discipline — the OPEN area: how cores CLAIM Ready
// moves without double-dispatch, LAND without racing, and FLIP dependents on the shared Work.
type Core struct {
	ID CoreID
	// per-core transient (COMB): the move being transitioned; a private steal-slice. deferred with the
	// concurrency model.
}

type CoreID struct{ Value string }

// Transition is ONE core's ISSUE pass over the SHARED endpoint (register-lifecycle-and-transition.md — the
// two-part reactor): claim-and-issue ready moves from host.Work — the atomic remove IS the claim IS the
// dispatch, FIRE-AND-FORGET (local host.Local.Perform, async; or remote host.Out.Egress). LANDINGS are the
// OTHER half, handled on arrival (write Registers, pop in-flight, flip dependents) — not here. Returns at
// quiescence (Work empty AND in-flight empty). (CONCURRENCY: the claim + land need synchronization once N>1.)
func (co *Core) Transition(host *CTU) c.Outcome {
	return c.NotYet // carried from v11 (single-core, tested there); re-derive + add the claim discipline
}

// Egress is the CTU's consumer-side port on the kernel barrier — the only OUTWARD call. kernel.Barrier
// satisfies it. Declared here so the CTU stays decoupled. A remote invoke sends a Move (the activation:
// identity + authority + TargetFd/Action + ReturnFd); the value returns as a Landing, not a Move.
type Egress interface {
	Egress(from c.Target, mv c.Move) c.Outcome
}

// ---- the trap interface: the CTU/kernel boundary (08-27 kernel-handler-block.md) ------------------------
//
// The CTU DEFINES this seam and implements NONE of it. On a fault it raises a VECTOR (a Kind), looks up
// Traps[Kind], and hands off a Fault frame; it never loads, persists, or provisions. The KERNEL fills the
// table at boot (Install) and supplies the Handler bodies. This is the IDT contract verbatim: the CPU
// defines the vector space + calling convention; the kernel fills IDT[i]. It MIRRORS the Driver seam (the
// Local arm) — CTU declares the interface, someone else implements: Driver for peers, Handler for the
// kernel. So the CTU is a pure dispatch fabric over two typed seams (Driver, Handler) + inline checks.
//
// There is NO CTU-side loader/pager/provisioner — the CPU has none. An i-cache HIT builds the register file
// INLINE (trivial Context construction); a MISS RAISES the Materialize vector (the page fault) whose handler
// IS the loader. Contrast Verify below, the MMU permission check — an inline mechanical check that DOES live
// on-chip (CTU-side). MMU inline, fault-handler in the kernel: the same cut Intel made.

// Kind is a FAULT VECTOR — the small, DENSE keyspace the CTU traps through (contrast a Target: the large,
// SPARSE keyspace it dispatches to). The CTU raises a Kind; the kernel's Traps[Kind] handles it. The vector
// keyspace is what makes the kernel a DIFFERENT LOOKUP than a peer, not a userspace entity (research #6a).
//
// OPEN (kernel-handler-block.md): the exact vector SET is NOT frozen. Materialize/Remember are grounded;
// ActWake (a contract/guard failure) and Provision (an authorization miss on a novel target) are candidates
// — NOT added until forced. Also OPEN: one gate or two (voluntary syscall vs involuntary fault).
type Kind int

const (
	Materialize Kind = iota // READ fault  — Body not resident: the page-in / loader
	Remember                // WRITE fault — a quiesced, Return context to persist: the page-out / pager
)

// Fault is the TRAP FRAME the CTU builds when it vectors — the uniform hand-off to a Handler. It carries
// what a handler might read; the handler makes state valid and returns an Outcome, and the CTU RE-DRIVES
// (idempotent on Move.ID — page-fault-and-retry) rather than the handler reaching into CTU memory.
//
// The RESULT CHANNEL (Materialize): the handler makes state resident and hands the provisioned Context BACK
// through Result — the CTU reads it, installs it into Cache, and re-drives. This keeps the boundary intact
// (the kernel returns a Context; it does NOT reach into CTU memory). nil for handlers that only make-resident.
//
// OPEN (frame shape still soft — NOT part of the frozen vocabulary): which fields each Kind reads. Set as the
// union of what the handlers want; extend freely as they are written.
type Fault struct {
	Kind   Kind        // which vector was raised
	Move   c.Move      // the faulting move — it IS the request (read faults: Materialize)
	Ctx    *c.Context  // the faulting context (write faults: Remember)
	Under  c.Horizon   // the horizon in force
	Result **c.Context // Materialize result channel: the handler writes the provisioned Context here; CTU installs it
}

// Handler is the kernel-side vector implementation, from the CTU's view: ONE uniform entry the CTU calls
// when it traps. The CTU owns this TYPE; the kernel supplies the bodies (kernel-handler-block.md). One
// method by design — a trap frame in, an Outcome out; the CTU never sees the handler's internals (loader /
// pager / provisioner), only the bit.
type Handler interface {
	Handle(f Fault) c.Outcome
}

// Install is the "write IDT[Kind]" op — the KERNEL calls it at boot to FILL the trap table (alongside driver
// registration at Handshake). The CTU never calls it on itself; it only READS Traps when it vectors. Pure
// boundary plumbing, not a handler.
func (ct *CTU) Install(k Kind, h Handler) {
	if ct.Traps == nil {
		ct.Traps = map[Kind]Handler{}
	}
	ct.Traps[k] = h
}

// Verify is the OFFLINE capability check at ingress — the MMU: sig ok, not expired, authorizes (t, action),
// faults on a violation. No kernel round-trip. This is what keeps the hot path kernel-free.
type Verify interface {
	Verify(cap c.Capability, t c.Target, action c.Name) bool
}

// Interpret is block selection over the CONTEXT's HORIZON: a name provisioned in the horizon → dispatch;
// a name that MISSES (unprovisioned) → FAULT. It answers AUTHORIZATION only; Core.Local answers LOCALITY.
type Interpret interface {
	Interpret(ctx *c.Context, mv *c.Move) (EligibleSet, c.Outcome)
}

type EligibleSet struct{ Targets []c.Target }

// Drivers advertise their mechanisms to the kernel at handshake (enumeration); the CTU never fires one
// except via a horizon.
type Drivers interface {
	Advertise() []DriverDecl
}

// Driver is an installed OS-specific mechanism. Decl is the inert kernel-facing advertisement; Perform is
// reachable only after Interpret established the horizon authorizes it.
type Driver interface {
	Decl() DriverDecl
	Perform(action c.Name, in []c.Observation) (c.Observation, c.Outcome)
}

// DriverDecl is a driver's declaration: the target it IS + the fine-grained actions it defines. Each
// (Target, Action) is one capability the kernel can put in a horizon. Permissions live on the drivers.
type DriverDecl struct {
	Target  c.Target
	Actions []c.Name
}
