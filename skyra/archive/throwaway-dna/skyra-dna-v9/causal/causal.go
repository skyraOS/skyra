// Package causal holds the types that CROSS boundaries in the possibility model —
// the causal state a CTU executes over, and the wire legs that travel between CTUs.
// Leaf package (imports nothing) so ctu/kernel/drivers/entity/fabric can share these
// shapes without an import cycle.
//
// AUTHORITY: skyra/specs/2026-08-23/ (live-possibility.md, possibility-graph.md,
// possibility-residency.md, execution-cycle.md), on the foundation of
// skyra/specs/2026-08-22/ and 2026-08-18/. This is v8's causal package evolved for
// the RUNNING SHAPE of a possibility:
//
//   - a Possibility is a REORDER BUFFER, not one instruction: a graph of Moves over a
//     register file, an in-flight cut index, and a retirement watermark. (live-possibility.md §2)
//   - the unit of execution is the MOVE NODE; a possibility is a grouping of moves
//     under one identity. (possibility-graph.md §1)
//   - the SLICE type is DISSOLVED: what a cut carries on its request leg is a narrow
//     child Possibility. (possibility-residency.md §1)
//   - a cut's wire leg is a TRANSFER; the node stays resident. (possibility-graph.md §8)
//
// See interface-rationale.md and claude-ideas.md for every choice; new terms are
// declared in the rationale.
//
// PRIMARY / SECONDARY split, as in v7/v8:
//
//	PRIMARY    appears directly in an interface signature — it CROSSES a boundary.
//	SECONDARY  a field of a primary type — it ties to an interface THROUGH its parent.
//
// Every type ties to at least one interface. An untied type = a concept not decomposed.
package causal

// ══════════════════════════════ PRIMARY TYPES ════════════════════════════════

// Possibility is the ARCHITECTURAL CONTEXT — the unit of execution AND of transport.
// It is a reorder buffer: a graph of Moves over a register file, the in-flight cut
// index, and the retirement watermark. It carries its own state, so DRIVE needs no
// other argument. It moves between CTUs as a unit — a migration is a context switch.
// Arena is CTU residency metadata, not a field here, so a transported child
// structurally cannot carry its parent's arena.
// live-possibility.md §1–2, possibility-graph.md §1.
// → ctu.CTU.Drive, .Greet (produced), .Think, .Orient; kernel.Kernel.Activate/.Mint
//
//	(produced), .Remember (param); field of Transfer (the crossing child)
type Possibility struct {
	// identity & lineage
	ID      PossibilityID // shared identity across the CTUs that hold its fragments (terms.md)
	Parent  *MoveID       // nil = ROOT; a remote child's one-hop link is the cut that created it
	Horizon Horizon       // authority held by this fragment
	Return  RegKey        // the register whose landing completes this fragment

	// the reorder buffer (live-possibility.md §2, possibility-graph.md §4)
	Moves     []Move              // THE PROGRAM — the compiled continuation as a graph of moves
	Registers map[RegKey]Register // THE REGISTER FILE — named+tagged slots (was: a flat Frame.Scope)
	InFlight  map[MoveID]RegKey   // each issued cut edge → the register its response will fill
	Retired   int                 // the retirement WATERMARK (program order; the precise-exception cut)
	Window    []Potential         // THINK's sliding available signatures (≤5)

	Result *Observation // the returning move's landed result (the kept result at a root)
}

