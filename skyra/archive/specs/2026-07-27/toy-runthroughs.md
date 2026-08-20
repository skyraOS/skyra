# Adversarial Toy Runthroughs

**Date:** 2026-07-27  
**Status:** TESTS FOR THE DESIGN. These are not proposed mechanisms. Each runthrough names a
failure the architecture must make visible before a mechanism is admitted to address it.

The five toys press on the boundaries of the stripped loop:

```text
observation  ──support──▶  expectation  ──policy──▶  action  ──▶  observation
```

A toy passes when the failure is either prevented or represented honestly as unresolved. It
does not pass merely because the loop continues running.

> **Two lineages now run against these toys.** Resolutions dated `2026-07-30` are against the
> `2026-07-27` → `2026-08-06` line, which is the live design. Sections marked **RUN AGAINST
> 2026-08-08** are against the whiteboard model in `specs/2026-08-08` — SELECT/RETURN, possibilities,
> edge-owned credit — which is not built and is not live. A toy may hold differently under each, and
> where it does, both readings are kept.

---

## Toy 1 — The blind interpreter converges on the wrong reading

### What it tests

Blindness prevents desire from leaking into interpretation. It does not, by itself, make the
interpretation correct.

### Setup

There are two expectations:

```text
E₁   "the cell is charging"
E₂   "the cell is discharging"
```

The world repeatedly returns a raw payload:

```text
status=connected · current=-2A
```

In this device vocabulary, negative current means current is leaving the cell: it is
discharging. The observation worker does not know that convention and consistently reads the
word `connected` as evidence of charging.

The action worker cannot contaminate this reading. It never sees the raw payload, and the
observation worker never sees the desired expectation or action.

### Run

```text
wave    assigned support       revision?

 1      c(o₁,E₁) = +0.8        initial reading
 2      c(o₂,E₁) = +0.8        no
 3      c(o₃,E₁) = +0.8        no
 ...
10      c(o₁₀,E₁) = +0.8       no
```

The interpreter is blind, stable, and wrong.

Its revision rate falls toward zero. Trust in `E₁` rises. Coherence stays high. Every currently
defined internal diagnostic makes the reading look healthy.

Later, a differently expressed observation arrives:

```text
battery percentage: 80 → 61
```

That observation bears negatively on `E₁` and positively on `E₂`. The graph is now contested,
but it still contains no oracle that declares which vocabulary reading was correct.

### Failure exposed

```text
STABILITY ≠ CORRECTNESS
BLINDNESS ≠ INDEPENDENCE
```

Two mutually blind workers can share the same learned misconception. A falling revision rate
only says the interpreter is becoming self-consistent.

### Pass condition

The system must not report the first ten readings as established correctness merely because
they are stable. When the percentage trajectory arrives, it must:

1. preserve the earlier readings and their provenance;
2. expose the conflict rather than silently rewriting history;
3. distinguish `stable`, `contested`, and `externally contradicted`;
4. allow the later cross-form evidence to revise which reading receives operational weight.

If no contradictory observation ever arrives, the correct state is **stable but not independently
validated**, not true by convergence.

### Measurement

At minimum the run should report separately:

```text
interpretation revision rate
cross-form contradiction rate
number of independent observation forms grounding the reading
```

### Open design question

What measurement complements revision rate without recreating a universal deterministic oracle?
Candidates may include cross-form prediction, independently generated readings, and deterministic
instruments where a boundary actually exposes one. This toy does not choose among them.

### RUN AGAINST 2026-08-08 — PASSES

Two decisions in that model bear on this toy.

**The edge carries the return it cited**, stored on change. Ten identical readings record as one
entry with `n=10`; an eleventh that differs is a **change event**. That satisfies ① — the earlier
readings and their provenance survive — which was impossible while only the fact of use was durable.

**Credit records carry scope**, and the three states the toy demands fall out of readings the model
already produces:

```text
STABLE                   no change event. n rising on one entry.
CONTESTED                a change event on the same edge.
EXTERNALLY CONTRADICTED  a long-range negative from a rejected chain.
```

*"Stable but not independently validated"* — the state the toy insists on when no contradiction ever
arrives — is expressible with no new mechanism: **short-range strong, long-range absent.**

**Not satisfied:** the measurement asks for the *number of independent observation forms grounding
the reading*. The edge set contains it. Nothing computes it.

---

## Toy 2 — Ten correlated records masquerade as ten grounds

### What it tests

`O(E)` currently contains distinct observation records. Distinct records are not necessarily
independent evidence.

### Setup

One upstream sensor publishes:

```text
source event S: "door is closed"
```

Ten subscribers copy the event into ten records:

