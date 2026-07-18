# Traversal as Excitation

**Status:** WORKING MECHANISM (2026-07-17 research). Conceptual; not implemented in the shared
`Space`/relation kernel.
**Companions:** [region-as-playbook.md](region-as-playbook.md) ·
[relational-region.md](relational-region.md) ·
[unification-truth-maintenance-and-field-combination.md](unification-truth-maintenance-and-field-combination.md) ·
[the-learning-loop.md](the-learning-loop.md)

This note records the emerging excitation model: excitation is not a separate cognitive operation
performed before traversal. The traversal itself produces the excitation and the active field.

## One traversal, many observation roots

Every currently excited observation does **not** receive an independent traversal. The excited
observations form the root set of **one traversal**:

```text
excited observations
    O1  O2  O3
     \  |  /
      one traversal
           |
           v
    one lit relational morphology
```

The kernel may represent this mechanically as a forest or as a synthetic active-field root whose
children are the excited observations. The synthetic root has no cognitive meaning; it only gives
the kernel one traversal origin.

The completed traversal is the active field:

> **The active field is the local relational workspace in which current state, target state, and
> the presently lit bridge between them coexist.**

The relational field remains the durable collection of regions and their histories. The active
field is the traversal-local morphology lit from the currently excited observation roots.

## Occurrence means occurrence in the current tree

When a cell checks whether it "exists elsewhere," the immediate meaning is:

> Does this same cell already occur elsewhere in the tree the kernel is currently building for
> this traversal?

It does **not** mean searching durable memory for a semantically similar cell. Keep these two
mechanisms separate:

```text
CURRENT-TRAVERSAL CONVERGENCE
    the same durable cell is reached through several paths in this traversal

REPRESENTATIONAL MATCHING
    different cells are treated as relationally equivalent
```

Traversal convergence is the mechanism under discussion here. Exact triples, compatible region
bindings, compression, and partial structural similarity belong to the later representation
question.

## Observe interacts with the traversal-so-far

Observe is not merely a notification that the kernel reached a cell. It is the interaction between
the cell and the tree currently being constructed.

Conceptually:

```text
kernel visits cell
    -> cell inspects the traversal-so-far
    -> cell detects its current-tree occurrences
    -> cell records traversal-local state on the Slice
    -> cell returns the spaces this traversal may follow next
    -> kernel extends the tree
```

The ownership boundary is:

```text
KERNEL
    owns tree construction, traversal order, admission, termination, and the fold

CELL
    owns what an occurrence means, how it affects local state, and which frontier it exposes

SLICE
    records the traversal-local occurrences and convergence
```

The kernel may know that a cell occurred several times through several paths. It must not know what
that count means cognitively.

## Tallies are a sufficient first record

The smallest useful traversal-local representation may be a tally:

```text
current traversal occurrences

observation-1     1
state-change-A    2
action-X          3
observation-9     1
```

This records a fact about the traversal: `action-X` was reached three times in the current tree. It
does not prematurely declare that the count means confidence, conjunction, relevance, scheduling
priority, or region membership.

Later, if required, an occurrence may retain its provenance:

```text
action-X
    occurrence 1: root O1, path P1, tick t1
    occurrence 2: root O2, path P2, tick t2
    occurrence 3: root O3, path P3, tick t3
```

Start with the tally. Preserve paths, roots, and ticks only when an experiment demonstrates that
the count alone loses necessary structure.

## Traversal convergence is structural excitation

A cell's excitation is produced by its participation in the current traversal:

```text
more convergent paths
    -> more occurrences in the current tree
    -> more traversal-local evidence on the Slice
    -> cell-specific behavior may change
```

This suggests the working principle:

> **A cell is excited insofar as it participates in the currently realized traversal.**

The Slice records the structural fact of participation. The cell or a later learning layer decides
how participation affects frontier, action selection, or learning.

Several observation roots may converge on one action:

```text
O1 -> state-change-1 --+
                       +--> action A
O2 -> state-change-2 --+
```

