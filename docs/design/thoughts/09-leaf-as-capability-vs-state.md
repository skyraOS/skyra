# Leaf as a capability, or invoke/return as a state?

Date: 2026-09-23  
Status: discussion. Follows [08](08-leaves.md). The phase table is superseded by
D9 ([11](11-the-potential.md)). Leaf as a Boundary capability is still open.

## The question

Should "leaf" stay a property of the registered executable, or should "invoke"
and "return" be states that a frame is currently in?

## They answer different questions

| | Capability | State |
|---|---|---|
| Asks | What may this frame **ever** do? | What may this frame do **now**? |
| Source | Registration ∩ Boundary | The frame's records in the tree |
| Changes | Never within a frame (a Boundary is fixed at release) | With each emission and each child's return |

Using only one of them loses something:

- **State only.** States change because of emissions. If nothing fixed forbids
  it, a tool "enters invoke state" just by emitting an Invoke. The guarantee
  that a tool can never invoke goes away, and so does the guarantee that resume
  fds end at leaves (08).
- **Capability only.** Nothing stops a non-leaf from returning while its
  children are still running, which orphans them and their resume fds.

## Proposal: both, in layers

### 1. Capability: leaf-ness lives in the Boundary, capped by registration

- The registry records whether an executable **can** invoke at all. `echo` never
  can. This is the ceiling.
- The parent decides per invocation, through `narrow`: a non-leaf can be invoked
  **as a leaf** (for example Codex with no tools) by giving it empty `grants`.
- "Leaf" is therefore not a type. It is a Boundary with empty `grants`, and a
  Boundary only narrows, so a frame never gains the ability to invoke.

### 2. State: a frame's phase is computed from the tree, never stored

| Phase | Derived from | May emit |
|---|---|---|
| `open` | No outstanding child slots | Invoke (if the Boundary allows), Return, Fault |
| `waiting` | One or more outstanding child slots | Invoke (fan-out, if allowed), Fault. **Not Return** |
| `settled` | `result` exists | Nothing. Any emission faults (already enforced by `O_EXCL`) |

- New rule: **a frame may Return only when it has no outstanding children.**
  This is structured concurrency: children never outlive the frame that owns
  their resume fds. A Return emitted while `waiting` is committed as a fault.
- Intake computes the phase by reading the frame's slots and result. There is no
  mutable state file, consistent with content-addressed records and "computed,
  not stored".

### How they combine at intake

```text
allowed(emission) = capability(Boundary) ∩ phase(tree)
```

The leaf rule from 08 falls out as a special case: a frame whose Boundary has
empty `grants` is permanently limited to Return or Fault, whatever phase it is
in.

## Open

- Should a Fault in `waiting` cancel the outstanding children, or fault them
  recursively? DNA-19 already faults descendants recursively.
- Should a Return in `waiting` be a fault, or an implicit cancel of the children?
  Faulting is stricter and easier to reason about.
