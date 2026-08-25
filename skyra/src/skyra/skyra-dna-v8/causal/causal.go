// Package causal holds the types that CROSS boundaries in the DISTRIBUTED
// possibility model — the causal state a CTU executes over, and that travels
// between CTUs. Leaf package (imports nothing) so ctu/kernel/drivers/entity/
// fabric can share these shapes without an import cycle.
//
// AUTHORITY: skyra/specs/2026-08-22/ (component-organized), on the foundation of
// skyra/specs/2026-08-18/. This is the v7 causal package evolved for one thread:
// a possibility is PARTITIONED across CTUs, a MOVE is the CUT between partitions,
// each CTU SIGNS the provenance segment it executed, REMEMBER GATHERS the cone to
// the root, and ORIENT reads a distributed FRONTIER. See interface-rationale.md
// and claude-ideas.md for every choice; new terms are declared in the rationale.
//
// PRIMARY / SECONDARY split, as in v7:
//
//	PRIMARY    appears directly in an interface signature — it CROSSES a boundary.
//	SECONDARY  a field of a primary type — it ties to an interface THROUGH its parent.
//
// Every type ties to at least one interface. An untied type = a concept not
// decomposed.
package causal

// ══════════════════════════════ PRIMARY TYPES ════════════════════════════════

// Possibility is the UNIT OF EXECUTION. "Give the CTU a possibility; it executes
// it" — it carries its own state, so run needs no other argument.
// ctu/possibility.md.
// → ctu.CTU.Execute, kernel.Kernel.Activate/.Remember, ctu.CTU.Greet (produced)
type Possibility struct {
	ID       PossibilityID // shared identity across the CTUs that hold its fragments (terms.md)
	Parent   *MoveID       // nil = ROOT; a remote child's one-hop link is the cut that created it
	Arena    Arena         // arenaID — the DENORMALIZED back-reference to the owning root
	Name     Name          // instruction this fragment executes
	Horizon  Horizon       // authority held by this fragment
	Payload  Payload       // opaque work carried by this fragment
	Frame    Frame         // observations in scope + the return contract
	Frontier Frontier      // this fragment's in-flight cut edges — ORIENT's local read
	Result   *Observation  // exactly one admitted result when this fragment completes
}

// Slice is a PARTITION of a possibility — the forwarded fragment that travels to
// the CTU that must resolve a dependency. "Ship the slice, not the whole."
// ctu/possibility.md, causal/terms.md.
// → field of Move (the request leg carries it); resumed as a child Possibility at GREET
type Slice struct {
	Of      PossibilityID // which possibility this fragment belongs to (crosses on the FORWARD)
	Frame   Frame         // the scope this fragment needs
	Horizon Horizon       // ATTENUATED to this slice — least-knowledge (ctu/possibility.md)
	Payload Payload       // the opaque work for the entity
	// NO Arena: the home arenaID does NOT travel with the slice (ctu/possibility.md — "a forwarded
	// slice never learns it"). The receiver gives the resumed child a LOCAL scope at GREET; the home
	// arena is rejoined only at the return, via MoveID. (Was a field; removed — see claude-ideas.md.)
}

// Move is custody transfer — the only transport — and, across CTUs, the CUT
// between two partitions. move.md, fabric/signed-provenance.md.
// → ctu.CTU.Move/.Greet/.Recognize, ctu.Interpreter.Interpret, fabric.Fabric.Approach
type Move struct {
	ID             MoveID          // the cut-edge identity (wire token; maps to a local possibility+edge)
	Header         Header          // Request (forward/pull) | Response (return at a join)
	Target         Target          // ENVELOPE — destination endpoint + name
	Horizon        Horizon         // ENVELOPE — authority bounding the cut, attenuated
	SysInstruction *SysInstruction // ENVELOPE, CTU-only — OPTIONAL placement narrowing (nil = absent)
	Slice          *Slice          // REQUEST-leg cargo (PUSH): the forwarded fragment. nil for a pull.
	Payload        Payload         // OPAQUE — a pull request's instruction content, entity-only
	Returned       *Observation    // RESPONSE-leg cargo: the result at the join. nil on a request.
	// SEAM: three cargo modes on one struct (push Slice / pull Payload / response Returned),
	// mutually exclusive by leg. A Cargo sum type is the cleaner model — claude-ideas.md.
}

// Outcome is the three-outcome result — S/F plus the NOT-YET a bool cannot carry.
// move.md. (Name is Claude's; the values are doc terms — claude-ideas.md.)
// → returned by almost every fallible crossing
type Outcome int

