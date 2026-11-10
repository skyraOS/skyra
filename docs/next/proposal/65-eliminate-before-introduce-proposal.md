# 65 — Amendment to 64: no assertion before an elimination

Date: 2026-11-10

Status: **draft**. Amends [64](64-dependence-and-trial-proposal.md);
everything in 64 it does not name stands. It takes 64's two primitives as
the whole of what the model writes, retires every clause around them, and
adds one requirement: **the model eliminates before it introduces.**
Nothing here is built. The syntax shown is provisional.

## Why: what has worked and what has not

**Eliminations and introductions have worked.** Across the live runs the
model cut what it needed (`$1[2]`, `$4.stdout[1]`) instead of copying it,
the copy check pushed every path and name into a cut, and the literals
left over were flags and filters. 59's seven asks at one frame agreed on
the instruction every time. What the model reads and what it constructs
have been stable.

**Clauses have not.** Every wrong answer so far came from a clause, a
statement *about* when something holds:

| Clause | Where | What it did |
|---|---|---|
| a guard's condition | 58, 61 | `size($4) > 0 ? "return 0 $4"`: true of nearly anything; *count the lines* reused *show the contents* |
| a guard's condition | 59 | seven guards for one frame split three ways on values never seen |
| a guard's condition, prompted | the spec runs | `$1.contains("ERROR")` worked for the specs tried, matches *every line but ERROR*, and guarded the `ls` step too |
| a class | 61 | `cat` and `wc` declared alike; `wc`'s counts came back as contents, exit 0 |
| a citation | 62, 63 | the model writes `Γ`; 64 shows a `Γ` without `not` asserts that `not` does not matter |

64 already removes the condition and the citation. This proposal says
why, in one rule: the model's only assertion is an introduction, and
everything the system decides about when it applies comes from
eliminations, which the system computes.

## The rule

**An introduction is the model's assertion**: a literal, an exit status,
or the operator (61, Variables: every operator is an introduction).
`return 0` asserts the goal is met; `'-a'` asserts the flag is wanted;
`$2 …` asserts this tool is the one to run. The system cannot check any
of them.

**No assertion before an elimination.** In a step, the first
introduction has an elimination before it. A step that opens with an
introduction is refused, `untried`, and the model is asked again:

```text
$1: "list all files in this directory:\n/tmp"      goal
$2: leaf ls                                         tool

eliminate $3 := $1[2]                               tried, and subtracts
eliminate $4 := $1[1]                               tried, and subtracts
introduce $5 := '-a'
$2 $5 $3
```

- **Every step eliminates.** The operator is an introduction and every
  step has one, so no step is accepted without an elimination.
- **Before means before.** A cut written inside the operator, `$2 $1[2]`,
  is an elimination made in the same breath as the assertion. It counts
  as an elimination, not as one before it. The requirement is order:
  read, then assert.
- **Once is enough.** Variables are append-only, so an elimination before
  the first introduction is before every later one. Whether each
  introduction needs one of its own is open question 1.

**An elimination need not subtract.** What is required is that the model
tried:

```text
eliminate $3 := $1                                  tried, subtracts nothing
return 0 $3
```

- **A whole elimination is a cut with no selector.** It reads its source
  whole, it is typed as its source is, and it is logged like any other.
- **It is safe.** Under 64 a cut splits its source's dependency in two,
  the part and the rest. A whole cut splits nothing, so the step depends
  on all of `$1`, equal, and is reused only where `$1` reads what it read.
  Not subtracting costs reuse and never correctness.
- **Subtracting is how a step earns reuse.** Only a part cut out can go
  among or free. A model that never subtracts is asked at every new
  value, and the asks are the signal (64, The signal) that it should.

**An elimination need not be used.** One that no introduction reads still
splits its source, so it is still where a value can come apart. It
records where the model looked (62's attention, recorded), with no claim
about what that look decided.

## What it is not

- **Not a citation.** An introduction names nothing. It does not say what
  it was derived from, so there is no `Γ` to get wrong. Which eliminations
  matter is found by sightings and trials (64, Dependence).
- **Not a condition.** An elimination selects nothing by itself. It votes
  only through the dependencies of the units after it.
- **Not a check of meaning.** The system checks that an elimination came
  first and that it reads something seen. It cannot check that the model
  looked at the right thing. A step that eliminates `$2` and asserts
  `-a` for *list all files* is accepted, and it depends on all of `$1`.

## Built first: passed parts free, the rest equal

64's dependencies need trials and `span` to move. Before that is built,
one fixed setting of them is enough to replace the guard:

- **A part passed into the operator is free.** `$3` above, the directory,
  is free, as 64's open question 9 (`passed` at 1) keeps it.
- **Everything else is equal.** The rest of each cut source, each
  elimination not passed, each whole `$n` seen, each tool.
- **A stored step votes where all of that holds.** This is the guard,
  derived from the step's eliminations, with nothing the model wrote
  about when.

On the spec runs this answers what the prompt fix answered with
keywords. Spec A cuts its directory from `$1[4]` and passes it. The rest
of `$1`, the spec's text, stays equal. Spec B's text differs, so A's step
does not vote and B startles. Spec A on another directory votes.

## What changes

| 64 | 65 |
|---|---|
| a step: eliminations, introductions, one operator | the same, and the first introduction has an elimination before it |
| refusals concerning a template: `fits`, `copies`, `join`, `unstored`, `unread` | and `untried` |
| a cut is a selector applied | and the empty selector, a whole elimination |
| `passed` at 1, an open question | built first, with every other dependency equal |

In the code: `drive`'s prompts ask for a step, not a CEL chain, and say
to eliminate first; `Machine.wrote` parses a step and refuses `untried`
before any template check; `recorded` keeps the step's own eliminations
in order, beside the cuts inlined from the operator; `tree.Cast` reads
the derived dependencies above in place of `Link.Cond`.

## Required Lean definitions and proofs

In `invariants/Invariants/Model.lean` and `Machine.lean`, before any Go:

- **A committed step tried:** every step `wrote` commits has an
  elimination at a lower `$n` than its first introduction.
- **`untried` is the only new refusal:** a step that opens with an
  introduction is refused `untried`, and nothing is appended.
- **A whole elimination reads its source:** `cut` with no selector is the
  identity, and its interface is its source's (`cut_sound`, extended).
- **A whole elimination splits nothing:** a step whose eliminations are
  all whole votes only at values equal to those it was authored at.
- **Passed parts are free, the rest equal:** two frames that agree on
  everything a step does not pass decide alike under it, whatever they
  pass.
- **Memory is still a fold of the log:** `step_replays` holds over steps.

## Open

1. **Per introduction.** Whether each introduction needs an elimination
   of its own between it and the last, so `-a` and the operator each
   follow a look. It is stricter and asks for eliminations nothing reads.
2. **What `return 0` must eliminate.** `exit(0)` asserts the goal is met.
   Whether it needs an elimination of the goal, `$1` or part of it, and
   not only of what arrived.
3. **Trying in the prompt.** Whether the model is told that a whole
   elimination is allowed, or only that it must eliminate. Told, it may
   always write `$1` whole and never subtract; untold, it may invent a
   cut to satisfy the rule.