// Move is the NODE of the possibility graph — one execution, one observation out.
// "Beneath the syntax, every line is a MOVE." It is either a LOCAL register read or a
// signed CROSS-CTU cut (its Site). This node IS the provenance node: it accretes a
// Segment as it retires. possibility-graph.md §1,4, continuation-language.md.
// → ctu.Interpreter.Interpret (param); field of Possibility.Moves
type Move struct {
	Position int             // the static TEMPLATE in the engram this instantiates (fixed at ACTIVATE)
	Target   Name            // the instruction this move dispatches — INTERPRET reads this + Sys, never Payload
	Payload  Payload         // OPAQUE per-instruction work; read ONLY by the entity
	Sys      *SysInstruction // OPTIONAL placement narrowing; ⊆ horizon. nil = absent
	Reads    []RegKey        // in-edges: the operand registers it consumes (its causes)
	Writes   RegKey          // out: the single register it produces (SSA: written once per (Name,Tag))
	Contract Contract        // return-type expectation for Writes, checked when the result lands
	State    InstrState      // Waiting → Ready → Issued → Retired
	Site     Site            // Local | Cut{ID,Target} — the ONLY wire-weight edge
	Segment  *Segment        // nil until Retired; then stamped + signed
}

// Transfer is the WIRE LEG of a cut — custody transfer across a trust boundary. A
// resident Move whose Site is Cut is realized on the wire as two Transfers: a REQUEST
// leg that pushes the narrow child Possibility, and a RESPONSE leg that returns an
// Observation at the join. It REPLACES v8's transport Move; the Slice is dissolved
// into the child Possibility. possibility-graph.md §8, possibility-residency.md §1.
// → ctu.CTU.Move/.Greet/.Recognize, fabric.Fabric.Approach, fabric.Ingress.Arrive
type Transfer struct {
	ID       MoveID       // the cut-edge identity (wire token; resolved to a local possibility+register)
	Header   Header       // Request (push a child) | Response (return an observation)
	Target   Target       // ENVELOPE — destination endpoint + name
	Horizon  Horizon      // ENVELOPE — authority bounding the cut, ATTENUATED to it
	Child    *Possibility // REQUEST leg: the narrow child possibility (the dissolved Slice). nil on a response.
	Returned *Observation // RESPONSE leg: the result at the join. nil on a request.
	// NO third cargo mode: v8's push/pull/return trio collapses to two legs. A pull never
	// crosses — RunHere is local (execution-cycle.md §3). Child XOR Returned, one per Header.
}

// Outcome is the three-outcome result — S/F plus the NOT-YET a bool cannot carry.
// move.md. (Name is Claude's; the values are doc terms.)
// → returned by almost every fallible crossing
type Outcome int

const (
	Success Outcome = iota // custody crossed / the move retired
	Failure                // custody did not cross
	NotYet                 // operands / resource / a remote leg unresolved — nothing happened, nothing broke.
	//                        A local name MISS is signalled here too: it is the FAULT (execution-cycle.md §3).
)

// Frontier is the OUTSTANDING FRONTIER — this fragment's in-flight cut edges. It is
// DERIVED from Possibility.InFlight (its keys), not stored, so it cannot skew from the
// truth. The whole-possibility view composes by following it. ctu/orient.md.
// → returned by ctu.CTU.Orient
type Frontier struct {
	InFlight []MoveID // the cut edges this fragment is outstanding on
}

// Selection is what SCHEDULE returns — placement within the eligible set: PUSH vs PULL.
// It never touches the graph; local and remote are the same instruction. decode-and-run.md,
// ctu/possibility.md, execution-cycle.md §3.
// → returned by ctu.Scheduler.Schedule
type Selection interface{ selection() }

// RunHere is PULL: dispatch on a local resource (a host BELIEVED, an entity STAMPED+CHECKED,
// or the KERNEL — which is a TRAP, not a run; see execution-cycle.md §3).
type RunHere struct{ Resource Resource }

func (RunHere) selection() {}

// SendTo is PUSH: forward the child possibility to this endpoint's CTU.
type SendTo struct{ Endpoint Target }

func (SendTo) selection() {}

// EligibleSet is what INTERPRET produces — resources a move may run on, filtered by the
// quadrant and any Sys so it NEVER spans the forbidden diagonal. dispatch.md, execution-cycle.md §3.
// → ctu.Interpreter.Interpret (return), ctu.Scheduler.Schedule (param)
type EligibleSet struct{ Resources []Resource }