The convergence must remain visible. It might eventually mean that both observations jointly
support the action, that either independently suggests it, or that the observations are redundant.
The kernel does not choose among those meanings.

## The field representation must supply the analogue of phase

Path magnitude or occurrence count is insufficient to determine what an overlap means. Two paths
may reinforce one another, compose into a larger structure, bind incompatible values into the same
relational position, remain independent, or leave the position unresolved.

The useful interference analogy is:

```text
QM-like idea                  Skyra mechanism

multiple paths                traversal paths from several roots
superposition                 several candidate structures remain live
constructive interference     compatible paths reinforce shared structure
destructive interference      incompatible paths weaken, block, or split structure
interference pattern          resulting activation morphology across the field
measurement/collapse          budgeted selection and committed expression
decoherence/decay             old traversal contributions leave the active field
```

Skyra does not thereby require complex-valued quantum phase. It does require a relational
combination rule that answers:

> **What happens when two traversal paths try to bind incompatible values into the same relational
> position?**

Conceptually:

```text
Combine(path contribution A, path contribution B)
    -> reinforce | compose | conflict | independent | unresolved
```

This rule belongs to the field/relational-region representation, not to the void kernel. The kernel
records paths, occurrences, identities, and convergence. The representation determines relational
compatibility and preserves contradictions rather than flattening every occurrence into a positive
tally.

The chosen compression representation must therefore preserve enough structure to distinguish:

```text
same binding        -> reinforce
compatible bindings -> compose
opposing bindings   -> conflict or split
unrelated bindings  -> remain independent
unknown binding     -> remain unresolved
```

This is a requirement on the future representation, not a decision that triples, signed weights,
complex amplitudes, or any other particular syntax is correct.

The strongest researched candidate decomposition is now: unification for binding compatibility,
constraint propagation for remaining domains and local consistency, an ATMS-like layer for minimal
support environments and contradiction provenance, and factor-graph-like message passing for local
aggregation. See
[unification-truth-maintenance-and-field-combination.md](unification-truth-maintenance-and-field-combination.md).

## Tree mechanics must preserve graph convergence

The traversal may be tree-shaped mechanically while several tree appearances refer to one durable
cell:

```text
tree appearances: 2
durable identity: action A
```

A global visited-once rule would destroy convergence by keeping only the first path that reaches
the shared cell. Cycle prevention should therefore be path-local:

```text
do not visit the same cell twice on one active path
```

not global:

```text
do not allow the same cell to occur through two different roots or paths
```

The Slice can preserve that the tree appearances converge on one durable identity.

## Observe should return the traversal-local frontier

The intended conceptual contract is:

```text
Observe(traversal context) -> map of onward Spaces
```

or an equivalent explicit frontier value.

This distinguishes:

```text
durable adjacency
    relationships retained by the cell across time

traversal-local frontier
    relationships this particular occurrence exposes to this traversal
```

The current shared code does not yet express this contract. `Space.Observe` returns nothing, and
the relation kernel calls `Children()` afterward. That encourages a cell to communicate its
per-wave frontier by mutating hidden child state. The standalone traversal experiment has a
`Frontier`-like contract, but it has not been collapsed into the shared `Space` interface.

This note records the desired semantics, not an implementation decision about the eventual Go
signature.

## Current state as traversal morphology

The direction suggests:

> **Current state is the lit morphology produced by one traversal rooted in the currently excited
> observations.**

The durable observations do not disappear when their excitation decays. What decays is their
participation in the active traversal:

```text
durable observation history
            |
            | excitation/admission
            v
current observation root set
            |
            | one traversal
            v
current active morphology
```

The exact relationship among decay, target excitation, observation admission, and the root set is
still open.

## Significance delta and episode segmentation

Keep an observation's durable content separate from its role in a particular traversal:

```text
CONTENT
    what the durable observation recorded

SIGNIFICANCE
    what role an occurrence of that observation plays in traversal T
```

