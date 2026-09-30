# Interpreted, not stated: invoke and return are derived by the runtime

Date: 2026-09-25  
Status: agreed in conversation. Follows [13](../design/thoughts/13-dependencies.md) and
[`../../faults.md`](../design/faults.md).

> **Revised 2026-09-30.** The core holds: the emitter supplies operands and
> the runtime derives the state. The `reply` capability is superseded
> ([24](24-the-room.md); improvements items 1 and 2): a return **addresses the
> slot its frame discharges** (the frame id, told to a target in
> `SKYRA_FRAME`), and output that addresses nothing is a **`leaf` fault**
> carrying the output. No name is reserved. A leaf shares its caller's
> Horizon; its reach is not "only `reply`". A fault is its own derived state
> and carries the observation it was derived from. The sections on the reply
> capability and "What falls out" items 3–7 are kept as history.
>
> ```text
> addresses the slot its frame discharges   → return
> names anything else in reach              → invoke
> names nothing in reach                    → fault (horizon)
> addresses nothing                         → fault (leaf), carrying the output
> not readable as operands                  → fault (malformed)
> fits no contract                          → fault (contract)
> ```
>
> Open item 1 is answered: `{"to": <slot | name | reference>, "payload":
> <JSON>}`, nothing else. Open item 2 becomes the tail-call question: a call
> whose return discharges its caller's slot.

## The idea

