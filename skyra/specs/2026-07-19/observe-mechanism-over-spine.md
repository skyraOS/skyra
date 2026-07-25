# Current Focus — Observe Mechanism over the Spine

**Date:** 2026-07-19
**Status:** Conceptual mechanism, pre-validation. Paper test (toy spine) pending before
implementation. Compression remains the parent research thread; a simpler mechanism may
still be hiding inside it.

> Paper test run 2026-07-19 — see [observe-paper-test.md](observe-paper-test.md).

## Summary

The observe pass is being formalized as spreading activation over an immutable causal
spine, with interpretation confined to two collapse bookends.

The durable substrate is the **spine**: an append-only, runtime-witnessed sequence of

```text
observation -> action -> observation -> action ...
```

Nothing above the spine is authoritative. Entities, lenses, and episodes are **derived
indexes** over spine history — deletable, rebuildable, and cheap to be wrong about.
Mutability tracks epistemics:

```text
witnessed   spine succession, settlement records        immutable
counted     causal statistics over spine pairs          updatable
interpreted resolution proposals, annotations           revisable
transient   per-traversal activation (Slice)            disposable
```

**Resolution is a spine event.** Entity-resolution collapse commits a reading of an
impulse; that commitment is an act, graded like any other act — by user correction or
downstream settlement. Confusion-then-correction episodes are retained permanently as
precedent, not overwritten. Lenses (entity -> episode pointer sets) are materialized
caches over settled resolution history, never hand-edited.

**Interpretation only at the bookends.** Between the resolution collapse (observe) and
the action collapse (express), everything is mechanics: conservation-based spreading
activation, path-local cycle prevention, additive Slice accumulation, commutative merge.

## Wave shape

```text
ingest       impulse -> observation node; candidate resolutions PROPOSED (plural allowed)
lens-hop     candidate entities -> episodes they participate in
descend      episodes -> spine spans
spine-walk   along causal succession, both directions
grade-read   at each action, read its settlement
converge     Slice accumulates; hottest span = precedent, settlement attached
collapse     resolution COMMITS as a spine event (strict, singular, falsifiable)
```

Strict at commit, plural at propose.

## Edge kinds (complete inventory for the observe pass)

```text
1. O -> A -> O            spine succession          witnessed
2. O -> entity            resolved-into (event)     spine fact
3. entity -> O            resolved-from             derived index
4. entity -> episode      participates-in           derived (settled resolutions)
5. episode -> O..O        spans                     derived (compressed hot spans)
6. A -> O                 settled-by                witnessed (prediction settlement)
```

> **2026-07-20 notice:** kind 6 revised to an append-only, many-to-many **settlement
> set** `A → {(O, grade, t), ...}` — see
> [settlement-set](../2026-07-20/settlement-set.md). Act-completion settles once;
> state-holding settles repeatedly (the renewal profile). Witnessed only via
> birth-stamped settlement conditions; these are the spine's long-range cross-span
> edges.

## Next step

Run the paper test: ~15-moment toy spine, five lenses, one impulse
("go get my water from the fridge"). Check that the correct precedent span heats and
ambiguous references resolve by convergence. Watch for unused edge kinds or layers that
always move together — candidates for deletion into a simpler compression-native
mechanism.

## QM-inspired 1:1

The analogy supplies mechanics, not a claim that the system is quantum.

```text
QM concept                    mechanism here

amplitude (magnitude)         activation energy arriving via a path
phase                         relational compatibility of bindings (Combine rule; open)
superposition                 plural candidate resolutions live before collapse
multiple paths                traversal paths from several impulse roots
constructive interference     compatible paths converging on shared spine spans
destructive interference      constraint conflict / contradicting bindings (signed
                              representation still an open decision)
interference pattern          activation heatmap over episodic spine history
measurement / collapse        the two bookends: resolution commit and action commit
Born rule (|amplitude|^2)     counted settlement statistics selecting among candidates
unitary evolution             the mechanical middle — deterministic, interpretation-free
irreversibility of collapse   committed resolutions/actions are spine events; never
                              rewritten, only graded and superseded
decoherence                   carry/budget decay; old contributions leave the active
                              field without being erased from the spine
conservation of energy        amplitude conservation across a cell's frontier
                              (fan effect; promiscuous nodes self-attenuate)
observer effect               resolution acts change the world they read (corrections
                              become new spine evidence)
```

Two honest divergences: activations are currently non-negative (no cancellation without
a signed representation), and the middle is not reversible in practice — only the
epistemic discipline (nothing writes durable truth but observation) is preserved, not
unitarity.

> **2026-07-20 notice:** the first divergence is resolved as design, not compromise —
> [paper test #3](../2026-07-20/observe-paper-test-3.md) retires the signed middle
> (three tests, nothing to cancel) and settles the phase row: compatibility =
> witnessed joint co-occurrence of committed bindings, checked at the collapse
> bookend; zero joint precedent = novelty flag.
