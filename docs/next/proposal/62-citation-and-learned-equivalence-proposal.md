# 62 — Amendment to 61: citation, and equivalence that is learned

Date: 2026-11-10

Status: **draft, being worked through**. Amends
[61](61-static-and-dynamic-proposal.md); everything in 61 it does not name
stands. It answers 61's open question 6 (goal-dependence) and most of 18
(asking at an introduction), and removes the program's `classes`. Nothing
here is built. The syntax shown is provisional.

Three changes:

1. **An introduction cites.** Every introduction names the eliminations it
   was derived from. A cited span is code, and the citation is compiled
   into the guard, so what was derived runs only where what it was derived
   from is read again.
2. **An introduction is a memory entry**: its key what was cited, its value
   what was derived. It is no longer a literal inlined into one template.
3. **Equivalence is learned, never declared.** `classes` leaves the
   program. Two alternatives at one position are held interchangeable only
   on what memory holds, counted against a new constant, `span`, and a
   trial still refutes them.

## Why

**A goal's interface is `string`, so every goal shares a key.** Under 61 a
frame's key is its scope's interfaces and its arrivals', never their
values (`Model.key_values`). Two goals under the same tools land on one
key, and the implementations stored there are tallied for both. Only a
guard's conditions tell the goals apart, and nothing makes a guard read
its goal.

The live eval shows it. The return Claude wrote with `cat` was
`size($4) > 0 ? "return 0 $4"` (61, What a trial cannot see): it reads
what arrived and never `$1`. At that key, *show the contents* and *count
the lines* decide alike, and the second returns the file's contents, exit
0. A goal the system has never seen is treated as seen. 61 leaves this as
open question 6, and 58 as its question 9.

