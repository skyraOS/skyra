# 2026-08-03

**The strip-down continues, and the first mechanism is now CODE.** `2026-07-27` stands as the
record of the design reasoning. This set carries forward only what has been built or is being
built against, and it inherits 07-27's rule unchanged:

> **Strip complexity, do not add it.** A mechanism earns its place by being required by the
> loop, not by being coherent with the rest of the doc set.

One thing has entered on its own terms rather than by being moved: the kernel, because it was
implemented and the implementation answered questions the doc could not.

---

## The loop

Unchanged from 07-27. It is the reason anything here exists.

```text
an OBSERVATION comes in
    │
    ▼
it is assigned a SUPPORT VALUE against an expectation
    │
    ▼
given that expectation, what is the next ACTION
    │
    ▼
given that action, what is the next EXPECTATION
    │
    └──────────────────────────▶ (an observation comes in)
```

Four causes, two node types:

```text
observation  ──▶  state change  ──▶  action  ──▶  observation
 (material)        (final)            (efficient)
```

## In this set

| doc | owns | status |
|---|---|---|
| [kernel](kernel.md) | the substrate. `Perturb` / `Observe` closed over `Space`, the returned transcript, and what the kernel is structurally incapable of doing. | **BUILT** — `src/skyra/spacetime/kernel` |
| [execution](execution.md) | durable graph vs. ephemeral traversal, the execution anchor, and **ONE STACK**. | **FOUNDATION** — rev. 08-06 |
| [families](families.md) | **one family of state, two edge kinds.** the interpretation node is gone: EXPECT is the only writer of state. | **SPEC** — rev. 08-06 |
| [graph](graph.md) | the expectation family's shape, and the perceive that makes every expectation graded. **superseded in part by [families](families.md)** — it assumed the expectation tree was the whole store. | **RULE SETTLED** — 08-03 |
| [phases](phases.md) | **`Perturb` reads. `Observe` writes.** the rule every behavior follows, and why eviction depends on it. | **RULE** — 08-05 |
| [behaviors](behaviors.md) | the four behaviors, and the one question every node asks: *am I closed / what would close me.* | **SPEC** — rev. 08-06 |
| [assembly](assembly.md) | retrieval and rendering. **a WALK with a budget**, not a behavior. budget down, accumulation up — and why grading is a different wave. | **SPEC** — rev. 08-06 |
| [closure](closure.md) | **closure is one observation, this execution. `P` is retrieval.** supersedes `2026-07-27/truth.md`'s claim that `P` measures closure. | **RESOLVED** — 08-06 |
| [questions](questions.md) | decisions not yet made, carried across sessions. | **LIVE** |

## What the kernel build settled

Three of `2026-07-27/kernel-alteration.md`'s open questions closed, and one of its own
statements turned out to be wrong. See [kernel](kernel.md) for the arguments.

```text
§13 Q2  MINIMAL TRANSCRIPT     RESOLVED   SpaceID + Parent + Depth + Promoted + Realized.
                                          parentage was derived while Perturb was one-shot;
                                          it iterates now, so the record is a tree.

§13 Q4  SLICE CONCURRENCY      DISSOLVED  the kernel writes nothing. the transcript is
                                          RETURNED, per-wave state is private, and the
                                          contended write it was worried about is gone.

§13 Q5  IS NIL SUFFICIENT      CONFIRMED  nil means "I promoted nothing." held in code.

§10     GROUPED MEMBER IDS     WRONG      it asked the transcript to carry grouped member
                                          IDs and boundaries; law 8 forbids the kernel from
                                          enumerating them. law 8 won. the transcript is a
                                          SPINE — what the kernel was handed, never what a
                                          Space contains.
```

## What was stripped on entry

`kernel-alteration.md` was admitted to 07-27 carrying vocabulary that set did not define, marked
rather than deleted because *"stripping it is a design decision, and this was a move."* The
decision is now made: it did not survive contact with the implementation, because nothing in
the loop asked for it.

```text
CreationPhase                 nothing needs a phase to be visible or invisible yet
amplitude / evidence phase    the kernel scaled nothing, so nothing carried amplitude
Cost / affordability          no budget exists to be spent
concrete / segmented          one Space interface, no refinement needed so far
```

If the loop turns out to require any of them, they come back the same way everything else
does — because something could not be represented without them.

## Not yet in this set

```text
THE BOUNDARY             nothing spends the outbox. no model call exists, so no support
                         values flow, the fold has never run, and calibration — which
                         families.md says is the ONLY thing making confidence edges safe —
                         is never computed. this is the keystone.

CLOSURE                  every node's first question is "am I closed?" and nothing answers
                         it. see closure.md: it needs a threshold on a perceive's grade,
                         and one has not been picked.

PROPER TIME              the old Slice carried the wave count. nothing owns it now.

THE IMPULSE QUEUE        one stack means impulses queue. nothing produces impulses yet, so
                         nothing queues them — but the rule is already written: FIFO, and
                         admission may not consult attention.

EVERYTHING IN 07-27      trust, actions, workers, dynamics, and the five toys are still the
                         live reasoning. truth.md is superseded in part by closure.md;
                         the rest has not been re-derived here.
```

## What has been ANSWERED since

```text
§13 Q3  GROUPING BEHAVIOR     the Space does not encode it — it reads the ORDERING POLICY
                              off the parent expectation. sequential/concurrent is a flag
                              the parent declares. open since 07-23.

        THE TYPES             built. src/skyra/cognitive-agent/skyra-dna-v4/graph.

        THE DURABLE SUBSTRATE Namespace — the base node, which IS the graph.

        MANY ENTRIES          a frontier is a chain, not a fan-out. no iterating Perturb
                              needed for it (though one was later adopted for pipelines).
```
