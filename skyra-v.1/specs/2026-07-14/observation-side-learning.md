# Understanding = Parsing — the Observation-Side Learning Loop

**Status:** MECHANISM (2026-07-15 research). High-level only — no implementation. The interpretation
axis (`observation → state change`, `P(ΔS | O)`) finally given a learning loop. **Parked here as a
foundation** while focus moves to the action loop.
**Companions:** [the-learning-loop.md](the-learning-loop.md) ·
[region-as-playbook.md](region-as-playbook.md) ·
[closure-duration-objective.md](closure-duration-objective.md) ·
[principle-falsifiable-boundary.md](principle-falsifiable-boundary.md)

For days we had a settle for the *action* side (does the action close the error and keep it closed)
and only a hand-wave for the *observation* side (how does an interpretation get graded when there is
no local ground truth for "the right understanding"). This is the observation side's loop, and it
turns out to be the **same** loop.

## Understanding is parsing into the grounded floor

The internal representation is fixed: the **grounded relational floor** — the storage boundary only
accepts relational structure, and its primitives are world-perceivable distinctions. Every external
language — natural language, JSON, a device protocol, one person's idioms — is met with a **parser**
that maps it into that floor.

> **Understanding = a successful parse into the grounded floor.**

Because the target is *grounded*, parsing is comprehension, not mere syntax. (Into an arbitrary
target it would be a category error; into a world-grounded one it is exactly understanding.) Parsers
go all the way down and terminate at the floor — the one thing not parsed, the bedrock.

## Parsers are learned, because speaking has no rules

Grammar has rules; **speaking does not.** So a parser cannot be rule-checked into correctness — it
has to be *learned from use*, per-language **and** per-person. A parser is born effortful (learning
Spanish, learning a new person) and migrates to transparent (you don't parse JSON, you read it) —
the same reflex-is-exhaust migration as everywhere else. Learning a language and getting fluent in a
specific person are the same mechanism; the human border stops being a fixed adapter and becomes a
relationship you get fluent in.

## The grade is downstream stability, never validity

"Successful parse" cannot mean "well-formed" — you can emit fluent, perfectly well-formed relational
structure that is *wrong* (grammatical Spanish that means the opposite). Because speaking has no
rulebook, syntactic validity is worthless as a signal. The only test is:

> **Did this understanding produce stable behavior?** (errors close and stay closed; predictions hold)

That is the closure / prediction-diff grade already defined for the action side. The parser **rides
the existing grade** — it needs no separate one. Understanding is confirmed not at parse time but
when what was parsed earns its keep.

## The learning loop

```text
observation → parse (understanding) → state change → action → behavior
                   ▲                                              │
                   └────────── unstable? destroy + error trace ───┘
```

- The parser emits an understanding (a relational structure).
- If the behavior it drives is **unstable**, the understanding is **destroyed** and linked back to
  the parser that made it, carrying an **error trace** of the instability — *"this structure: nothing
  downstream could do anything correct with the information shaped this way."*
- The parser learns to structure inputs so they produce stable behavior. This is the **same loop as
  the playbook**: emit a structure, grade it by downstream stability, keep what works, destroy-and-
  trace what doesn't.

## Destroy is safe — the observation is immutable

You never destroy the observation, **only the parse**. The raw thing that crossed the boundary is
append-only (LAW 1); the understanding is a *re-derivable fold* over it. So destroying an
understanding loses nothing — the observation stays, and a better, more fluent parser can **re-parse
it later**. Ground truth is permanent; interpretations are disposable. A bad understanding is not a
loss, it is a fold you will redo when you are more fluent.

## Same mechanism as the action side — the pair is complete

The parser learns exactly like the playbook. There are not two learning systems, one for
understanding and one for acting — there is **one loop, applied at two points** (the parse and the
action) and graded end-to-end. That is the settle for the interpretation axis that had no local
ground truth: it settles by *downstream stability with co-variation blame* — not by a rulebook, not
by a judge. Both axes of learning now stand on the same footing.

## Open — nail these next

- **Blame localization (the one to get right first).** Unstable behavior is a **global** alarm over
  `observation → parse → state change → action → world`. Attributing it needs co-variation: the
  parser is indicted only when *its* structures fail **systematically**, across different playbooks
  and contexts. One failure indicts nothing; a pattern under a parser indicts the parser.
- **The three-way split.** When behavior is unstable, what decides **re-parse** (indict the
  understanding) vs **re-plan** (indict the playbook) vs **indict nothing** (the world is just
  hard)? Get this right and the loop is real; get it wrong and it either thrashes or never learns.
- **Parser exploration.** The error trace says "this failed"; improving needs *trying a different
  structure* for similar inputs — the generate case. Unspecified.
- **Floor extension.** A genuinely new concept the relational floor cannot express requires **growing
  a primitive** — language-learning becomes concept-learning. The vocabulary-extension frontier.