// Resource is one thing the CTU can dispatch to; its concrete type says WHICH TABLE
// (block) supplied it — the trust decision. dispatch.md: "named, not flagged."
type Resource interface{ ResourceName() Name }

// HostResource came from the HOST block (system, trusted). Output BELIEVED.
type HostResource struct{ Name Name }

func (r HostResource) ResourceName() Name { return r.Name }

// EntityResource came from the ENTITY block (user, untrusted). Output STAMPED + CHECKED.
type EntityResource struct{ Name Name }

func (r EntityResource) ResourceName() Name { return r.Name }

// KernelResource NAMES the kernel. It is NOT a block — dispatching it is a TRAP to
// kernel mode (privilege), not a run. execution-cycle.md §3.
type KernelResource struct{ Name Name }

func (r KernelResource) ResourceName() Name { return r.Name }

// RemoteResource is reached by forwarding the child to another CTU. Placement, not a
// trust or privilege claim.
type RemoteResource struct{ Target Target }

func (r RemoteResource) ResourceName() Name { return r.Target.Name }

// Content is RAW entity output — untrusted, unstamped. The CTU turns it into an
// Observation. decode-and-run.md.
// → entity.Entity.Run (return)
type Content struct{ Bytes []byte }

// Observation is a returned value WITH provenance, checked against a Contract when it
// lands in its register. Its provenance crosses a cut as a signed Segment.
// continuation-language.md, fabric/signed-provenance.md.
// → ctu.Scheduler.Schedule (operands), kernel.Remember (keep); field of Transfer.Returned, Register
type Observation struct {
	Content Content
	Segment Segment // stamped + signed by the CTU, never constructed by the entity
}

// Goal is what THINK is given. think.md.
// → ctu.CTU.Think
type Goal struct{ Statement string }

// Potential is a proposed child signature from THINK — contract-shaped, at most five.
// think.md.
// → ctu.CTU.Think (return); field of Possibility.Window
type Potential struct{ Signature string }

// Horizon is the authority — the token. ENVELOPE; narrowing-only; attenuable to a
// slice; a SIGNATURE CHAIN (macaroon/biscuit), never an unsigned {Root,Caveats}.
// horizon-and-root.md, ctu/possibility.md.
// → ctu.CTU.Bound, kernel.Activate/.Mint/.Define, fabric.Authority/.HorizonMinter;
//
//	field of Transfer, Possibility
type Horizon struct {
	Root    Name        // which account this descends from
	Caveats []Caveat    // ordered, append-only; each one NARROWS
	Keys    [][]byte    // proof-key chain, one public key per signature
	Chain   []Signature // root attestation followed by one signature per caveat
	Proof   []byte      // current attenuation proof; never the root minting key
}

// Caveat narrows a Horizon. Append-only. horizon-and-root.md.
// → ctu.CTU.Bound, fabric.Authority.Attenuate
type Caveat struct{ Constraint string }

// Name is an entry in the kernel-owned namespace: name → meaning. boundaries.md.
// → kernel.Activate/.Mint, drivers.Driver.Activate, fabric.HorizonMinter.Mint
type Name struct{ Value string }

// Engram is durable behaviour = child contracts + compiled continuation.
// Body closes the interface set's largest gap with explicit MoveTemplates;
// Materialize instantiates them at a Tag. primitive-details.md,
// live-possibility.md §8, possibility-graph.md §8.
// → drivers.Driver.Activate (return)
type Engram struct {
	Name      Name
	Contracts []Contract
	Body      []MoveTemplate // compiled continuation materialized by ACTIVATE
	Return    RegKey
}

// MoveTemplate is an engram's static compiled-continuation position. Materialize
// turns it into a runtime Move at a Tag. TEMPLATE and POSITION are doc terms from
// possibility-graph.md; this is the previously missing ENGRAM BODY.
type MoveTemplate struct {
	Position int
	Target   Name
	Payload  Payload
	Sys      *SysInstruction
	Reads    []Name
	Writes   Name
	Contract Contract
	Site     Site
}

