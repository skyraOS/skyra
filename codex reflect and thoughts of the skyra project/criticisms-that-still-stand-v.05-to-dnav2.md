amend: ⟦LAW≈CONVENTION⟧
add: ⟦ONTOLOGY-FRAGMENTATION⟧
Delete: ⟦MEMORY≠SELF⟧

# Criticisms That Still Stand from v.05 to dnav2

## Standard of criticism

The governing criticism remains:

> The nouns are ahead of the evidence.

This must be applied in both directions.

The builder cannot use ambitious nouns as proof that the system possesses the
properties those nouns imply. The critic cannot translate every unfamiliar
abstraction back into conventional software vocabulary and treat that
translation as proof that nothing new is being attempted.

The relevant question is not whether Skyra already deserves every noun, nor
whether each mechanism can be redescribed using familiar engineering terms.
The question is whether the implementation produces observable evidence for
the verbs implied by those nouns.

Can it learn? Can it develop? Can it preserve identity? Can it revise a belief
because of consequences? Can it form and maintain relationships? Can it act
more competently because of its particular history?

Across v.05 and dnav2, the following criticisms still stand.

## 1. ⟦Δ≠DEV⟧ State change has not yet established development

Both systems accumulate state.

v.05 accumulates exchanges, memories, relationships, threads, tasks, and
thought history. dnav2 accumulates episodes, frames, causal nodes,
predictions, actions, grades, activation records, and query-budget history.

But accumulating state is not sufficient evidence of development. Development
requires a durable, consequential change in future behavior or judgment. The
system must demonstrate that a later action is better or meaningfully different
because of a specific earlier experience.

dnav2 makes this criticism less severe by adding explicit feedback and
weighted structures. It does not eliminate it. Numeric movement inside a graph
is evidence that an update occurred, not that the being developed.

The missing demonstration remains longitudinal:

1. Establish the being's behavior before an experience.
2. Expose it to a consequential event.
3. Identify the internal state changes caused by that event.
4. Show a later behavioral difference caused by those changes.
5. Compare the result against a stateless or history-only baseline.
6. Show that the difference is appropriate rather than merely different.

Until then, development is a target, not a demonstrated property.

## 2. ⟦LLM→MEANING⟧ The LLM remains the principal source of meaning

Both architectures organize LLM calls rather than replace their semantic
authority.

In v.05, LLMs generate private thought, public action, routing decisions, memory
operations, and relationship-facing language.

In dnav2, LLMs decompose observations, construct frames, invent causal
hypotheses, predict futures, propose actions, judge causal agreement, judge
prediction matches, judge execution, and grade coherence.

dnav2 gives these calls narrower contracts and better transactional handling.
That is an engineering improvement. But much of the system's apparent causal
knowledge is still model-generated content evaluated by other model calls.

The machinery can ensure that a judgment is well-formed. It cannot, by itself,
ensure that the judgment is grounded, independent, or true.

This is especially important when one LLM-generated claim becomes evidence for
another. A graph of mutually consistent assertions can become increasingly
structured without becoming increasingly accurate.

## 3. ⟦FEEDBACK↺⟧ Feedback is not yet sufficiently independent

Both versions lack a strong, general source of external ground truth.

v.05 has little explicit evaluation beyond whether routing and protocol
operations complete. dnav2 improves this by comparing predictions with later
messages and actions with environmental returns. However, the comparison is
frequently interpreted by an LLM judge.

That produces feedback, but not necessarily independent feedback. The same
class of model may generate the hypothesis, interpret the observation, and
grade the relationship between them.

The strongest future evidence would come from consequences that do not depend
on model interpretation:

- Did a command succeed?
- Did the object reach the requested location?
- Did the prediction identify a measurable future event?
- Did the user later correct the same belief again?
- Did task performance improve on a held-out recurrence?
- Did the system avoid a previously observed failure without being reminded?

LLM judgment may remain useful, but it should not be the only court hearing the
case.