The LLM does not declare whether its output is an invoke or a return. **The
runtime derives it.** Everything an instance needs comes from state (memory, the
stack, its Horizon) and its operands. This is D1 ("the emitter is never the
checker") taken all the way: the emitter does not even name what its output is.

Invokes and returns are still recorded (in the record, and eventually in a
procedure), but they are the runtime's interpretation, never the LLM's claim.

## States are derived, not stated

This changes the model underneath D9. D9 said the states are the Moves, but the
emitter still *wrote* the Move, so the states were declared. Now:

```text
emitter supplies:  (which capability?, payload)          raw operands, nothing more
runtime derives:   interpret(state, output) → invoke | return | fault | interrupt
record stores:     the derived state
replay:            derives it again, and must get the same answer
```

**Invoke, return, fault and interrupt are states the system is in, never things
anyone says.** The state machine belongs entirely to the runtime:

| State | Derived from |
|---|---|
| **invoke** | the address is a capability in reach, and the payload fits its contract |
| **return** | the address is `reply` (explicit or by default), and the payload fits the return contract |
| **fault** | the runtime proved something broke |
| **interrupt** | memory lacked what the interpretation needed |
| **resolve** | the stack shows every dependency discharged |

Nothing an untrusted party emits can put the system into a state. It can only
supply operands that the runtime judges. The naming test, "the system is in a
state of ___", becomes literally true: it is the system's state, not the
message's.

## The address decides

The runtime already knows each instance's parent, from the custody chain. So:

- output sent to **a capability in reach** is an **invoke**;
- output sent to **the parent** (through `reply`; see the refinement below) is a
  **return**.

The LLM chooses only *which capability* and *what payload*. It never knows or
claims whether that is a call or a return.

### Refinement: the reply's slot, not the parent

State the return rule as **"the target is the slot this instance's `reply`
names"**, not "the target is the parent". The runtime still derives it from the
custody chain; the difference is what it compares against.

- **Tail calls.** Once `reply` has been passed on, its holder returns to its
  *grandparent's* slot, not its own parent. Comparing the target with the parent
  gets this wrong. Following `reply` gets it right.
- **The parent stays opaque.** The emitter never needs to know who its parent
  is. It addresses `reply`, or leaves the output unaddressed, and the runtime
  resolves where that goes. If it could name its parent, part of the tree's
  structure would appear in the frame, and parent identity would become an
  operand the emitter chooses.

Both forms derive the return from protocol rules. Only the `reply` form still
holds once capabilities are passed on.

## Return is invoking the reply capability

When a parent invokes a child, it hands the child one extra capability: **`reply`**,
a handle to the slot the parent is waiting on. Returning means invoking it.

```text
child's Horizon:  reach = { toolA, toolB, …, reply }      reply = the parent's slot, single use
emit(toolA, payload)  → a reachable Potential → open a dependency   (invoke)
emit(reply, payload)  → the reply capability  → discharge the slot  (return)
```

This is continuation-passing style (returning is calling your continuation), and
it is how the E language does promises (a promise is fulfilled by sending to its
resolver). `reply` may be represented differently from ordinary tools, for
example as a reserved name, but semantically it is a capability like the others.

## Unaddressed output is a reply

If the output names no target, it is a reply. **An unspecified target falls to
the least authority available**, and `reply` is the least authority there is: it
can discharge only the one slot it was given, once. Defaulting to it can never
escalate anything.

| Case | Interpretation |
|---|---|
| **Leaf** (reach is only `reply`) | any output is a reply. A tool just prints its data |
| **Continuation**, output names a capability in reach | invoke |
| **Continuation**, unaddressed output (for example the final answer) | reply |
| Output names something **not in reach**, shown or not | horizon fault. Visibility grants nothing |
| **No output** | custody fault ("target exited without emitting") |

## Interpretation is a total, pure function

```text
interpret(state, output):
  not parseable                                   → malformed move fault
  names a capability not in reach                 → horizon fault
  names a reachable capability, or reply (explicit or by default):
      payload fits that capability's input contract  → invoke, or return if it is reply
      fits nothing known                             → novelty interrupt
      fits more than one contract                    → novelty interrupt (ask, never guess)
```

It depends only on state and the output, so `step` stays pure and replay
reproduces every interpretation.

## What falls out

1. **One primitive.** Everything is `invoke(capability, payload)`. This is the
   CTU pivot's "one transport verb" in its simplest form.
2. **The model cannot forge control flow.** Output that says "return success"
   is not a return. Only a payload that fits the reply capability's contract is.
   This belongs to the enforcement side (what is invokable), not the narrowing
   side (what appears in the frame).
3. **Exactly-once returns are structural.** `reply` is single use: an `O_EXCL`
   claim on the slot.
4. **An instance can only return to the one who asked.** It holds exactly one
   reply capability.
5. **Tail calls are free.** "My result is the result of X" is passing your reply
   capability to X.
6. **Contracts unify.** Every capability has an input contract, and **the reply
   capability's input contract is the return contract.** Novelty checking and
   input checking are the same check: does this payload fit the contract of the
   capability it was sent to?
7. **Faults fit.** A runtime-proven fault is the runtime invoking the reply
   capability on the instance's behalf, with a fault payload.
8. **Targets never emit a `Fault` Move, or any Move.** They emit output. The
   Move variant is the runtime's interpretation, stored in the record.

## Where control flow lives

The LLM no longer declares control flow; it chooses capabilities and supplies
payloads. Repeated patterns of invokes and returns are recorded into procedures
(keyed by input contract), so a known procedure can drive the same sequence
again. A missing procedure is a procedure interrupt, and a procedure an LLM
proposes is itself output, interpreted by whether it fits the procedure schema.

**Capability risk:** agents like Codex decide tool calls mid-turn. Under this
model a tool request is output, interpreted against the input contracts in
reach. It stays dynamic, but only through contracts the runtime knows. How often
real agent output fails to fit, and interrupts, is an empirical question.

## Change to the minimal shape

- Echo prints its value (`hello`), not `{"return":{"value":"hello"}}`.
- Intake interprets it: echo is a leaf, so unaddressed output is a reply. It
  checks the payload against echo's return contract, then discharges the slot.
- `move.Parse` becomes part of interpretation, not the wire format. The record
  stores the interpreted variant.

## Open

1. **The addressing format** for a continuation's output: how it names a
   capability (a tool-call structure, a reserved field), and how `reply` is
   written explicitly.
2. **Handing `reply` to a child** (tail calls): the child's Horizon then holds its
   grandparent's slot. Does the invoker lose its own reply at that moment?
3. **Several outputs from one target** (a streaming continuation): each is
   interpreted separately; only one can be a reply, and it ends the instance.
