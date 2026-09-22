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

// Entity is WHO — a userspace identity. It is MULTI-HOMED: it points to one or more home hosts, any of
// which can resolve its address via host(entity) locally (the home-registry / HLR pattern). The human
// is an Entity too — structurally uniform, special only in authority. (bootstrap/sequence.md A)
type Entity struct {
	ID    EntityID
	Homes []HostID // the home hosts that can resolve this entity's address; a Target picks one.
}

type Name struct{ Value string }

// Target names a move's destination: an entity, resolved via its host, by name.
type Target struct {
	Entity EntityID
	Host   HostID
	Name   Name
}

// ---- capability (macaroon-shaped) = the unit of a horizon --------------------

// Capability is FINE-GRAINED and TIMESTAMPED: it authorizes ONE action on ONE target, until ExpiresAt.
// (Target, Action) is DEFINED by the target's driver (ctu.DriverDecl — permissions live on the drivers).
// The user HOLDS it and PRESENTS it; verified OFFLINE (sig + not-expired + target/action match) at
// ingress. Minted only by the kernel. NO CAVEATS — attenuation is by REMOVAL, not narrowing.
type Capability struct {
	Target    Target
	Action    Name  // the fine-grained action, declared by the target's driver. (Target,Action) = one capability.
	ExpiresAt int64 // offline lease; removal handles the rest.
	Sig       Signature
}

type Signature struct{ Bytes []byte }

// Horizon is the PROVISIONED VIEW — the fine-grained capabilities a Context carries (D4: Horizon =
// []Capability). A view of the kernel's table, carried with the work, not held by the CTU.
// ATTENUATION IS BY REMOVAL: a child/cut gets a SUBSET — never a narrowed grant. only the kernel widens.
type Horizon struct{ Capabilities []Capability }

// ---- Body (compiled) vs Context (live) — the CPU split -----------------------

// Body is the ENGRAM: a compiled move graph. STATIC, content-addressed, shareable, read-only. THINK's
// output; what REMEMBER caches. The "compiled artifact." (cpu-execution-model; think-act-paths)
type Body struct {
	ID     BodyID         // content address (digest)
	Moves  []MoveTemplate // the compiled dataflow graph (explicit deps — a dataflow machine, not von Neumann)
	Return RegKey         // the exit register
}

type BodyID struct{ Value string }

// Context is the live POSSIBILITY: register file + the carried horizon + result. MUTABLE, one per run.
// The "process." It holds NO move state — moves live in the CTU's global WorkList (loose coupling), so a
// Context is pure data. Its outstanding children (fds) are just the cut subset of Work[Issued] filtered by
// this Possibility. It is LIVE STATE until REMEMBERed (persisted) — that persist IS its retirement; there
// is NO per-move watermark. Materialize brings a Context IN, Remember writes it OUT — symmetric. Two runs
// of one Body = two Contexts.
type Context struct {
	ID        PossibilityID
	Body      BodyID // by CONTENT ADDRESS — the CTU resolves it from its body cache (an i-cache). the wire carries this reference, not the whole graph.
	Registers map[RegKey]Register
	Horizon   Horizon // carried; the provisioned view — what it CAN reach (Work[Issued] is what it IS reaching)
	Result    *Observation
	ReturnTo  *Route // server-side backward pointer: destination + the parent's fd
}

// PossibilityID identifies a live Context (a possibility) — used to wire-scope a MoveID.
type PossibilityID struct{ Value string }

// MoveID identifies a move by the register it WRITES. Each move writes exactly one register and RegKeys
// are unique (Name, Tag), so Writes IS the move's identity — and its fd, since the response lands in
// Writes. Scoped by its Possibility for the wire. A cut IS a move; this is its wire token. The WorkList
// keys on it (it IS the fd); fire-and-forget moves (Return=false) never park in Issued.
type MoveID struct {
	Possibility PossibilityID
	Writes      RegKey
}

// Route is the server-side backward pointer carried by a request. To locates the caller; Fd identifies
// the exact parent move whose register the response fills. A child echoes Fd as the response Move.ID.
type Route struct {
	To Target
	Fd MoveID
}

// ---- moves -------------------------------------------------------------------

// MoveTemplate is a node in the compiled Body — static. Contract is the RUNNABLE half THINK authors
// (target + shape); the input/return TYPE is negotiated at ACT. (think-act-signature)
type MoveTemplate struct {
	Target   Target
	Action   Name // the action on the target — checked against the horizon as (Target, Action). driver-declared.
	Reads    []RegKey
	Writes   RegKey
	Contract Contract
	Return   bool // does this move expect a return? (structural — the block participation switch)
}

// Move is the unit of work AND the unit that crosses the barrier — the old Transfer struct folded in
// (a move IS the wire form; it already knows its Target and its Writes/fd). It is a MoveTemplate + runtime
// state + the three WIRE fields (Header/Returned/ReturnTo). Whether it ran locally or is an outstanding
// cut is NOT a stored flag: an Issued move is in flight — a cut parks in Issued awaiting a Response over
// the wire, a local call pops within the pass. A move is transferred IN via CTU.Transfer. A Request carries
// the lean Context seed needed to materialize its live possibility; resident moves read that state from Cache.
type Move struct {
	ID       MoveID // {Possibility, Writes} — its identity and its fd (where a Response lands)
	Template MoveTemplate
	State    MoveState
	Header   Header       // Request | Response — which leg is crossing
	Returned *Observation // Response leg: the landed observation (nil on a Request / a resident move)
	Seed     *Context     // Request leg: Body + Horizon + seed Registers + ReturnTo; nil on a Response
}

// MoveState — a move's runtime state, and the KEY of the WorkList: a state change IS a move between
// buckets. There is NO Retired: a done move POPS from the list (out-of-order, no watermark). Retirement
// is a CONTEXT event (Remember), not a per-move one.
type MoveState int

const (
	Waiting MoveState = iota // operands not all landed
	Ready                    // operands satisfied, not yet dispatched — the RUN QUEUE bucket
	Issued                   // dispatched — running local, or a parked cut awaiting a Response. when its result lands, the move POPS.
)

// WorkList is the CTU's global move work-list, bucketed by STATE across ALL contexts (moves are loosely
// coupled — MoveID.Possibility says which Context to read/write). A state change is a move between
// buckets; the Ready bucket is the run queue. Keyed within a bucket by MoveID for O(1) landing. The one
// Transition core steps this. (ctu: Core.Transition)
type WorkList map[MoveState]map[MoveID]Move

// Cache is the CTU's resident live possibilities — the RAM tier, keyed by PossibilityID. CTU-owned:
// Materialize writes into it, a done Context is Remembered out of it, eviction reclaims it. (context map)
type Cache map[PossibilityID]*Context

// BodyCache is the i-cache: content-addressed compiled bodies. A MISS on a needed BodyID is the page
// fault that reaches the kernel (ACTIVATE from disk / MINT fresh via THINK).
type BodyCache map[BodyID]Body

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
