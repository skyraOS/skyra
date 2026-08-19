# Causal Recurrence

**Status:** Working formalization

## Abstract

An accepted observation can carry a causal proof with the same recursive shape as the computation
that produced it. Each invocation consumes attributable observations under a specific artifact,
entity, identity, and authority context, then produces a new observation. The runtime—not the
worker—constructs the provenance around that computation.

The result is a recursively justified observation:

\[
C(o)=\left\langle
o,\;v,\;e,\;i,\;A,\;
[C(o_1),\ldots,C(o_n)]
\right\rangle
\]

Only the causal cone selected by outer acceptance becomes durable. Content-addressing these proof
nodes produces a Merkle-like DAG, enabling integrity verification, shared causal structure, and
selective disclosure without making proof material part of the worker's input.

## 1. The recurrence

Let an invocation \(q\) produce observation \(o\). Define:

- \(O(o)\): a canonical commitment to the observation;
- \(V(q)\): the exact artifact or position version interpreted;
- \(E(q)\): the interpreting entity;
- \(I(q)\): the invocation identity and causal position;
- \(K(q)\): a commitment to the relevant bounded execution context;
- \(D(q)\): the ordered, labeled observations admitted to the invocation by the substrate.

The causal recurrence is:

\[
C(o)=
\begin{cases}
H(\texttt{ingress/v1},O(o),S(o))
& \text{if }o\text{ entered from outside the recursive computation},\\[6pt]
H(\texttt{causal/v1},O(o),V(q),E(q),I(q),K(q),
[(\rho_j,C(x_j))]_{x_j\in D(q)})
& \text{if invocation }q\text{ produced }o.
\end{cases}
\]

Here \(S(o)\) is the substrate-held source provenance for an ingress observation, and \(\rho_j\)
records each dependency's role. Useful roles include:

```text
parameter          supplied as an invocation input
candidate          the observation proposing the selected behavior
cite               the visible observation prompting the selection
child-return       an observation returned from a child position
revision           evidence that caused a parent-owned contract or draft to change
```

The role labels preserve meaning that an unordered collection of hashes would lose.

## 2. Observable causality, not reported influence

The phrase "causal inputs actually used" requires care. A substrate cannot prove which inputs
altered a worker's hidden internal state or by how much. It can prove:

```text
which observations it placed in the worker's frame;
which candidate and child position the parent selected;
which entity and artifact revision ran;
which authority bounded the invocation;
which observation came back;
which parent-owned boundaries admitted that observation;
which returned results were remembered.
```

Therefore \(D(q)\) is constructed by the substrate from observable exposure and causal operations.
It is never supplied by the worker as an explanation after the fact.

This makes the causal claim precise:

> This invocation, under these recorded and bounded conditions, received these attributable
> observations and produced this observation.

Whether an input was counterfactually necessary is a separate experimental question. Because the
invocation conditions are recorded, necessity can be tested later through replay and ablation.

## 3. Blind action

The worker does not need recursive proofs in order to act. Its semantic input can remain:

\[
(\text{artifact revision},\text{observations},\text{expectation},\text{Horizon})
\]

The runtime carries invocation identity, provenance, proof references, and causal position around
that input. The worker consumes values; the causal substrate constructs provenance.

This separation has two consequences:

1. Proof construction does not increase the worker's semantic authority or context.
2. A worker cannot manufacture its own causal standing by writing provenance labels into content.

## 4. Recursive acceptance

When a parent admits a child's return and advances its continuation, the same recurrence applies at
the next level:

\[
C(o_{parent})=
F_{parent}(C(o_{local,1}),\ldots,C(o_{local,m}),C(o_{child}))
\]

The recurrence continues toward the causal origin until the outer perspective remembers a result.
The proof is thus the recursively attributable shadow of the continuation tree.

In practice the structure is a DAG rather than a strict tree: several invocations may share an
input observation, concurrent branches may join, and one observation may participate in several
later computations.

## 5. Occurrence and remembrance

Proof follows witnessed occurrence, while durable causal memory follows acceptance. These must not
be collapsed.

Let:

\[
G_{occurred}=\text{all substrate-witnessed causal events in the live branch}
\]

For the outermost accepted observation \(o^*\), define:

\[
G_{remembered}(o^*)=
\operatorname{BackwardReachable}
(G_{occurred},o^*,\text{Remember edges})
\]

Then:

\[
G_{remembered}(o^*)\subseteq G_{occurred}
\]

This is the formal form of two laws:

```text
Something may occur without being remembered.
Nothing may be remembered without having occurred.
```

If a chain \(A\rightarrow B\rightarrow C\) executes but the relevant return is rejected, that work
may have transient audit evidence and real cost, but it is not recursively incorporated into the
proof of an unrelated accepted result.

The durable closure commitment for branch \(b\) can therefore be:

\[
R_b=H(\texttt{closure/v1},\text{namespace},\text{span},C(o^*))
\]

The root commits to exactly the accepted causal cone.

## 6. Merkle-DAG representation

