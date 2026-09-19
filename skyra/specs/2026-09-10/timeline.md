# Timeline — the smallest durable working demo (~2026-09-10 → 2026-09-24)

**Status: PLAN.** Two-week arc for DNA 16. The goal is one thing: a minimal
*durable* working loop. Not a feature set — a single demo that earns the word
"durable." Ordered by dependency, not effort. `interpret` stays stubbed the
whole way; it is compiled reasoning and is deferred to outside resources, not
solved here.

New generation: [`skyra-dna-16`](../../src/skyra/skyra-dna-16/). The v15 contract
surface is already consolidated into one
[`contracts.go`](../../src/skyra/skyra-dna-16/contracts.go) (build + vet green).
That is the seed; everything below fills bodies within it.

## Definition of done

One demo, one sentence:

```text
Submit a root through the user driver. Let the agent driver fan out and dispatch a
child. Kill the process AFTER the child is dispatched and BEFORE the parent resumes.
Restart. The workflow runs to completion from persisted state and the user driver
receives the root result.
```

That kill-and-resume is the whole point. Persistence that is only written and never
restored from does not count. Durability means the resident continuation state
survives process death and the loop picks up where it stopped.

## Ordering and why

```text
1. kernel first      the kernel owns the durable-operation boundary; durability is
                     a privileged operation someone owns, not a property sprinkled
                     on the loop. Nothing durable is real until the kernel is.
2. drivers second    user + agent are the archetypes the loop exists to serve;
                     they replace the carried-over echo/root fixtures.
3. prove the loop    minimal kill/resume proof, interpret stubbed.
```

## Week 1 — re-establish the loop, then flesh the kernel

**W16-0 — port v15 bodies into the single package.** Bring registry, controller
(dispatch/execute), scheduler (`transition`/`write`/`runtime`/`continuation`/
`continuationOpcode`/`authorize`), ctu boot, and the support helpers into
`skyra-dna-16` as bodies against the seeded contracts. Re-establish a green
`build` / `vet` / `test -race` on the fixtures before changing behavior. This is
a move, not a redesign.
- Opportunistic: decompose `continuationOpcode` while moving it, per
  [`focus.md`](focus.md) — attribute ownership, behavior frozen. Only if it stays
  a pure code-move; otherwise leave it and keep the port clean.

**W16-1 — flesh the kernel into a real first driver.** Promote `FixtureKernel`
(registration-only) toward a kernel that owns a durable-operation boundary. The
kernel is special in *privilege*, not in bypassing the protocol: the crossing
stays explicit, source-bound, and observable.
```text
kernel address
privileged descriptor
durable-operation boundary   <- the new piece
authority/capability check
completion and fault path
```

**W16-2 — decide the minimal persistence seam.** Settle, in the smallest form:
```text
WHAT persists   the continuations map + each Handoff (sequence, placement,
                pending output, consumed) + Result — enough to resume the gate
WHERE           behind the kernel's durable-operation boundary, not scattered in
                the scheduler; scheduler stays the sole in-memory owner
WHEN            on committed transition under the mutex; restore rebuilds the map
                before the loop accepts new events
```
Record the decision before wiring it. This is the seam most likely to grow — keep
it to the demo.

## Week 2 — user/agent drivers and the durable proof

**W16-3 — user driver.** Submits the root Invoke and receives the root result
(the `complete` boundary). This is the demo's mouth: in one end, result out the
other.

**W16-4 — agent driver.** Decomposes an Invoke into a child batch and, on
resumption with observations, returns the aggregate. Same shape the `RootDriver`
fixture proved; now it is a named archetype, not a test prop. Echo becomes the
trivial leaf endpoint the agent dispatches to.

**W16-5 — the durable proof.** Deterministic kill/resume test:
```text
dispatch a child, persist, simulate process death (drop in-memory state),
reconstruct from the durable boundary, resume, complete, deliver to user driver.
```
Plus the v15 proofs still pass (immediate/reversed/concurrent completion, one
resumption, placement gate, batch atomicity, source binding, duplicate/stale
feedback, capability separation).

## Checks

```sh
go build ./src/skyra/skyra-dna-16/...
go vet   ./src/skyra/skyra-dna-16/...
go test -race ./src/skyra/skyra-dna-16/...
```
Green before each item and after. Deterministic coordination, no timing sleeps.
Do not touch v13/v14/v15 or module config; v15 is read-only reference.

## Explicitly deferred (do not let these creep in)

```text
interpret / compiled reasoning   the hard research piece — stub returns Waiting,
                                 resumption rests on CompleteRegister. Deferred to
                                 outside resources; do not attempt it here.
retries, multi-wave fan-out, driver replacement/draining/generations
network transport (the demo is durable, single-node; distribution stays a design bet)
production authority verification (fixture grants remain trusted)
```

## Open

- Persistence encoding: JSON snapshot of the continuation map is enough for the
  demo; anything richer is scope creep. Confirm before W16-2 wiring.
- Exactly which continuation attributes must be rehydrated for the gate to resume
  correctly — verify against the `Handoff` + `Result` + `Resumed` set, not by
  guessing. This is the same attribute-ownership question `focus.md` raises, seen
  from the durability side.
- Whether the user driver's result delivery is itself a durable boundary (does a
  crash *after* completion but *before* delivery re-deliver?). Note it; the
  minimal demo can accept at-least-once and log it as known.