The observation does not change when a later interaction arrives, but its significance may:

```text
O0@T0 -> confirmation
O0@T1 -> contradiction
O0@T2 -> background
```

This makes the delta between traversals a candidate endogenous episode-segmentation signal. The
important change is not merely which durable nodes appeared, but how the overlap pattern and the
roles of its occurrences changed:

```text
same patterns retain significance
    -> the episode may continue

patterns reorganize, split, disappear, or reverse role
    -> candidate episode boundary
```

The collapse/carry operation determines which patterns from one traversal participate as roots in
the next. A disappearance is not automatically a resolution; it may instead be decay, budget loss,
compression loss, suppression, or supersession.

## A recurrent traversal loop

The emerging recurrence is:

```text
incoming interaction It
+ structures carried from T(t-1)
            |
            v
        root set Rt
            |
            v
        traversal Tt
            |
            v
   interference patterns Ft
            |
            v
       collapse Ct
        /          \
       v            v
carry into T(t+1)  expression At
                         |
                         v
                   observation I(t+1)
```

At cold start, the interaction introduced into the field may be the whole root set. On later turns,
the new interaction perturbs structures carried from the preceding traversal:

```text
Rt = Carry(T(t-1)) + Interaction(t)
```

The next context is reconstructed, not appended. A later observation may cause an earlier
observation to occupy a completely different relational role without rewriting its durable
content.

Patterns created by overlap at `Tt` are candidates to become roots at `T(t+1)`. Not every raw
observation needs to roll forward; the carried object may be a compressed interference pattern.
Without a budget or collapse, recursively carrying every pattern would make the active field grow
without bound.

## A region may first be an interference pattern

A relational region need not begin as a durable container waiting to be activated. It may first be
the pattern that appears when several traversal paths overlap:

```text
path from O0 --+
               +--> overlap pattern --> region occurrence --> behavior
path from O1 --+
```

This suggests two levels:

```text
REGION OCCURRENCE
    one interference pattern appearing in a particular traversal

DURABLE REGION
    a compression of a pattern that recurs across traversals
```

Repeated region occurrences may eventually justify a reusable representation:

```text
T0 produces pattern P0
T1 produces similar pattern P1
T2 produces similar pattern P2
            |
            v
durable compressed region P
```

The compression representation must decide which differences are irrelevant and which are
behaviorally load-bearing. The field representation's phase analogue—the combination rule for
reinforcing, composing, conflicting, independent, and unresolved contributions—is therefore part
of region formation.

## Current and target may be roles, not separate objects

The newer working direction does not require a separately materialized `CurrentState` or
`TargetState`.

Current state may be the **observed side** of an active relational pattern. Target state may be the
**directional or desired role** played by another part of that same pattern:

```text
active pattern
    observed:
        what interaction and world observations establish

    directional:
        what the interaction requests, requires, or leaves unresolved

    behavioral:
        expressions associated with resolving that pattern
```

For an observed request, the desired relation may already be present in the decomposed interaction.
Copying it into a second target object may add no mechanism.

The current/target distinction can remain a useful explanatory projection over the active pattern
without becoming two stores, node types, or kernel primitives:

> A target is not necessarily a separate object. It may be a role played by part of the active
> pattern. Likewise, current state may not be a separate object; it may be the observed side of the
> active pattern.

This is a working refinement, not yet a deletion of every use of target language. The design still
needs the functions previously supplied by target state.

## Removing target does not remove the objective

If explicit target state disappears, the mechanism must still answer:

- Why was one action selected?
- What observable result would count as success?
- When does the interaction stop recruiting action?
- How is an expression graded?
- How is resolution distinguished from simple decay?
- What prevents collapse from forgetting an unresolved interaction?

The objective cannot disappear. It may move into the relational pattern, the field representation,
and the collapse function.

Instead of:

```text
CurrentState != TargetState
```

the active field may contain:

```text
an unresolved directional pattern that still affords expression
```

Instead of:

```text
current matches target -> closure
```

