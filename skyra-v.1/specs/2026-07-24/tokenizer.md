# Tokenizer — the Stable Sensory Alphabet

**Date:** 2026-07-23  
**Status:** DECIDED at the representation boundary. The exact production codec and
storage encoding remain implementation choices.

**Reads on:** [v4-data-model](v4-data-model.md) for the observation/belief/action
spine, [traversal-spine](../2026-07-22/traversal-spine.md) for the distinction between an
authoritative event and the derived view selected from it, and
[belief-expression-inference](belief-expression-inference.md) for the call that
owns contextual entity resolution.

---

## 1. Decision

Unstructured textual content is stored as an ordered sequence of token IDs under
one fixed, reversible, versioned tokenizer.

The separately stored raw string is removed. It is a derived view:

```text
Decode(codec, Encode(codec, input)) == input
```

The tokenizer is Skyra's stable sensory alphabet. Human input may be misspelled,
run together, inconsistently capitalized, or irregularly punctuated; the input is
not corrected or normalized at the observation boundary. The fixed codec gives
that exact signal a deterministic identity while preserving it losslessly.

```text
human signal
    ↓ fixed reversible codec
ordered token IDs
    ↓ interpretation / compression
beliefs
    ↓ expression
actions
    ↓ world return
ordered token IDs
```

Tokens are the discrete material substrate. A segmented token field is the
distributed representation; a concrete Observation, Belief, or Action is the
bounded whole assembled from that substrate. Model hidden state is ephemeral and
is not the durable meaning of "latent" in this design.

---

## 2. What opinion the tokenizer introduces

The tokenizer is not neutral. Its trained vocabulary and merge rules decide which
recurring byte sequences receive individual token identities.

The opinion resides in tokenizer training, not in runtime tokenization:

```text
TRAINING      chooses the vocabulary and merge structure once
RUNTIME       applies that frozen structure deterministically
```

That prior is acceptable because:

1. tokenization is reversible;
2. the original signal is not normalized or discarded;
3. the codec is fixed and explicitly identified;
4. meaning is not assigned by token boundaries;
5. similarity can confer attention but cannot settle a claim.

The tokenizer shapes retrieval geometry. It does not rewrite evidence.

---

## 3. Representation

Candidate contract:

```go
type CodecID string
type TokenID uint32

type TokenSequence struct {
    Codec  CodecID
    Tokens []TokenID
}
```

The sequence is ordered. Repeated tokens remain repeated. No delimiter is inserted
when decoding; each token contributes its original bytes directly.

For example:

```text
["hell", "oh", "imy", "ism", "ike"]
```

reconstructs:

```text
"hell" + "oh" + "imy" + "ism" + "ike"
= "hellohimyismike"
```

Spaces, when present, belong to token bytes:

```text
["hello", " my", " name", " is", " mike"]
```

reconstructs exactly:

```text
hello my name is mike
```

### Placement in the v4 structs

There must be one stored token sequence for one witnessed textual payload. The v4
port must choose its exact field placement between `PerceivedAct` and
`Observation`; it must not store the sequence on both.

The ownership test is:

> Which record owns the exact content that crossed the world/being boundary?

Whichever record answers that question owns the `TokenSequence`. The other reaches
it through its existing ID reference.

Typed mechanical data stays typed:

```text
IDs · addresses · primitive · source · time · coordinates · truth
```

Unstructured content is tokenized:

```text
utterances · messages · file bodies · command output · action payloads
```

Tokenization does not turn exact world addresses or control fields into token soup.

### The concrete/segmented cycle

The cognitive loop alternates between concrete wholes and segmented token
fields:

```text
ConcreteObservation
    ↓ deterministic tokenization / perturbation
SegmentedObservation
    ↓ belief inference / collapse
ConcreteBeliefs
    ↓ deterministic tokenization / perturbation
SegmentedBeliefs
    ↓ expression inference / collapse
ConcreteActions
    ↓ deterministic tokenization / perturbation
SegmentedActions
    ↓ execution, world return, and perception / collapse
ConcreteObservations
```

This is the complete representation cycle. It contains three field
transformations:

```text
SegmentedObservations → ConcreteBeliefs
SegmentedBeliefs      → ConcreteActions
SegmentedActions      → ConcreteObservations
```

and three deterministic decompositions that expose each concrete result to the
next field:

```text
ConcreteObservation → SegmentedObservation
ConcreteBelief      → SegmentedBelief
ConcreteAction      → SegmentedAction
```

