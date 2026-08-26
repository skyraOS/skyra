// Package causal holds the types that cross boundaries in v11.
//
// FRESH BUILD. Nothing is imported from prior generations — re-derived from the v11 specs so the
// SEAMS show. The v11 focus is the EXECUTION MODEL: the two paths (cold/THINK, hot/ACT), the
// Body/Context split, the Move barrier, and the capability-as-horizon.
//
// AUTHORITY: skyra/specs/2026-08-24/skyra-dna-v11/ — research/ (cpu-execution-model, think-act-*,
// move-and-the-kernel-barrier, capability-table-and-horizon, authorization-and-registration),
// ownership.md, bootstrap/. Docs win.
package causal

// Outcome is the total result of a step. FAULT = a name resolved to nothing PROVISIONED (a page fault
// → cross to the kernel). NOTYET = operands / a parked cut not yet landed. (capability-table-and-horizon)
type Outcome int

const (
	NotYet Outcome = iota
	Success
	Failure
	Fault
)

// ---- identity & addressing: entity → host → address (bootstrap A) ----------

type EntityID struct{ Value string }
type HostID struct{ Value string }
type Address struct{ Value string } // VOLATILE, host-resolved; never stored on the entity.

// Entity is WHO — a userspace identity. It points to its HOME host; the address is resolved by
// host(entity) at the host, locally (the home-registry / HLR pattern). The human is an Entity too —
// structurally uniform, special only in authority. (bootstrap/sequence.md A)
type Entity struct {
	ID   EntityID
	Home HostID // the home host that resolves this entity's address
}

type Name struct{ Value string }

// Target names a move's destination: an entity, resolved via its host, by name.
type Target struct {
	Entity EntityID
	Host   HostID
	Name   Name
}

// ---- capability (macaroon-shaped) = the unit of a horizon --------------------

// Capability is a macaroon: what it authorizes + a chain of NARROWING caveats + an offline-verifiable
// signature. It IS a grant. The user HOLDS it and PRESENTS it on the hot path; the CTU verifies it
// offline (no kernel). Minted only by the kernel. (authorization-and-registration; think-act-paths)
type Capability struct {
	Target  Target
	Caveats []Caveat // attenuation — narrows only (Bound).
	Sig     Signature
}

type Caveat struct{ Constraint string }
type Signature struct{ Bytes []byte }

// Horizon is the PROVISIONED VIEW — the capabilities a Context carries. D4: Horizon = []Grant; here a
// grant is a Capability. It is a view of the kernel's table, carried with the work, not held by the CTU.
type Horizon struct{ Grants []Capability }

// ---- Body (compiled) vs Context (live) — the CPU split -----------------------

// Body is the ENGRAM: a compiled move graph. STATIC, content-addressed, shareable, read-only. THINK's
// output; what REMEMBER caches. The "compiled artifact." (cpu-execution-model; think-act-paths)
type Body struct {
	ID     BodyID         // content address (digest)
	Moves  []MoveTemplate // the compiled dataflow graph (explicit deps — a dataflow machine, not von Neumann)
	Return RegKey         // the exit register
}

type BodyID struct{ Value string }

// Context is the live POSSIBILITY: register file + in-flight fds + retirement watermark + the carried
// horizon. MUTABLE, one per run. The "process." Two runs of one Body = two Contexts.
type Context struct {
	Body      *Body
	Registers map[RegKey]Register
	InFlight  map[CutID]RegKey // the FD TABLE: cut-ID (an fd) → the register it fills (per-possibility)
	Retired   int              // retirement watermark
	Horizon   Horizon          // carried; the provisioned view
	Result    *Observation
	ReturnTo  *Target // server-side backward pointer (set when this Context is a resident request)
}

// ---- moves -------------------------------------------------------------------

// MoveTemplate is a node in the compiled Body — static. Contract is the RUNNABLE half THINK authors
// (target + shape); the input/return TYPE is negotiated at ACT. (think-act-signature)
type MoveTemplate struct {
	Target   Target
	Reads    []RegKey
	Writes   RegKey
	Contract Contract
	Return   bool // does this move expect a return? (structural — the block participation switch)
}

// Move is a live move in a Context — a MoveTemplate plus runtime state.
type Move struct {
	Template MoveTemplate
	State    MoveState
	Site     Site
}

type MoveState int

const (
	Waiting MoveState = iota
	Ready
	Issued
	Retired
)

// Transfer is the wire form crossing the barrier (ingress/egress). A Request PUSHES a child Context; a
// Response returns an observation. Capability is the carried horizon the user presents (hot path).
type Transfer struct {
	ID         CutID       // the fd
	Header     Header      // Request | Response
	Target     Target
	Capability Capability   // carried horizon; verified offline at the CTU
	Child      *Context     // Request leg (nil on Response)
	Returned   *Observation // Response leg (nil on Request)
	ReturnTo   Target       // Request leg: the return route (server-side backward pointer)
}

type Header int

const (
	Request Header = iota
	Response
)

type Register struct {
	Contract Contract
	Value    *Observation // nil ⇒ not yet landed
}

type Observation struct {
	Content Content
	Segment Segment
}

type Content struct{ Bytes []byte }
type Segment struct{ By Name }

// Contract is the judgeable return-type expectation. (think.md)
type Contract struct{ Name Name }

type RegKey struct {
	Name Name
	Tag  Tag
}
type Tag struct{ Iter int }
type CutID struct{ Value string } // an fd

type Site interface{ site() }

type Local struct{}

func (Local) site() {}

type Cut struct {
	ID     CutID
	Target Target
}

func (Cut) site() {}
