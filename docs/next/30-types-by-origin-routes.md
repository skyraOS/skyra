# Types by origin, and routes

Proposal: **In the proposal**, section F of [next/proposal](proposal/README.md);
the implementation plan is [30-types-by-origin-routes-proposal](proposal/30-types-by-origin-routes-proposal.md).

Date: 2026-10-01  
Status: **built** on branch `self-extension` (2026-10-01), F1–F9 of the
[proposal](proposal/30-types-by-origin-routes-proposal.md#as-built), which
records the decisions taken and where the build differs. "Wake" is named
**interrupt** throughout: the runtime raises it, never the parent (29).
Builds on [21](21-observation-passing.md),
[24](24-the-room.md), [28](28-types-observations-checks.md) and
[29](29-actors-own-their-state.md). Replaces 28's checks with routes: the
parent does not judge a return, it says where the return goes. Replaces the
declared output contract with a type the runtime computes.

## The rule

> **The runtime computes types; models write routes.** A type is the actor,
> the contract it was invoked through, and the shape of what it returned,
> identified by a hash the runtime authors. A model never sees the hash; it
> sees the type's presentation. A return whose type the parent has never seen
> is a novelty interrupt, and a novelty interrupt produces a route: given this
> type, send this observation to this target, or interrupt me. The route is kept,
> so the next return of that type does not interrupt.

Three layers, and where meaning lives:

| Layer | What it is | Authored by |
|---|---|---|
| **Shape** | the structure of a JSON value: its keys and its leaf kinds, not its leaf values | the runtime, while parsing |
| **Identity** | the type hash, `H(actor, contract, shape)` | the runtime |
| **Presentation** | `actor.contract` and the shape, as text a model reads | the runtime, when it shows a type |
| *Meaning* | what this type is for, and where it goes | the parent, as a route |

"Presentation", not "representation": in PL a value's representation is its
concrete encoding, which here is the identity layer. The presentation is the
printed form, as Haskell's `show` is.

## JSON is the floor

JSON is the minimal way to take part in the protocol. An actor, including an
external one that never agreed to our protocol (the Codex app-server, a web
API), returns a JSON value. It declares nothing else.

- **The runtime produces the primitive types.** It walks every return and
  computes the shape; leaves are plain values of a JSON kind. It does not
  guess what a string means: `"kind": "file"` is a string, not an enum.
- **Meaning is not the runtime's job.** Deciding what a return is for is the
  model's, at the novelty interrupt, and what it decides is a route.

## Identity: actor, contract, shape

```text
type hash     = H("skyra/type"  ‖ actor identity ‖ contract hash ‖ shape hash)
presentation  = actor.contract → shape            e.g.  twitter.search → {text: string, media: [{url: string}]}
```

- **The actor is its identity, not its name.** Its registration: for a binary,
  the surface's `sha256`; for a host surface, its registration entry. Names are
  labels on hashes and never part of one. If the name `twitter` were in the
  hash, re-registering a different program as `twitter` would reuse every old
  type, and the old routes would run on the new program's output with no
  interrupt. A new version is a new registration (improvements item 6), so its
  returns are new types, and their routes are written again, against the log.
- **The contract is the one invoked:** the hash of the input contract the call
  went through. `twitter.search` and `twitter.post` can return the same shape
  and still be different types.
- **The origin is stamped by the runtime,** from the record of who emitted the
  bytes, never read from the payload. An actor cannot claim another's origin.
- **Same shape, different origin, different type.** A string from `twitter` is
  not a string from `macbook`. Moving a value from one origin's type into
  another's input is a route, authored and recorded. Text an external actor
  returns never becomes another program's argument without one.

## Two hashes per node

The runtime walks the decoded value once, bottom-up, and gives every node two
hashes: its **shape** (leaf values left out) and its **content** (leaf values
in). A node stores its children's hashes, not its children (Merkle).

```text
shape(string) = H("s")                 content(string) = H("s" ‖ len ‖ bytes)
shape(number) = H("n")                 content(number) = H("n" ‖ len ‖ digits as written)
shape(bool)   = H("b")                 content(bool)   = H("b" ‖ 0|1)
shape(null)   = H("z")                 content(null)   = H("z")
shape(object) = H("o" ‖ n ‖ for each key, sorted: len(k) ‖ k ‖ shape(v[k]))
content(object) = H("o" ‖ n ‖ for each key, sorted: len(k) ‖ k ‖ content(v[k]))
shape(array)  = H("a" ‖ m ‖ the distinct element shapes, sorted)
content(array)  = H("a" ‖ n ‖ content(e₀) ‖ … ‖ content(eₙ₋₁))
```

- **Unique** up to SHA-256 collisions, because the encoding is injective: a
  kind tag on every node, length prefixes on every variable part, keys sorted.
- **Keys are shape; values are content.** `{"text": "hi"}` and
  `{"text": "yo"}` have one shape and two contents.
- **An array's shape does not depend on its length.** A listing of 3 entries
  and a listing of 300 have one shape, so a directory listing is not novel each
  time. Its content does depend on length and order.
- **Verifying a type is recomputing the root.** The root hash commits to the
  whole tree, so one comparison covers every type inside it. The runtime
  computes it while parsing, so verifying costs nothing extra.

This is E3's idea (two hashes per step: a content id and a stamp) applied to
every node of a value.

## Diffs come from the tree

Two returns of the same type are compared by walking both trees together:
where content hashes match, the whole subtree is skipped; where they differ,
the walk descends. It ends at exactly the leaves that changed, by path.

| Measure | What it is |
|---|---|
| **changed** | the paths of the leaves whose content differs |
| **ratio** | changed leaves ÷ leaves |
| **added, removed** | paths present in one tree and not the other, when shapes differ |

A content hash says *whether* a string changed, not *how much*: `"hi"` →
`"hj"` and `"hi"` → an essay are both one changed leaf. How much a string
changed is a separate measure over the strings themselves, which the runtime
has; it is not in v1.

## Hashes never reach a model

The security boundary is the entry point to the system, and it rests on one
invariant:

> **The system does not show hashes to a model or to an external actor.**
> Hashes go to workers (trusted programs the runtime starts), never into a
> prompt and never across the boundary. A model sees presentations.

A model writes routes against presentations and paths (`$.patch`); the
runtime binds them to hashes when it records the route. Where hashes are
shown today, to be audited (F1):

| Where | Shows | To whom |
|---|---|---|
| `SKYRA_FRAME` (A4) | the frame id, a stamp | the worker, such as the Codex adapter. Not the model, as long as the adapter does not forward it |
| a continuation's stdin (`Continuation.Frame`, `StepView`) | frame and step ids | an LLM continuation, if its program puts them in the prompt |
| the novelty interrupt | the return's type | the model writing the route: shown as its presentation |
| observations by hash (21) | observation hashes | whoever passes them on |

### The audit (F1, 2026-10-01)

Every place a hash reaches a model or an external actor today, read from the
code, with a fix or an accepted exception:

| Where | What crosses | To whom | Verdict |
|---|---|---|---|
| `machine.runOf`: `SKYRA_FRAME` | the frame id | every target's environment (a worker) | **Accepted.** A target addresses its return with it. `echo`, `codex/io` and `read` copy it into `to`, which the runtime reads, never a model |
| `Continuation.Frame` | the frame id | a continuation's stdin (a worker) | **Accepted.** `targets/agent` and `targets/user` use it only as the `to` of their return. `codex/adapter` and `codex/resident` do not put it in the turn input (that is `input.text`) or in a tool result |
| `codex/adapter`, `codex/resident`: JSON-RPC ids | `<frame>:thread`, `<frame>:turn`, `<frame>:tool:<n>`, `<frame>:interrupt` | the Codex app-server (an external actor) | **Accepted exception.** They are protocol correlation ids, and the actor host dedups deliveries by them, so they must be unique per frame and stable across a resumed step. The app-server echoes them in responses and does not put them in the model's context. To fix with 29's actor state, where the host can map ids |
| `system.narrow`: the reference | was the invoke's stamp | the continuation, as a step's value and a tool name (`Continuation.Tools`) | **Fixed.** A reference is now named `system.ref.<n>`, a counter in the runtime's namespace, unique in the Horizon that holds it |
| `StepView.Value`, `StepView.Fault` | a child's return, and fault causes | a continuation; Codex's tool results | **Accepted.** No runtime-authored cause quotes a stamp. A return can carry a hash only if a target put one there, and the only one that did was the reference, now fixed |
| `codex/resident` start: `profile`, `incarnation` | the profile's sha256; a pid-and-time label | the caller's step value (`targets/user`, a worker) | **Accepted.** Not forwarded to the model: `targets/user` shows the human text only |
| `hosts/web`: `calls[].stamp`, parked frame ids | invoke and fault stamps | the browser (the human, outside the boundary) | **Accepted exception.** The human is not a model, and the page keys its calls by stamp. Recorded: the page is outside the boundary |
| observation hashes in payloads (21) | — | — | **None passed today.** No target or adapter puts an observation hash in a payload |
| the novelty interrupt, the route contract (F7) | — | the model writing a route | **By construction:** a presentation, step indexes and names. A test asserts no 64-hex string appears in a `route` field |

Hashes inside a payload are references to observations already recorded (21).
The type of what a reference points at was fixed when it was recorded; the
actor can pass it on, not change it.

## A novelty interrupt produces a route

```text
child returns ─▶ runtime computes the type ─▶ lookup (binding, type hash)
                                              ├─ hit:  run the route
                                              └─ miss: novelty interrupt
```

1. The child runs; its return is committed as an observation, as always.
2. The runtime computes the return's type and finds no route for this binding
   and this type: **novelty**.
3. It is an interrupt, not a fault. The slot stays discharged and the child's
   work is kept.
4. The parent, typically a model, is woken and shown the **route contract**
   (below): the return's presentation, the conditions it can test, the
   decisions it can make, its Horizon, and what can happen to it. It writes a
   route.
5. The route is recorded, authored by the parent, and keyed to the type.
6. The next return of that type finds the route.

**A hit is an exact type hash.** A shape the parent has not seen, even one
differing from a known shape by an optional field, interrupts: the model has
not seen it. One interrupt per distinct shape is the cost, paid once.

A deterministic parent (the Codex adapter, `user`, `agent`) cannot answer an
interrupt, so it declares its routes at registration.

## Routes

**Routes are keyed on the type, which is shape; they decide on content.** A
route keyed on content would make every distinct value a miss, and every
return would wake the model. So the key says *which* route runs, and the route
says *whether to wake*, from the content diff:

```python
# route for (watcher, twitter.search → {text: string, media: [{url: string}]})
def route():
    if skyra.previous() == None or skyra.changed("$.text") or skyra.diff() > 0.3:
        return skyra.interrupt()                  # wake the parent with this return
    return skyra.send("log", {"line": skyra.get("$.text")})
```

This is "what to wake me up for next time": the model, at the interrupt,
writes down when it wants to see this type again.

- **A route returns one decision:** interrupt the parent, send an observation to a
  target, or fault. A send is the envelope every actor already emits,
  `{"to": ..., "payload": ...}`; the runtime derives invoke or fault from `to`
  exactly as it does today (A1), under the parent's Horizon.
- **Refusing is a route.** A reviewer that will not take a truncated read
  routes it to a fault. There is no separate check.
- **The observation is a filled input:** the target's input contract, its
  slots filled from the return.

### Routes are Starlark, one invocation, no loops

A route is a [Starlark](https://github.com/bazelbuild/starlark) program run
once per return. We do not author a language.

- **No `while`, no recursion.** Starlark's defaults, kept: branching is `if`.
  A route is a single invocation that decides and returns. The step cap
  (`SetMaxExecutionSteps`) is the backstop.
- **No ambient authority, deterministic.** No files, network, clock or
  randomness. The same return, the same previous return and the same route give
  the same decision, so replay runs it again.
- **Run as a runtime-owned surface,** as 28 describes for checks: a pinned
  binary embedding `go.starlark.net`, under an empty Horizon, through the
  ordinary execve path. The controller stays standard-library Go.

### The `skyra` package

A predeclared module, and a route's only way to touch the system:

| | |
|---|---|
| `skyra.get(path)` | a field of the return, read-only |
| `skyra.previous()` | the previous return at this binding and type, or `None` |
| `skyra.changed(path)` | whether the content at `path` differs from the previous return |
| `skyra.diff()` | the ratio of changed leaves |
| `skyra.interrupt()` | deliver the return to the parent: an interrupt of kind `route` |
| `skyra.send(to, payload)` | send an observation to a target in the parent's reach |
| `skyra.fault(cause)` | refuse, carrying the return as the observation |

It has no reads of state, no writes and no I/O. The previous return is
supplied by the runtime from the log, so a route stays pure.

## What the model sees: the route contract

28 named this the check contract ("what a check receives and what verdicts it
may give") and left it open. At a novelty interrupt the model is shown five
things, all as presentations, never hashes:

| # | What | Shown as |
|---|---|---|
| 1 | **The return** | its presentation: every path with its leaf kind, `$.text: string`, `$.media[]: {url: string}`, under `actor.contract`. These are what a condition can test |
| 2 | **The conditions** | `get(path)` → a value of that path's kind; `previous()` → the previous return of this type, or `None`; `changed(path)` → bool; `diff()` → float, the ratio of changed leaves |
| 3 | **The decisions** | `interrupt()`, `send(to, payload)`, `fault(cause)`; a route returns exactly one |
| 4 | **Its Horizon** | the whole Horizon, with its hashes left out: every name in reach with its input contract, every reference by name, and its authority (read and write grants with their scope, mach names, connect sockets) |
| 5 | **What can happen to it** | 29's first interface: the fault kinds (`malformed`, `horizon`, `contract`, `custody`, `leaf`) and the interrupt kinds: `novelty` (no route for this type), `route` (the route asked) and `depth` (eight sends in a row) |

- **The whole Horizon, for now.** Today a continuation is shown only the input
  contracts in its reach (`machine.tools`). The route contract shows the
  authority too, so the model can route on what it may do, not only on what it
  may call. Whether a model should see less (an offer list, improvements item
  3, direction 6) is later.
- **Hashes are stripped.** A Horizon holds hashes (`Stamp`, `Frame`,
  `Resume`, each surface's `sha256`, each reference's stamp). None is shown:
  surfaces by name and input contract, references by name, nothing of
  `Resume`.
- 29's second interface, state methods, joins when actor state exists.

## The log is the test suite

28's rule, applied to routes. Every time a route runs, the runtime records the
type, the return, the previous return and the decision. A rewritten route is
run against every recorded case for its (binding, type); if a decision
changes, the parent declares the change, and the flip is recorded with the
parent as author.

## What changes

| Where | Today | Here |
|---|---|---|
| `core/contract` | `Schema`, `Fits` | plus `Shape`, `Content`, `Diff` and `Present` over decoded JSON |
| `contract.json` `"output"` | declared by the registrant | not needed: the type is computed from the return |
| `interpret` | a return fits the child's output contract | a return addressing the slot is committed; its type is computed from it |
| `core/machine` | a discharge wakes the parent | a discharge is routed: a route, or a novelty interrupt that wakes the parent |
| 28's checks and handlers | a check judges, a handler picks a restart | one route: interrupt, send or fault |
| a new runtime-owned surface | — | the Starlark route runner with the `skyra` package |
| the room's facets | `system.narrow` | plus `system.route`: how a parent records a route |
| records | — | types, routes and route runs, with improvements item 9's encoding |

## Prior art

- **Hidden classes and shapes** (Self; V8's maps; SpiderMonkey's shapes).
  Objects with the same keys share a shape, computed by the runtime from the
  object itself, never declared. The closest precedent for the shape layer.
- **Inline caches** (Deutsch and Schiffman, Smalltalk-80, 1984; polymorphic
  inline caches, Hölzle, Chambers and Ungar, 1991). Dispatch keyed on a
  shape: a miss takes the slow path and caches what it found, a hit is a
  lookup. The novelty interrupt is the miss; the route is the cached entry.
- **`doesNotUnderstand:`** (Smalltalk). A message an object has no method for
  is reified and handed to a handler rather than crashing: novelty at the
  input edge (29).
- **Content-addressed definitions** (Unison). Types and code identified by the
  hash of their structure; names are labels on hashes. Why the actor's
  identity, not its name, is in the type hash.
- **Hash-consing** (Ershov, 1958; Filliâtre and Conchon, 2006). Structurally
  equal values share one representative, so structural equality is identity.
- **Merkle trees, IPLD, and Git.** A node stores its children's hashes; the
  root commits to the whole tree, and two trees are diffed by descending only
  where hashes differ.
- **Types from data** (Petricek, Guerra and Syme, *Types from Data*, PLDI
  2016; F# type providers). Types inferred from JSON samples. They merge
  samples into one type; we do not: a new shape is novelty.
- **Information-flow labels and taint** (Denning, 1976; Perl's taint mode;
  JIF, Myers, 1999). Data carries where it came from, and crossing a boundary
  needs an explicit act. Here, the origin in the type and the route.
- **Content-based routing** (Hohpe and Woolf, *Enterprise Integration
  Patterns*, 2003: the Content-Based Router and the Message Translator). A
  message goes where its content says, reshaped for the receiver.
- **Pattern-matching receive** (Erlang). An actor dispatches each message by
  its shape.
- **Restricted languages for decisions** (Starlark in Bazel; Google's CEL in
  Kubernetes admission and Envoy). Deterministic, bounded, no ambient
  authority, written by people who are not the host.
- **Kept model-written code** (Voyager, Wang et al., 2023: a skill library of
  programs the model wrote and reuses). Here, routes, recorded with their
  author.

## Open questions

- **Scalar kinds.** *Decided (2026-10-01):* JSON's four scalar kinds are leaf
  shapes, so `1` and `"1"` are two shapes.
- **`for` over the return.** *Decided:* only `if`. `for` and comprehensions
  are refused before a route runs, like `while`.
- **Which previous return.** *Decided for v1:* the last return at this binding
  and type in this frame. Across frames needs an index over the log.
- **What a send's result does.** *Decided:* it is routed again, up to eight
  sends in a row at one waiter; the next return is a `depth` interrupt.
- **Inputs.** Input contracts stay declared for now. Whether an input type is
  also computed, and who answers novelty at an input (29).
- **Observation reach.** Whether a payload may reference any recorded
  observation, or only those handed to it: D3's c-list, applied to
  observations.
- **Record encoding.** Types, routes and route runs as records belong with
  improvements item 9 (E).

## Order

1. F1, the audit: where hashes reach a model or an external actor today.
2. Shapes, contents, diffs and presentations in `core/contract`.
3. The type of every return, computed and recorded.
4. The Starlark route runner and the `skyra` package, under an empty Horizon.
5. Routes declared at registration, for deterministic parents.
6. The novelty interrupt and `system.route`.
7. The log as the test suite for routes.
8. Agent-written tools as Starlark, and `system.register` (28).