const (
	Success Outcome = iota // custody crossed
	Failure                // custody did not cross
	NotYet                 // operands / resource / a remote leg unresolved — nothing happened, nothing broke
)

// Frontier is the OUTSTANDING FRONTIER — this fragment's in-flight cut edges. The
// whole-possibility view composes by following them (ORIENT is distributed).
// ctu/orient.md, causal/terms.md.
// → returned by ctu.CTU.Orient; field of Possibility
type Frontier struct {
	InFlight []MoveID // the cut edges this fragment is outstanding on
}

// Segment is a SIGNED PROVENANCE SEGMENT — the provenance one CTU produced, signed
// so a downstream CTU can RECOGNIZE it without trusting the sender's word.
// fabric/signed-provenance.md.
// → fabric.Custody.Sign/.Verify, fabric.Fabric.Fetch (return), kernel gather
type Segment struct {
	Provenance Provenance // the causal history this CTU produced
	By         Name       // which CTU produced and signed it
	Signature  Signature  // makes it unforgeable across a trust boundary
	Ref        SegmentRef // content-address — how the gather fetches and dedups it
}

// SegmentRef is a CONTENT-ADDRESS handle to a Segment. Content-addressing is what
// makes the gather tractable (dedup, skip-if-durable). implementation.md,
// kernel/remember-gather.md.
// → fabric.Fabric.Fetch (param); field of Segment
type SegmentRef struct{ Hash string }

// Selection is what SCHEDULE returns — the WHERE decision, which now includes
// PUSH vs PULL. decode-and-run.md ("selection"), ctu/possibility.md ("push or pull").
// → returned by ctu.Scheduler.Schedule
type Selection interface{ selection() }

// RunHere is PULL: run on a local eligible resource.
type RunHere struct{ Resource Resource }

func (RunHere) selection() {}

// SendTo is PUSH: forward the slice to this endpoint's CTU.
type SendTo struct{ Endpoint Target }

func (SendTo) selection() {}

// EligibleSet is what INTERPRET produces — resources a MOVE may run on, filtered by
// the quadrant and any SysInstruction. decode-and-run.md, dispatch.md.
// → ctu.Interpreter.Interpret (return), ctu.Scheduler.Schedule (param)
type EligibleSet struct{ Resources []Resource }

// Resource is one thing the CTU can dispatch to; its concrete type says which table
// supplied it. dispatch.md: "named, not flagged." (Carried from v7.)
type Resource interface{ ResourceName() Name }

// HostResource came from the trusted host table.
type HostResource struct{ Name Name }

func (r HostResource) ResourceName() Name { return r.Name }

// EntityResource came from the untrusted entity table.
type EntityResource struct{ Name Name }

func (r EntityResource) ResourceName() Name { return r.Name }

// KernelResource is the one named privileged executor.
type KernelResource struct{ Name Name }

func (r KernelResource) ResourceName() Name { return r.Name }

// RemoteResource is reached by forwarding the slice to another CTU. It is a
// placement result, not a trust or privilege claim.
type RemoteResource struct{ Target Target }

func (r RemoteResource) ResourceName() Name { return r.Target.Name }

// Content is RAW entity output — untrusted, unstamped. The CTU turns it into an
// Observation. decode-and-run.md.
// → entity.Entity.Run (return)
type Content struct{ Bytes []byte }

// Observation is a returned value WITH provenance, checked against a Contract at
// RECOGNIZE. continuation-language.md. In the distributed model its provenance is
// carried as a signed Segment across a cut. fabric/signed-provenance.md.
// → ctu.CTU.Recognize, ctu.Scheduler.Schedule (operands), kernel.Remember, Move.Returned
type Observation struct {
	Content Content
	Segment Segment // stamped + signed by the CTU, never constructed by the entity
}

// Goal is what THINK is given alongside the frame. think.md.
// → ctu.CTU.Think
type Goal struct{ Statement string }

// Potential is a proposed child signature from THINK — contract-shaped, at most five.
// think.md.
// → ctu.CTU.Think (return)
type Potential struct{ Signature string }

// Horizon is the authority — the token. ENVELOPE; narrowing-only; attenuable to a
// slice. horizon-and-root.md, ctu/possibility.md (least-knowledge).
// → ctu.CTU.Bound, kernel.Activate, kernel.Define; field of Move, Slice
type Horizon struct {
	Root    Name        // which account this descends from
	Caveats []Caveat    // ordered, append-only; each one NARROWS
	Keys    [][]byte    // proof-key chain, one public key per signature
	Chain   []Signature // root attestation followed by one signature per caveat
	Proof   []byte      // current attenuation proof; never the root minting key
}

