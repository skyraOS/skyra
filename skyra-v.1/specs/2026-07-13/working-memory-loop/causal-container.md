# Causal Container and Surprise-Gated Meaning

**Date:** 2026-07-13  
**Status:** ROUGH IDEA — exploratory note, not yet an implementation spec  
**Parent:** [The Working-Memory Loop](README.md)

This note explores a possible structure for causal nodes inside working memory.
The central idea is that a working-memory situation is a stable material/formal
container fed by many efficient causes and oriented by one active final cause.
Ordinary cognition may replace the final cause, but it may not rewrite the
material/formal container. A material/formal rewrite is exceptional and must be
authorized by a large positive delta in surprise.

## 1. Shape

For a given working-memory container `M`:

```text
efficient causes                 material/formal              active final

A ---------+
B ---------+----> M ----------------------------------------> D
C ---------+
     * : 1                     1 : 1                         1 : 1
```

- `A`, `B`, and `C` are efficient causes participating in the situation.
- `M` is the situation's one material/formal container and current summary.
- `D` is its one active final cause: the meaning, direction, or outcome that
  currently closes the situation.
- Efficient causes are many-to-one with respect to `M`.
- A container has one current material/formal summary.
- A container has one active final cause at a time.

The `1:1` final invariant means **one active final**, not necessarily one final
candidate for the entire lifetime of the container. Rejected candidates may be
retained as proposal history without being simultaneously asserted as final
causes.

```text
M
|-- D1  score +0.05  rejected
|-- D2  score -0.48  contradicted
`-- D3  score +0.61  active
```

## 2. Evidence belongs to relationships

A final cause is graded with respect to each efficient cause in the container.
Evidence is therefore not one global grade stored on `D`, nor a fact stored only
on `A`, `B`, `C`, or `M`. It is a set of relational assessments:

```text
grade(D | A)
grade(D | B)
grade(D | C)
```

Each assessment records how strongly that efficient cause supports or defeats
the proposed final cause. The active final's aggregate support is computed from
those assessments and the amplitudes carried by the efficient causes.

A provisional normalized score is:

```text
score(D) = sum(amplitude(Ei) * grade(D | Ei))
           -----------------------------------
                    sum(abs(amplitude(Ei)))
```

Normalization prevents a container from appearing more meaningful merely
because it accumulated more efficient causes.

Example:

```text
A  amplitude .60  grade(D | A) +.80
B  amplitude .25  grade(D | B) +.20
C  amplitude .15  grade(D | C) -.40

score(D) = .47
```

The exact formula and threshold remain open. A provisional closure threshold is
`score >= +0.35`.

Sign and magnitude must not be collapsed:

```text
score >= +0.35              supported final cause
score <= -0.35              strongly defeated final cause
-0.35 < score < +0.35       weak or unresolved
```

The magnitude may be salient, but a strongly negative grade is refutation, not
successful meaning extraction.

## 3. Ordinary meaning search keeps M fixed

Incoming evidence may append or reweight efficient causes and may change the
grades of final candidates. It does not ordinarily rewrite `M`.

```text
new efficient cause or evidence arrives
        |
        v
grade active final D against the updated efficient set
        |
        v
D remains above closure threshold? ---- yes ---> retain M and D
        |
        no
        v
propose another final candidate against the same M
        |
        v
candidate closes? ---------------------- yes ---> replace active final only
        |
        no
        v
retain the unresolved state; do not rewrite M
```

The container is always trying to leave ordinary settlement with a supported
final cause. If `D1` scores only `+0.05`, cognition may append and grade `D2`.
This search must be bounded. Exhausting the candidate or inference budget yields
an explicitly unresolved container; it does not grant permission to invent a
meaning or mutate the material/formal summary until something passes.

Possible stop conditions include:

- a maximum number of final candidates;
- no improvement greater than a small epsilon;
- inference/query budget exhaustion;
- an explicit unresolved or deferred state.

## 4. Surprise gates material/formal change

Rewriting `M` is a separate, slower operation. It occurs only when new evidence
causes an exceptional positive jump in surprise.

One provisional signal is derived from the support of the active final:

```text
support(t)       = aggregate support of the active final
surprise(t)      = 1 - support(t)
deltaSurprise(t) = surprise(t) - surprise(t-1)
```

`M` may be revised only when:

```text
deltaSurprise(t) > dynamicShockThreshold
```

This comparison is directional rather than absolute. A large decrease in
surprise means the situation became better explained and should not recruit a
material/formal rewrite.

The dynamic threshold might combine recent positive surprise deltas with an
absolute floor:

```text
dynamicShockThreshold =
    mean(recent positive deltas) + k * deviation(recent positive deltas)
