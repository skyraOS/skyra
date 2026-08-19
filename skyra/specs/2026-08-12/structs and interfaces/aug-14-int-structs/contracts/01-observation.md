# contracts/01 · The Observation

**Package:** `observation` (proposed; `content` remains its own package)

The observation is the only semantic currency. Everything the seven primitives pass, return, or
retain is one.

---

## 1 · The shape

```go
type Observation struct {
    ID       ids.ObservationID
    Content  content.Content
    Producer Provenance          // SUBSTRATE-SET. see §2.
}
```

Two fields the current runtime's `graph.Observation` also has, and one that changes meaning.

```text
CARRIED FROM ../../03-graph.md    ID and Content, unchanged.
CHANGED                        Producer. was `ids.EngramID`, with EMPTY meaning ingress.
```

`Content` and its closed `Value` set carry over from `../../02-content.md` intact — `Str` `Num` `Bool`
`List` `Obj` `Ref` `WorldRef`, closed by an unexported `isValue()`. **The closure is still
load-bearing** for the same reason: `Cites` is derived by scanning content for `Ref`s, and a scan
over an open set cannot be exhaustive.

One member of that set does not carry over. See §4.

---

## 2 · Provenance is substrate-known, not self-reported

`engram-contracts-and-composition.md` is unambiguous:

> *"An observation carries content. The substrate carries its provenance."*

and states the failure it is refusing:

```text
NOT SUFFICIENT     {"type": "reach"}
LOAD-BEARING       this observation is known BY THE SUBSTRATE to have been produced by the
                   behavior whose observations this contract accepts
```

So provenance must be a field the substrate sets and content cannot express:

```go
type Provenance struct {
    Engram     ids.EngramID       // which behavior produced it
    Invocation ids.SpanID         // which opening of it          ← NEW vs the runtime
    Position   Position           // the child position it occupied, if any
    Origin     Origin
}

type Origin uint8

const (
    OriginReturn   Origin = iota   // returned by a child possibility
    OriginApproach                 // entered from outside. no producer exists.
    OriginThink                    // authored by a thinking behavior beneath Think
    OriginOrient                   // constructed by the kernel from live topology
)
```

**Why the invocation and not just the engram.** A contract may accept observations from a behavior;
necessity testing in `../../../causality/observable-causality.md` requires knowing *which run* produced
one. The current runtime carries `Producer ids.EngramID` alone, so two openings of the same engram
are indistinguishable in provenance, and ablation cannot name its subject.

**Why `Origin` is an enum and not an empty-value convention.** `../../03-graph.md` records
`Producer ids.EngramID // EMPTY means INGRESS`. An empty string standing for a semantic case is the
shape that lets a zero value pass for a decision. Approach observations genuinely have no producer —
`primitives.md` §1: *"the observation enters with an empty producer because no possibility produced
it"* — so the case is real and deserves a name rather than an absence.

```text
FOR REVIEW    OriginThink and OriginOrient are proposed, not carried. they exist so a contract can
              distinguish a candidate from a return without reading content. whether that
              distinction belongs in Origin or in the contract vocabulary is open —
              see open-questions.md H4.
```

---

## 3 · What the observation may not carry

```text
NO TYPE TAG          P4. a content field named `type` is data, never typing.
NO HORIZON           P14 and think-interface-proposal's rejected shape: "HORIZON TRAVELS AS AN
                     OBSERVATION — authority would become caller-authored data."
NO ADDRESS           P12.
NO VERSION POINTER   ../../../reach/age.md: "THE POINTER LIVES ON THE ENGRAM, NEVER IN CONTENT."
                     if the credit walk could descend it, every edit would inherit standing.
```

The last one is structural rather than argued, and worth stating as such: `Cites()` scans content,
the version pointer is an engram field, and a scan of content cannot reach an engram field. No
citation can name a predecessor.

---

## 4 · Finding — provenance typing dissolves `EngramRef`

`../../02-content.md` introduces the one genuine addition to the closed Value set:

```go
type EngramRef struct { ID ids.EngramID; Name string }   // ADDED
```

It exists to enforce one distinction, quoted there from `grammar.md`:

```text
DELEGATION   reach arriving from an entity I CHOSE to invoke.   legitimate.
INJECTION    reach arriving in the data channel.                whoever can talk to me picks
                                                                what I can touch.
```

— and that document's own note says *"Same bytes, different provenance, and only the type keeps them
apart."*

**In the new model, provenance keeps them apart, because provenance is now a substrate-set field.**
The distinction is exactly delegation-versus-injection, which is exactly a question about who
produced an observation, which is exactly what `Provenance` answers. A parent-owned contract can
accept observations with `Origin == OriginThink` from a known reach-thinking behavior and accept
nothing else at that position. An injected one has the provenance of whoever sent it and fails.

This has three consequences.

