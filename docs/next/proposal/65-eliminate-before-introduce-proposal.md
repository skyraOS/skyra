# 65 — Amendment to 64: every introduction is from an elimination

Date: 2026-11-10

Status: **draft**. Amends [64](64-dependence-and-trial-proposal.md);
everything in 64 it does not name stands. It takes 64's two primitives as
the whole of what is written, retires every clause around them (CEL among
them), and fixes the terms. **A step is a substitution: each elimination
reads a part of what the frame received, and owns the one introduction
put in its place.** Nothing here is built. The syntax shown is
provisional.

## Terms

These are the foundations everything else is written in, so they are
natural deduction's, and used only as defined here.

| Term | Means | In natural deduction |
|---|---|---|
| **introduction** | a value asserted: by the person (the request), the model (a literal, an exit status, a tool chosen) or a tool (its return) | an introduction rule's conclusion |
| **elimination** | a part read out of an introduction: a run of words, a run of lines, a field, or the whole | an elimination rule's conclusion |
| **E eliminates V** | E reads a part of V. One value is eliminated by many. | V is E's major premise |
| **I is introduced from E** | I is put in the place of what E read. Within a frame, each introduction is from exactly one elimination, and each elimination owns exactly one introduction. | E is I's premise |
| **premise** | the elimination an introduction is from | the same |
| **composition** | the introductions a step puts together into the one value it sends: an act, a return, a delegation | an introduction rule with several premises |

- **The methods are `introduce` and `eliminate`**, in Lean and in Go:
  `eliminate` reads a part of a value, `introduce` puts a value in its
  place. Nothing else makes a value in a frame.
- **An address is not an elimination.** `$1.goal`, `$6.stdout[1]` say
  where a part is; they produce nothing. An elimination is an address and
  a part: `$1.goal / "move"`.
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

## A step is a substitution

**Each elimination reads a part of what the frame received, and the
introduction from it is put in that part's place.** Run, the step
substitutes: what was read is replaced by what was introduced.

```text
$1: request("Move that file from my desktop to this folder.", tools: [ls, mv])

eliminate $2 := $1.goal / "that file"
eliminate $3 := $1.goal / "my desktop"
introduce $4 := $1.tools[1] from $2            ls, to find which file
introduce $5 := '~/Desktop' from $3            in place of "my desktop"
$4 $5                                          ls ~/Desktop

$6: {exit: 0, stdout: "file.txt\n", stderr: ""}   arrived

eliminate $7 := $1.goal / "Move"
eliminate $8 := $6.stdout[1]                   file.txt
eliminate $9 := $1.goal / "this folder"
introduce $10 := $1.tools[2] from $7           mv, in place of "Move"
introduce $11 := $5'/'$8 from $8               in place of "that file"
introduce $12 := '.' from $9                   in place of "this folder"
$10 $11 $12                                    mv ~/Desktop/file.txt .
```

**An introduction's value comes from one of three places**, and which
one decides how it is reused:

| Value | Example | Next time |
|---|---|---|
| a literal | `mv` for "move", `-a` for "all files" | the same: a constant |
| the part read, in its own place | `/tmp` for "/tmp": `introduce $7 from $4` | whatever text is there: passed |
| a value held, at an address | a tool, `$1.tools[2]`; a line of what arrived, `$6.stdout[1]` | whatever the frame holds there |

**A reference is resolved by looking, never by a literal.** "That file"
is not in the request as a name. Introduced as the literal `file.txt`,
the next "move that file" would move `file.txt` whatever is on the
desktop; the copy check refuses it, rightly. It is resolved from what
arrived, as "move" is resolved to `mv`, so it takes two steps: look,
then substitute.

**Within a frame, an elimination owns one introduction. Across frames it
does not.** "Move" introduces `mv` in one frame and a moving company in
another. That is two steps, held apart by memory: at different keys, or
where the words nobody eliminated differ. Two such steps at one frame
are a dissent, and the tally settles it.

## The rule

### Syntax: one value received, n eliminations, two introductions or more

```text
1 introduction  →  n eliminations  →  ≥ 2 introductions, composed into 1
```

- **A frame receives one value**: its parent's composition, or for the
  root the person's request with the tools they provisioned. Nothing
  else: no goal of the parent's, no tool the parent kept. This is 61's
  narrowing, made the shape of the frame.
- **Every introduction is from exactly one elimination**, of this frame,
  at a lower `$n`. Two introductions are never from one elimination.
- **Every elimination owns exactly one introduction.** An elimination
  owning none is a reading that leads nowhere.
- **A step composes two introductions or more.** It is a subject and a
  predicate, and no complete sentence lacks either; an exec is a binary
  and its arguments. The composition line is made only of `$n`, each an
  introduction (61: an instruction made entirely of variables).

| Composition | Head, the predicate | Arguments, the subject |
|---|---|---|
| act | the tool run | what it is run on |
| return | the exit status, `0` or `1` | what is returned, and why |
| delegate | the goal passed | the tools |