```text
o₁ ... o₁₀
```

Each record carries the same content and the observation worker assigns the same support:

```text
c(oᵢ,Eclosed) = +0.8
```

All ten records ultimately derive from `S`. No subscriber looked at the door.

For comparison, another expectation has one genuinely independent camera observation:

```text
c(camera,Eopen) = +0.9
```

### Naive evaluation

```text
T(Eclosed) = 10 × 0.8 = 8.0
T(Eopen)   =  1 × 0.9 = 0.9
```

Trust ranks the copied report almost nine times above the direct observation. Attention then
spends more of its frame on `Eclosed`, increasing the practical effect of the inflation.

### Failure exposed

The graph counted transport fan-out as epistemic breadth.

Simple content deduplication is not a solution:

```text
"I checked the door ten times over an hour"
```

may be ten useful observations of sustained closure. The relevant distinction is shared cause,
not repeated text.

### Pass condition

The ten copied records must remain ten auditable events, but they may not contribute ten units of
independent grounding. The evaluation must be able to recover that:

```text
o₁ ... o₁₀ share one evidential origin S
camera is a second evidential origin
```

Repeated direct measurements at meaningfully different times must remain capable of accumulating
mass. A fix that collapses every repetition into one observation fails the other side of the toy.

### Measurement

Report both raw and effective breadth:

```text
record count
independent-ground count, or effective independent mass
largest provenance cluster
trust before and after correlation handling
```

### Open design question

What makes observations independent: provenance, time separation, intervention, measured
co-occurrence, or some combination? Until that is defined, Trust is an upper bound on independent
grounding rather than a finished measurement of it.

### RUN AGAINST 2026-08-08 — DISSOLVES for credit · DORMANT for admission

The failure required counting **records**. That model counts **invocations** — an edge accumulates a
record when the agent made a select and cited the return.

```text
TRANSPORT FAN-OUT   happens OUTSIDE the agent. produces no selects. INFLATES NOTHING.
TEN GENUINE CHECKS  ten selects. accumulates — which the toy explicitly requires.
```

Both sides hold, and by one structural property rather than a deduplication rule, which is the fix
this toy warns against.

**The residual is dormant, not gone.** A fan-out source can still deliver ten copies at the
*boundary*. That matters only if **recurrence** is chosen as the ingress mechanism for Toy 5, where
volume would force admission. So Toy 2 still blocks that specific candidate and nothing else.

---

## Toy 3 — One action produces several expectation changes

### What it tests

The rule says both directions pair exactly once:

```text
ACTION → EXPECTATION   exactly one
EXPECTATION → ACTION   exactly one
```

Physical actions commonly have more than one observable effect.

### Setup

One switch controls a light and a fan:

```text
Aflip   flip switch S once

Elight  "the light changes from off to on"
Efan    "the fan changes from stopped to spinning"
```

The world executes one physical action and returns observations bearing on both changes.

### Three tempting representations

**Attach the action to both expectations**

```text
Elight ──action──▶ Aflip
Efan   ──action──▶ Aflip
```

This violates the stated one-to-one pairing unless both edges may point to the same immutable
action occurrence.

**Duplicate the action**

```text
Elight ──▶ Aflip₁
Efan   ──▶ Aflip₂
```

This falsely records two executions and may physically flip the switch twice, undoing the change.

**Create one compound expectation**

```text
Ecompound   "light turns on AND fan starts"
```

This preserves cardinality but loses the fact that the light may succeed while the fan fails.
The two changes need separate positions and separate falsification.

### Run

The switch closes, the light turns on, and the fan motor is broken:

```text
ΔP(Elight) > 0
ΔP(Efan)   = 0
```

One action occurrence has two distinguishable consequences.

### Failure exposed

One-to-one pairing is unambiguous only if `action` means something narrower than a physical action
occurrence, or if one occurrence may be referenced by several expectation records without being
owned by several of them.

Deleting credit splitting is valuable, but cardinality cannot erase real multi-effect actions.

### Pass condition

The representation must:

1. execute the physical action exactly once;
2. preserve separate positions for `Elight` and `Efan`;
3. learn separate efficacy values for the two changes;
4. record that both measurements arose after the same action occurrence;
5. avoid turning partial success into compound success or compound failure.

### CLOSED — 2026-07-30

The one-to-one rule is gone. An action has `≥1` expectation and is a leaf; an expectation has
`0:*` actions and `0:*` children ([expectations](expectations.md)). The switch decomposes:

```text
E "the switch changes open → closed"          actions: [flip]
   ├── E "the light changes off → on"          actions: [perceive_light]
   └── E "the fan changes stopped → spinning"  actions: [perceive_fan]
```