```

The exact estimator, window, `k`, and floor require live calibration.

When the shock threshold is crossed:

```text
new evidence sharply defeats the current orientation
        |
        v
authorize material/formal revision
        |
        v
produce a new version of M's summary
        |
        v
regroup/regrade efficient causes under that version
        |
        v
search for one supported final cause
        |
        v
atomically commit new M version + active final
```

This yields two speeds:

| Event | Efficient causes | Active final | Material/formal summary |
|---|---|---|---|
| Ordinary evidence | append/reweight | regrade or replace | frozen |
| Weak final | unchanged or newly assessed | try another candidate | frozen |
| Surprise shock | may regroup | rebuilt | new version |

Failure to find a final cause is not, by itself, permission to change the
material/formal account of the situation.

## 5. Surprise and repression

The intended dynamic is:

- incoming material introduces unexplained amplitude and can increase surprise;
- the active final attempts to absorb or explain that material;
- inner cognition represses ordinary surprise by searching for a better final
  while holding `M` fixed;
- only an abnormal residual recruits the machinery allowed to revise `M`;
- learning occurs when a newly supported final explains material that prior
  finals could not.

The environment supplies the challenge. The system should not be instructed to
manufacture contradictions merely to maximize surprise.

## 6. Bootstrap: the observed-given edge

Every observed impulse becomes material/formal input. If an impulse has no
existing home, a new container is born around it:

```text
observed impulse A has no home
        |
        v
create M1
        |
        v
efficient set = {A}
summary = "A occurred"
grounding = A grounds its own occurrence
```

The immediate consequence of saying something is that it was said. This is a
valid bootstrap identity, but it should not masquerade as learned causal
evidence. Its relationship should carry special provenance such as `given`,
`birth`, or `identity`, rather than an ordinary graded efficient-cause edge.
Otherwise self-grounding could accumulate artificial confidence or create a
causal reinforcement loop.

## 7. Summary versioning

`M` has stable identity, but its summary should be versioned rather than silently
overwritten:

```text
M/v1  "Michael said something"
M/v2  "Michael requested movement"
M/v3  "Michael requested movement after a failed attempt"
```

Grades and active-final decisions retain the summary version against which they
were made. A shock-driven rewrite atomically advances the active version. This
preserves the meaning of historical evidence and prevents a later summary from
retroactively changing what an earlier grade assessed.

## 8. Provisional invariants

1. Many efficient causes may feed one material/formal container.
2. A container has one active summary version at a time.
3. A container has at most one active final cause at a time.
4. Final candidates are graded independently against each efficient cause.
5. Final support is normalized so node count alone cannot increase confidence.
6. Ordinary evidence may replace the final cause but may not rewrite the
   material/formal summary.
7. Only an exceptional positive surprise delta authorizes a new summary version.
8. A strong negative score is refutation, not closure.
9. Unresolved is a legitimate terminal state for a bounded search.
10. Bootstrap self-grounding has explicit non-learned provenance.

## 9. Open questions

1. Is `+0.35` the right closure threshold, and should it vary by container?
2. What exactly owns efficient-cause amplitude: the cause-to-container edge, the
   cause node's activation, or a traversal snapshot?
3. Does replacing the active final require hysteresis so two close candidates do
   not oscillate?
4. Should strongly defeated finals remain active as explicit conflict content or
   become inactive history immediately?
5. How many candidate finals may ordinary cognition try before deferring?
6. How are candidate finals generated without optimizing toward a threshold or
   hallucinating closure?
7. What surprise estimator best distinguishes a true shock from noisy grading?
8. Can efficient causes be regrouped only during an `M` rewrite, or can ordinary
   evidence retire them without changing the summary?
9. Is the final cause best understood as meaning, predicted outcome, intended
   direction, or a typed union of those roles?
10. How does a supported final become reusable learned structure in later
    containers?

