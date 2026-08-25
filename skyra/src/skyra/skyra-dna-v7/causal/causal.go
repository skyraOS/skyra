// Package causal holds the types that CROSS boundaries — the causal state the
// CTU executes over. Leaf package (imports nothing) so ctu/kernel/drivers/entity
// can share these shapes without an import cycle.
//
// NEW TERMINOLOGY NOTE: `causal` is not a struct term; it is from "causal state"
// (causal-transition-unit/README.md). Every TYPE below is a doc term, cited
// inline. See interface-rationale.md and claude-ideas.md.
//
// The file is split PRIMARY / SECONDARY:
//
//	PRIMARY    appears directly in an interface signature — it CROSSES a boundary.
//	SECONDARY  a field of a primary type — it ties to an interface THROUGH its
//	           parent, never on its own.
//
// Every type ties to at least one interface. If one didn't, a concept wasn't
// decomposed (that is how kernel.Define was found — Account/Pin were orphans).
package causal

// ══════════════════════════════ PRIMARY TYPES ════════════════════════════════

// Move is custody transfer — the only transport. Two instructions under one
// authority. move.md.
// → ctu.CTU.Move / .Greet, ctu.Interpreter.Interpret
type Move struct {
	Header         Header          // ENVELOPE — routes to GREET / RECOGNIZE
	Target         Target          // ENVELOPE — user-instruction (the address)
	Horizon        Horizon         // ENVELOPE — the authority bounding BOTH instructions
	SysInstruction *SysInstruction // ENVELOPE, CTU-only — OPTIONAL (nil = absent)
	Payload        Payload         // OPAQUE — user-instruction (the content), entity-only
	Correlation    Correlation     // SEAM: envelope vs payload (see Payload)
}

// Payload is OPAQUE — carried, never read by fabric or CTU; read ONLY by the
// entity. move.md.
// SEAM: move.md says the payload ALSO holds Header + Correlation; traffic-routing.md
// puts both in the envelope. Recorded, unresolved — interface-rationale.md.
// → entity.Entity.Run
type Payload struct{ Body []byte }

// Outcome is the three-outcome result — S/F PLUS the NOT-YET a bool cannot carry.
// move.md. (The name is mine; the values are doc terms — claude-ideas.md §3.)
// → returned by ctu.Move/.Greet/.Recognize, kernel.Remember, drivers.Remember
type Outcome int

const (
	Success Outcome = iota // custody crossed
	Failure                // custody did not cross
	NotYet                 // operands / resource unresolved — nothing happened, nothing went wrong
)

// Goal is what THINK is given alongside the frame. think.md.
// → ctu.CTU.Think
type Goal struct{ Statement string }

// Frame is the environment the runtime supplies: observations in scope + the
// return contract. continuation-language.md.
// → ctu.CTU.Think
type Frame struct {
	Scope  []Observation
	Return Contract
}

// Potential is a proposed child signature from THINK — a method signature,
// contract-shaped (no values), at most five per firing. think.md.
// → returned by ctu.CTU.Think
type Potential struct{ Signature string }

// Horizon is the authority — the token. ENVELOPE; the fabric refuses on it.
// Narrowing-only. horizon-and-root.md.
// → ctu.CTU.Bound, kernel.Activate, kernel.Define
type Horizon struct {
	Root    Name     // which account this descends from (horizon-and-root §4)
	Caveats []Caveat // ordered, append-only; each one NARROWS
}

// Caveat narrows a Horizon. Append-only; never widens. horizon-and-root §4.
// → ctu.CTU.Bound
type Caveat struct{ Constraint string }

// EligibleSet is what INTERPRET produces — resources a MOVE may run on, already
// filtered by the quadrant and any SysInstruction. decode-and-run.md, dispatch.md.
// → returned by ctu.Interpreter.Interpret, param of ctu.Scheduler.Schedule
type EligibleSet struct{ Resources []Resource }

// Resource is one thing the CTU can dispatch to. Its concrete type says which
// table supplied it; trust is not a boolean claim carried by the resource.
// dispatch.md: "which table answers is the trust decision."
type Resource interface{ ResourceName() Name }

