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
| **variable** | an identity and a level: `$n` at every frame with that identity | a variable, bound once in its scope |
| **binding** | a frame and a level: the one value `$n` holds in that frame | a variable's value in one activation |
| **reference** | an occurrence of `$n` in a step: names a variable, and read in a frame gives its binding | an occurrence of a variable |
| **level** | the `n` in `$n`, counted from the frame's start | a de Bruijn level |
| **location** | 61's: an absolute address, a hash; what a binding resolves to | an address in the store |
| **composition** | the introductions a step puts together into the one value it emits: an act, a return, a delegation, a join | an introduction rule with several premises |
| **emission** | the operator that sends a composition: `act`, `return`, `delegate`, and a join's separator. Not an introduction, and outside the rules below | the rule's name, not a premise |

**Introduction and elimination are forms, not a seventh pair.** 61's
pairs stand: interface and observation, implementation and decision.
Introduction and elimination are what both are made of, each with a
static side and a dynamic side:

| | Static (definition) | Dynamic (instantiation) |
|---|---|---|
| a value | interface | observation |
| what decides | implementation | decision |
| an introduction | as written in a step: a literal, a reference placed, `null` | the observation it made in a frame |
| an elimination | as written: an address and a part | the observation it read, bound at its level |

```text
static:   interface   → [ implementation: eliminate …, introduce … ] →  interface
dynamic:  observation → [ decision: each elimination reads, each introduction makes ] → observation
```

- **An interface is defined by introductions**: the type of what was
  introduced, as a type theory defines a type by its introduction forms.
  The interface a frame receives is that of its parent's composition;
  the one it emits, that of its own.
- **An implementation is defined by its eliminations**: what it reads,
  and so where it votes. Its introductions define the next interface,
  not it.
- **The model writes forms; the system derives the rest.** It derives
  interfaces from introductions, the model's and also the person's and a
  leaf's, synthesizing each type and never asking for one. It derives
  where an implementation applies from the model's eliminations. The
  model writes no type and no condition.
- **So memory's key is introductions and its value eliminations**: a key
  is the interfaces received, a value the implementations whose
  eliminations decide when each applies.
- **Syntax is the static side, data the dynamic.** One received, n
  eliminations, two introductions or more is the rule a step obeys as
  written; n to n to n is the graph the log holds.