```text
1  THE VALUE SET LOSES ITS ONE ADDITION.  content.Value returns to exactly what
   2026-08-09/interfaces.md declares. ../../02-content.md's "FOR REVIEW" is answered by removal
   rather than by a decision to keep it.

2  THE E1/E2 STOP DISSOLVES.  ../../../overnightBuild/stops.md stopped on: "Raw is text, and reach
   cannot be written as text" — three resolvers commit content containing an EngramRef, and
   grammar/parse.go:172 toValue produces only Ref, WorldRef, or Str. With reach carried as
   provenance rather than as a content value, there is nothing for the grammar to need a surface
   form for. Step can return text.

3  questions.md §15's WRINKLE NARROWS.  ../../02-content.md records that because reach is not a Ref,
   "a judge that picks among candidates is invisible to credit." Under provenance typing the
   judging behavior is a producer with a span, so it is visible in the causal cone. Whether it
   should receive CREDIT is a separate question and is NOT settled here — credit descends
   observations, and this only establishes that the judge is nameable.
```

```text
FOR REVIEW, AND IT IS THE BIGGEST CALL IN THIS DIRECTORY.
Removing EngramRef is a subtraction from a closed set, which is the safe direction, but it moves
an enforcement from the type system into a substrate-set field. The type system refuses at compile
time; a field is only as good as the code that refuses to trust content over it. The argument for
doing it anyway is that the field is unforgeable BY CONSTRUCTION — content cannot express
Provenance — whereas EngramRef was a content value and therefore always one parse away from being
constructible by whoever could write content.
```

---

## 4a · The other pointer type, and why it stays

§4 rebuilt the reasoning for one member of the closed Value set and left the other carried by §1's
blanket *"carry over intact."* `WorldRef` is that other one, and it deserves the same pass.

```go
type WorldRef struct { Host ids.HostID; Handle string }   // KEPT
```

**Is it an address, and therefore forbidden by P12?** No — and the criterion is sharper than the
question:

```text
AN ADDRESS RESOLVES.     ../../../misc/os-parallels.md 8: "the scheme IS the resolution procedure."
A WORLDREF DOES NOT.     ../../02-content.md: taking one to another host is "an ERROR, never a
                         lookup." It has no resolution procedure, so it is not a name in a
                         namespace.
```

P12 governs how you REACH something. A `WorldRef` presupposes the reach already happened; it is
what exists afterward.

> **You open a program on your desktop and it does not change the desktop. It opens, and it sits on
> top of it.**

That is the layering, and it is the whole justification. A `WorldRef` points UP the stack at
something instantiated above a host, never DOWN into how that host is contacted. Addresses and
transport stay below the Reach boundary; a receipt for what was opened above it is ordinary semantic
currency.

**And the `HostID` is not a P13 violation.** P13 forbids locality as an ENTITY PROPERTY. This is not
that: it is the desk the window sits on, and it travels with the receipt because the receipt means
nothing elsewhere. Which also makes `Fingerprint`'s `w:handle@host` correct rather than an exception
to *recorded, not keyed* — two receipts on two hosts are genuinely not the same material.

**The direction rule this makes explicit**, stated nowhere but observed in two places:

```text
kernel/07 §3   THE KERNEL HANDS BACK HANDLES.  Head.Engram is an ID; the reader dereferences it.
kernel/03 §3   THINK REFUSES TO.               "the candidate is an observation, NOT a handle" —
                                               a handle would imply something had been opened.
HERE           A POSSIBILITY MAY PRODUCE ONE.  because by then something HAS been opened.
```

Three positions, one rule: a handle may only exist where the thing it names already does.

```text
HOLE — LIFETIME. The desktop analogy has one asymmetry, and it is the durable half.

A program can be closed. The graph cannot forget the receipt: append-only, and there is no
`delete()` in graph or kernel outside tests. So a WorldRef written into a durable observation
outlives what it points at, permanently and by construction.

../../../misc/os-parallels.md 14 flagged dangling targets as an UNTESTED symlink problem. For
WorldRef it is not untested, it is EVENTUAL. Whether a stale receipt is a defect, or simply an
accurate record that something was open once, is not decided. RECORDED AS H14.
```

---

## 5 · The folds

Carried from `../../02-content.md`, unchanged in definition:

```go
func (c content.Content) Cites() []ids.ObservationID
func (c content.Content) Shape() string
func (c content.Content) Render() string

func Fingerprint(c content.Content, d Deref) string
func FingerprintSet(cs []content.Content, d Deref) string
```

`Content.Reach()` is removed with `EngramRef`.

All remain **read-time and derived**, so replacing a definition requires no migration. `Shape` stays
the sorted set of field paths paired with value kinds, with a nested `Ref` as a shape element of
kind `ref` and never dereferenced. `Fingerprint` still dereferences, bounded at depth 4, marking
rather than silently degrading past the bound.

```text
UNCHANGED CONCERN    ../../02-content.md notes Cites scans INSIDE STRINGS for embedded <obs-N>,
                     forced by a real model return putting its reference in prose. That widens
                     what is derived and changes nothing about what is authorized, since the scope
                     check refuses any reference the emitter never held. Carried as-is.
```

---

## 6 · Reconciliation with the runtime

```text
graph.Observation        →  becomes observation.Observation. Producer widens to Provenance.
content.EngramRef        →  REMOVED. see §4.
content.Content.Reach()  →  REMOVED with it.
content.Value set        →  unchanged, and back to exactly interfaces.md's members.
content.Deref            →  unchanged. still exists so content does not import graph.
```
