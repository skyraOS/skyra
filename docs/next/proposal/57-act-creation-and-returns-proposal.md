# 57 — Act creation and returns

Date: 2026-11-05

Status: **proposed; not implemented.** Extends [55](55-one-tool-trees-invariant-proposal.md)
and [56](56-tree-machine-proposal.md). Act creates input observations as part
of invoking an effect. Its return privileges depend on the actual receiving
boundary: a delegate receives passed observations; root may also receive
new content.

How instructions are admitted, typed and continued (`sense`, base types,
refinements, the scope table, and the store) is
[58](58-type-and-transition-machine-proposal.md), split from this proposal.
Observation references (`$n`), their escaping, and the canonical
base-relative form of an instruction are defined there. This proposal
depends on 58; 58 does not depend on this one.

## Effect invocation

The agreed proposed model envelope is:

```json
{
  "perform_effect": {
    "contract": 3,
    "create": "Summarize this document in three sentences.",
    "input": {"name": 0}
  }
}
```

`create` is optional. It is the content to create, not an object with an
`instruction` field. `input` passes an existing observation by name. The two
fields keep newly authored content separate from passed observations.
Omitting `create` passes the input without creating replacement content:

```json
{
  "perform_effect": {
    "contract": 3,
    "input": {"name": 0}
  }
}
```

The runtime performs observation creation automatically under act's
authority. The model no longer needs a separate `create_input` call before
dispatch. Creation still goes through observation admission and is recorded
for replay; it is implicit in the model interface, not absent from the
machine. Passed observations retain their identity and provenance. An input
composed from new content and a passed observation must retain the latter
as a source member, rather than flattening its value into untracked text.

This is a proposed interface. Today's `perform_effect` takes `input` as a
numeric name or an `earlier` reference, and requires separate
`create_input` instructions for new content. The proposed `{"name": 0}`
reference must be implemented explicitly, not assumed to work already.

## Return privileges

| Instruction | Allowed destination | Create new content? | Pass observations? |
|---|---|---|---|
| `act_return_delegate` | Parent delegate | No | Yes |
| `act_return_root` | Root boundary | Yes | Yes |

## Return to a delegate

Replace act's generic return with a destination-specific instruction:

```json
{
  "act_return_delegate": {
    "members": [2]
  }
}
```

This passes observation 2. It has no `create` field and no creation
privileges. Reject attempts to include newly authored content. Preserve the
existing restrictions on returning observations the act created as inputs,
and on returning received lineage. Effect results remain returnable when
they satisfy those restrictions.

This restriction applies to the return. An act beneath a delegate still
has creation privileges when making inputs for effects.

## Return to root

An act returning across the tree's root boundary may create content as well
as pass observations:

```json
{
  "act_return_root": {
    "create": "Here is my answer.",
    "members": [2]
  }
}
```

`create` is optional here too. `members` retains its observation-passing
meaning. A content-only root reply should be expressible with `create` and
an empty `members` list; a passing-only reply omits `create`. Reject a return
with neither content nor observations.

Record new root-return content as model-created under the act's authority,
separately from passed observations. Do not disguise it as an external
effect result. Passing still preserves identity and provenance; permission
to create does not grant permission to pass an out-of-scope observation or
bypass existing restrictions on passed members.

This deliberately changes the blanket prohibition on model-created
returns. The exception is new content explicitly created by
`act_return_root`, not an unrestricted exemption for every observation
the act previously created.

## Destination is runtime state

Root means the receiving boundary outside the tree, not any ancestor the
model chooses to call root. The runtime derives the destination from the
node's actual position and parentage. It exposes the matching return
instruction and independently rejects the wrong one during interpretation.

An act whose immediate receiver is a delegate cannot acquire root creation
privileges by emitting `act_return_root`, even if that delegate is the root
node. A root act may use `act_return_root`. Delegate-node return behavior is
outside this proposal; delegates do not acquire content creation privileges.

## Required Lean definitions and proofs

Write Lean for this proposal in `controller/apprentice/typesys-prototype/invariants/` before
implementing the corresponding Go behavior. Define the two act return
instructions, their destination-dependent admission rules, and optional
creation in `perform_effect`. Prove that:

- Effect-input creation is available only to act.
- A return to a delegate cannot create content.
- Root-return creation is admitted only at the actual root boundary;
  selecting an instruction cannot change the node's destination or privileges.
- Passing observations preserves their identity, source membership, and
  scope restrictions, including when combined with newly created content.
- The root-return exception does not weaken the restrictions on passed
  members or make previously created inputs freely returnable.

Extend the Lean-generated conformance vectors and Go checks for these
rules. The existing proofs do not establish this proposal: the new
definitions and proofs must typecheck, and the Go implementation must agree
with their vectors.

## Review: unresolved boundaries

1. **Input composition.** The canonical, hashed form of an instruction is
   base-relative ([58](58-type-and-transition-machine-proposal.md),
   question 1). The `create`/`input` envelope is a model-facing form
   translated into it. Specify that translation, how `create` and
   `input` are assembled into the invoked contract's payload, and how the
   structured `input` reference relates to 58's `$n` expression syntax.

## Implementation and unresolved details

Update the invariant specification in `controller/apprentice/typesys-prototype/invariants/`
first, then the corresponding Go admission and return checks, instruction
parser, shell tool schemas, prompts, and provider instruction handling.
Keep replay records explicit about creation, composition, dispatch, and
return destination. Update the affected statements in proposals 55 and 56
when implementing this change; this proposal does not claim their current
proofs already cover the root-return exception.

Before implementation, specify how `create` and `input` are assembled into
the invoked contract's typed payload. The model-facing shape above is fixed;
it does not imply that every existing contract accepts a new two-field
payload. Also settle the supported value types of `create`, creation-only
effect calls, literal-versus-reference escaping, and compatibility for
existing `create_input` calls and recorded instruction indices. Do not
silently reinterpret old records under the new grammar.

## Checks and measurements

Deterministic tests must demonstrate:

- Only act can create effect inputs through the combined instruction.
- Omitting `create` passes the existing input without creating a substitute.
- Mixed creation and passing preserve source membership and observation identity.
- `act_return_delegate` rejects creation, including an explicit `create` field.
- `act_return_root` accepts creation only at the actual root boundary.
- Both return instructions reject the wrong destination and out-of-scope members.
- Existing restrictions on passed return members remain enforced.
- Root content is recorded as model-created, never as an external result.
- Replaying accepted records reproduces the same observations and destination checks.

Measure model calls, input-creation instructions, effect invocations, return
turns, and elapsed time separately on the same task before and after the
change. Combining creation and dispatch removes one explicit model tool
call per invocation that previously needed `create_input`; it does not
guarantee one fewer network request when a provider already batches calls.
Measure latency rather than inferring a fixed speedup. Root content replies
should no longer need an echo-like effect merely to become returnable.
