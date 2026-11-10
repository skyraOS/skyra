# 65 — Amendment to 64: every introduction is from an elimination

Date: 2026-11-10

Status: **draft**. Amends [64](64-dependence-and-trial-proposal.md);
everything in 64 it does not name stands. It takes 64's two primitives as
the whole of what is written, retires every clause around them (CEL among
them), and fixes the terms. **Every introduction is from exactly one
elimination, and every elimination reaches an introduction.** Nothing
here is built. The syntax shown is provisional.

## Terms

These are the foundations everything else is written in, so they are
natural deduction's, and used only as defined here.

| Term | Means | In natural deduction |
|---|---|---|
| **introduction** | a value asserted: by the person (the request), the model (a literal, an exit status, a head) or a tool (its return) | an introduction rule's conclusion |
| **elimination** | a value read out of another: a field, a run of lines, a run of words, or the whole | an elimination rule's conclusion |
| **E eliminates V** | E reads from V. One value is eliminated by many. | V is E's major premise |
| **I is introduced from E** | E is the reason for I. Each introduction is from exactly one elimination; each elimination is the premise of at most one. | E is I's premise |
| **premise** | the elimination an introduction is from | the same |

- **The methods are `introduce` and `eliminate`**, in Lean and in Go:
  `introduce` makes an introduction from a premise, `eliminate` reads one
  value out of another. Nothing else makes a value in a frame.
- **An assertion** is an introduction, said plainly: what the system
  cannot check. It is not a separate term.
- **"Cut" is retired.** In proof theory a cut is an introduction met at
  once by an elimination of it, which here is a tree edge: a parent
  introduces, a child eliminates. The code's `Cut`, `CutAll` and
  `tree/cut.go` are renamed `eliminate` as they are touched.
- **"Decision" keeps 61's meaning**: what the tally gives at a frame. A
  step is what the model writes; the decision is which step runs.

## Why: what has worked and what has not

**Eliminations and introductions have worked.** Across the live runs the
model eliminated what it needed (`$1[2]`, `$4.stdout[1]`) instead of
copying it, the copy check turned every path and name into an
elimination, and the literals left over were flags and filters. 59's
seven asks at one frame agreed on the instruction every time. What the
model reads and what it asserts have been stable.

**Clauses have not.** Every wrong answer so far came from a clause, a
statement *about* when something holds:

| Clause | Where | What it did |
|---|---|---|
| a guard's condition | 58, 61 | `size($4) > 0 ? "return 0 $4"`: true of nearly anything; *count the lines* reused *show the contents* |
| a guard's condition | 59 | seven guards for one frame split three ways on values never seen |
| a guard's condition, prompted | the spec runs | `$1.contains("ERROR")` worked for the specs tried, matches *every line but ERROR*, and guarded the `ls` step too |
| a class | 61 | `cat` and `wc` declared alike; `wc`'s counts came back as contents, exit 0 |
| a citation | 62, 63 | the model writes `Γ`; 64 shows a `Γ` without `not` asserts that `not` does not matter |

64 already removes the condition and the citation. This proposal removes
the language they were written in: when a step applies is found by
reading its eliminations again, as text.

**An introduction alone means nothing.** A model, a person or a program
that prints `yes` has said nothing until something around it is
eliminated: yes to what. `exit(0)` is the same `yes`. It means the goal
it answers, and only if the goal was eliminated to answer it.

## The rule

### Every introduction is from exactly one elimination

```text
$1: request("list all files in this directory:\n/tmp", tools: [ls])

eliminate $2 := $1.goal
eliminate $3 := $1.tools[1]                     ls
eliminate $4 := $2[2]                           the directory
eliminate $5 := $2 / "list"
eliminate $6 := $2 / "all files"
introduce $7 := '-a' from $6
$3 $7 $4 from $5                                ls -a /tmp
```

- **Exactly one.** An introduction from no elimination, or from two, is
  refused: a step is a list of single derivations, each readable alone.
- **Its own.** Two introductions are never from one elimination. `ls` and
  `-a` above are each from their own reading of the goal.
- **Before it.** Variables are append-only, and an introduction's premise
  is at a lower `$n`.
- **Refused as `untried`**: an introduction from no elimination, from
  more than one, from one another introduction is from, or from one
  after it. The model is asked again, told which.

