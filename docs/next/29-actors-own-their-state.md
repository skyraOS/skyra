# Actors own their state

Proposal: **Not in the proposal.** Written after the proposal. Actor state, its methods, and the three interfaces of a novelty interrupt are not yet in [next/proposal](proposal/README.md).

Date: 2026-09-30  
Status: agreed in discussion, not built. Builds on [26](26-agents-and-facets.md),
[27](27-a-horizon-is-a-process.md) and [28](28-types-observations-checks.md).
Gives 28's novelty interrupt what it shows, and 26's creator facets and 25's
resident fields a place to go.

## The rule

> **The runtime owns control flow. An actor owns its state.** An actor never
> says whether it invokes or returns: the runtime derives that. What an actor
> owns is its state and the methods over it, which live in a place the kernel
> provisions and in no Horizon.

## Three interfaces

An actor acts through exactly three interfaces. They are what a novelty
interrupt (28) shows the actor, and what it writes code against.

| Interface | Owned by | What the actor does with it |
|---|---|---|
| **Faults and interrupts** | the runtime | receives them: what can happen to it. It never raises one |
| **State methods** | the actor, which authors them | reads and writes its own state |
| **Input contracts in its Horizon** | whoever it reaches | fills one to ask something of another program |

The first is fixed by the runtime. The second is the actor's own. The third is
its Horizon (24): what it may reach, and nothing more.

## Actors own their state

- **State is not authority.** A Horizon is authority over what is outside an
  actor. State is what the actor is. An actor's state methods are in no
  Horizon, so no one can be granted them, narrow them, or invoke them.
- **The kernel provisions the place.** At registration the kernel sets aside
  where the actor's state lives and where its read and write methods live. The
  registered actor writes the rest.
- **Every read and write is recorded.** The place is in the tree and reached
  only through the methods, the way `.skyra/` is reached only through the
  runtime today. A read or write that bypassed the log would break replay, and
  with it the log as the test suite (28).
- **Races are the actor's.** Two frames can call into the same actor at once.
  For now the actor is responsible for its own race conditions; the kernel
  does not serialize its turns.