Against the five pass conditions: `flip` executes once ①; `Elight` and `Efan` hold separate
positions ②; each learns its own efficacy through its own leaf ③; both sit beneath the same
action occurrence, which the Slice transcript records ④; and partial success stays partial,
since neither inherits the other's closure ⑤.

The question the toy asked — *operator, occurrence, prediction edge, or physical causation* —
is answered: **the cardinality was never about physical causation.** It constrains the edge.
An action occurrence may be referenced by several expectations without being owned by any of
them, because actions own nothing.

### RUN AGAINST 2026-08-08 — STILL CLOSED, and ④ improves

One select for `flip`, then separate selects for `perceive_light` and `perceive_fan`. One execution
①; separate positions ②; separate edges and therefore separate efficacy ③; partial success stays
partial because the two edges are independent ⑤.

④ — *both measurements arose after the same action occurrence* — is served by **span**, which
already groups everything in one invocation. Co-occurrence is free rather than reconstructed from a
transcript.

---

## Toy 4 — One expectation requires several physical actions

### What it tests

A state change may require a sequence or conjunction of acts even when the design assigns it one
direct action.

### Setup

The expectation is:

```text
Eopen   "the sealed door changes from closed to open"
```

Opening it requires:

```text
Aunlock   unlock the deadbolt
Aturn     turn the handle
Apull     pull the door
```

No single action is sufficient in the initial state.

### First decomposition

The sequence can be expressed as intermediate changes:

```text
Eunlocked  ──▶ Aunlock
Eunlatched ──▶ Aturn       supported by Eunlocked
Eopen      ──▶ Apull       supported by Eunlatched
```

This is the strongest interpretation of the current design. Each expectation has one proximate
action, while prior physical actions survive as grounded prerequisite changes rather than a list
of joint act IDs.

### Run A — decomposition works

```text
Aunlock   closes Eunlocked
Aturn     closes Eunlatched
Apull     closes Eopen
```

The field represents three changes and gives each one a separately falsifiable action. No credit
split is needed.

### Run B — hidden conjunction

Now the door requires two people to pull opposite releases at the same time:

```text
Arelease-left
Arelease-right
```

Neither produces a stable intermediate world change. The door opens only under their temporal
conjunction.

Inventing `Eleft-released` and `Eright-released` would assert state that the world does not
independently sustain. Assigning only the last action as the direct cause erases the other
necessary act.

### Failure exposed

Sequential composition may fit the one-action rule through honest intermediate expectations.
Irreducibly joint or simultaneous action does not obviously decompose that way.

The architecture must distinguish:

```text
multiple steps with observable intermediate changes
multiple acts forming one irreducible intervention
```

### Pass condition

For Run A, the existing decomposition must work without an implicit joint-act list.

For Run B, the system must either represent the joint intervention explicitly or remain
unresolved. It may not:

- fabricate intermediate changes;
- credit only whichever act happened last;
- represent two physical executions when there was one coordinated intervention;
- claim that a later `perceive` caused the door to open.

### RESOLVED for Run A · OUT OF SCOPE for Run B — 2026-07-30

**Run A passes**, and more cleanly than the first decomposition managed. Each step is an
expectation with its own leaf action, chained, so execution is sequential by structure. No joint
act list, no invented intermediate world state.

**Run B is declared unrepresentable.** The ordering rule is *siblings concurrent, chains
sequential* — an action sequence runs one at a time. Irreducibly simultaneous physical action
has no representation, and the being cannot express a two-hand coordinated release.

The toy permits this. Its pass condition allows the system to *"remain unresolved,"* and none of
the four forbidden moves is taken: nothing intermediate is fabricated, no last act is credited
alone, one intervention is not recorded as two executions, and no later `perceive` is claimed to
have caused the door.

**Stated explicitly so it is a choice rather than an accident.** The being cannot perform
irreducibly joint action. If that limitation ever needs lifting, the mechanism is already
sketched in [kernel-alteration](kernel-alteration.md) §2 — a grouping Space privately carrying
*fan-out and join behavior* — and the cost is that concurrency inside one intervention reopens
attribution.

### RUN AGAINST 2026-08-08 — A PASSES MORE CLEANLY · B MAY BE REPRESENTABLE

**Run A** needs no ordering policy at all. A body is a sequence of emissions, so sequence *is*
emission order. Both *siblings concurrent, chains sequential* and the parent-declared policy it
required become unnecessary — the property survives and the field that carried it is gone.