### Every elimination reaches an introduction

**An elimination is written for an introduction, or it is incoherent.**
Each elimination is a premise, or is placed in an introduction as a value
(`$4`, the directory, above), or is eliminated by one that is. `$2` above
is eliminated by `$4`, `$5` and `$6`, each of which reaches an
introduction.

- **Refused as `unused`**: an elimination that reaches no introduction.
  Reading a value and doing nothing with it is a look that leads nowhere;
  the step it was in is not a derivation of anything.
- **Refused as `repeated`**: an elimination alike another, the same
  source and selector. Two alike are one reading.
- **So the step is a derivation.** Every value in it leads from what
  arrived to what is asserted, with nothing left over.

### A step is a sentence

**A step's operator is a head and at least one argument: two
introductions or more.** It is a subject and a predicate, and no complete
sentence lacks either. An exec is the same: a binary and its arguments.

| Operator | Head, the predicate | Arguments, the subject |
|---|---|---|
| act | the tool run | what it is run on |
| return | the exit status, `0` or `1` | what is returned, and why |
| delegate | the delegation | the goal passed, and the tools |

- **The operator is counted by its parts.** 61 counts an operator as one
  introduction. Here its head is one, and each argument is one more: the
  assertion that this value goes in this place. 64's agreement already
  counts this way: the same operator on the same position.
- **The head's premise is written after the operator**: `from $5`, the
  reading of `list` that `ls` is introduced from.
- **An argument is an introduction too.** A literal argument is
  introduced on its own line from its premise. An eliminated argument is
  its own premise: passing what was read costs nothing more.
- **So the floor is cheap where the model reads**, and costs only where
  it writes text of its own.
- **Refused as `fragment`**: a head with no argument. It is a predicate
  whose subject is implicit. `ls` alone reads the working directory, which
  the frame never saw, and `return 0` alone is a `yes` to nothing.

**Nothing is introduced, never left out.** A model that holds that a head
needs no subject says so: it introduces `null` from an elimination, and
the system runs it as nothing.

```text
$1: request("what time is it?", tools: [date])

eliminate $2 := $1.goal
eliminate $3 := $1.tools[1]                     date
eliminate $4 := $2 / "what time"
eliminate $5 := $2 / "it"                       nothing here is an argument
introduce $6 := null from $5
$3 $6 from $4                                   date
```

- **`null` is an introduction**, so it counts toward the two, and it is
  from its own elimination: the reading that found nothing to pass.
- **So the model claims null against an elimination.** Leaving an
  argument out was a choice the log never showed. `null from $5` is a
  claim of absence at a place, and a trial that varies `$5` and gets an
  argument refutes it.
- **The system sees nothing.** An act passes no argument for it; a
  return returns nothing for it (`rm`'s `return 0 null`: done, nothing to
  show). It is not the string `'null'`, and it copies no text.
- **`fragment` is only an omission.** A head with no argument is refused;
  a head with `null` is a claim.

### Two to the parent, one to the child

**A step is two introductions or more where it is written, and one where
it arrives.** The parent writes a head and its arguments. They compose
into one value of one composite interface (61, Interface and
observation: every operator makes one), and that is all the child
receives.

```text
parent    $1: request("count the lines of this file:\n/tmp/x", tools: [cat, wc])
          eliminate $2 := $1.goal
          eliminate $3 := $1.tools[2]               wc
          eliminate $4 := $2 / "count"
          delegate($2, $3) from $4                  three introductions

child     $1: delegation(goal: string, tools: [wc])  one introduction
          eliminate $2 := $1.goal
          eliminate $3 := $1.tools[1]
          eliminate $4 := $2[2]                     the file
          eliminate $5 := $2 / "lines"
          eliminate $6 := $2 / "count"
          introduce $7 := '-l' from $5
          $3 $7 $4 from $6                          wc -l /tmp/x
```

- **The request is the person's introduction.** The root frame starts
  with one value, as every frame does: the goal the person asserts and
  the tools they provisioned. The person is the root's parent, and takes
  or rejects what it returns (64: credit is anchored there).
- **A frame starts with one value, nothing else**: no goal of the
  parent's, no tool the parent kept. This is 61's narrowing, made the
  shape of the frame.