**Refusals.** The model is asked again, told which:

| Refusal | A step that |
|---|---|
| `untried` | introduces from no elimination, from two, from one another introduction is from, or from one after it |
| `unused` | eliminates something and introduces nothing from it |
| `repeated` | eliminates the same part at the same address twice |
| `fragment` | composes fewer than two introductions |

**Why the rule is one to n to two.** The system must be able to isolate.
Each introduction is from one elimination and each elimination owns one,
so a trial that varies one elimination can change at most one
introduction, and what changed is traced to what was varied. A premise
shared, or an introduction from two, and that trace is lost.

### Data: n introductions, n eliminations, n introductions

**The one value a frame receives is a hypothesis.** It is the parent's
two introductions or more, composed: the parent's claim that these
belong together as one request. The child sees one value; the log keeps
the parts, each with the elimination it is from.

- **The data is recorded as it presents itself**: a graph of
  introductions and eliminations, with edges *eliminates*, *from* and
  *composed of*. Across frames and steps it is many to many.
- **The syntax is a projection of it.** A frame's start, its key, and
  the rule above are read from the graph, never stored beside it.
- **The child is where hypotheses are tested.** Every part of the
  child's one value traces to exactly one of the parent's eliminations,
  so a trial at the child varies one part and tests one of the parent's
  readings, with everything else held.
- **The one is corroborated by the child's return.** A return that holds
  corroborates the composition. One refuted puts the composition in
  question: asked again, a child that writes the same step was faithful,
  and the error is the parent's (64, Whose error).

### Nothing is introduced, never left out

A model that holds that a head needs no subject says so: it introduces
`null` from an elimination, and the system runs it as nothing.

```text
$1: request("what time is it?", tools: [date])

eliminate $2 := $1.goal / "what time"
eliminate $3 := $1.goal / "it"                 nothing here is an argument
introduce $4 := $1.tools[1] from $2            date
introduce $5 := null from $3
$4 $5                                          date
```

- **`null` is an introduction**, so it counts toward the two, and it is
  from its own elimination: the reading that found nothing to pass.
- **So the model claims null against an elimination.** Leaving an
  argument out was a choice the log never showed. `null from $3` is a
  claim of absence at a place, and a trial that varies `$3` and gets an
  argument refutes it.
- **The system sees nothing.** An act passes no argument for it; a
  return returns nothing for it (`rm`'s `return 0 null`: done, nothing to
  show). It is not the string `'null'`, and it copies no text.
- **`fragment` is only an omission.** A head alone is refused; a head
  with `null` is a claim.

### An elimination need not subtract

What is required is that the model read before it asserted. Reading a
value whole is a reading:

```text
$6: {exit: 0, stdout: "yes\n", stderr: ""}     arrived

eliminate $7 := $1.goal[1]                     the question
eliminate $8 := $6.stdout                      whole
introduce $9 := exit(0) from $7
introduce $10 from $8
return $9 $10
```

- **A whole elimination has no part selected.** It reads its address
  whole, is typed as what is there, and is logged like any other.
- **It is safe.** Nothing is left out of what the step depends on
  (Memory, below). Not subtracting costs reuse, never correctness.
- **One to n makes it subtract.** An elimination owns one introduction,
  and two alike are refused, so an address can be read whole once. A
  step with more introductions than addresses must read parts.
- **It needs runs of words.** "Move" and "this folder" are on one line.
  With only lines and fields (61) they cannot each be eliminated. 62's
  runs of words (Lexical units), matched by text, are built with this.

## Memory: introductions are keys, eliminations are values

**Memory is set up as it is.** 61's key is what a frame has received,
and its value what was learned there. In these terms:

- **A key is the type of the introductions a frame has received**: its
  start, then each arrival. A type, never a value: a request's text is
  not in its key (61, The rule), or no two goals would share anything.
- **A value is the steps written there**, each with its eliminations and
  the text each read.
- **The request's identity is checked in the value.** A step depends on
  what it read, so a request asking something else reads differently,
  and the step does not vote. Identity is held where it is checked, not
  where it is looked up.
- **`from` decides nothing.** Which steps vote is a function of what
  their eliminations read, never of which introduction each owns. `from`
  is what makes a step well formed and what a trial varies first; it is
  not a condition.

### Built first: when a step votes

64's dependencies need trials and `span` to move. Before those are built,
one fixed setting of them replaces the guard. A stored step votes at a
frame where:

- **every elimination whose introduction is a literal or a value held
  reads the same text again**: no "move", no `mv`;
- **every elimination whose introduction is the part itself reads
  anything**: passed, free, as 64's open question 9 (`passed` at 1) keeps
  it;
- **every word no elimination read is the same again.** "Don't move that
  file" holds "move" and "that file", and does not vote for `mv`, since
  "Don't" was never read. 64 starts every dependency here; trials free a
  word later.