// Caveat narrows a Horizon. Append-only. horizon-and-root.md.
// → ctu.CTU.Bound
type Caveat struct{ Constraint string }

// Name is an entry in the kernel-owned namespace: name → meaning. boundaries.md.
// → kernel.Activate, drivers.Driver.Activate
type Name struct{ Value string }

// Engram is durable behaviour = child contracts + compiled continuation.
// primitive-details.md.
// → drivers.Driver.Activate (return)
type Engram struct {
	Name      Name
	Contracts []Contract
}

// Cone is everything a kept Observation depended on — what REMEMBER writes. In the
// distributed model it is GATHERED across CTUs. kernel/remember-gather.md.
// → kernel.Remember (return), drivers.Driver.Remember (param)
type Cone struct{ Segments []Segment } // the signed segments the gather assembled

// Account is the durable identity a Pin unlocks; owns a namespace + durable graph.
// horizon-and-root.md.
// → kernel.Define (return)
type Account struct{ Root Name }

// Pin authenticates an Account and unlocks its root Horizon. horizon-and-root.md.
// → kernel.Define
type Pin struct{ Secret string }

// ═════════════════════════════ SECONDARY TYPES ═══════════════════════════════
// Each is a FIELD of a primary type; it ties to an interface through its parent.

// PossibilityID is the LOCAL identity of a live possibility. It crosses only inside
// a forwarded Slice (so a fragment knows its possibility); it is NOT on a response —
// the issuing CTU resolves its own possibility from the cut edge. terms.md,
// fabric/signed-provenance.md.
// → field of Possibility, Slice
type PossibilityID struct{ Value string }

// MoveID is the CUT-EDGE identity: the wire token that keys the owning possibility's
// outstanding edge at RECOGNIZE. terms.md ("the per-edge form ... the v7 build's
// MoveID"); Correlation is dissolved into it (changes/changes.md).
// SEAM: terms.md's logical identity is (possibility identity, edge), but
// signed-provenance.md keeps PossibilityID off the wire — so the TOKEN is
// edge-scoped and resolved to a possibility locally.
// → field of Move, Frontier
type MoveID struct{ Value string }

// Arena is the arenaID — a possibility's back-reference to its owning ROOT. It is
// INERT for commit (a non-root names no arena FOR COMMIT — REMEMBER stays
// structurally root-only). It does NOT travel with a Slice (ctu/possibility.md).
// ctu/possibility.md.
// → field of Possibility (only — not Slice)
type Arena struct{ Root Name }

// Signature makes a Segment unforgeable across a trust boundary. The concept
// "signed" is the doc's; Signature is the mechanism noun (declared in rationale).
// fabric/signed-provenance.md.
// → field of Segment; Custody.Sign returns the completed Segment
type Signature struct{ Bytes []byte }

// Header is the MOVE's leg — two values → exactly two ports. move.md.
// → field of Move
type Header int

const (
	Request  Header = iota // arrives at GREET — forwards a slice (push) or a pull instruction
	Response               // arrives at RECOGNIZE — returns an observation at a join
)

// Target is the address a MOVE names: transport endpoint + within-runtime name.
// move.md, traffic-routing.md.
// → field of Move, of causal.SendTo
type Target struct {
	Endpoint string // runtime transport endpoint (which CTU)
	Name     Name   // the within-runtime name
}

// SysInstruction is the OPTIONAL system-space placement — narrows the eligible set;
// ⊆ horizon. move.md.
// → field of Move (as *SysInstruction)
type SysInstruction struct{ Constraint string }

// Payload is OPAQUE — carried, read ONLY by the entity. move.md.
// → field of Move, Slice; entity.Entity.Run
type Payload struct{ Body []byte }

// Provenance is what the CTU stamps onto Content: which resource/invocation produced
// it, and its causal antecedents. The entity cannot forge it; a Segment signs it to
// cross a cut. decode-and-run.md, fabric/signed-provenance.md.
// → field of Segment (Observation carries that Segment)
type Provenance struct {
	Resource   Name
	Invocation string
	Causes     []SegmentRef // antecedents by content-address — the lineage the gather walks
}

// Contract is the parent-owned return-type expectation; RECOGNIZE checks an
// Observation against it at the correlated join. continuation-language.md.
// → field of Engram, Frame (frame-resident)
type Contract struct{ Name Name }

// Frame is the environment the runtime supplies. continuation-language.md.
// → field of Possibility, Slice; ctu.CTU.Think
type Frame struct {
	Scope  []Observation
	Return Contract
}