The recursive proof need not be transmitted as a giant nested payload. Each proof node can be
content-addressed:

\[
P(o)=H(O(o),V(q),E(q),I(q),K(q),[P(x_1),\ldots,P(x_n)])
\]

The result is a Merkle-like causal DAG:

```text
accepted root
├── invocation commitment
│   ├── artifact/version commitment
│   ├── authority-context commitment
│   ├── input proof reference
│   └── child-return proof reference
└── returned observation commitment
```

Sharing causality means sharing a root and the reachable proof nodes the recipient is authorized to
inspect. Common subgraphs are transmitted once and referenced thereafter.

Domain-separated encodings such as `ingress/v1`, `causal/v1`, and `closure/v1` prevent structurally
different objects from accidentally sharing the same hash interpretation. Canonical serialization
is required: the same semantic node must produce the same bytes, independent of map order,
platform, or implementation language.

## 7. Selective disclosure

A holder may disclose only part of a causal proof:

```text
P(result)
├── P(a)        disclosed
├── P(b)        disclosed
└── P(c)        commitment disclosed; observation withheld
```

The recipient can verify that \(P(c)\) occupied the committed causal position without receiving
the underlying observation. Access to the observation remains a separate authorization decision.

Plain hashes may not adequately hide low-entropy private values. A concealed observation should
normally use a hiding commitment, for example:

\[
O(o)=H(\text{canonical observation}\;\|\;\text{random nonce})
\]

The nonce is disclosed only when the holder chooses to open the commitment. More advanced proof
systems could later prove properties of a concealed observation without revealing it, but they are
not required by the recurrence itself.

## 8. What a Merkle proof establishes

A Merkle proof establishes integrity and membership relative to a root:

> Given this root, these disclosed nodes were committed in this exact structural relationship.

It does not by itself prove:

```text
that the physical invocation happened;
that the runtime honestly constructed the node;
that unfavorable history was not omitted;
that the disclosed root represents all history ever produced.
```

The first two claims depend on trusting or attesting the causal substrate. Roots should therefore
be signed by a runtime or namespace identity. The latter two require a non-equivocation mechanism,
such as an externally witnessed append-only transparency log.

## 9. Optional ledger anchoring

The causal structure is naturally a Merkle DAG. It does not inherently require a blockchain.

```text
Causal Merkle DAG       proves structural derivation and supports selective disclosure
Signed closure root     identifies which substrate or namespace witnessed the commitment
Transparency log        makes rewriting and equivocation detectable over time
Blockchain              optionally supplies shared ordering and finality for published roots
```

A blockchain, if used, should witness commitments to causal computation rather than pretend to be
the computation. It introduces consensus, global ordering, validator policy, replication, and
finality—none of which is necessary for local causal derivation.

A distributed deployment can instead maintain many privately owned causal DAGs. Their owners may
exchange signed roots or subproofs and anchor selected roots into a common log only when they need
stronger shared evidence.

## 10. Verifiable operational history

The causal record can be queried across several independent coordinates:

\[
\text{Entity}\times\text{ArtifactVersion}\times
\text{ResolvedEndpoint}\times\text{Runtime/Host}
\]

These coordinates must remain distinct. An endpoint is volatile and is not an entity's identity.
Recording which endpoint was selected for a particular occurrence is valid; treating endpoint
history as entity identity is not.

An accepted causal DAG can support verifiable claims about:

```text
invocations that contributed to retained results;
accepted returns;
artifact revisions in accepted cones;
entities and endpoints participating in those cones;
causal distance through successive remembrance.
```

It cannot alone support claims about total attempts, total Reach failures, rejected returns, or
total incurred cost. Those events may be absent from accepted memory by construction.

Supporting broader statements such as:

> This endpoint participated in 18,421 invocations, including 518 failures and 17,903 retained
> returns.

requires a separate, durable commitment to occurrence or cost records. That stream must remain
epistemically distinct from accepted causal memory:

```text
ACCEPTED CAUSAL GRAPH     what survived outer remembrance
OCCURRENCE/AUDIT LOG      what the substrate witnessed running
COST ACCOUNT              what the namespace paid, including discarded exploration
```

Whether to retain the second and third streams is an architectural decision. Adding them changes
the privacy, abandonment, and crash semantics of a system in which rejected execution would
otherwise evaporate.

## 11. Summary

The primitive recursive form is:

\[
C(o)=\left\langle
o,\text{artifact version},\text{entity},\text{invocation},
\text{bounded context},[C(o_1),\ldots,C(o_n)]
\right\rangle
\]

with three qualifications:

1. The dependencies are substrate-witnessed inputs and causal operations, not worker-reported
   influence.
2. Durable proof is the backward-reachable cone selected by outer remembrance, not every event
   that occurred.
3. Content addressing proves integrity relative to a root; signatures and append-only external
   witnesses are needed for authorship, history, and omission resistance.

The resulting property is the central one:

> The proof has the same recursive shape as the computation. Provenance is not a separate language
> bolted onto execution; it is the content-addressed, recursively attributable shadow of accepted
> execution.