- **So a frame must eliminate before it can introduce anything.** Its
  only value is an introduction it did not make, and every premise is an
  elimination of it, or of what arrives later.
- **A return is the same, upward.** The child's `return` is its head and
  arguments; the parent receives one arrival, `{exit, stdout, stderr}`,
  and eliminates it before introducing anything from it. A leaf is a
  child: `ls` receives one argv and gives one return.
- **This forces composition.** Every edge of the tree carries one
  composite value. Each frame eliminates what arrived, introduces from
  the parts, and composes two or more introductions into what it sends.
  A frame cannot pass a value on untouched.
- **Every edge is a redex**: an introduction, the parent's, met by an
  elimination, the child's. A product is introduced by its parts and
  eliminated by projection; the child's first eliminations are the
  projections.

```text
$1: request("is /tmp/x a directory? answer yes or no:\n/tmp/x", tools: [test])
…
$6: {exit: 0, stdout: "yes\n", stderr: ""}       arrived
eliminate $7 := $2[1]                            the question
eliminate $8 := $6.stdout
return 0 $8 from $7
```

### An elimination need not subtract

What is required is that the model read before it asserted. Reading a
value whole is a reading:

```text
eliminate $2 := $1.goal
eliminate $9 := $2                               whole, subtracts nothing
…
return 0 $8 from $9
```

- **A whole elimination has no selector.** It reads its source whole, it
  is typed as its source is, and it is logged like any other.
- **It is safe.** A whole elimination leaves nothing out of what the
  step depends on (Memory, below). Not subtracting costs reuse, never
  correctness.
- **One for one makes it subtract.** Two eliminations alike are refused,
  so a source can be read whole once. A step with more introductions than
  sources must take parts, and each part is a place a value can come
  apart.
- **It needs runs of words.** `ls` and `-a` both come from the goal's
  first line. With only lines and fields (61) they cannot each have an
  elimination. 62's runs of words (Lexical units), matched by text, are
  built with this.

## Memory: introductions are keys, eliminations are values

**Memory is set up as it is.** 61's key is what a frame has received,
and its value what was learned there. In these terms:

- **A key is the type of the introductions a frame has received**: its
  start, then each arrival. A type, never a value: a request's text is
  not in its key (61, The rule), or no two goals would share anything.
- **A value is the steps written there**, each with its eliminations and
  the text each read when the step was written.
- **A step votes where its eliminations read the same text again**,
  except those placed as arguments. That is the whole of when a step
  applies. Two strings compared, with no expression around them.
- **The request's identity is checked in the value.** A step depends on
  the parts of the request it did not pass, so a request asking
  something else reads differently there, and the step does not vote.
  Identity is held where it is checked, not where it is looked up.
- **`from` decides nothing.** Which steps vote is a function of what
  their eliminations read, never of which introduction each is the
  premise of. `from` is what makes a step well formed and what a trial
  varies first; it is not a condition.

### Built first: arguments free, the rest equal

64's dependencies need trials and `span` to move. Before those are built,
one fixed setting of them replaces the guard:

- **What is placed as an argument is free.** `$4` above, the directory,
  is free, as 64's open question 9 (`passed` at 1) keeps it.
- **Everything else is equal.** The rest of each value an elimination
  reads from, each elimination not placed as an argument, each tool.
- **A stored step votes where all of that holds.** This is the guard,
  derived from what the step eliminated, with nothing written about when.

On the spec runs this answers what the prompt fix answered with
keywords. Spec A eliminates its directory from line 4 and passes it. The
rest of the goal, the spec's text, stays equal. Spec B's text differs, so
A's step does not vote and B startles. Spec A on another directory votes.

**Each condition is a set of strings at a place.** *Equal* is a set of
one, 64's *among* is a larger set, and *free* is no condition. So 64's
trials later change how the sets grow, not how a vote is counted.

## What it is not

- **Not a citation.** 62's citation was a set the model wrote, compiled
  into the guard, so it decided when the link selected. A premise is one
  elimination, and it decides nothing: what matters is found by reading
  again, and later by trials (64, Dependence).
- **Not a condition.** An elimination selects nothing by itself. A step
  votes only where its eliminations read what they read before.
- **Not a check of meaning.** The system checks that every introduction
  is from its own elimination and that each reads something seen. It
  cannot check that the model read the right thing. `-a` from `$4`, the
  directory, is accepted, and that premise is the first a trial varies.