// Materialize instantiates a compiled continuation at one occurrence tag.
func Materialize(body []MoveTemplate, tag Tag) ([]Move, map[RegKey]Register) {
	moves := make([]Move, 0, len(body))
	registers := make(map[RegKey]Register)
	for _, position := range body {
		reads := make([]RegKey, len(position.Reads))
		for i, name := range position.Reads {
			reads[i] = RegKey{Name: name, Tag: tag}
			if _, exists := registers[reads[i]]; !exists {
				registers[reads[i]] = Register{}
			}
		}
		writes := RegKey{Name: position.Writes, Tag: tag}
		registers[writes] = Register{Contract: position.Contract}
		site := position.Site
		if site == nil {
			site = Local{}
		}
		moves = append(moves, Move{
			Position: position.Position, Target: position.Target, Payload: position.Payload,
			Sys: position.Sys, Reads: reads, Writes: writes, Contract: position.Contract,
			State: Waiting, Site: site,
		})
	}
	return moves, registers
}

// Cone is everything a kept Observation depended on — REMEMBER's output, GATHERED across
// CTUs as signed segments. kernel/remember-gather.md.
// → kernel.Remember (return), drivers.Driver.Remember (param)
type Cone struct{ Segments []Segment }

// Account is the durable identity a Pin unlocks; owns a namespace + durable graph.
// horizon-and-root.md.
// → kernel.Define (return)
type Account struct{ Root Name }

// Pin authenticates an Account and unlocks its root Horizon (and origination). horizon-and-root.md.
// → kernel.Define, kernel.Mint
type Pin struct{ Secret string }

// Segment is a SIGNED PROVENANCE SEGMENT — the provenance one CTU produced, signed so a
// downstream CTU can RECOGNIZE it without trusting the sender's word. fabric/signed-provenance.md.
// → fabric.Custody.Sign/.Verify, fabric.Fabric.Fetch (return), fabric.SegmentVerifier.Verify;
//
//	field of Observation, Move
type Segment struct {
	Provenance Provenance // the causal history this move produced
	By         Name       // which CTU produced and signed it
	Signature  Signature  // makes it unforgeable across a trust boundary
	Ref        SegmentRef // content-address — how the gather fetches and dedups it
}

// SegmentRef is a CONTENT-ADDRESS handle to a Segment — what makes the gather tractable.
// implementation.md, kernel/remember-gather.md.
// → fabric.Fabric.Fetch (param); field of Segment, Provenance.Causes
type SegmentRef struct{ Hash string }

// ═════════════════════════════ SECONDARY TYPES ═══════════════════════════════
// Each is a FIELD of a primary type; it ties to an interface through its parent.

// Register is one slot of the register file: a typed destination with a landed value.
// live-possibility.md §2.
// → field of Possibility.Registers
type Register struct {
	Contract Contract     // the expectation the landing move's Observation is checked against
	Value    *Observation // nil ⇒ not yet landed
}

// RegKey keys the register file and the Reads/Writes edges: a name plus an OCCURRENCE
// tag. Tag is zero-valued until loops exist; keying by (Name,Tag) now is the recommended
// forward-compatible shape. execution-cycle.md §5, possibility-graph.md §8.
// → key of Possibility.Registers; field of Move.Reads/.Writes
type RegKey struct {
	Name Name
	Tag  Tag
}

// Tag is the occurrence label a loop increments — a fresh instantiation of the body's
// positions per iteration. Ctx is the loop-nest context; Iter the iteration index.
// A loop UNFOLDS by minting tags; it is not a back-edge. execution-cycle.md §5.
// → field of RegKey
type Tag struct {
	Ctx  uint64 // loop-nest context (a stack, once recursion exists — §open)
	Iter uint32 // iteration index within Ctx
}

// InstrState is a move's lifecycle through the reorder buffer. live-possibility.md §2.
// → field of Move
type InstrState int