## 4. ⟦NOUN>EVIDENCE⟧ The metaphors still imply more than the mechanisms provide

v.05 uses *universe*, *physics*, *being*, *relationship*, and *mortality*. DNA
v2 uses *DNA*, *genes*, *transcription*, *membrane*, *nucleus*, and *causes*.

These metaphors are productive because they suggest architectural properties.
They are also dangerous because readers naturally import properties that have
not been implemented.

A `Reality` does not establish a universe. A routing constraint does not
establish physics. Persistent persona data does not establish a being. A Go
function registered at startup does not establish a gene in the evolutionary
sense. A model-generated causal sentence does not establish causality.

The criticism is not that these names must be abandoned. It is that every name
creates an evidentiary debt. The project should keep a precise ledger:

- Which properties of the metaphor are implemented?
- Which are design intentions?
- Which are explicitly not claimed?
- What observation would justify promoting the term from metaphor to technical
  description?

## 5. ⟦LAW≈CONVENTION⟧ Substrate laws are stronger; semantic policy remains conventional

This criticism needs narrowing. v.05 leaves claimed laws in prompts or inactive
conventions. dnav2's atomic joins, one-reframe rail, bounded correction,
transactional commits, directed edges, typed workflow conditions, and fresh
contextual causal edges survive model confusion; at that level they are laws.

Semantic policy remains conventional. The model still chooses frames, causal
hypotheses, and grades. The substrate can enforce form without proving that a
valid interpretation is right. Label each claim as enforced invariant, model
policy, or empirical hypothesis, and test the first independently of the model.

## 6. ⟦PERSIST≠CONTINUITY⟧ Persistence is not yet equivalent to continuity

Both projects treat retained state as central, but continuity has several
levels:

- survival across turns;
- survival across threads;
- survival across process restarts;
- survival across model changes;
- survival across schema and runtime upgrades;
- preservation of provenance and meaning after compaction;
- selective forgetting without destruction of identity.

v.05 persists some memories and skills to the filesystem, while much runtime
state remains process-local. dnav2 repeatedly calls its indexes durable, but
within the inspected package the causal and activation stores are primarily
in-memory structures; explicit disk durability is most visible in run tracing.

Neither version yet demonstrates that a being survives long periods of real
operation while retaining a coherent, migratable identity.

## 7. ⟦ONTOLOGY-FRAGMENTATION⟧ Referential identity can split accumulated learning

dnav2 relies on node reuse, yet resolution is mostly exact-match and term-to-
entity merging is deferred. One referent observed or asserted under different
labels can remain multiple nodes unless the model supplies the right stable ID.

Edges, evidence, grades, indexes, and causal lineage then accumulate on separate
fragments, so learning through one identity may not transfer through another.
The system needs aliases, merge/split operations, provenance-preserving
migration, contradiction handling, and tests that learning survives correction.

## 8. ⟦COMPLEXITY>PROOF⟧ Complexity is growing faster than demonstrated capability

The conceptual surface expands across both versions:

- worlds and nested realities;
- beings and relationships;
- threads and exchanges;
- private and public cognition;
- memory and skills;
- economics and governance;
- episodes and frames;
- causal nodes and aitia sets;
- activation, alignment, salience, familiarity, and closure;
- adaptive query budgets and causal lineage.

Each addition may be locally coherent. The risk is that internal consistency
becomes the project's primary evidence.

A smaller system with one demonstrated developmental effect would currently
support the thesis better than a larger system with many plausible but
unvalidated mechanisms.

The project needs periodic compression: identify which mechanisms have earned
their place through behavioral evidence, which remain necessary experiments,
and which are conceptual ornament.

## 9. ⟦CLAIM¬FALSIFIED⟧ The abstractions remain difficult to falsify at the product level

The code contains many testable local invariants. That is good. Tests can show
that a join fires once, a graph preserves edge direction, a malformed model
response does not commit, or a grade changes a budget.

Those tests do not establish the central product claims.