**Run B** was declared unrepresentable because ordering was sequential. Under that model a
possibility's **body** may perform the joint intervention atomically, reached with one select.
Against the four forbidden moves: nothing intermediate is fabricated, no last act is credited alone,
one intervention is one execution, and no `perceive` is claimed as cause. All four hold.

The cost is that the two releases are invisible inside the body, so nothing can be learned about
them separately — arguably correct, since an irreducible intervention has no separable parts.

---

## Toy 5 — Dynamics closes the door on contradiction

### What it tests

Dynamics uses trust and convergence to narrow attention and lower exploration. That controller
can create a self-sealing loop if narrowing prevents contradictory evidence from entering the
observation worker's frame.

### Setup

An expectation has accumulated apparently healthy grounding:

```text
E       "the route remains safe"
T(E)    high
P(E)    high
coherence high
ΔP      shrinking
```

Dynamics correctly reads this as convergence:

```text
narrow attention
reduce context depth
lower exploration
commit
```

The environment then changes. A raw observation arrives:

```text
o*   "bridge inspection: structural fracture"
```

Activation considers `o*` only weakly related to the vocabulary around `E`. Under the reduced
frame budget, `E` and `o*` are never presented together. The observation worker scores what is in
frame and creates nothing, so it assigns no negative support.

### Runaway

```text
high trust + small ΔP
        ↓
narrower frame
        ↓
contradiction receives no edge
        ↓
P stays flat · coherence stays high
        ↓
controller reads stronger convergence
        └──────────────────────────────▶ narrower frame
```

Every local decision follows the current documents. Together they make the belief harder to
dislodge precisely because it was previously trusted.

Logging the controller makes the failure auditable after the fact, but does not prevent it.

### Failure exposed

Attention is allowed to select relevance, but if it can reduce contradiction ingress to zero,
Trust becomes self-protecting. A controller that changes what can be observed indirectly changes
what can become true for the system, even though it never writes an edge itself.

### Pass condition

Contradictory raw evidence must retain a path into evaluation that cannot be eliminated by prior
trust in the thing it contradicts. When `o*` arrives, the run must eventually:

1. place it in a frame where it can bear on `E`, or expose it as unresolved novelty;
2. prevent high prior trust from suppressing it indefinitely;
3. log the control state and selection decisions that delayed or admitted it;
4. widen again when previously coherent material becomes contested.

The test should run long enough to prove eventual admission, not merely one-wave availability.

### Measurement

```text
raw observations admitted / received
unresolved novelty age
frame exclusion count by expectation and observation
minimum exploration and attention breadth reached
time-to-admit contradictory evidence
```

### STILL OPEN, and relocated — 2026-07-30

Two changes bear on this toy. Neither solves it; together they move where it fails.

**1 · Contestation is no longer available.** Under grading-at-leaves, `o*` cannot attach to
`E "the route remains safe"` at all — that expectation is grounded through its children, not
directly. So pass conditions ② and ④ as written are **unsatisfiable**. They assume `o*` can
bear on `E` and make it contested. It cannot.

The remaining path is condition ①'s second clause: expose it as unresolved novelty. `o*` routes
to the expectation worker, which creates a **rival** expectation rather than contesting the
existing one. Revision by rivalry, not by accumulation of disconfirmation.

Rewritten, the run must eventually:

```text
① route o* to the expectation worker as unresolved novelty
② create a rival expectation rather than leave o* unhomed
③ log the control state and selection decisions that delayed or admitted it
④ let the rival accumulate grounding on its own leaves, and surface the conflict
   between two expectations that cannot both close
```

**2 · The gap detector inherits the same failure.** The expectation worker's frame is
`activation HIGH + support ≈ 0` ([workers](workers.md)). Toy 5's premise is that activation
rates `o*` as only weakly related to `E`'s vocabulary — which is `activation LOW`, and that
classifies as **noise**, not as a gap.

So the new machinery does not rescue this toy. It relocates the failure:

```text
BEFORE   dynamics narrows the frame → o* is never scored → no contradiction
NOW      dynamics narrows the frame → activation stays low → o* reads as NOISE
                                    → the expectation worker never sees a gap
```

The self-sealing loop now runs through the gap detector as well as the grader. It is arguably
**worse**, because a designed novelty channel that is gated by the same quantity dynamics
suppresses gives the appearance of an ingress guarantee without providing one.

### Open design question

What guarantees contradiction ingress: a nonzero exploration floor, an audit lane, novelty aging,
periodic breadth reset, or another bounded controller property? The requirement is architectural;
the mechanism is not selected here.

**Sharpened:** whatever the mechanism is, it cannot key on activation, because activation is
what the controller suppresses. It needs an input the controller does not modulate.

### PARTIAL ANSWER — confidence