- **The child's eliminations are in harmony with the parent's
  introductions** (natural deduction's *harmony*): eliminating gives back
  only what was introduced, so a child can read nothing its parent did not
  give it.

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
- **Variable and binding replace 61's position and slot.** 61's
  *position* (interface + `$n`, the key through `$n`) is a variable; its
  *slot* (`id(frame)#n`) is a binding, which resolves to a location. A
  reference `$4` names level 4; in a frame, the binding at level 4 gives
  a location, the hash of the value. The code's `model/positions.go` and
  `tree.Position` are renamed as they are touched.
- **A level is not a location.** A level is relative, meaningful only in
  a scope; a location is absolute. A frame holds locations, at levels.
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
introduce $11 from $8                          file.txt, in place of "that file"
introduce $12 := '.' from $9                   in place of "this folder"
$10 $5'/'$11 $12                               mv ~/Desktop/file.txt .
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

**It is the same in both directions.** An invocation and a return are
one shape: whoever receives sees one introduction, eliminates it, and
emits two introductions or more, composed into one for whoever receives
next.

| Direction | Received, one introduction | Emitted, two or more composed |
|---|---|---|
| invoke | the child's start: the parent's composition | the child's act, delegation or return |
| return | the parent's arrival: the child's return | the parent's next act, delegation or return |

- **A frame receives one value at a time**: its start, the parent's
  composition or for the root the person's request with the tools they
  provisioned; then each arrival, one return. Nothing else: no goal of
  the parent's, no tool the parent kept. This is 61's narrowing, made the
  shape of the frame.
- **Every introduction is from exactly one elimination**, of this frame,
  at a lower `$n`. Two introductions are never from one elimination.
- **Every elimination owns exactly one introduction.** An elimination
  owning none is a reading that leads nowhere.
- **A step composes two introductions or more.** It is a subject and a
  predicate, and no complete sentence lacks either; an exec is a binary
  and its arguments.
- **Emission does not count.** `act`, `return`, `delegate` and a join's
  separator (`'/'` in `$5'/'$11`) are how a composition is sent, not
  introductions: they need no elimination and count toward nothing. The
  composition line is made only of references, each to an introduction,
  and emission (61: an instruction made entirely of variables).
- **An introduction may be composed more than once.** `$5`, `~/Desktop`,
  is in the `ls` and the `mv`. It is still from one elimination; what is
  owned is the introduction, not each use of it.

| Composition | Head, the predicate | Arguments, the subject |
|---|---|---|
| act | the tool run | what it is run on |
| return | line items, each a goal assertion and a child assertion; the first is the exit status and its evidence | (see Returns are line items) |
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

## What the model is told

**The model's words are cut and assert.** The system's terms stand
(introduce, eliminate), and the prompt maps them once:

| The model reads | The system's term | Why this word |
|---|---|---|
| **cut** | eliminate | it means *take a part of this*, and the model took parts with it: in the live runs (below) no step cut the goal whole, where with *read* and *eliminate* most did |
| **assert** | introduce | it says the value is a claim; with it, `exit(0)` was asserted from the goal (*show the contents*), where before it was from `cat`'s exit status |

"Cut" is retired as a system term (Terms) because proof theory gives it
another meaning; the model is not asked to use it in that sense, and the
mapping is stated wherever the prompt is.

**What the prompt says**, in this order:

```text
Definitions
  Cut:        a signal of what you think matters.
  Assertion:  your inference from a cut.

Premise
  The more honest you are with your cuts, the better your memory will serve you in the future.

Rules
  1. Assertions can never be cut. Cut only the goal and what arrived.
  2. When you act, the goal is cut at least once.
  3. Every assertion is derived from a single cut, and every cut gives a single assertion.
  4. Every act sends at least two assertions, and at least one of them is derived from a
     cut of the goal. A return is line items: each is two assertions, the first from a cut
     of the goal (what it answers), the second from a cut of what arrived (the answer).
     The first line item says whether the goal is done: exit(0) or exit(1), and what
     arrived that shows it.
  5. Never leave an argument out: where a tool needs none, assert null from the cut that
     shows it needs none.
  6. exit(0) asserts the goal is done; exit(1) asserts it is not.
  7. A literal is remembered as written: assert only what holds for every goal, a flag,
     never a name, a path, or text you could cut.

Grammar
  the forms (cut, assert, act, return ($g $a) ($g $a) …), addresses, values, and an
  example of an act and of a return
```

- **The order is the order a step is written in**: what may be cut, how
  much of the goal, what an assertion is from, what is sent, what a
  return means, what a literal costs.
- **Rule 4 is not rules 2 and 3 again.** An assertion may be made and
  not sent; rule 4 makes what is sent answer to the goal.

### Returns are line items

**Each line item is what it answers and the answer**: an assertion from
a cut of the goal, then one from a cut of what arrived.

```text
cut $19 := $1.goal / "show"          assert $23 := exit(0) from $19
cut $20 := $18.exit                  assert $24 from $20
cut $21 := $1.goal / "contents"      assert $25 from $21
cut $22 := $18.stdout                assert $26 from $22
return ($23 $24) ($25 $26)
```

- **The exit status is its own line item.** "The goal is done", from the
  goal, beside "the tool exited 0", from the child: both claims kept, so
  neither stands in for the other. The `yes` test is the item's first
  half (Why).
- **A return says only what it has.** A frame whose answer is that it is
  done (`mv`, `rm`) returns one item, the exit and its evidence, with
  nothing made up to fill a count. A return cuts the goal once per item,
  so a return with a payload cuts it at least twice.
- **The parent receives a record labelled by the goal**: each item names
  the part of the goal it answers, so the parent can cut by those words.
  The labels are text, values: the key holds the record's shape, how
  many items and each item's types, never the labels (61, The rule).
- **The verdict can fall on an item**: "the contents" refuted while
  "done" held is a finer signal (66) than the return refuted whole.
- **What an act needs of the goal is one cut**, the reason for what it
  runs; a return needs one per item.
- **Every rule is checked**: a step that breaks one is refused, told
  which, and asked again. Rules 1 to 5 and 7 are 65's syntax; rule 6 is
  the meaning of the return's head.

**What the live runs showed** (`ls`/`cat` on two directories, `wc`,
`date`; 2026-11-10):

- **The forms were written from one example**, with two or three
  refusals a run at most, the same mistake each time: cutting its own
  earlier assertion. Rule 1 is that refusal, written down.
- **Without the goal rules, premises were tautological**: `ls` asserted
  from a cut of `$1.tools[1]`, itself. With rules 2 and 4, every chosen
  value was asserted from a cut of the goal: `cat` from *show the
  contents*, `-l` from *lines*, `date` from *what time*.
- **`null` was claimed against a cut** at the first `date`: `null` from
  *is it*, and `date` ran with no argument.
- **What arrived was sometimes cut whole** (`$6.stdout`, newline and
  all), the act failed, and the next step cut `$6.stdout[1]`. The world
  corrected it in one step, as 66 expects.
- **With returns as line items** the padding went: `mv` and `rm`
  returned one item, `(exit(0)` from the goal, the exit status`)`, and
  `wc` returned `("the lines of this file", "2")`. All six programs right,
  one refusal (a tool cut), fixed on the next ask.
- **The experiment is kept** in
  `controller/apprentice/typesys-prototype/experiment/65-cut-and-assert/`:
  the driver, the prompt levels, a setup script, and every transcript
  under `data/`, by round.
- **One layout of the prompt was declined by the API** before any
  output, for content that is not unsafe: the rules numbered under
  headings, with the grammar after them. The same seven rules, the
  definitions and the premise, added one at a time to the layout that ran
  cleanly (rules after the grammar), were never declined.
- **With all seven rules, six programs ran to a right answer** (`ls` and
  `cat` on two directories, `wc`, `date`, `mv`, `rm` in a sandbox) with no
  step refused. `mv` and `rm` each looked with `ls` before acting, and
  asserted the file from what arrived, never a literal.
- **Rule 4 is met on a return by filler** when `exit(0)` is asserted from
  what arrived: `mv` returned the goal's words *into the second folder*,
  and `rm` returned its exit status, to have something sent from the goal.

## Memory: one store, through the S

### The S

Every role receives one introduction and emits introductions. Only the
middle one is required to eliminate between them:

```text
root          frame (delegate / act)                 leaf          frame                                  root
introduce  →  [ introduce → eliminate → introduce ]  →  introduce  →  [ introduce → eliminate → introduce ]  →  introduce  → …
request        received    read parts   emitted        return         received    read parts   emitted        verdict, next request
```

- **Root and leaf: the system requires introductions only.** The person
  asserts a request and takes or rejects what returns; a binary takes
  its argv and gives its return. Neither shows the system an
  elimination, and none is asked for.
- **Frames: the system requires an elimination before every
  introduction.** A frame is the only place the system decides, so it is
  the only place it can demand a derivation.
- **Memory is a never-ending S through it.** Each hop writes into the
  memory of the role that receives.

### Identity: where, who, role, what

**A key is an identity, and its prefix is where, who and role:**

| Part | Is |
|---|---|
| **where** | the OS |
| **who** | one binary |
| **role** | root, frame (delegate or act), or leaf |
| **what** | depends on the role |

**One identity is one binary.** What makes up a *who* (a model, its
version, its settings, a tool and its flags' dialect) is provisioned by
the administrator, and the system does not standardize it. It sees one
thing: the binary, by content hash. Claude Opus 5 and Claude Opus 5.5
are two binaries, each wrapped by the administrator, and two identities.

- **The model is in a frame's *who*.** A frame's who is the binary of the
  model that writes its steps, for isolation's sake: a trial measures a
  model (64, Trials), and votes from two models under one identity
  cannot be told apart.
- **Sharing what was learned is policy, not data.** The data keeps every
  *who* apart, so that whatever is shared can be traced. Whether steps
  one binary learned are offered to another, and how they are counted,
  is a policy over that data, never a property of the key. 61 shared
  across models by leaving the model out of the key; here the key keeps
  it, and sharing is decided above it.
- **The working directory is not in *where*.** A frame knows its
  directory only by calling something that shows it, `ls` or `pwd`, and
  then it is an introduction a leaf asserted, in the frame's path like
  any arrival. A leaf's return under the same argv in another directory
  is a different return at the same identity: a surprise (64,
  Expectation), recorded.
- **What a directory means for an effect is the administrator's.** A
  binary that must act only in one place is provisioned to: its wrapper
  sets its directory, and that is in its hash. Anything more, such as
  requiring an effect to follow a look, is policy per *who* (Open, 5).

### What the data owes a policy

Policies are not designed here, and the system has no API for them yet.
The data must not stand in the way of one. A policy has two places to
act, and both exist:

| Hook | Where | A policy decides |
|---|---|---|
| **the fold** | memory, a fold of the log | which events count toward a key: this *who*'s only, or another binary's too |
| **the gate** | between a decision and its act, at `trust` | whether this act may run |

So the data guarantees:

- **Every event carries its identity**: where, who and role. A fold can
  select by *who* only if the *who* was written, and a log written
  without it cannot be read under a policy later.
- **Memory stays a pure fold.** A policy changes which events a fold
  reads, never the events, so memory is the log's replay under any
  policy.
- **The gate sees the step**: its eliminations and what each read, not
  only its composition, so a rule such as *an effect follows a look* can
  be checked there.

**Role says how a key is used:**

| Role | who | what | Writes | Used for |
|---|---|---|---|---|
| **root** | the person's entry, the `skyra` binary | the request, as asserted | the request; then the verdict, taken or rejected | where credit is anchored (64, The signal) |
| **frame** | the model's binary | the types of what it has received: its start, then each arrival | steps | decisions: the tally |
| **leaf** | the tool's binary | its composition, literally: the argv | the return observed | expectation, trials and shadows; for a read-only tool, what to expect without running |

- **One store, not three.** Role is in the key, so there is one memory,
  one key shape, and one fold of the log (`step_replays`).
- ***What* is a value at root and leaf, and a type at a frame.** Root
  and leaf are observed: their what is the request and the argv as they
  were. A frame decides, so its what stays a type, or nothing would be
  reused (61, The rule). A key holds no value only where it is reused.
- **A frame's identity is what it has received, in order.** Its start,
  then each return: 61's trie path. When a child returns, the frame's
  identity is no longer its start alone: it is its start and the child's
  introduction. The step stored at the shorter path stays true of it; it
  is the step that made the call.
- **Invoke and return are one operation:** a frame receives one
  introduction, and memory appends it to the frame's path. A child's path
  starts with its parent's composition; a parent's grows by its child's
  return. The parent cannot tell whether its child was a frame or a
  leaf.
- **Each arrival is isolated.** It is one introduction at one position
  of the path, so a trial at a parent varies one child's return and
  holds the rest.

### What a frame's value holds

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

### Variables and bindings

**A value is found by its frame and its level.** Levels are append-only,
so `$n` in a frame is written once and never changes: a binding never
dangles and never comes to mean something else. A part of it is found
from there by elimination (61, Reference and location).

| Address | Gives | Is |
|---|---|---|
| frame + level | one value | a **binding**: a row, one sighting |
| identity + level | every value bound at that level, at every frame with that identity | a **variable**: a column, all the sightings |

- **A frame is named by its process and itself.** Two runs each have a
  frame 1; a binding's frame is unique across processes, so the store's
  ids are (process, frame).
- **A binding's identity follows from it**: the frame's path up to the
  level, its start and the arrivals before `$n`. Carried in an address,
  it is a check.
- **A variable is stable by construction.** It is an identity and a
  level, both static, so it names the same thing for as long as its
  identity exists. Every sighting a policy counts is a binding of it.
- **The API exposes values by binding, and stability is computed over
  variables.**

### What memory knows: the reverse index

**Eliminations index the introductions they have owned.** Read the other
way, the log's many-to-many relation gives, for each text an elimination
has read, every introduction put in its place and where:

```text
"move"        →  mv                     at who = opus-5.5's binary, 14 sightings
              →  a moving company's API  1 sighting
"all files"   →  -a
"this folder" →  .
```

- **It is derived**: a fold of the log, as memory is. Nothing new is
  stored.
- **It is what memory knows**, and its counts are what is most probable,
  relative to the identity it is read under. Which identities count, one
  *who* or several, is the fold's policy (What the data owes a policy).
- **It proposes and never decides.** It may offer candidates at an ask,
  and substitutions at a trial (64, Trials). The tally counts only what
  was written.
- **Shown candidates leak.** A step written with candidates shown may
  agree with the prompt, not with the goal, so it is recorded as shown,
  and the history policy decides what such a vote counts (as 64 counts a
  retry).
- **It is not read at a program's boot.** Boot is decided as it is (61),
  with nothing proposed from what other runs learned.

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

### Deciding: three policies

**A decision is made at a stable identity, on a settled history, by
agreement.** Each is a policy, set at boot and changeable over the
program's life. The system exposes what each reads: the identity, each
step's eliminations with what each read, its introductions, and two
counts.

| Policy | Asks | Reads | Today |
|---|---|---|---|
| **identity is stable** | does this identity behave the same each time it is met? | the sightings of its variables, and what followed each: a hesitate, a surprise | nothing; `hesitate` is its coarse form |
| **history is stable** | has what was decided here settled? | the votes cast at these values, and how they split | `quorum`, counted on the ballot (63, section 1) |
| **the steps agree** | which step, by how much? | the steps voting | 59's `group`, `levels`, `floor` and margin `required(ι, cast)` |

- **Identity stability is about the world**: whether the same identity
  predicts what follows it. `date` never is; `ls ~/Desktop` is until the
  desktop changes.
- **History stability is about the decisions**: whether the steps
  written at an identity agree. A model can agree with itself at an
  identity whose world keeps changing, and a stable `ls` can draw three
  steps from a model not yet settled. The two fail apart.
- **Any one failing is `null`**, and the frame asks.

**What no policy changes:**

- **A decision selects only a step written at its identity.** A policy
  chooses or withholds; it never authors.
- **No vote is `null`, and `null` asks.**
- **The tally's own properties hold under every policy**: order does not
  matter (`tally_perm`), at most one winner (`wins_unique`), unanimous
  votes select (`tally_unanimous`).
- **A stricter stability policy only withholds** (`decideAt_quorum_mono`,
  generalized).

**A policy change is an event in the log.** A decision is a function of
memory and the policy in force; with the policy outside the log, a
replay would not give the same decisions. So it is appended, as a
`forget` is, and memory and every decision replay exactly.

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
| 61: a key is a frame's scope and arrivals; the model is never in it | where (the OS), who (one binary, the model's for a frame), role, then what |
| 61: one memory for frames | one store through the S: root, frame and leaf keys, by role |
| the model writes eliminate and introduce | it reads cut and assert, mapped once; seven rules, each checked |
| `exit(0)` from anything | from a cut of the goal: rule 4, and rule 6's meaning |
| 63, section 1: quorum counted on the ballot | built, as the default history policy |
| 59: the tally's constants are the program's | three policies, set at boot and changed by an event in the log |
| 61: position and slot; "variable" for `$n` loosely | variable (identity, level), binding (frame, level), reference, level, location |
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
- **A committed return is line items:** each item is two introductions,
  the first from an elimination of the goal and the second from one of
  an arrival; the first item's first introduction is an exit status.
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
- **A binding is unique:** (process, frame, level) gives at most one
  value, and it never changes once written.
- **A variable is static:** (identity, level) is a function of definitions
  only, and every binding of it is at a frame with that identity.
- **Three policies decide:** a `Policy` record of identity stability,
  history stability and agreement, with `quorum` counted on the ballot
  and 59's equations as the default.
- **The floor holds under every policy:** a decision selects only a step
  written at its identity; no vote is `null`; `tally_perm`, `wins_unique`
  and `tally_unanimous` hold; a stricter stability policy only withholds.
- **A policy is in the log:** decisions replay exactly across a policy
  change.
- **A key's prefix is where, who and role:** two keys differing in any
  of them share nothing, and a frame's who is its model's binary.
- **Roles bind differently:** root and leaf events carry introductions
  only and are never refused for lack of an elimination; frame steps are
  refused without one.
- **Every event carries where, who and role**, and memory is a fold of
  the log under any selection of its events by them (`step_replays`,
  restated over a filtered log).
- **The gate's input is the step:** the edge an act runs from is a
  function of the decided step, its eliminations and what each read.
- **Invoke and return are one append:** a frame's key after an arrival is
  its key before, extended by that arrival's type.

## Open

1. **A premise the sightings contradict.** What is done when an
   introduction is unchanged whatever its premise reads: shown to the
   person, counted against the model, or moved to the dependency that
   stayed equal.
2. **Predicates.** A threshold (`restarts > 2`) is not text read again.
   Nothing in `ls`, `cat`, `wc`, `date`, `mv` or `rm` needs one. Whether
   one returns as a kind of elimination, with no expression around it.
3. **Policy per *who*.** What may be shared between binaries, and what
   an effect must have read first (an `rm null` learned where the frame
   never looked runs wherever its binary is started), are policies over
   the data, set per *who*. Not designed here.
