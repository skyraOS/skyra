# 65 — Amendment to 64: every assertion has its own elimination

Date: 2026-11-10

Status: **draft**. Amends [64](64-dependence-and-trial-proposal.md);
everything in 64 it does not name stands. It takes 64's two primitives as
the whole of what the model writes, retires every clause around them, and
adds one requirement: **each introduction is paired with exactly one
elimination of its own, written before it.** Nothing here is built. The
syntax shown is provisional.

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

**An assertion alone means nothing.** A model, a person or a program that
prints `yes` has said nothing until something around it is eliminated:
yes to what. `exit(0)` is the same `yes`. Its meaning is the goal it
answers, and a return that eliminated none of the goal has asserted it
was met without looking at what was asked.

## The rule

**An introduction is the model's assertion**: a literal, an exit status,
or the operator (61, Variables: every operator is an introduction).
`return 0` asserts the goal is met; `'-a'` asserts the flag is wanted;
`$2 …` asserts this tool is the one to run. The system cannot check any
of them.

**One for one.** Each introduction names exactly one elimination, its
own, written before it:

```text
$1: "list all files in this directory:\n/tmp"      goal
$2: leaf ls                                         tool

eliminate $3 := $1[2]                               the directory
eliminate $4 := $1 / "list"
eliminate $5 := $1 / "all files"
introduce $6 := '-a' from $5
$2 $6 $3 from $4
```

```text
$1: "is /tmp/x a directory? answer yes or no:\n/tmp/x"
$3: {exit: 0, stdout: "yes\n", stderr: ""}          arrived

eliminate $4 := $1[1]                               the question
return 0 $3 from $4
```

- **Exactly one.** An introduction with no elimination is refused, and so
  is one naming two: a step is a list of single derivations, each
  readable on its own.
- **Its own.** Two introductions never share an elimination. `-a` and
  `ls` above each have their own look at the goal; one look at `$1[1]`
  cannot be the reason for both.
- **Before it.** Variables are append-only, and an introduction names an
  elimination at a lower `$n`. A cut written inside the operator,
  `$2 $1[2]`, is an argument, not the operator's elimination.
- **Every step eliminates.** The operator is an introduction and every
  step has one.
- **Refused as `untried`**: an introduction with no elimination, with
  more than one, with one another introduction holds, or with one after
  it. The model is asked again, told which.

**The pairing is a record, not a clause.** It does not say the
introduction depends only on its elimination, and it is compiled into
nothing. Under 64 the unit still depends on everything its frame has
seen. The pairing is the model documenting its own assertion: which look
it made for which claim. The system uses it in two places:

- **It is read first.** A trial varies a unit's paired elimination before
  any other dependency (64, Trials), since it is where the model said to
  look.
- **It is checked against the world.** A unit whose dependency on its
  paired elimination goes free, the same assertion whatever that part
  reads, has a pairing the sightings contradict. That is counted against
  the model's documentation, and shown (Open, 2).

**An elimination need not subtract.** What is required is that the model
tried:

```text
eliminate $3 := $1                                  tried, subtracts nothing
return 0 $2 from $3
```

- **A whole elimination is a cut with no selector.** It reads its source
  whole, it is typed as its source is, and it is logged like any other.
- **It is safe.** Under 64 a cut splits its source's dependency in two,
  the part and the rest. A whole cut splits nothing, so the step depends
  on all of `$1`, equal, and is reused only where `$1` reads what it read.
  Not subtracting costs reuse and never correctness.
- **One for one makes it subtract.** Two eliminations alike are one look,
  and the second is refused. So a source can be taken whole once. A step
  with more assertions than sources must cut, and every cut is a place a
  value can come apart.
- **It needs word cuts.** `-a` and `ls` both come from the goal's first
  line. With only lines and fields (61), they cannot have an elimination
  each. 62's runs of words (Lexical units), matched by text, are built
  with this.

## What it is not

- **Not a citation.** 62's citation was a set the model wrote, compiled
  into the guard, so it decided when the link selected. A pairing is one
  elimination, and it decides nothing: what matters is still found by
  sightings and trials (64, Dependence).
- **Not a condition.** An elimination selects nothing by itself. It votes
  only through the dependencies of the units after it.
- **Not a check of meaning.** The system checks that every introduction
  has its own elimination and that each reads something seen. It cannot
  check that the model looked at the right thing. A step that pairs `-a`
  with `$1[2]`, the directory, is accepted, and the pairing is the first
  thing a trial tests.

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
| a step: eliminations, introductions, one operator | the same, each introduction naming one elimination of its own, before it |
| refusals concerning a template: `fits`, `copies`, `join`, `unstored`, `unread` | and `untried` |
| a cut is a selector applied | and the empty selector, a whole elimination; two alike are one |
| runs of words, 62's | built with this: one for one needs them |
| a trial varies a dependency | the paired elimination first |
| `passed` at 1, an open question | built first, with every other dependency equal |

In the code: `drive`'s prompts ask for a step, not a CEL chain, and say
to pair each introduction with its own elimination; `Machine.wrote`
parses a step and refuses `untried` before any template check; `Sel`
gains a word run matched by text; `recorded` keeps the step's
eliminations in order with their pairings, beside the cuts inlined from
the operator; `tree.Cast` reads the derived dependencies above in place
of `Link.Cond`.

## Required Lean definitions and proofs

In `invariants/Invariants/Model.lean` and `Machine.lean`, before any Go:

- **A committed step is paired:** every step `wrote` commits maps its
  introductions injectively to eliminations, each at a lower `$n` than
  its introduction.
- **`untried` is the only new refusal:** a step that is not so paired is
  refused `untried`, and nothing is appended.
- **Two eliminations alike are refused:** a committed step's
  eliminations are pairwise distinct as (source, selector).
- **A whole elimination reads its source:** `cut` with no selector is the
  identity, and its interface is its source's (`cut_sound`, extended).
- **A whole elimination splits nothing:** a step whose eliminations are
  all whole votes only at values equal to those it was authored at.
- **The pairing decides nothing:** two steps differing only in their
  pairings vote alike at every frame.
- **Passed parts are free, the rest equal:** two frames that agree on
  everything a step does not pass decide alike under it, whatever they
  pass.
- **Memory is still a fold of the log:** `step_replays` holds over steps.

## Open

1. **How many introductions a decision needs.** A floor would make every
   decision document more than one assertion. As 61 counts them,
   `return 0 $3` has two (the return and `exit(0)`) and `$2 $3` one (the
   operator), so a floor of three refuses both unless the model writes a
   literal it does not need, and an unneeded flag is an assertion that
   reaches the world.
2. **A pairing the sightings contradict.** What is done when a unit's
   paired elimination goes free: shown to the person, counted against the
   model, or the pairing moved to the dependency that stayed equal.
3. **What `return 0` must pair with.** Whether `exit(0)` must pair with an
   elimination of the goal, `$1` or part of it, as the `yes` example
   does, and not of what arrived.
4. **Trying in the prompt.** Whether the model is told that a whole
   elimination is allowed, or only that it must eliminate. Told, it may
   take each source whole and stop; untold, it may invent a cut to
   satisfy the rule.