This is the guard, derived from what the step eliminated, with nothing
written about when. On the spec runs it answers what the prompt fix
answered with keywords: spec A passes its directory, and the rest of its
text stays equal, so spec B does not vote for A's step and startles, and
spec A on another directory votes.

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
  is from its own elimination and that each reads something received. It
  cannot check that the model read the right thing. `mv` from "this
  folder" is accepted, and that premise is the first a trial varies.

## What changes

| 58 to 64 | 65 |
|---|---|
| 58: the model writes a CEL chain | a step: eliminations, introductions, one composition; no CEL |
| 64: a predicate is an elimination the system computes | none for now: an elimination reads text (Open, 3) |
| a step's introductions, unrelated to its eliminations | a substitution: each introduction from exactly one elimination, each elimination owning one |
| 61: an operator is one introduction | a composition of two introductions or more |
| an argument left out | `null`, introduced from an elimination: the system passes nothing |
| 61: a frame starts with its scope, `$1 … $n` | with one value, its parent's composition; the root's is the person's request |
| 61: a cut is a selector applied | an address, and `eliminate`, a part read at it |
| refusals concerning a template: `fits`, `copies`, `join`, `unstored`, `unread` | and `untried`, `unused`, `repeated`, `fragment`; those concerning a condition go |
| 63, section 1: quorum counted on the ballot | built: needed once steps vote only at the text they read |
| `passed` at 1, an open question | built first, with every word not read equal |

In the code: `introduce` and `eliminate` are the methods in `model`; the
log records the graph, and a frame is its projection; `drive`'s prompts
ask for a step and say each introduction replaces what one elimination
read; `Machine.wrote` parses a step and refuses `untried`, `unused`,
`repeated` and `fragment` before any template check; an address gains a
run of words matched by text; a branch's condition is derived from its
step; `tree`'s `StoreConds`, `CheckGuard`, `CommitChain`, `Written` and
CEL go, and `-show` prints steps.

## Required Lean definitions and proofs

In `invariants/Invariants/Refine.lean`, `Model.lean` and `Machine.lean`,
before any Go. 59's tally stands, with its theorems: `tally_perm`,
`tally_unanimous`, `tally_null`, `wins_unique`, `tally_decides_mono`.

- **The graph:** introductions and eliminations, with edges
  *eliminates*, *from* and *composed of*; `introduce` and `eliminate` are
  the only definitions that add a node.
- **A frame is a projection:** its start, its key and its steps are
  functions of the graph; replaying the log gives the same graph, so
  `step_replays` holds over it.
- **One received:** a frame's start is exactly one value, its parent's
  composition or the person's request.
- **A committed step is a substitution:** `from` is a bijection between
  its eliminations and its introductions, each premise at a lower `$n`,
  within the frame.
- **A committed step composes two or more.**
- **Isolation** (`isolate`): in a committed step, varying what one
  elimination reads changes at most the one introduction it owns.
- **The child traces to the parent:** each part of a child's start is an
  introduction of the parent's, from exactly one of its eliminations.
- **`null` is nothing:** an act or return with a `null` argument runs as
  one without it.
- **The refusals are the four:** a step not so formed is refused
  `untried`, `unused`, `repeated` or `fragment`, and nothing is appended.
- **A whole elimination reads its address:** with no part selected it is
  the identity, typed as what is there.
- **A condition is derived** (`derive`): a conjunction of (place, set of
  strings): one per elimination not passed, and one for the words no
  elimination read. The chain's definitions and theorems go.
- **A step votes where it was written** (`derive_self`, replacing
  `commit_witness`): with no other voter its commit selects it, and
  `afterStartle`'s second ask and `unselected` go.
- **Passed is free** (`passed_free`): two frames alike but for what a
  step passes decide alike under it.
- **A word not read is held** (`unread_held`): a frame differing from a
  step's at a word none of its eliminations read is not voted for by it.
- **`from` decides nothing** (`premise_irrelevant`): two steps alike but
  for their premises vote alike at every frame.
- **Quorum counts votes cast on the values**, not steps at the key.

## Open

1. **A premise the sightings contradict.** What is done when an
   introduction is unchanged whatever its premise reads: shown to the
   person, counted against the model, or moved to the dependency that
   stayed equal.
2. **What `return 0` must be from.** Whether `exit(0)` must be introduced
   from an elimination of the goal, as the `yes` example's is, and not of
   what arrived.
3. **Predicates.** A threshold (`restarts > 2`) is not text read again.
   Nothing in `ls`, `cat`, `wc`, `date`, `mv` or `rm` needs one. Whether
   one returns as a kind of elimination, with no expression around it.
4. **The whole elimination in the prompt.** Whether the model is told it
   may read a value whole. Told, it may read each address whole and stop;
   untold, it may invent a part to satisfy the rule.
5. **Joins.** `$5'/'$8` puts an earlier introduction into the value of a
   later one. Whether a value held may be an introduction of this frame,
   or only what was received and what arrived.