a returned observation may transform the pattern into an observably resolved morphology.

That resolution must remain observation-authored. A pattern leaving the next context is not enough.
The carry/collapse representation must preserve distinctions such as:

```text
RESOLVED
    a returned observation established resolution

DROPPED
    budget, decay, or compression removed the pattern from the active field

UNRESOLVED
    the pattern remains directional but is not currently selected

SUPERSEDED
    a later observation revised or canceled the direction
```

Only `RESOLVED` supplies closure evidence.

## Observe builds; Express proceeds from the result

The current high-level split is:

```text
OBSERVE
    each visited cell reads the traversal-so-far
    records its occurrence
    returns its traversal-local frontier

TRAVERSE
    the kernel builds one multi-root tree
    the Slice records repeated occurrences and convergence

EXPRESS
    proceeds from the completed traversal/active field
```

Do not yet fix whether Express:

- executes during post-order ascent;
- returns action candidates into a final fold;
- selects one path at a root;
- executes one next action;
- or executes a partially ordered runnable set.

Those are consequences of the planning order, base case, and episode-segmentation mechanisms, which
remain open.

## Collapse may have two stages — leave the boundary open

Two operations may be hiding under the word collapse:

```text
REPRESENTATIONAL COLLAPSE
    large traversal field
        -> compressed structures fitting the context budget

BEHAVIORAL COLLAPSE
    live candidate behaviors
        -> committed expression/action
```

The field representation is expected to eliminate or compile most behavioral ambiguity when stable
patterns exist. A parser may produce the action directly. An LLM may handle a novel or unresolved
field. It remains open whether representational and behavioral collapse are two explicit stages,
one combined operation, or different implementations of one interface.

Do not let either operation write world state. Collapse may carry structure or emit expression; a
returned observation is still required to change durable evidence or establish resolution.

## The base case is deliberately unsettled

Do not fix the traversal base case yet.

Candidate formulations have included reaching an explicit target, satisfying a target predicate,
meeting observation-grounded current morphology during backward planning, and reaching a resolved
active pattern that no longer affords corrective expression. The move from explicit current/target
objects toward roles inside an interference pattern changes the stopping question.

The base case must eventually distinguish at least:

```text
resolved by observation
unresolved but actionable
unresolved and requiring perception/inference
conflicting
dropped or decayed without resolution
```

Until the field/compression representation can express those conditions, selecting a base case
would be premature.

## Ordering caution

If a cell behaves differently on its first and later occurrences, traversal order can affect the
result:

```text
O1 reaches A, then O2 reaches A
```

versus:

```text
O2 reaches A, then O1 reaches A
```

The final tally is commutative (`A occurred twice`), but "expand only on first occurrence" is not.
Root ordering, sibling ordering, traversal ticks, and cell behavior therefore need either
deterministic rules or order-independent convergence.

Slice ticks may record occurrence chronology, but tick order must not automatically be treated as
causal or executable action order. Historical event order, current traversal order, and playbook
dependency order are distinct.

## Open

- What admits a durable observation into the excited root set?
- What exactly decays: root eligibility, occurrence contribution, edge activation, or all three?
- What current-tree view does a cell receive during Observe?
- Is a tally enough, or must each occurrence retain root/path provenance?
- How does a repeated occurrence change the frontier without making results traversal-order
  dependent?
- How are AND support, OR support, and redundant evidence distinguished?
- What field representation supplies the phase analogue: the compatibility rule for reinforcing,
  composing, conflicting, independent, and unresolved path contributions?
- Is target a separate object anywhere, or only a directional role inside an active pattern?
- Where do direction, stopping, grading, and unresolved carry live if target is not explicit?
- What is the base case? Explicitly unsettled pending the field/compression representation.
- How does Express turn the completed traversal into planning/execution order?
- Does an episode segment correspond to the lifetime of one target-conditioned traversal or to
  several linked traversals?

These questions should be settled at the mechanism level before choosing triples, compression, or
another representation for relational regions.