This is the interpreter and shell split from `core/machine`, applied to an
actor. The actor's code is `Step(state, input) → (state', requests)`, pure. The
kernel is the shell: it applies the writes, performs the requests through
intake, and derives the state of what comes out.

## Two ways out, both to the invoker

What an actor emits is derived exactly as today (A1, A2):

| The actor | The runtime derives | The invoker receives |
|---|---|---|
| addresses its slot, with a payload that fits its output contract | **return** | a formal value |
| addresses nothing | **leaf fault** | the raw output, as the fault's observation |

A leaf fault is a return to the invoker: it discharges the slot, and the
invoker sees it as a step with the output as its observation. A1, A2 and A4
stay as built.

## Where the check comes in

The second row is where the parent's check (28) does its work. A leaf fault
hands the parent an observation; the parent's check decides what to do with
it:

- accept it;
- reshape it into a value of its own;
- fill another child's input contract with it, which the runtime derives as an
  invoke.

The check does not choose invoke or return either. It fills an input contract
in its Horizon, or produces a value, and the runtime derives the rest. Chained,
this reaches a formal return when the checks along the way are compiled
correctly. Nothing here guarantees that the chain ends: see the open
questions.

## Novelty at both edges

28 places the novelty interrupt at a return: the parent has no check for what
came back. There is a second edge, the first push:

```text
invoke a child with contract C ─▶ the child has no code for C        → novelty (input)
the child returns type R       ─▶ the parent has no check for R      → novelty (return)
```

An invocation does not determine what the child does with a contract it has
never seen. Both are the same interrupt: no compiled code for this binding and
this contract, at this edge. Whoever can write it is summoned, is shown the
three interfaces, and writes against them. What it writes is recorded and
kept, so the next crossing finds it.

## Example: the Codex adapter

The adapter (`controller/codex/adapter/main.go`), not the app-server behind
it. The adapter has state today; it has nowhere to keep it. It "keeps no state
between runs", so each run rebuilds its state by reading the actor host from
the start and delivering every line again, relying on the host's
at-most-once delivery by id:

| Its state | How each run recovers it today |
|---|---|
| thread id | delivers `thread/start` again under `<frame>:thread` and reads the response |
| turn id | delivers `turn/start` again under `<frame>:turn` |
| event cursor | the sequence number the redelivery returns |
| tool calls answered | counts `item/tool/call` events against the steps, answering each again |
| final message | scans the turn's events again |

Recovering state by redelivery works only if every line is byte-identical to
the first time. On 2026-09-30 it was not: `dynamicTools` builds the tool list
from a Go map, so a resumed round delivered `thread/start` with the tools in
another order, and the host refused it as a conflict. Both fake-Codex demo
scenarios faulted. (Fixed in `d1389c7`: the adapter now offers only its
model-facing tools, in a fixed order. The fragility of recovering state by
redelivery remains.)

Under this document:

- **State**, in the place the kernel provisions: `{thread, turn, cursor, answered}`.
- **State methods**, authored by the adapter: open a thread, start a turn,
  record an answered call, advance the cursor.
- **Input contracts in its Horizon**: the actor host's socket and the tools in reach.
- **Faults and interrupts**: what it receives, such as a fault on a tool call.

A run reads its state, reads new events from its cursor, and then either fills
an input contract (a Codex tool call becomes an invoke) or addresses its slot
(the final message, or a native attempt as an observation). A resumed round
reads `thread` from its state and never delivers `thread/start` again. The
host's delivery log goes back to its own job, delivering each line at most
once, and stops holding the adapter's state.

The adapter also shows a race the actor owns: two frames of one conversation
share the thread id `conversation:<id>:thread`, so two concurrent turns on one
conversation race. The host's at-most-once delivery makes that loud, as a
conflict, not silent.

## What changes

| Where | Today | Here |
|---|---|---|
| actor state | none; each run re-derives it (the adapter by redelivery) | a place the kernel provisions at registration, reached only through the actor's methods, every read and write recorded |
| 26's creator facets (start, exchange, inspect, stop) and `ResourceOwnerDepth` | facets in the creator's reach (D6) | the actor's state methods, in no Horizon |
| 25's resident fields (C4) | on the surface | state the kernel provisions for the actor |
| 28's novelty interrupt | at a return | at a return and at an input; shows the three interfaces |
| the Codex adapter | stateless; state recovered by redelivery | `{thread, turn, cursor, answered}` as its state |
| A1, A2, A4 | built | unchanged: an unaddressed output is a leaf fault, a return to the invoker |

## Open questions

- **Who answers an input novelty.** The child's author (its registrant), the
  child itself when it is an agent that can write its own handler, or the
  invoking parent.
- **What the place is.** Its shape in the tree, how it is keyed (per
  registration, per conversation, per Horizon), and its record encoding, which
  belongs with improvements item 9.
- **What the methods are written in.** Starlark, as checks are (28), so they
  are pure and replayable; the kernel supplies reads and writes as the only
  effects.
- **State inside a resident process.** The app-server keeps its own state
  (threads, history, caches) in a home the kernel already provisions
  (`ResidentPaths`). The kernel can record what crosses the connection, not
  what happens inside. Whether that is a second kind of actor state, or state
  the actor's methods only wrap.
- **Inspection.** With state methods in no Horizon, who can read an actor's
  state: the kernel, through the admin facet (26), or no one but the actor.
- **Termination.** A chain of checks feeding children can cycle. Starlark's
  step cap bounds one run, not the chain; a budget charged to the Horizon
  (improvements item 7) would, and exhausting it should interrupt, not fault
  silently.
- **Races later.** Whether the kernel eventually serializes an actor's turns,
  one at a time as in E's vats, or leaves it to the actor for good.

## Order

1. The place: the kernel provisions state for a registered actor, in the tree,
   every read and write recorded and replayed.
2. The Codex adapter keeps `{thread, turn, cursor, answered}` there, and stops
   recovering state by redelivery.
3. State methods as Starlark, run like checks (28).
4. The novelty interrupt shows the three interfaces; novelty at an input as well
   as a return.
5. 26's creator facets and 25's resident fields move into actor state.
