# Observation passing: author on the way down, pass on the way up

Proposal: **Not in the proposal.** Its open choices were answered by the proposal; building the observation register is not scheduled in it.

Date: 2026-09-28  
Status: proposed rule, not implemented. Follows [13 — Dependencies](../design/thoughts/13-dependencies.md) and 14 — Interpreted, not stated.

> **Revised 2026-09-30.** Three open choices are answered by
> [24](24-the-room.md), 25 and improvements item 9:
>
> - **Faults:** both. A fault is its own derived state, and it **carries the
>   observation of why it faulted** as its payload (built). The parent passes
>   the whole fault as one observation, or singles out the observation inside it.
> - **A model that makes no calls:** `Surface.Leaf()` is gone. Authority is not
>   coupled to leaf or continuation; whether a node authored or passed its
>   return is read from the realized tree, which is this document's own
>   definition.
> - **Author identity:** a principal's stable id is the stamp of its first
>   declaration; an observation's origin is the step that authored it
>   (improvements item 9).
>
> The register lives in the durable stack of 24. One case is added: **7. A
> leaf fault**, a program that printed without addressing its slot; the
> fault's observation is what it printed, authored by that leaf, and the
> parent may pass it on.

## The rule

> A node may create observations for an invocation on the way down. A
> continuation returning to its parent may only pass observations it received
> from children. It cannot return an observation it authored itself.

The direction is an edge in the invocation tree, not a point in time. A
continuation may learn from one child, author a new observation for another
child, and invoke that child. The restriction applies when it returns upward.
The root may author the initial observations. A leaf, which has no child result
to pass, may author the observation it returns.

This replaces the earlier candidate rule, “a continuation cannot create
natural language after it has invoked a tool.” That rule would also forbid
useful instructions to later tools. The directional rule allows those
instructions while keeping the returned result's origin in the subtree that
produced it.

## Observation IDs

An observation has a content-addressed ID. Invocation operands are recorded
as observations and passed by ID. A tool receives those IDs as typed inputs:

```text
some_operand(obs: observation[])
```

For example, Codex can invoke a writing leaf with IDs for the results it wants
turned into prose:

```json
{"to":"write_answer","payload":{"obs":["<result-hash-1>","<result-hash-2>","<brief-instruction-hash>"]}}
```

`observation` in the input contract means a hash reference, not inline text.
The Codex-facing tool schema exposes that input. The runtime checks each ID
against observations available in the invoking frame, then supplies the
referenced observations to the child. A JSON Schema string check alone cannot
establish that an ID exists or is in scope.

The ID must bind the observation's content **and its origin**. Hashing only
the text would merge identical text authored by different nodes and lose the
fact needed to enforce the return rule. The exact canonical observation
record, including how it names its author and source observations, remains to
be specified.

## The stack register

Each frame has a logical **observation register** in the stack. It tracks the
observation IDs that frame can use as operands, with their origin and the edge
through which they arrived. The register is the local view of the lineage,
not a global bag of hashes and not authority supplied by the model.

```text
frame register:
  ID → observation record, author, received-from, available-for-invoke/return
```

At descent, the runtime commits any newly authored operands as observations,
adds their IDs to the invoking frame's register, and passes the selected IDs
into the child's register. At discharge, the child's returned IDs enter the
waiting frame's register, marked as received through that child slot. A
continuation's upward return is checked against those entries: an ID it
authored for a downward call is available for later invocations, but it is
not eligible as that continuation's own return.

The register is a **derived view** of committed observations, invocations and
discharges. It must be reconstructible after a crash and by replay; a mutable
in-process map cannot be the source of truth. Today's frame stack carries
directory descriptors and dependency slots. The observation register would
extend what that stack represents without changing who holds custody of it.

## Down and up

| Crossing | What the node may do | Runtime check |
|---|---|---|
| Root → child | Author initial operand observations and invoke a child with their IDs | Commit them, check the child's input contract, populate its register |
| Continuation → child | Author operand observations for that invocation; pass IDs it already holds | Commit new ones, check each ID in its register, populate the child's register |
| Leaf → parent | Author a result observation | Commit it as the leaf's result and enter its ID in the parent's register through the discharged slot |
| Continuation → parent | Pass child-result observation IDs without changing their contents or origin | Each ID is marked in its register as received through a discharged child slot, and was not authored by this continuation |

A continuation may select and order the observations it passes. Selection is
its decision; the observation's words and origin remain the author's. If it
wants a synthesis, summary, explanation, or other new natural language, it
invokes a leaf that accepts the relevant IDs and authors a new observation.
The continuation can then return that new ID.

```text
root authors task observation A
  → Codex invokes search with [A]
    ← search authors result B
  → Codex invokes write_answer with [B] and its downward instruction C
    ← write_answer authors answer D
  ← Codex passes D, without authoring a replacement answer
```

Passing a hash does not grant access by itself. The runtime must check the
invoking frame's register and the target's contract. A child receives only
the observation IDs passed in its invocation, under its narrowed Horizon.

## What this changes from the current controller

Today an invocation carries inline JSON payloads, a return carries inline
JSON data, and `machine.Continuation.Steps` gives the resumed continuation
each child's inline value. The Codex adapter can write fresh final text after
a tool result. None of these paths records an observation author, maintains
an observation register, or checks that a continuation's return passes a
child's result. This rule therefore needs a new observation reference in the
contracts, records, continuation input, and return check. Prompt instructions
alone cannot enforce it.

The emitter still supplies operands, never a verdict. The receiving runtime
resolves IDs, checks provenance and contracts, derives invoke or return, and
commits the result before anything acts. Replay must re-derive those checks
from the recorded observations and return the same verdict.

## Cases the rule must decide

1. **A continuation returns its own text after a child ran:** refuse the
   return. The text is not a passed child observation.
2. **A continuation passes a child observation unchanged:** accept, if the ID
   is in scope and fits the reply contract.
3. **A continuation writes a new instruction for another child:** accept it on
   the downward invocation; record its author.
4. **A continuation passes an unknown, sibling-private, or forged hash:**
   refuse it before the child runs; the ID is absent from its register.
5. **A child echoes its parent's observation ID:** that does not change the
   original observation's author. The parent cannot use the echo to return its
   own observation as though the child wrote it.
6. **A writing leaf uses child results to author prose:** its new observation
   has the leaf as author and names the inputs it used. The continuation may
   return its ID.

## Open design choices

- **Minting a downward observation:** a model cannot be trusted to declare
  its own hash. Decide whether authoring and invoking are separate steps, or
  one output contains new observation bytes that the runtime commits and
  replaces with IDs before invoking.
- **Return shape:** decide whether `reply` passes one observation ID or an
  ordered list. Tool inputs already need `observation[]`.
- **A model that makes no calls:** decide whether it counts as a leaf of the
  realized invocation tree and may author its direct answer. The current
  `Surface.Leaf()` means *empty reach*, which is a different definition.
- **Faults:** the runtime derives faults from evidence. Decide whether a
  fault is a separate discharge type or a system-authored observation; do not
  silently attribute it to the continuation or a leaf.
- **Author identity:** specify the durable frame or record identity used in
  the observation hash, and the exact fold that reconstructs each frame's
  register from committed observations and discharges.

## Done when

- Every operand used as a tool input has a checked observation ID, recorded
  origin, and an entry in the invoking frame's derived register.
- A continuation can author inputs on descent and cannot return its own
  observation on ascent.
- A leaf can author a return; a continuation can pass that return by ID.
- Forged and out-of-scope IDs are refused before execution.
- Replay re-derives the author and scope checks from the record tree.
