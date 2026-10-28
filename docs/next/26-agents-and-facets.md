# Agents and facets

Proposal: **In the proposal.** Tracked in [next/proposal](proposal/README.md) as D1–D6; built in part.

Date: 2026-09-30  
Status: agreed in discussion; `system.narrow` and references joining the caller's reach built on 2026-09-30 (see the [proposal status](proposal/README.md#status)). Builds on [24](24-the-room.md) and
25; extends the adapter model of
15.

## The rule

> **An agent is a Horizon. A facet is a group of contracts over it. Reach is
> which facets a node holds.** What is public is what has been handed out;
> what is private is what never was.

| Term | Meaning | Agoric |
|---|---|---|
| **Agent** | a Horizon: the unit of authority, state and identity | a vat |
| **Contract** | one method: one input contract, one output contract | a method |
| **Facet** | a named group of contracts over one Horizon | a facet: one object over shared state, exposing a subset of methods |
| **Reach** | the facets a node holds | the c-list (imports) |
| **The room** | derives Horizons, records every facet handed out | the kernel |

One input contract, one output contract is unchanged. A facet groups
contracts; it does not merge them.

## Why

The system has contracts (one per surface) and Horizons (one per program),
but nothing in between says which contracts belong to the same agent, or
which a given caller should hold. The Codex adapter shows the gap:
`codex.start`, `codex.exchange`, `codex.inspect` and `codex.stop` are four
surfaces over one resident, and nothing in the registry says they belong
together, or that `codex.stop` is the owner's.

A facet is that missing grouping. It is an extension of the adapter model,
not a replacement: an adapter becomes an agent that presents its facets.

## Registration presents facets

An agent presents its facets to the runtime when it is registered:

```text
register codex
  creator facet:  codex.start, codex.exchange, codex.inspect, codex.stop
  public facet:   codex.exchange
```

- **The runtime knows the grouping** from the declaration: which contracts
  are over which Horizon, and which facet each belongs to.
- **Registration creates the agent's Horizon,** and the room records it
  (24: registration is an invoke that creates a Horizon).
- **The registrant is handed the facets back,** as references. It keeps the
  creator facet in its own reach and passes the public facet to whoever
  should have it. This is Zoe's `startInstance`, which returns
  `{creatorFacet, publicFacet}` to whoever started it.

The declaration says what is *meant* to be public. What *is* public is
whatever has been handed out.

## Public and private

There is no visibility flag. Privacy is not handing a reference out.

- **Public:** facets of an agent that sit in someone else's reach.
- **Private:** facets the agent keeps in its own reach and never hands out.

In Agoric a private method is a function inside the vat: the call never
reaches the kernel and is not recorded, so it is only safe because the whole
vat is deterministic and replayed from a transcript. Bringing that here
would mean unrecorded execution inside a Horizon, which would have to be
trusted and deterministic.

We do not need it. **A private method is an ordinary step under the agent's
own Horizon:** recorded, sandboxed and replayed like any other. The only
difference from a public one is who holds the reference. It costs a step per
internal call, which is what every call costs today.

## Shared state is the Horizon's durable state

In Agoric, facets share mutable state in the vat's heap. Agents here are
stateless continuations, so the state their facets share is the Horizon's
durable state: its stack and its log, read, computed on and written once,
through the runtime (24). No new mechanism.

## Calling another agent crosses into its Horizon

If an agent is a Horizon, invoking another agent's facet runs under **the
callee's** Horizon, not the caller's. A vat never runs with its caller's
authority. This refines 24, which had one way to change Horizons:

| Call | Runs under |
|---|---|
| a facet of your own agent (private or public) | your Horizon, shared |
| a facet of another agent | that agent's Horizon |
| through `system.narrow` | a new, narrowed Horizon |

An agent's authority is fixed when it is registered, not by who calls it.
The caller's authority still bounds what it can *reach* (it must hold the
facet), but not what the callee can *do*.

## The room is the root, and invokable through its facets

`system` is the room: the root Horizon, the one position holding all
authority. A caller never invokes the room itself. It invokes a **facet** of
the room, an ordinary invoke with an ordinary frame. The derivation then runs
in the room's position: the room is woken, derives, and exits, as it is on a
return.

- **Admin facet** (widen a Horizon, change the table): in the human's reach
  only. The creator facet.
- **Public facets** (`system.narrow`, `system.register`): in the reach of
  whoever the room grants them to. Withholding them is how a node is kept
  from narrowing or registering.

These are the only surfaces the runtime implements itself, since creating a
Horizon is a runtime operation. They join the exclusions in 24: no one can
register a surface in their place.

### Narrowing creates, then you talk

Narrowing creates a context; it does not speak to the actor. So it is two
steps:

```text
1. {"to": "system.narrow", "payload": {"actor": "read", "narrow": {...}}}
   → the room creates a narrowed Horizon for read and returns a reference to it

2. {"to": <that reference>, "payload": <read's input>}
   → an ordinary invoke; the child runs under the narrowed Horizon for its duration
```

This is Agoric's `E(vatAdmin).createVat(bundle)`, which returns
`{root, adminNode}`, and only then `E(root).method(args)`. Intake checks each
step with faults that already exist: an actor not in reach is a horizon
fault, a narrowing that asks for anything not held is a horizon fault, and a
payload that does not fit is a contract fault.

The envelope stays `{"to", "payload"}`. This replaces the separate `system`
field that 24 briefly had.

### Not a special return

Making `system` a special return that unwinds to the nearest Horizon
checkpoint was considered and rejected. A return means "done"; a return that
also means "ask the system" means two things, the `"reply"` problem again.
Unwinding to whoever holds authority already exists: faults climbing (24).
That is the path for **widening**, which must go up. Narrowing and
registering never need more than you hold, so they do not climb; they are
invokes of the room's public facets.

## Codex: the first agent with explicit facets

The resident already has the shape:

| Facet | Contracts | Held by |
|---|---|---|
| creator | `codex.start`, `codex.exchange`, `codex.inspect`, `codex.stop` | the frame that started the resident |
| public | the conversation turn (today `codex`) | whoever the creator hands it to |

Making this explicit replaces the implicit ownership now carried by
`ResourceOwnerDepth` and the resident fields on the surface (25, item 3).

## Open questions

- **Can reach grow by receiving references?** Step 1 of a narrowing returns
  a reference, and step 2 invokes it. So `to` must accept references a node
  has **received**, not only names fixed at Horizon creation. That is
  Agoric's c-list: receiving a reference in a message grants it, and the
  kernel translates every reference. Here, the room would record every
  reference carried by a return and add it to the receiver's reach, within
  its Horizon's authority. It also answers how a registered tool is shared
  upward (24): return its reference. It is the largest change in this doc.
- **How a reference is written in `to`.** A name, a facet id, or the Horizon
  id plus the facet name.
- **Revocation.** Agoric makes it a pattern (a caretaker facet the creator
  can cut). Whether a creator facet here includes revoking what it handed
  out.
- **What a facet declaration looks like** in the registry and in a
  `contract.json`, and whether a single-contract agent needs to declare one
  at all (a default facet of everything).

## Order

After 24 and 25 items 1–2. The c-list question first, since narrowing
depends on it; then facets at registration; then Codex as the first agent,
together with 25 item 3 and improvements item 4.