**A person's assertion is the other source of wrong answers.** With `cat`
and `wc` declared one class, `wc`'s counts came back as the file's
contents, exit 0; undeclared, the model asked at the `wc` child returned 1
(drive's `TestClaudeClasses`). The declaration turned an honest failure
into a wrong answer, and the system cannot check what a tool means.

Both are one gap: what was learned is reused where nothing ties it to what
it was learned from. Citation is the tie.

## What stands

- **`string` is the type of what has not been understood** (61, Interface
  and observation). A goal never seen arrives as `$1: string`. That is the
  type of a situation never seen, and it is not changed.
- **The rule.** A definition never holds a value. A cited span is code, as
  a guard's `x > 7` is and as `exit(n)` is a singleton type; so the key
  stays a function of types alone.
- **Elimination and introduction**, told apart by form; variables
  append-only; an instruction made entirely of variables; the copy check
  (what a cut takes is cut, not written).
- **One map.** There is no input and return split, so everything below
  holds for a goal, a context value and an arrival alike.

## Citation

**Every introduction cites.** An introduction names one or more
eliminations, each a cut of something the frame has seen:

```text
$1: "list all files in /tmp"                 goal
$2: leaf ls                                  tool

eliminate $4 := $1 / "all files"             cited
eliminate $5 := $1 / "/tmp"                  passed
introduce $6 := "-a" citing $4
$2 $6 $5
```

An introduction is a relationship between what the model was given and
what it derived from it. The citation records the first half; the literal
is the second. An introduction that cites nothing is refused, with the
reason, and the model is asked again.

**A span is cited or passed.**

| The span is | Used as | In the reading (below) |
|---|---|---|
| **cited** | the reason for an introduction | a literal: code |
| **passed** | an argument, a `$n` in the template | a hole: `string` |
| neither | nothing | not carried forward |

**Not every span is carried.** What the model neither cites nor passes
does not reach the decision: a stored decision is a function of only the
spans its implementation cited or passed. So the citations are the model's
attention, recorded: what it held worth carrying forward, visible in the
log and in memory.

**A citation is compiled into the guard.** At commit, each citation
becomes a condition of the link that introduces: `-a` is introduced only
when `$1` reads `all files`. Change that span and the link selects
nothing. A goal none of the stored implementations' citations read gives
an empty ballot: `null`, `startle`, the path that exists. This is what
makes a goal never seen startle, with no change to the key.

**Citation works anywhere.** An arrival is a `$n` as the goal is. Citing a
cut of stderr to introduce `exit(1)` is the same act as citing the goal to
introduce `-a`:

```text
$3: {exit: 1, stdout, stderr}                          arrived
eliminate $7 := $3 / .stderr / "No such file"          cited
introduce $8 := exit(1) citing $7
return $8 $3.stderr
```

A return's exit status is an introduction (61), so it cites too:
`exit(0)`, the claim that the goal is met, cites the spans of the goal it
answers. The `cat` return above could not be stored as written: its
`exit(0)` would cite *show the contents*, and would not select for *count
the lines*.

**What the system checks, and what it cannot.** It checks that every
introduction cites, that each citation is an elimination of something
seen, and that the cited text is read where the link runs. It cannot check
that `all files` means `-a`. A wrong citation is trusted until a tally or
a trial goes against it, as any implementation is.

### Lexical units

**A span is a run of words, not of a model's tokens.** A tokenizer belongs
to one model, and the model is never in the key: implementations learned
under one count under another (61, Scope and frame). Token boundaries also
move with a leading space or a neighbour, and the shell sees text, never
tokens. A word is defined once, in the language version, which memory's
root already hashes: whitespace-delimited, so `/tmp/x` is one word, as it
is one argument.

This takes the selector grammar one level down. 61 cuts runs of lines and
field paths; 62 adds runs of words within a line.

**A cited span is matched by its text, not its index.** Word positions
move when an earlier passed span changes length; the literal `all files`
is found wherever it lands. A cited selector is therefore the text itself
(`$1 / "all files"`), matched on word boundaries, and reads nothing when
the text is absent.

### The reading

**Citations refine `string`.** 61 learns how bytes are read (Reading what
arrives): the first arrival at a key is text, the elimination the model
wrote is stored, and later arrivals are typed on arrival. A goal is read
the same way. An implementation's cited and passed spans give a
**reading** of the value they cut: its literals, in order, with a hole at
each passed span.

```text
"list all files in /tmp"      read as      … "all files" … ⟨string⟩
```

**Order is in the reading**, as it is in the key: the literals are a
sequence, and a goal holding them in another order is another reading.

Promotion (61, step 3) then applies: a goal a learned reading fits is
typed by it on arrival, and the reading is in the key; a goal no reading
fits stays `string`. The compiled conditions above already give the right
decisions without it. Promotion is what moves the difference from the
guard into the trie, so that goals of one reading share a prefix and reuse
rules can see how far two goals agree (open question 3).

**The exact key keeps everything.** 61's principle holds: the key is
everything seen, never a projection. Matching on cited spans alone is a
projection, so it is a **reuse rule** over exact keys and not the key
itself. Under-attention is the risk it carries: a model that never cites
`not` in *do not delete* writes an implementation that selects with or
without it. Quorum is the existing defense, several voters attending
differently, and the rule is a hypothesis as every wider rule is (open
question 2).

## An introduction is a memory entry

Under 61 an introduction is inlined into its template at commit, as a
literal (`Model.inline_seen`). The literal is then in one frame's
implementation, and nothing records where it came from.

**An introduction is stored as an implementation is**: something was seen,
the model derived something from it, and the pair is kept.

```text
key:    $1 / "all files"  →  def ls      what was cited, then what it is introduced for
value:  "-a"                             what was derived
```

- **The key is a path, not one hash**: the interface of each cited cut, in
  order, then the definition the introduction is an argument to. 61
  already keeps a cut undivided in the key (`I / sel`), and an
  elimination's interface is synthesized, never asked for. Everything
  derived from `all files` shares a node and parts by destination: `-a`
  for `ls`, something else for `find`, `exit(0)` for a return.
- **Both sides are code.** The entry is a definition and holds no value.
- **The template names the entry**, not the literal. At `trust` the entry
  is looked up and its value bound, as a reference is resolved.
- **It is tallied.** Each `startle` that introduces at a key is one voter
  there. `-a` and `-la` derived from one span are a conflict at one key,
  settled as any conflict is. Quorum and inhibition apply unchanged.
- **It is shared.** Any frame that cites the span for the same destination
  finds what was derived, with no ask. This is what lets the system type
  an introduction without the model (61, open question 18): memory holds
  it.

Under 61, an introduction is in effect keyed by everything its frame saw.
A citation narrows that key to what mattered, and the narrower key is
reused more widely.

## Equivalence is learned

**`classes` is removed from the program**, with the `-class` flag, boot's
checks on classes, and a class's location as the hash of its members. A
program asserts nothing about which tools do the same job.

**Natural language is syntax to the model, and tools are too.** Two
spellings are held the same when different inputs arrive at one output:
`all files` and `every file` both deriving `-a` for `ls`; `cat` and `head`
both filling an act's first position under one template. These are one
fact. Two keys that differ at one position and hold the same value make
that position a **sum** (61's `I₁ + … + Iₙ`, the alternatives one position
takes), now written by memory instead of by a person.

**Held, not proven.** Agreement on the inputs seen is not equivalence. 61
says so of trials: one selecting what the other does is no evidence, since
a weak guard selects anything. So:

- A learned equivalence is a hypothesis and is never final. A refutation
  is final, kept in memory beside the keys, by the pair and the position.
- It is only as strong as the guards behind it. Compiled citations make a
  guard harder to satisfy by accident, which is what the `cat` and `wc`
  guard lacked.
- **Borrowing, trials and refutation stay** as 61 has them (Reuse rules):
  when a frame's own key selects nothing, the key with one alternative
  swapped is tallied; a borrowed act runs the original as a shadow; outputs
  that refine differently refute. What changes is who proposes the pair.
- **`readOnly` stays.** It declares a tool's effects, not its likeness to
  another, and a shadow is safe only under it. A tool not declared
  read-only is never borrowed for: the model is asked.

### `span`

Quorum is how many sightings at one key before its tally is trusted:
depth. It already gates a borrow, since the swapped key is decided under
the same quorum. It does not count in how many places two alternatives
agreed: breadth. A pair that agreed five times at one key has been
compared in one context.

```text
constants: {
  …
  quorum:  nat ≥ 1      implementations a key needs before its tally is trusted
  span:    nat ≥ 1      distinct keys at which two alternatives hold the same
                        value before one is borrowed for the other elsewhere
}
```

- **A constant of the program**, set at boot, read through the static
  link, recorded with boot in the log, and out of memory's key. Programs
  differing only in `span` share memory.
- **It only withholds.** Raising it never changes what a borrow selects,
  only whether it happens, as raising the quorum does
  (`decideAt_quorum_mono`).
- **Counted per pair, per position.** `cat` and `head` agreeing as an
  act's tool says nothing of them elsewhere.
- **A refutation overrides any count.**

**The cost.** Each new tool or spelling takes at least one ask before
anything is shared. Declared classes saved the second child's ask from the
first run (61: `cat` and `head`, `ls` and `find`). That saving is traded
for never being wrong because of what a person asserted.

## What changes in 61

| 61 | 62 |
|---|---|
| Variables: an introduction constructs from nothing | constructs from what it cites; one citing nothing is refused |
| Variables: an introduction's interface is asked of the model until memory holds it | memory holds it: the entry at its cited key |
| Variables: inlining at commit replaces an introduction by its literal (`Model.inline_seen`) | by the entry's key |
| Reading what arrives: a learned decoding (`/ json`) | also a learned reading: literals and holes, in order |
| Selectors: field paths, runs of lines | and runs of words; a cited span by its text |
| Program: `classes` | removed; `constants.span` added |
| Boot, step 1: each class has two members or more, all read-only, each member in one class | removed; `span ≥ 1` |
| Reuse rules: classes are a reuse rule, tried | learned sums are; the trial is unchanged |
| Open question 6, goal-dependence | settled: a citation is compiled into the guard |
| Open question 18, asking at an introduction | settled where memory holds the entry; its deferred goal (replace introductions with eliminations) is what a citation does |

In the code: `Program.Classes` and `Ids.Class` go, with `cmd/skyra`'s
`-class`; `Constants` gains `Span`; `Entry.Introduce` gains what it cites;
`Sel` gains word runs and text; `Frame.Inline` stores an introduction's
key; refutations are keyed by pair and position.

## Open

1. **How a passed span is found.** A cited span is matched by its text. A
   passed span has no text of its own: it is found by word position, which
   moves, or by the literals beside it, which makes its neighbours part of
   the reading whether or not they were cited. Two passed spans side by
   side have no boundary between them.
2. **Uncited text: literal or hole.** Whether the cited-spans reuse rule
   may ignore everything uncited, or only what several voters agree to
   ignore. *do not delete* is the case to get right.
3. **Promotion of a reading.** When a reading enters the key, and how
   implementations stored under `string` are carried to the read key
   without relearning (61, open question 19, asks the same of `/ json`).
4. **An implementation that neither cites nor passes its goal.** An act
   with no introduction (`$2 $3`, both arrivals) reads no goal span and
   selects for every goal. Whether each link must read its goal at least
   once, or whether the return's `exit(0)` citing it is enough.
5. **A cited text that occurs twice**, or not on a word boundary.
6. **What a response to an arrival cites.** An introduction answering a
   tool's output cites the arrival. Whether it must also cite the goal, or
   whether the tree's links back to the goal are enough.
7. **Distinct, for `span`.** Keys differing outside the swapped position;
   or, stricter, keys of different goal readings. Two keys sharing a long
   prefix are barely two contexts.
8. **`span`'s default.** Any finite default borrows with no declaration,
   where 61 borrows only under a declared class. 1 is the most permissive;
   off unless set keeps today's behavior.
9. **The copy check's false positives** (61, open question 17) remain. A
   citation gives a refused introduction a second reading, *written, and
   derived from this span*, which may be what lets the model override.
10. **Several citations.** An introduction citing two spans is keyed by
    both, in order. Whether order is the order cited or the order seen.
11. **Rephrasing.** Any new spelling startles, which is right by
    definition and costs an ask each. A reuse rule over near readings is
    the next thing to design, and `span` is its gate.

## Required Lean definitions and proofs

In `invariants/Invariants/Model.lean` and `Machine.lean`, before any Go:

- **Every introduction cites:** a committed implementation holds no
  introduction with an empty citation; each citation is an elimination of
  a `$n` seen.
- **A citation is a condition:** a link that introduces selects only when
  each cited cut reads its text; so an implementation whose citations a
  value does not read casts no vote for it.
- **Definitions still hold no value:** an introduction's key is a function
  of cut interfaces and a definition; `key_values` holds with readings in
  the key.
- **Uncited spans do not reach the decision:** two values agreeing on
  every span an implementation cites or passes decide alike under it.
- **Word cuts:** a word selector's interface is its source's with the
  selector applied; a text selector reads its text or nothing
  (`cut_sound`, extended).
- **An introduction is a fold of the log:** replaying the committed
  `startle`s gives the same entries, as it gives the same memory.
- **Exact reuse of an entry aligns by identity.**
- **No class in the program:** `Program.wellFormed` reads no class; boot's
  theorems hold without one.
- **Borrowing is sound without a declaration:** `borrow_sound` and
  `go_refuted`, restated over a pair at a position.
- **`span` only withholds:** raising it never changes what a borrow
  selects (`decideAt_quorum_mono`'s sibling); a refuted pair is never
  borrowed for, whatever the count.
