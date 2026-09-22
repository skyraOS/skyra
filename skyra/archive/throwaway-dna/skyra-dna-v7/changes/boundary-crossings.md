# Overnight build — findings (interfaces that did not hold)

**Date:** 2026-08-22
**Status: THROWAWAY IMPLEMENTATION NOTES.** The design-relevant half of the overnight build's
morning trail — the **new information** it surfaced. Split here from the build's execution snapshot,
which stays with the code at `../boundary-crossings.md` ("what runs"). The authority remains
`skyra/specs/2026-08-18/`.

## Interfaces that did not hold

### Fallible crossings had no outcome

`drivers.Activate`, `kernel.Activate`, `kernel.Define`, `entity.Run`, `Interpreter.Interpret`, and
`Scheduler.Schedule` all returned a zero value on failure. That made missing durable names,
authentication refusal, malformed execution, no eligible resource, and a resource wait
indistinguishable from valid zero values. The specs say the machine is total and name three
outcomes, so all six now return `Outcome`.

`kernel.Remember` now returns `(Cone, Outcome)` and takes the root `Possibility`. The old
`Remember(Observation) Cone` could neither make root-only availability expressible nor report a
durability failure.

### `Target{Endpoint}` could reach a runtime but not an instruction

Traffic routing explicitly separates the host endpoint from the envelope target within that
runtime. `Target` therefore grew a `Name`. `Endpoint` routes to a CTU; `Name` resolves the engram
and selects the instruction's eligible resources.

### GREET could admit work but could not return it

`Greet(Move)` had no source/connection operand. A request could arrive, but the receiving endpoint
had no target for its mandatory response. It is now `Greet(from, Move)`. `from` is supplied by
normalized ingress, below the entity payload.

### A correlated contract had no concrete home

RECOGNIZE requires the parent's frame-resident contract at the correlated destination, but the
proposed `Frame` contains only one return contract and no correlation map. The running CTU adds an
`Expect(correlation, contract)` ledger. This is deliberately concrete rather than promoted into a
new primitive. The language/compiler still needs to settle the actual frame shape.

### Resource booleans contradicted “named, not flagged”

Claude proposal §6 was rejected. `Resource{Name, Trusted, Kernel bool}` allowed impossible states
and made trust/privilege look asserted per resource. It is now a sum by concrete type:
`HostResource`, `EntityResource`, or the singular `KernelResource`. Which dispatch table answered
is structural provenance; the MOVE still carries no privilege bit.

## Documents that collide

### Header and correlation: payload versus envelope

`move.md` puts both inside the opaque payload. `traffic-routing.md` requires both for routing and
response matching. The build keeps the interface proposal's envelope placement. Otherwise the
fabric cannot select GREET versus RECOGNIZE without reading entity-private bytes. This is a design
choice forced by the routing document, not a resolved canon change.

> **Update — 2026-08-22.** Resolved since, in `specs/2026-08-22/fabric/signed-provenance.md`: the
> routing identity (Endpoint, MoveID, Header) lives in the **envelope**; the payload stays opaque.
> MoveID is the cut edge between partitions.

### Observation transport has no settled layer

An entity returns raw `Content`; its CTU stamps `Observation`; the parent CTU must receive that
provenance to RECOGNIZE it; and MOVE may not read the payload. The documents never say where the
stamp travels. The local fabric currently carries the stamped observation beside the opaque MOVE
through the same `Approach`/normalized-arrival mechanism. This keeps the entity from forging it
and keeps MOVE from opening the body, but it is an unresolved wire-shape seam, not a design answer.

> **Update — 2026-08-22.** Given a shape in `specs/2026-08-22/fabric/signed-provenance.md`: each CTU **signs the
> provenance segment it executed**, so a downstream CTU RECOGNIZEs a stamp it did not make without
> trusting the sender's word. Where the signed segment travels on the wire is still the open edge.

### Request admission and response admission are different completions

The synchronous prototype returns `Success` once GREET admits the request. A later refused response
cannot retroactively return request custody, so its outcome lands in the parent's correlated
`Result`. Delivery acknowledgment and retry remain open in `traffic-routing.md`; this build does
not pretend synchronous function return settles them.

## Concepts that needed a home

- The within-runtime instruction name now lives in `Target.Name`.
- The response destination comes from normalized ingress, not entity content.
- Provenance gained `Causes`; REMEMBER walks that trace to compute the cone instead of asking the
  persistence driver what should survive.
- `Function` is only a Go adapter for exercising foreign entity hardware. It is mechanism, not a
  new machine concept.
- The preliminary caveat language is not invented here. The port subsystem accepts an `Authority`
  mechanism; without one it safely refuses every caveated Horizon.

## Deliberately unresolved after the run

- The preliminary `Horizon{Root, Caveats}` shape can describe scope but cannot prove that the
  bearer received it from the pin. `Root` is explicitly a public identifier, no chain/signature is
  represented, and `kernel.Activate` currently checks only that it is non-empty. The port subsystem
  can inject an `Authority` verifier for caveats, but an uncaveated root bypasses it. Consequently
  the running `GREET` path demonstrates where authority admission belongs, not an implementation of
  the docs' stronger "cannot be forged" boundary. Adding a boolean or trusting the root name would
  merely hide the gap; the token mechanism must first be decomposed.
- `kernel.Remember(root, keep)` can identify the names in the kept observation's provenance cone,
  but neither argument carries a new `Engram` (compiled continuation plus contracts). The concrete
  kernel can therefore reload and durably rewrite existing names, but cannot perform the specified
  write-resolution step in which a contribution installs a new name → meaning binding. This is an
  interface failure, not a missing filesystem method: the driver correctly owns HOW to persist,
  while the current kernel surface never receives enough information to decide WHAT new meaning to
  persist.
- Scheduling is local and deterministic. Distributed eligibility/resource readiness remains open.
- THINK enumerates registered instruction signatures in arrival order and enforces the cap of five;
  it does not implement model-backed discovery or branch prediction.
- Contract satisfaction is the smallest provenance-grounded check: expected name equals producing
  engram name. The global-name/local-satisfaction seam in `boundaries.md` remains real.
- ACTIVATE can construct a root, but the proposed signature carries no parent operand for child
  activation. The test can reject an explicitly non-root REMEMBER; the full live possibility-tree
  ownership and child-linking mechanism still has no interface home.

> **Update — 2026-08-22.** The possibility-tree home is now
> `specs/2026-08-22/ctu/possibility.md`: the tree is partitioned across CTUs, a forwarded slice is a child
> possibility, and a cut edge is a MOVE. Child-linking and the parent operand are the interface
> work that remains.

- The local fabric has no ambiguous-disconnect retry, external ingress, encryption, or replay
  defense.
- If a response cannot be delivered, the parent can remain outstanding. The transport spec already
  names acknowledgment/retry as open; solving it would be design work beyond this throwaway path.
- No “process” type, registry, lifecycle, or OS-process semantic was introduced.