A concrete form owns identity, boundary, order, provenance, and its complete
`TokenSequence`. A segmented form exposes individually addressable token
occurrences while retaining the concrete parent and token position:

```go
type SegmentedToken struct {
    ParentID string
    Position int
    Token    TokenID
}
```

Conceptually:

```text
ConcreteObservation obs-17
    body: [24912, 922, 1308]

SegmentedObservation field
    (obs-17, 0) → 24912
    (obs-17, 1) → 922
    (obs-17, 2) → 1308
```

The segmented form does not replace or fragment the authoritative concrete
record. It is the field-addressable form of that record. Decoding the parent's
ordered `TokenSequence` reconstructs its exact unsegmented textual content.

An inference payload containing several concrete records is a third thing: a
temporary bounded assembly selected from their segmented fields. It may be kept
in the payload log for provenance, but it is not itself the authoritative
Observation, Belief, or Action.

The governing distinction is:

```text
CONCRETE     bounded, ordered, identifiable whole
SEGMENTED    distributed, individually addressable token material
PAYLOAD      temporary selection assembled to cross a field
```

Within one wave, [Perturb](perturb-observe-traversal.md) traverses the segmented
Observation–Belief–Action field forward. `Observe` then unwinds the same stack
through concrete Action–Belief–Observation counterparts. No second concrete
traversal is required. Collapse is a representational boundary and attention
decision, not a declaration of truth.

---

## 4. Evidence addresses

Beliefs cite exact token ranges in exact observations:

```go
type TokenSpan struct {
    ObservationID ObservationID
    Start         int // inclusive token position
    End           int // exclusive token position
}

type Belief struct {
    ID       BeliefID
    Triples  []Triple
    Evidence []TokenSpan
}
```

Competing beliefs may cite the same span. One belief may cite several spans.
Token boundaries are addressing units, not semantic boundaries.

```text
TOKENIZATION             mechanical encoding segmentation
OBSERVATION SEGMENTATION one witnessed event or moment
BELIEF SEGMENTATION      learned interpretation over token spans
```

These three segmentations must not be collapsed.

---

## 5. Indexing and similarity

The first exact reverse index is:

```text
TokenID → occurrences → beliefs
```

An occurrence includes at least the owning observation and position:

```go
type TokenOccurrence struct {
    ObservationID ObservationID
    Position      int
}
```

Exact token identity is not semantic identity. Missing spaces and typos normally
produce different token sequences:

```text
hello my name is mike
hellomynameismike
hello my nmae is mike
```

Derived similarity indexes may retrieve related beliefs:

```text
token overlap
character n-grams
normalized routing forms
later learned relations
```

They are attention mechanisms only.

> Approximate similarity may decide what gets considered. It may not decide what
> an observation meant, and it may never settle a prediction.

### Entity resolution belongs to belief inference

**Status: DEFERRED.** The tokenizer and first observation field do not implement
an entity-resolution index, alias registry, candidate table, or durable
`token → entity` link. This section preserves the eventual ownership boundary so
the first implementation does not put resolution into the observation field for
convenience.

Token matching detects possible mentions. It does not establish identity.

The exact pass gathers candidates and evidence spans:

```text
observation token span "Mike"
    ├── candidate entity:mike-1
    └── candidate entity:mike-2
```

Spacing, casing, spelling variation, shared names, pronouns, and contextual
references are resolved during the observation-time belief inference call. That
call receives:

- the observation's ordered token sequence;
- the exact token spans that triggered entity candidates;
- the candidate entity IDs and their known aliases;
- the current entity/predicate vocabulary and desire-blind observation context.

It must not receive the target, open expectations, their confidences, or the
answer wanted by planning. Entity resolution is part of the belief-reading pass
and therefore remains under the same blindness boundary as every other
interpretation.

The call emits beliefs whose entity references cite the token spans that licensed
them:

```go
type EntityReference struct {
    EntityID EntityID
    Evidence TokenSpan
}
```

When context distinguishes the referent, the inferred belief may name one entity.
When it does not, the output preserves multiple competing beliefs or entity
candidates with their own confidence. The inference call is allowed to resolve
the ambiguity; it is not required to manufacture certainty.

```text
TOKEN MATCHING       possible mentions and candidate entities
BELIEF INFERENCE     which candidate this occurrence appears to denote
SETTLEMENT           whether the inferred belief later earned support
```

This inference occurs at observation time, never at grading time. A selected
entity ID is an interpreted belief, not a raw fact merely because an alias
matched.

### V0 boundary while entity resolution is deferred

The first observation layer remains:

```text
ObservationID → ordered TokenSequence
TokenID       → TokenOccurrences
```

Heated token spans may enter belief inference without an intermediate entity
resolution object. If the model emits a belief containing a referent, that
referent is provisional belief content tied to its evidence span; it does not
create a reusable token/entity hard-link.

Exact world addresses remain typed and require no natural-language entity
resolution.

The guard for V0 is:

> Do not add `token → entity` to the observation or excitation field merely to
> simplify the first implementation.

---

## 6. Compression

The representation forms a compression hierarchy:

```text
bytes
  ↓ tokenizer vocabulary
token IDs
  ↓ interpretation
beliefs
  ↓ expression
actions
```

Tokenizer compression replaces recurring byte fragments with vocabulary IDs.
Belief learning performs the more important compression: recurring token sequences
become compact, falsifiable structure while retaining exact evidence spans.

Tokenization does not guarantee smaller physical storage. A `uint32` token array
may be larger than a short UTF-8 string. Packed integers, interning, or ordinary
lossless compression are separate storage decisions.

The representation decision is about stable identity and one cognitive path, not
about claiming a disk-size win.

---

## 7. Codec law

Once token IDs are durable, the tokenizer is part of the storage schema.

The `CodecID` must identify:

- algorithm;
- vocabulary;
- merge rules;
- special-token policy;
- normalization policy;
- version;
- content hash of the complete codec definition.

Required invariants:

1. decoding reconstructs the exact original byte sequence;
2. the codec is immutable after durable observations use it;
3. every token sequence names its codec;
4. old observations are never silently retokenized;
5. changing inference providers does not change the durable codec;
6. a future codec creates a new version or derived index;
7. token positions remain stable for the lifetime of their codec.

A model provider's tokenizer may be selected as the initial codec, but it becomes
Skyra-owned once pinned. Provider updates do not silently enter the store.

---

## 8. Measured probe

A disposable local experiment used `tiktoken` 0.13.0 with `o200k_base`. This is
evidence about feasibility, not the production-codec decision.

```text
INPUT   "hello my name is mike"
IDS     [24912, 922, 1308, 382, 150980]
PIECES  ["hello", " my", " name", " is", " mike"]

INPUT   "hellomynameismike"
IDS     [10844, 35927, 897, 2367, 970]
PIECES  ["hell", "omy", "name", "ism", "ike"]

INPUT   "hellohimyismike"
IDS     [10844, 2308, 87947, 2367, 970]
PIECES  ["hell", "oh", "imy", "ism", "ike"]

INPUT   "hello my nmae is mike"
IDS     [24912, 922, 297, 809, 68, 382, 150980]
PIECES  ["hello", " my", " n", "ma", "e", " is", " mike"]

INPUT   "Hey,  grab MY water!"
IDS     [25216, 11, 220, 18013, 30096, 3411, 0]
PIECES  ["Hey", ",", " ", " grab", " MY", " water", "!"]
```

Every sample round-tripped exactly.

Measured on the development machine:

```text
initial vocabulary download/load      ~2066 ms
new process, cached vocabulary load    ~202 ms
first encode after initialization        ~7.7 ms
steady-state short-input encode           2.2 µs
```

The steady-state result is the mean of 100,000 encodes of
`"hello my name is mike"`. A persistent personal-AI process pays codec
initialization once; steady-state tokenization is not expected to be a material
part of inference latency.

---

## 9. What remains open

1. Which reversible byte-level codec is the production V0 codec.
2. Whether the codec vocabulary is embedded in the binary or loaded from a
   content-addressed local artifact.
3. The packed on-disk representation of `[]TokenID`.
4. The exact v4 owner of `TokenSequence`: `PerceivedAct` or `Observation`.
5. Which approximate routing indexes are justified after the exact-token path is
   measured.
6. Whether Skyra eventually learns a successor codec from its own accumulated
   observations. Such a codec would be a new version, never an in-place rewrite.
7. Entity resolution: candidate representation, alias ownership, ambiguity
   persistence, and the location of any derived resolution index. Explicitly
   deferred until a concrete run requires it.

---

## 10. One-paragraph statement

Skyra stores unstructured observations as ordered token IDs under one fixed,
reversible, versioned codec. The token sequence is the authoritative textual
representation; raw language is reconstructed on demand and is not stored a
second time. Tokenizer training contributes a frozen prior about recurring byte
patterns, while runtime tokenization is deterministic and lossless. Tokens
provide stable sensory identity and evidence addresses; beliefs provide meaning,
and only independently settled observations provide truth.