// HostResource came from the trusted host table.
type HostResource struct{ Name Name }

func (r HostResource) ResourceName() Name { return r.Name }

// EntityResource came from the untrusted entity table.
type EntityResource struct{ Name Name }

func (r EntityResource) ResourceName() Name { return r.Name }

// KernelResource is the one named privileged executor. It is a distinct type,
// not a privilege bit on a MOVE or a generic resource.
type KernelResource struct{ Name Name }

func (r KernelResource) ResourceName() Name { return r.Name }

// Content is RAW entity output — untrusted, unstamped, unchecked. The CTU turns
// it into an Observation. decode-and-run.md.
// → returned by entity.Entity.Run
type Content struct{ Bytes []byte }

// Observation is a returned value WITH provenance, checked against a Contract at
// RECOGNIZE. Not self-typed. continuation-language.md.
// → ctu.Recognize, ctu.Scheduler.Schedule, entity.Entity.Run (operands), kernel.Remember
type Observation struct {
	Content    Content
	Provenance Provenance // stamped by the CTU, never by the entity
}

// Possibility is the live process — a running frame. Parent==nil IS the root
// (the arena) — a position, not a kind. primitive-details.md, cpu-to-ctu-correspondence.md.
// → returned by kernel.Activate
type Possibility struct {
	Parent      *Possibility // nil = ROOT / the arena
	Name        Name
	Frame       Frame
	Outstanding Outstanding
}

// Outstanding is the in-flight children — the custody you do NOT hold. move.md.
// → returned by ctu.CTU.Orient
type Outstanding struct{ Count int }

// Name is an entry in the kernel-owned namespace: name → meaning. boundaries.md.
// → kernel.Activate, drivers.Driver.Activate
type Name struct{ Value string }

// Engram is durable behaviour = child contracts + compiled continuation.
// primitive-details.md.
// → returned by drivers.Driver.Activate
type Engram struct {
	Name      Name
	Contracts []Contract
}

// Cone is everything an Engram depends on — what REMEMBER writes. "the chain is
// the cone." primitive-details.md.
// → returned by kernel.Remember, param of drivers.Driver.Remember
type Cone struct{ Engrams []Engram }

// Account is the durable identity a Pin unlocks; owns a namespace + durable graph.
// horizon-and-root.md.
// → returned by kernel.Define   (was an ORPHAN until Define existed — that is how
//
//	the missing authentication decomposition was found)
type Account struct{ Root Name }

// Pin authenticates an Account and unlocks its root Horizon. horizon-and-root.md.
// → kernel.Define
type Pin struct{ Secret string }

// ═════════════════════════════ SECONDARY TYPES ═══════════════════════════════
// Each is a FIELD of a primary type; it ties to an interface through its parent.

// Header is the MOVE's direction — two values → exactly two ports. move.md.
// → field of Move
type Header int

const (
	Request  Header = iota // arrives at GREET
	Response               // arrives at RECOGNIZE
)

// Target is the user-space address a MOVE names. ENVELOPE. move.md, the-quadrant.md.
// → field of Move
type Target struct {
	Endpoint string // runtime transport endpoint
	Name     Name   // user-instruction within that runtime
}

// SysInstruction is the OPTIONAL system-space instruction — placement. ENVELOPE,
// CTU-only; narrows the eligible set; ⊆ horizon. move.md.
// → field of Move (as *SysInstruction)
type SysInstruction struct{ Constraint string }

// Correlation matches a RESPONSE to the outstanding REQUEST it answers.
// SEAM: traffic-routing.md (envelope) vs move.md (payload).
// → field of Move
type Correlation struct{ ID string }

// Provenance is what the CTU stamps onto Content: which resource/invocation
// produced it. The entity cannot forge it. decode-and-run.md.
// → field of Observation
type Provenance struct {
	Resource   Name
	Invocation string
	Causes     []Provenance
}

// Contract is the parent-owned return-type expectation; RECOGNIZE checks an
// Observation against it. continuation-language.md.
// → field of Engram and Frame (never in a signature directly — frame-resident)
type Contract struct{ Name Name }
