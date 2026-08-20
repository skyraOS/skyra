# Locality — and the loop that drives the kernel

**Date:** 2026-08-07
**Status: UNDECIDED.** Nothing here is settled and nothing should be built from it yet. Written
to hold a direction while it is being thought about, not to enter the set.

> **No component holds a model of the whole tree.**

---

## The rule is not "nothing is central"

There is already one central component and it is fine: the kernel sees every wave. What makes it
safe is that it **cannot** enumerate, compare, or interpret — safe by incapacity, never by
restraint ([kernel.md](kernel.md) §2).

That is the license, and anything else central has to earn it the same way.

```text
CENTRAL        unavoidable. something hands out work, something records that it happened.
ORCHESTRATOR   holds a model of the whole tree and decides from it. THIS is what is refused.
```

Stated as *"no central orchestrator"* the rule will eventually be read as *"nothing central,"* and
someone will distribute the recorder — which destroys the only honest witness. Say **model of the
whole**, not **central**.

## The outer loop

Execution stays inside the kernel. What drives it is a while loop that produces one ephemeral
Space per iteration and invokes the kernel on it.

```text
produce an ephemeral Space  →  invoke the kernel  →  take what accumulated
                            →  fire it            →  results arrive as impulses
                            →  loop
```

**It is the exchange.** Not a resemblance — the same object. A wave ends when it promotes nothing,
so a second invocation only makes sense if something new arrived, and something new only arrives
if something fired. A loop that does not fire has nothing to loop over.

[questions.md](questions.md) Q1 already described it: *"not a Space, not a behavior — it DRIVES
the kernel rather than being driven by it."* Reached from the interface side instead of the
grading side.

```text
INHERITS Q1's RENAME PROBLEM   three things are called the boundary. this does not fix that,
                               and should not be named until it is.
```

### What it already fits

```text
RealizeConcrete   §5 named "the one caller entitled to the root's concrete return" before
                  anything existed to be it. this is that caller. Realize stays what a cell
                  calls, and the perception invariant holds as method reachability.

NOTHING FIRES     firing happens BETWEEN invocations, so the eviction guarantee stops being
INSIDE A WAVE     a rule and becomes geometry.

ONE STACK         the loop serializes invocations. one at a time, by construction.

TERMINATION       the kernel pops a frame that promotes nothing; the loop exits on a wave
                  that accumulates nothing with an empty queue. the same fixed point, twice.
```

### It may not reassemble the tree

The post-order fold stays in the kernel. A loop that rebuilds the tree is enumerating
composition — Law 7 inverted — and it re-inherits the result-tree apparatus the current kernel
deleted. **The return path being the tree is the kernel's whole claim to existence.**

### It may not choose the entry

```text
DEFAULT   the standing root. chroot is the exception, for a resumed execution.
NEVER     the loop picking a subgraph root, which is choosing what the agent attends to.
```

`execution.md` — *"whoever picks the entry has chosen what the agent attends to, by hand."*

This is the **fourth** place the sealing failure has tried to appear: queue admission,
interception-as-filter, the ranker multiplier, now loop entry. Four independent recurrences is
enough to treat it as one rule with a name rather than four local warnings.

## A meta recording layer

Bookkeeping sits **above** the kernel rather than inside it, and wraps both levels.

The argument is externality, not economy. §4 already holds that a thing recording its own
execution can omit itself; kernel-side bookkeeping is safe only because the kernel is trusted, and
a wrapping layer makes it structural instead. Implementing it once is the second reason, not the
first.

**It generalizes.** A kernel hop and a loop iteration are the same record:

```text
an invocation happened  ·  its lineage  ·  whether it produced anything
```

`Hop` is close to that shape already.

```text
THE CONSTRAINT   strings and bools, at BOTH levels.
```

The recorder sees the accumulation going out and the results coming back, which makes it the
best-positioned component in the system to become a results channel. §5's rule was written for the
transcript; here it is load-bearing in a way it was not there, because this layer sees what no
transcript could.

## Locality is already law

[phases.md](phases.md), derived from eviction safety:

> `Observe` reads ONLY WHAT IT IS HANDED — its accumulation and the child's realization.
> **It does not go looking.**

That is *the receiver does not worry about what is above it; it does the work and returns*,
already written, from a different direction. Same convergence pattern the set already trusts.

Top-down holds too: budget DESCENDS and accumulation ASCENDS ([behaviors.md](behaviors.md)). What
a node may spend is decided by its causal past, and nothing it returns depends on knowing who
asked.

## What this implicates

**Local learning and dropping the fold are the same decision.**

```text
CLOSURE   one perceive, THIS execution, no history.        LOCAL.
P         a lifetime aggregate over everything a node      GLOBAL IN TIME.
          ever saw.
```

`P` is the only mechanism in the set that learns by aggregating across a whole history. Under this
rule it is not merely expensive — it is the wrong shape. So the grading question is not a separate
simplification to weigh; **this decides it**, if this is adopted.

## The exception, held consciously

**Retrieval is global.** The token index spans the graph and an impulse may pull a candidate from
anywhere ([search.md](search.md)). That is genuinely non-local.

Probably fine: read-only, heats nothing, and already declared derived, rebuildable, and deletable.
But it is the exception, so it is named as one rather than discovered later.

> **Learning is local writes. Caching may be global reads.**

One line, and it stops someone later "fixing" the index into something that also learns.

---

## Open

```text
WHAT WITNESSES THE LOOP   per-wave transcripts are honest about each invocation. the
                          SEQUENCE across them — which impulse was taken, what was fired,
                          in what order — is outside all of them. the meta layer is the
                          candidate. undecided whether that is where it belongs.

THE ONE RECORD SHAPE      "hop and iteration are the same record" is asserted here, not
                          demonstrated. if forcing one shape generalizes Hop into something
                          meaningless, two shapes is the better answer and implement-once
                          was never the real argument.

THE FUEL COUNTER          the loop is the only place that sees every wave, so a step budget
                          there would bound unbounded re-execution (Q7), the mount cycle,
                          and a Space that never returns nil — in one place. execution.md
                          guessed those wanted one answer. this is where it would go, and
                          nobody has decided to put it there.

THE NAME                  blocked on Q1's rename.
```