## What changes

| 58 to 64 | 65 |
|---|---|
| 58: the model writes a CEL chain | a step: eliminations, introductions, one operator; no CEL |
| 64: a predicate is an elimination the system computes | none for now: an elimination reads text (Open, 3) |
| a step's introductions, unrelated to its eliminations | each from exactly one, before it; each elimination reaching one |
| 61: an operator is one introduction | its head and each argument: a step is two or more |
| an argument left out | `null`, introduced from an elimination: the system passes nothing |
| 61: a frame starts with its scope, `$1 … $n` | with one value, its parent's introduction; the root's is the person's request |
| refusals concerning a template: `fits`, `copies`, `join`, `unstored`, `unread` | and `untried`, `unused`, `repeated`, `fragment`; those concerning a condition go |
| a cut is a selector applied | `eliminate`, and with no selector, a whole elimination |
| 63, section 1: quorum counted on the ballot | built: needed once steps vote only at the text they read |
| `passed` at 1, an open question | built first, with every other dependency equal |

In the code: `introduce` and `eliminate` are the methods in `model`;
`drive`'s prompts ask for a step and say every introduction is from its
own elimination; `Machine.wrote` parses a step and refuses `untried`,
`unused`, `repeated` and `fragment` before any template check; `Sel`
gains a run of words matched by text; a branch's condition is derived
from its step; `tree`'s `StoreConds`, `CheckGuard`, `CommitChain`,
`Written` and CEL go, and `-show` prints steps.

## Required Lean definitions and proofs

In `invariants/Invariants/Refine.lean`, `Model.lean` and `Machine.lean`,
before any Go. 59's tally stands, with its theorems: `tally_perm`,
`tally_unanimous`, `tally_null`, `wins_unique`, `tally_decides_mono`.

- **`introduce` and `eliminate`** are the only definitions that add a
  value to a frame.
- **A committed step is from its eliminations:** its introductions map
  injectively to eliminations, each at a lower `$n`.
- **A committed step is coherent:** every elimination in it reaches an
  introduction, and no two are alike.
- **A committed step is a sentence:** its operator has a head and at
  least one argument.
- **One to the child:** a child frame starts with exactly one value, the
  one its parent's operator made; every value a frame holds is that, an
  arrival, or made by `introduce` or `eliminate` in the frame.
- **`null` is nothing:** an act or return with a `null` argument runs as
  one without it.
- **The refusals are the four:** a step that is not so formed is refused
  `untried`, `unused`, `repeated` or `fragment`, and nothing is appended.
- **A whole elimination reads its source:** `eliminate` with no selector
  is the identity, and its interface is its source's.
- **A condition is derived** (`derive`): a conjunction of (place, set of
  strings), one for each place the step read and did not place as an
  argument. The chain's definitions and theorems go.
- **A step votes where it was written** (`derive_self`, replacing
  `commit_witness`): its derived condition holds at the values it was
  written at, so with no other voter its commit selects it, and
  `afterStartle`'s second ask and `unselected` go.
- **Arguments are free, the rest equal** (`passed_free`): two frames that
  agree on every place a step read and did not place as an argument
  decide alike under it.
- **A request never seen gets no vote:** a frame whose start differs from
  a step's at a place it read and did not pass is not voted for by it.
- **`from` decides nothing** (`premise_irrelevant`): two steps alike but
  for their premises vote alike at every frame.
- **Quorum counts votes cast on the values**, not steps at the key.
- **Memory is still a fold of the log:** `step_replays` holds over steps.

## Open

1. **A premise the sightings contradict.** What is done when an
   introduction is unchanged whatever its premise reads: shown to the
   person, counted against the model, or moved to the dependency that
   stayed equal.
2. **What `return 0` must be from.** Whether `exit(0)` must be introduced
   from an elimination of the goal, as the `yes` example's is, and not of
   what arrived.
3. **Predicates.** A threshold (`restarts > 2`) is not text read again.
   Nothing in `ls`, `cat`, `wc`, `date` or `rm` needs one. Whether one
   returns as a kind of elimination, with no expression around it.
4. **The whole elimination in the prompt.** Whether the model is told it
   may read a value whole. Told, it may read each source whole and stop;
   untold, it may invent a part to satisfy the rule.
