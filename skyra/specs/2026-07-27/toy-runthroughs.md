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

### Open design question

Is the one-to-one rule about an **authored operator**, an **execution occurrence**, an
**action-to-change prediction edge**, or physical causation itself? The type boundary must be
stated before the cardinality claim can be tested.

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

### Open design question

Is an action allowed to be a compound operator whose execution contains coordinated physical
steps? If so, what makes that operator one action rather than a hidden plan, and how does its
support prediction retain attribution to the constituent failures?

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

### Open design question

What guarantees contradiction ingress: a nonzero exploration floor, an audit lane, novelty aging,
periodic breadth reset, or another bounded controller property? The requirement is architectural;
the mechanism is not selected here.

---

## What the five toys jointly require

```text
Toy 1   stability must not impersonate correctness
Toy 2   record count must not impersonate independent grounding
Toy 3   cardinality must not impersonate physical causation
Toy 4   decomposition must not invent intermediate state
Toy 5   attention must not become an epistemic gate that seals itself
```

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
