# Body is the register and return specification

> **LATEST REFINEMENT:** [One register per fd](register-frame.md) records the accepted live shape:
> Register has Contract plus ordered Children carrying child fds and returned observations. Contract is
> now schema-only; Action/Params stay on invoke. This supersedes the older Action-plus-Schema section
> and single-value/contract-placement assumptions below. Durable Body/Engram encoding remains open.

**Date:** 2026-09-06  
**Status: DIRECTION AGREED / details to work.** Records the Body discussion without introducing a final
Go struct, new binding subsystem, or runtime implementation. Builds on
[Focus and Move](focus-and-move.md) and [memory/execution ownership](engram-loading-and-fault-return.md).

## The return obligation

**Later refinement:** [Parent and child contracts on the engram](#parent-and-child-contracts-on-the-engram)
names the obligations explicitly. Body still describes register requirements; durable parent/child
contract fields now have a home on Engram. The returned value is an Observation, not a Return object.

Body records:

> What observations do I need, in which registers, to produce the observation I return to my parent?

Body **is the register specification for fulfilling that return**, rather than a separate operand-binding
recipe alongside the registers. The same slots carry the input roles, their observation requirements,
and their relationship to the return. Exact fields remain to be designed.

An illustrative body:

```text
Return: Comparison

Required registers:
  original : Document
  revised  : Document
```

These are observation names/roles, not a claim that the first declaration already knows their full
structural types. The contract lifecycle below explains how structure is discovered.

The engram's immediate dependency references identify the executables supplying those slots. Focus
contains available choices; it does not make every visible capability a required producer. Committing
a producer connects it to a register requirement. The exact dependency-to-slot representation is open;
do not duplicate the same binding in a second independent structure.

## Specification and live values

| Surface | Responsibility |
|---|---|
| Body | Register requirements and return specification |
| Live registers in Memory | Actual observations satisfying those requirements for this execution |
| Op | Behavior that fulfills the return using those observations |
| Child contracts / dependencies | Obligations between this move and its selected children |
| Executable dependency references | Identify the producers supplying those obligations/slots |

This keeps Body durable/reusable while Memory holds execution-specific values. A value landing does
not mutate the compiled engram. This is a distinction between a register's specification and its live
contents, not a new set of competing register or binding abstractions.

Observation type and register identity have different jobs. In the example, both inputs are Documents,
but original and revised are distinct roles. An arriving Document must land at its designated slot for
the correct execution; its type alone cannot select either slot. The existing execution/register
identities supply that distinction; their concrete mapping to Body remains to be designed.

## Firing fulfills the return

The trigger is expressed through the return obligation: what must be satisfied to return this move's
observation to its parent? For an ordinary operation requiring all its inputs:

```text
required registers satisfied
  → eligible to fire
  → scheduler claims and dispatches
  → CTU performs Op using Memory
  → produce and return the observation
```

Eligibility is not itself completion. A pure composition can fire by assembling and returning existing
observations. Comparison, generated code, or driver work still needs the existing Op to produce its
result. Body specifies what is needed; Op supplies the behavior. There is no new trigger executable
introduced by this framing.

Memory maintains slot consistency and publication. Kernel/scheduler owns readiness admission, claims,
and the consequences of CTU return events. Input availability alone does not bypass access checks or
fault handling. Conditional-input requirements and non-value completion obligations remain to be worked;
do not assume that every possible branch must fill a slot before return.

## Contracts: declaration names, execution discovers, REMEMBER records

This fits [contracts-and-remember.md](../2026-08-27/contracts-and-remember.md). It does not reverse the
decision that the caller does not declare a new child's full structural return type in advance.

| Moment | What is known or recorded |
|---|---|
| Declaration / first call | Observation names, register roles, and whatever contract information is already known; new structural details can remain unknown |
| Cold execution / return | Observe what dependencies actually return and what this move composes or produces; structural information is discovered bottom-up |
| REMEMBER | Record the discovered structural contract with the remembered executable; exact encoding and location remain open |
| Warm execution | Use remembered structural assumptions and guard returned observations against them; changed structure follows the existing deopt/re-Remember direction |

The body can express the obligation before the first execution without pretending that discovery has
already happened. A nominal requirement such as Document names what is needed; it is not automatically
a complete field schema. Existing known contracts can inform declaration, but do not manufacture
knowledge about an unseen child's result.

We agree this lifecycle fits the register/return framing. The relationship needs further definition,
especially how a declaration becomes a remembered specification. This document does not settle the
structural type system, guard placement, or compiler ownership questions elsewhere in the specs.

## Next discussion

The parent/child contract refinement below gives durable obligations an explicit home. These questions
remain about their contents and relation to Body/registers, rather than whether they have a home at all.

[Register decision review](register-decision-review.md) collects prior register shapes, lifecycle,
matching, tags, fan-out, and contract decisions with current/superseded/open assessments. It is a
review inventory, not a new set of decisions or automatic retirement of older proposals.

1. What exactly does a declared register record: its role/name, expected observation name, known contract
   information, and producer reference? Which existing types can express that without duplicate bindings?
2. What does the return specification record before execution versus after discovery? How does it
   describe composition or an Op-produced result without duplicating Op's behavior?
3. How do unknown/partially known structural contracts become recorded contracts at REMEMBER? Where
   does that feedback live durably, and how does it participate in EngramID/versioning?
4. What does "satisfied" mean: value present, structurally acceptable, or further semantic fulfillment?
   Keep runtime availability and semantic validation explicit while working the firing rule.
5. How do warm guards distinguish a changed valid shape from an invalid observation, and how does
   recovery preserve the original register roles and return obligation?

No final Body/Engram struct or runtime register representation is selected here. The agreed collapse is
that Body records the register requirements for the return, using the existing execution concepts.

## Parent and child contracts on the engram

**Direction agreed:** name the two perspectives explicitly and record them as fields on the engram:

| Field / perspective | Meaning |
|---|---|
| parentContract | What this executable owes its parent |
| childContracts | What this executable needs its selected children to return in order to fulfill its own obligation |

Dependencies, in the obligation sense, are the contracts between a move and its children. Executable
references still identify who supplies them; a contract is not itself an EngramID. Focus remains the
set of possible invocations, not a set of already-established child contracts.

For a particular parent/child relationship, the parent's childContract and the child's parentContract
are **two perspectives on the same return obligation**. They must not evolve as independent,
contradictory agreements. This does not settle how reusable artifacts expose a return contract versus
how a particular caller binds it; durable references, compatibility, and versioning need further design.

```text
Parent.childContract for this child
               ↕ same return obligation
Child.parentContract
```

**Refined: the shared underlying type is Contract = Action + Schema**, defined below. This supersedes
the earlier proposal to use Return itself as the contract type. Both parentContract and each childContract
use Contract; do not invent unrelated parent/child contract types. Existing engine code calls the return
opcode Ret and supplies `Value c.Observation`. That opcode delivers the value; it is not the schema.

**The returned value is an Observation, not a Return object.** Contract describes the obligation;
executing the return delivers an Observation to the parent. An observation arriving is a runtime fact;
whether it satisfies the contract is a judgment whose exact rules remain open. Do not change the
result/register value type from Observation to Return as a consequence of the naming.

The engram stores the durable contract specification, alongside Body and immediate executable references.
This extends the earlier body-plus-children field sketch. Memory still holds live observations and
execution-specific fulfillment; immutable engrams are not updated on each arrival. How registers refer
to these contracts without duplicating binding truth remains open, as does the childContracts collection
shape and its key. The full Engram/register structs are not finalized here.

The existing lifecycle still fits: declaration can name obligations with incomplete structural knowledge;
execution discovers returned observation structure; REMEMBER records the learned contract; warm runs
use it. The unresolved work is to make those stages precise within the shared Contract, not to
assume a full structural type exists before a child's first return.

## Contract = Action + Schema

**Historical: superseded by [schema-only Contract](register-frame.md#contract-and-invoke).** The Action
semantics and 160-token limit survive on invoke/Capability; Contract no longer duplicates Action.

**Decision:** a contract has exactly the two conceptual fields Action and Schema. The agreed Go shape
uses the existing Name carrier and opaque JSON bytes from encoding/json:

```go
type Contract struct {
    Action Name
    Schema json.RawMessage
}
```

**Action is the single model-visible method name / contract header, capped at 160 tokens.** There is
no separate tool name, display title, or description. For example, `Search customers by name or email`
is the action itself, not a description attached to a hidden `search_customers` identifier. Focus already
exposes capabilities with Action; no OperationDecl or parallel descriptive fields are needed.

The existing Name type is merely the Go carrier for Action, not another Contract.Name field. The action
header is not to be split back into JSON Schema title/name/description annotations. Target remains the
existing addressing/authority context; this decision does not make action text globally unique or use
it in place of EngramID. The tokenizer and enforcement point for the 160-token limit remain to be chosen.

**Schema is an opaque JSON schema describing the returned observation's structure.** json.RawMessage
is the carrier, not a validator or a fixed Go type hierarchy. The schema dialect, validation interface,
canonical encoding, and structural discovery/update mechanics remain open.

An Observation is the concrete returned value conforming to that schema, with its provenance. It need
not embed a copy of the schema; the exact association is still to design. Schema conformance establishes
structural validity; additional semantic fulfillment rules are not settled by this shape.

The lifecycle remains declaration → discovery → REMEMBER → warm validation. Action can identify the
obligation before the full schema is known. How an unknown or partial schema is represented is open;
do not silently treat `{}` as a special "unknown" marker. Learned schema changes must be reconciled
with immutable engram versioning rather than mutating an existing artifact during execution.

This replaces v12's current `Contract struct{ Name Name }` in the design and the earlier shared-Return
proposal, but does not yet change runtime Go code. parentContract and childContracts retain their agreed
home on the engram and both use this same Contract shape.