const (
	Waiting InstrState = iota // some Reads register is unlanded
	Ready                     // all Reads landed; awaiting SCHEDULE
	Issued                    // dispatched: a local run in progress, or a cut on the wire
	Retired                   // landed + signed; below the watermark
)

// Site is the one real structural distinction on an edge: a LOCAL register read, or a
// signed CROSS-CTU cut. The MoveID, the Target, and (via the node's Segment) the
// signature live on the Cut. possibility-graph.md §3–4.
// → field of Move
type Site interface{ site() }

// Local is a register read within one CTU — no wire token, no signature.
type Local struct{}

func (Local) site() {}

// Cut is a cross-CTU custody transfer: the node is outstanding on edge ID toward Target.
// Its two legs travel as Transfers referencing ID. possibility-graph.md §3.
type Cut struct {
	ID     MoveID
	Target Target
}

func (Cut) site() {}

// PossibilityID is the LOCAL identity of a live possibility. It crosses only inside a
// forwarded child (so a fragment knows its possibility); it is NOT on a response — the
// issuing CTU resolves its own possibility from the cut edge. terms.md, signed-provenance.md.
// → field of Possibility
type PossibilityID struct{ Value string }

// MoveID is the CUT-EDGE identity: the wire token that keys the owning possibility's
// outstanding edge at RECOGNIZE. Correlation is dissolved into it. terms.md, changes.md.
// SEAM (carried from v8): logical identity is (possibility identity, edge), but
// signed-provenance.md keeps PossibilityID off the wire — so the TOKEN is edge-scoped and
// resolved to a possibility locally.
// → field of Move.Site(Cut), Transfer, Frontier; key of Possibility.InFlight; Parent
type MoveID struct{ Value string }

// Arena is the arenaID — residency metadata held by the CTU, not crossing state.
// Moving it out of Possibility resolves the one-type versus no-arena-on-wire
// contradiction found by the build; see boundary-crossings.md.
// possibility-residency.md §1–2.
// → field of ctu.residency (never a field of Transfer.Child)
type Arena struct{ Root Name }

// Signature makes a Segment unforgeable across a trust boundary. "signed" is the doc's;
// Signature is the mechanism noun. fabric/signed-provenance.md.
// → field of Segment (Custody.Sign returns the completed Segment)
type Signature struct{ Bytes []byte }

// Header is a Transfer's leg — two values → exactly two receiving ports. move.md.
// → field of Transfer
type Header int

const (
	Request  Header = iota // arrives at GREET — pushes a child possibility (continuation)
	Response               // arrives at RECOGNIZE — returns an observation at a join
)

// Target is the address a cut names: transport endpoint + within-runtime name.
// move.md, traffic-routing.md.
// → field of Transfer, Move.Site(Cut), causal.SendTo
type Target struct {
	Endpoint string // runtime transport endpoint (which CTU)
	Name     Name   // the within-runtime name
}

// SysInstruction is the OPTIONAL system-space placement narrowing — narrows the eligible
// set; ⊆ horizon. Read by INTERPRET, never the entity. move.md, execution-cycle.md §3.
// → field of Move (as *SysInstruction)
type SysInstruction struct{ Constraint string }

// Payload is OPAQUE — carried, read ONLY by the entity. move.md.
// → field of Move; entity.Entity.Run
type Payload struct{ Body []byte }

// Provenance is what the CTU stamps onto a move's result: which resource/invocation
// produced it, and its causal antecedents by content-address. The entity cannot forge
// it; a Segment signs it to cross a cut. decode-and-run.md, fabric/signed-provenance.md.
// → field of Segment
type Provenance struct {
	Resource   Name
	Invocation string
	Causes     []SegmentRef // antecedents (the Reads' producers' segments) — the lineage the gather walks
}

// Contract is the return-type expectation RECOGNIZE (or a local landing) checks an
// Observation against. continuation-language.md.
// → field of Move, Register, Possibility.Return, Engram
type Contract struct{ Name Name }
