# Symbols That Still Stand in dnav3

## Method

A criticism still stands when dnav3 has not yet supplied the kind of evidence
the original criticism demanded. This does not mean dnav3 made no progress. A
criticism may become less severe while remaining unresolved.

The distinction matters. If a mechanism now exists but has only been exercised
inside controlled fixtures, the criticism should move from “there is no
mechanism” to “the mechanism has not yet established the claimed behavior.” It
should not be repeated as though nothing changed, and it should not be retired
as though implementation were validation.

Four original symbols still stand.

## ⟦Δ≠DEV⟧ State change has not yet established development

This remains the central unresolved criticism.

dnav3 contains much more meaningful state change than v.05 or dnav2. It records
world-derived observations, model-authored expressions, responsible acts,
falsifiable expectations, settlement evidence, conditional-probability edges,
activation fields, interpretation claims, calibration histories, and planning
metrics. Its updates are no longer arbitrary movement inside a graph. They are
connected to events that crossed a world boundary.

That is a serious answer to the first half of the criticism: dnav3 now has a
plausible mechanical account of how experience could alter later cognition.

It has not yet completed the second half: demonstrating development.

The flagship behavioral thesis is explicit:

```text
passes-to-closure  ↓
prediction error   ↓
```

But the inspected evidence does not yet show both curves falling for a live
general-purpose model over varied, repeated, held-out tasks. The clearest
passes-to-closure test uses a mock inference function deliberately written to
return a complete plan whenever `PRECEDENT` appears in the prompt. That test
proves that precedent can be stored, retrieved, rendered, noticed by the mock,
and converted into a cheaper plan. It does not prove that an ordinary model
will use precedent correctly, distinguish relevant precedent from misleading
similarity, or improve across a distribution it was not scripted to recognize.

The original longitudinal demand therefore remains:

1. Establish behavior before a consequential experience.
2. Record the particular internal changes caused by that experience.
3. Demonstrate an appropriate later behavioral difference caused by those
   changes.
4. Compare against stateless, raw-history, and ablated-memory baselines.
5. Repeat across non-identical instances and held-out environments.
6. Show that improvement survives restarts, model changes, and irrelevant
   intervening experience.

dnav3 has made development testable. It has not yet made development an
observed property.

Status: **standing, substantially weakened by real mechanical progress**.

## ⟦PERSIST≠CONTINUITY⟧ Persistence is not yet equivalent to continuity

This criticism still stands almost unchanged at the process boundary.

dnav3 makes a more thoughtful distinction between transient and retained state.
The observation graph belongs to a world episode. The expression graph survives
`NewEpisode` and is treated as learned memory. The activation field is a
continuously replaced working projection rather than the durable history itself.
That is a stronger internal account of continuity than simply appending more
conversation history.

However, the inspected stores are still in-memory Go structures. The code shows
continuity across turns and across an in-process episode replacement. It does
not show continuity across:

- process termination and restoration;
- schema or runtime migration;
- model-provider replacement;
- partial writes or corruption;
- compaction of very long histories;
- deletion and selective forgetting;
- reconciliation between multiple machines;
- years of accumulated experience.

The being can enter a fresh toy world without forgetting its expression graph.
That is useful evidence for where the boundary should be. It is not yet evidence
that a being can survive the death and reconstruction of the process that holds
the graph.

The criticism is not “nothing persists.” Things clearly persist within the
running object. The criticism is that the architecture continues to use words
such as *durable memory* while the strongest forms of continuity remain
unimplemented and untested.

Status: **standing**.

## ⟦COMPLEXITY>PROOF⟧ Complexity is growing faster than demonstrated capability

This criticism still stands, although dnav3 also contains evidence of healthy
compression.

dnav3 deleted dnav2's annotated causal-node apparatus, active *aitia* set,
causal-belief expansion, and LLM grading path. That is exactly the kind of
architectural compression the earlier criticism demanded: a weaker foundation
was removed rather than protected by another layer.

At the same time, the current research surface includes:

- expression graphs and observation graphs;
- conditional-probability edges;
- span-scoped eligibility;
- continuous activation fields;
- active-memory token density;
- interpretation routes;
- ambiguity-preserving candidate groups;
- density pockets;
- independently falsifiable interpretation claims;
- multidimensional prediction fields;
- evidence attribution;
- held-out calibration;
- same-claim sensor fusion;
- two spacetime kernels;
- flat and spherical world implementations;
- a separate user boundary.

Many of these mechanisms answer a specific contamination or leakage failure.
The tests are unusually explicit about those failures. Nevertheless, the
integrated behavioral evidence remains much smaller than the conceptual
surface.

The interpretation system illustrates the tension. It contains sophisticated
protections against ambiguity collapse, lucky-hit credit, self-labeling,
selection/evaluation overlap, and current-span leakage. Some mechanisms remain
proposal-stage, intentionally provisional, or callable mainly through direct
test construction rather than the main planning loop. The conceptual account
is advancing faster than an end-to-end being can exercise it.

This symbol should now demand ablation, not mere simplification:

- What behavior measurably worsens when a mechanism is removed?
- Does it lower prediction error on held-out cases?
- Does it lower planning cost without hiding failures?
- Does it prevent a reproduced contamination bug?
- Does its benefit survive outside the fixture that motivated it?

Without those comparisons, internal rigor may again become the project's
primary evidence.

Status: **standing, with evidence that the project is capable of responding to
it through deletion and compression**.

## ⟦PROTO≠PROD⟧ Production readiness remains distant

This criticism stands clearly.

The inspected code calls dnav3 an active cognition experiment. The
interpretation document labels itself a proposal. The distributed-agent README
admits that “distributed” is forward-looking and that both worlds are local and
in-process. This honesty prevents the criticism from becoming an accusation of
misrepresentation, but it does not make the systems production-ready.

The current architecture remains transitional:

- dnav3 is deliberately single-threaded and does not use the spacetime kernel;
- the spatial worlds use older spacetime contracts;
- the relation kernel has return perception explicitly suspended;
- a newer stack kernel exists alongside it;
- the multi-stack runner is documented as currently unreachable;
- durable cognitive state is not persisted across process restarts;
- the user agent is a minimal interaction log;
- the four supplied architectural regions are not yet one demonstrated runtime.

There is also a concrete safety gap. The dnav3 filesystem realm describes itself
as sandboxed, but its `exec` primitive invokes `sh -c` with only the working
directory changed. That does not prevent absolute-path access, parent traversal,
network access, environment-variable access, or use of any capability available
to the host process. The path resolver's string-prefix boundary is also weaker
than a component-aware containment check.

The realm is suitable as a controlled experimental fixture. It is not a security
sandbox.

Production readiness would require evidence around isolation, authentication,
authorization, persistence, migrations, cancellation, recovery, idempotency,
remote failures, observability, provider degradation, and long-running resource
behavior.

Status: **standing**.

## Summary

The symbols that still stand are:

```text
⟦Δ≠DEV⟧
⟦PERSIST≠CONTINUITY⟧
⟦COMPLEXITY>PROOF⟧
⟦PROTO≠PROD⟧
```

Their common structure is no longer “there is no mechanism.” dnav3 has supplied
substantial mechanisms. Their common structure is now “the mechanism has not yet
established the full longitudinal, durable, integrated behavior claimed.”