Splitting the observation worker's output into `relationship ∈ [-1,1]` and
`confidence ∈ [0,1]` ([workers](workers.md)) supplies one such input.

`o*` scored against whatever is in frame returns `relationship ≈ 0` **and** `confidence low`
across the board — "I read this and could not place it." That pattern is a property of the
reading, not of `o*`'s relationship to `E`, so narrowing the frame does not erase it.

It also weakens what the runtime must guarantee, from impossible to cheap:

```text
BEFORE   guarantee o* is paired with the expectation it bears on
NOW      guarantee o* is scored against SOME frame. any frame.
```

**What it does not fix.** Confidence is self-reported, so it catches *known unknowns* and is
blind to *confident misreadings* — which is [Toy 1](#toy-1--the-blind-interpreter-converges-on-the-wrong-reading)
exactly. An interpreter that reads `connected` as charging ten times without hesitating reports
high confidence throughout.

So the ingress guarantee still needs a second channel that is not self-reported. **Recurrence**
remains the candidate: an observation that keeps arriving and keeps failing to find a home is
measurable without anyone judging it relevant and without the reader judging itself.

That channel is blocked on [Toy 2](#toy-2--ten-correlated-records-masquerade-as-ten-grounds).
Recurrence cannot be used for admission until *"the same observation arriving ten times"* can be
told from *"ten copies of one event"* — otherwise a single fan-out source can force admission
indefinitely.

### RUN AGAINST 2026-08-08 — LOCALIZED, NOT PREVENTED

Passes under this file's **second** clause — *represented honestly as unresolved*. It should not be
read as handled.

**What changed.** Contradiction no longer has to reach the thing it contradicts. An observation
lands in a frame and is placed or not, and *cannot place* is a purely **local** fact: it licenses
discovery, whose rejection licenses creation. Structure is created where the observation landed, and
running that structure builds the edges that make future co-presence possible. The old failure
assumed a matching problem across the graph. There isn't one.

**What that model adds.** The **absence** case — a glob stops matching, nothing is edited, no event
fires — is unreachable by any ingress mechanism, because there is nothing to pipe. The edge carrying
its cited returns is what makes `480 passed` comparable to `500 passed` at all.

**What remains is one line.** Three separate failures collapse into a single precondition:

> An observation in frame is bound, cited, or routed to discovery. **Never silently dropped.**

Stated; nothing enforces it. Competition needs a competitor, competitors come from placement failure
or from ingress, and both run through that rule.

### And one limit is structural rather than open

A possibility at the **terminus of accepted chains** has nothing above it that can reject and nothing
that fails to place. It is unfalsifiable from inside.

```text
run-tests invoked directly · returns "480 passed" · cited · accepted by a human reading the
same output. short-range positive. long-range positive. both wrong, and wrong together
because they read one instrument.
```

Correction there is **external and discrete** — a later event from the world — never gradual
starvation. Use does not kill a bad implementation; use is the positive signal, so use protects it.

This is Toy 1's *blindness ≠ independence* arriving at the credit layer: two channels that cannot
influence each other can still be wrong together if they read the same instrument.

---

## What the five toys jointly require

```text
                                                        2026-07-27 → 08-06     2026-08-08
                                                        (LIVE)                 (WHITEBOARD)

Toy 1   stability ≠ correctness                         OPEN                   PASSES
Toy 2   record count ≠ independent grounding            OPEN                   DISSOLVES ·
                                                                               dormant at ingress
Toy 3   cardinality ≠ physical causation                CLOSED 07-30           CLOSED, ④ better
Toy 4   decomposition must not invent state             A CLOSED · B OUT       A CLEANER ·
                                                        OF SCOPE               B REPRESENTABLE
Toy 5   attention must not become a gate that seals     OPEN, sharpened —      LOCALIZED, not
                                                        the gap detector       prevented — one
                                                        inherits the failure   precondition
```

**What the 08-08 column costs to claim.** Toys 1 and 2 move because of two decisions — *the edge
carries the return it cited* and *credit counts invocations, not records* — and both are recent and
unbuilt. Toy 5's entry is the second clause of the pass rule, not the first.

These are not five unrelated edge cases. Each asks whether a useful compression has erased the
distinction it was meant to measure:

```text
blindness       removes intent leakage       but not shared error
distinctness    removes path duplication     but not source correlation
one-to-one      removes ambiguous credit     but not real causal multiplicity
decomposition   removes implicit plans       but not irreducible conjunction
dynamics        removes fixed control        but not feedback capture
```

The stripped design should admit a mechanism only when one of these runs proves the loop cannot
represent the distinction without it.