The system still needs end-to-end criteria capable of proving the architecture
wrong. Examples include:

- A world-aware being should outperform an identical stateless model on a
  recurring relationship-dependent task.
- Causal learning should improve prediction calibration across related
  episodes.
- Reframing should reduce repeated failure rather than merely produce a new
  description.
- A learned causal path should predict behavior better than retrieval of the
  raw exchanges from which it was derived.
- Removing the proposed physics should measurably degrade continuity or task
  performance.

Without comparisons of this kind, successful local mechanisms can coexist with
an unsuccessful overall theory.

## 10. ⟦EMERGENCE?⟧ Emergence remains asserted before it is observed

Both versions hope that behavior will emerge from structure rather than be
directly orchestrated.

Yet v.05 explicitly routes relations through authored layers, and dnav2 uses a
static authored workflow whose genes contain detailed prompts and transition
conditions. The systems may eventually enable emergent behavior, but their
current behavior is substantially designed and orchestrated.

Emergence should be reserved for stable behavior that was not directly encoded
and that appears across varied conditions because of lower-level interactions.
Unexpected model output alone is not emergence. Complexity alone is not
emergence. An evocative explanation after the fact is not emergence.

## 11. ⟦PRODUCT≠SCIENCE⟧ The boundary between product language and scientific language remains unclear

Skyra sometimes speaks as a product, sometimes as a philosophy, sometimes as a
cognitive architecture, and sometimes in language adjacent to neuroscience or
biology.

These modes have different standards of evidence.

“A world that feels alive” can be a product ambition. “A being develops” is a
behavioral claim. “This is a nervous system” or “this learns causality” sounds
like a scientific claim. Moving among these registers without marking the
boundary makes both building and criticism less precise.

The project would benefit from labeling claims explicitly:

- metaphor;
- product experience goal;
- implemented mechanism;
- observed behavior;
- tested causal claim;
- future hypothesis.

## 12. ⟦PROTO≠PROD⟧ Production readiness remains distant

Neither architecture is yet evidence of a reliable, secure, maintainable
product system.

The inspected implementations expose unresolved concerns around error
propagation, external authentication, state durability, migrations, model
failure, operational recovery, termination, calibration, and long-running
behavior. dnav2 improves internal tracing and concurrency discipline, but it
is a package dependent on a surrounding substrate rather than a standalone
validated system.

This criticism should not be confused with a rejection of the research. A
research architecture can be valuable before it is production-ready. The fault
would be presenting prototype evidence as operational maturity.

## What has improved without resolving the criticism

dnav2 does not merely rename v.05. It materially reduces several weaknesses:

- Development has a proposed mechanical loop.
- Predictions can be closed against later observations.
- Invalid model output is handled more transactionally.
- Provenance distinguishes observation from assertion.
- Causal structures and evidence links are inspectable.
- Concurrency and joins have explicit invariants.
- Internal conceptual contradictions are documented rather than hidden.
- The test surface is much broader and closer to the architecture's claims.

These improvements matter. A criticism can remain valid while becoming less
damning. Saying that dnav2 has not demonstrated development is not the same as
saying it has made no progress toward development.

## The criticism that should govern the next phase

The project no longer most urgently needs additional nouns. It needs decisive
verbs with measurements attached.

Choose a small number of claims:

- learns from consequence;
- maintains a relationship-specific identity;
- improves prediction;
- revises a mistaken causal belief;
- carries an intention across time;
- avoids repeating a learned failure.

For each claim, specify:

1. The observable behavior.
2. The relevant internal state transition.
3. The baseline architecture without the proposed mechanism.
4. The conditions under which the claim should fail.
5. The evidence required before using the corresponding noun literally.

The enduring criticism is therefore not that Skyra's language is too ambitious.
Ambitious language can guide invention.

The criticism is that the project must continuously close the distance between
its language and its behavior—and that critics must measure that distance
rather than pretending either endpoint is the whole truth.

## Line count inspection

before: 329
after: 328
