# Open Protocol, Proprietary Behavior

**Date:** 2026-08-13
**Status: WORKING NOTE.** The protocol is open and free to implement. Behavior compiled into an
engram may remain proprietary. This document states what must remain observable when the behavior
itself is opaque.

> **The behavior may be secret. Its identity, authority, inputs, outputs, and exits cannot be.**

---

## The boundary

Skyra does not need to own a principal's memory in order to coordinate useful work. A principal's
namespace, observations, history, and accepted causal records remain under that principal's
control. The protocol that gives those records meaning is open and portable.

The behavior compiled into an engram is a different object. It may contain authored workflows,
domain knowledge, selection strategies, translations, optimizations, or other proprietary work.
The protocol does not require that implementation to be disclosed.

```text
OPEN AND PORTABLE                    MAY REMAIN PROPRIETARY
────────────────────────────────────────────────────────────────────────────
Decide / Remember semantics          compiled engram behavior
causal-envelope format               compiler and optimizer
accepted-cone representation         authored workflow implementation
grant and delegation semantics       domain-specific behavior libraries
entity and reach contracts           private ranking and translation logic
effect disposition semantics         implementation techniques
conformance fixtures
```

Openness applies to the seam. Secrecy may apply to what runs behind it.

An implementation must be replaceable without changing the meaning of a principal's durable
record. A proprietary behavior must be replaceable without pretending it was the same behavior.

---

## An opaque engram still has a public identity

The compiled body need not be readable to be identifiable. A proprietary engram is distributed or
reached as a content-addressed, signed artifact whose public envelope contains enough information
to authorize and verify its execution without revealing its implementation.

Conceptually:

```text
EngramArtifact {
    artifact_hash
    signer
    format_version
    compiler_identity
    entity_identity
    required_grants
    permitted_endpoints
    input_shape
    output_shape
}
```

This is not a declaration that the behavior is good, correct, or capable. Those remain matters of
observed history. It is an execution contract:

```text
ARTIFACT HASH        which exact compiled behavior was authorized and run.
SIGNER               who published that artifact.
FORMAT / COMPILER    how its identity and executable form are interpreted.
ENTITY               who interpreted it.
GRANTS               which world effects it may request.
ENDPOINTS             through which mediated exits it may communicate.
SHAPES                what the crossing can accept and return, not whether the result is good.
```

Signing prevents silent substitution. Content addressing makes two different compiled behaviors
two different artifacts even when both carry the same human-facing name.

---

## Three proofs

The privacy and accountability claim is not one proof. It is three separate proofs with different
subjects. None may be used as a substitute for another.

### 1 · Causal proof

```text
QUESTION     What computation did the accepted result come from?
PROVES       the observable causal conditions and the cone retained by the outer perspective.
DOES NOT     prove that the behavior was correct, honest, or free of hidden egress.
```

For each retained invocation, the record establishes:

```text
this artifact and expectation
interpreted by this entity
received exactly these observations through the frame
under this parent decision
in this emitted and causal position
and produced this result
which this parent remembered
and which survived, or did not survive, outer remembrance
```

The causal proof is grounded by substrate-held events rather than an entity's retrospective account
of its reasoning. It establishes exposure, choice, order, output, remembrance, and acceptance.

It does not claim access to private reasoning. It does not prove counterfactual necessity from one
execution; necessity is tested by replay and intervention against the recorded conditions.

### 2 · Confinement proof

```text
QUESTION     Where could information from this computation leave?
PROVES       the complete set of exits the runtime made available and what happened at each one.
DOES NOT     prove what an unconstrained or incorrectly isolated entity did outside the runtime.
```

Knowing which entity answered or which endpoint received the final request is not sufficient.
Secret behavior can encode data into any channel it is allowed to use. A claim that data was not
collected through a back door requires every possible egress channel to be mediated by the
substrate.

The confinement record must cover at least:

```text
NETWORK           every address and transport reached; no ambient socket access.
STORAGE           every readable or writable path/handle; no ambient filesystem access.
EXECUTION         every spawned entity and the authority handed to it.
WORLD EFFECTS     every staged, landed, failed, and discarded effect.
CREDENTIALS       which customer-controlled handles were made available, never their secret bytes.
INHERITANCE       environment, descriptors, sessions, caches, and other continuity crossing in.
DENIALS           every attempted effect refused by the runtime.
```

The proof is therefore not merely:

```text
we know which entity answered
```

It is:

```text
this computation had exactly these observable exits,
and the enforcement boundary made no others available
```

This claim depends on isolation. Ambient filesystem access, inherited descriptors, unrestricted
DNS, telemetry, undeclared subprocesses, provider-side memory, or any other back-channel makes the
exit set incomplete. That is not a minor attribution defect; it invalidates the confinement proof.

### 3 · Artifact proof

```text
QUESTION     Which proprietary behavior actually executed?
PROVES       the authorized artifact and measured runtime correspond to the execution receipt.
DOES NOT     reveal the artifact's implementation or establish that its behavior is good.
```

The artifact proof binds the execution to:

```text
artifact hash
publisher signature
compiler and format identity
entity identity
runtime measurement
protocol version
causal invocation identity
```

The customer may not be able to inspect the compiled behavior, but can detect replacement of the
artifact it authorized. An audit, certification, prior evaluation, or commercial agreement can
refer to the hash rather than to a mutable product name.

Where the artifact executes remotely, a receipt signed only by the artifact vendor proves what the
vendor asserted. Stronger artifact proof requires a customer-verifiable execution boundary, such as
a measured customer-controlled runtime or independently verifiable attestation.

---

## How the proofs compose

Each proof answers one question:

```text
CAUSAL PROOF        what contributed to the retained result?
CONFINEMENT PROOF   where could the supplied information go?
ARTIFACT PROOF      which opaque behavior ran?
```

Together they can establish:

```text
artifact H ran
in measured runtime T
as entity E
with observations [A, B]
under grants G
with exits limited to [X, Y]
produced result R
caused effects [landed, discarded, failed]
and its result survived along this accepted causal cone
```

They still do not establish timeless truth, semantic correctness, or benevolent intent. Those are
evaluated by outcomes, competing behavior, later observations, and the accepting perspective.

```text
ONE WITHOUT THE OTHERS
────────────────────────────────────────────────────────────────────────────
causal without confinement     proves what was supplied, not where it may have leaked.
confinement without artifact   proves the exits, not which proprietary behavior used them.
artifact without causality     proves what ran, not what it received or what was retained.
```

No product claim should compress the three into one undifferentiated word such as "verified."

---

## Deployment strength

The same protocol can support different trust boundaries. Their proofs are not equivalent and must
not be presented as if they were.

### Customer-controlled execution

```text
ARTIFACT       signed proprietary behavior executes in the customer's runtime.
DATA           remains in infrastructure controlled by the customer.
POLICY         the customer controls grants, endpoints, credentials, and retained memory.
PROOF          the customer directly observes confinement and verifies the artifact hash.
```

This is the strongest form of the claim that the behavior provider does not receive the customer's
data. The provider supplies behavior; the customer supplies the execution boundary.

### Attested execution

```text
ARTIFACT       executes in a measured confidential environment.
DATA           is released only after the customer verifies the measurement and artifact.
POLICY         exits remain constrained by the attested runtime.
PROOF          depends on the attestation root and the completeness of the measured boundary.
```

This permits secret remote behavior with a customer-verifiable claim about what ran and which exits
were available.

### Vendor-controlled execution

```text
ARTIFACT       executes in infrastructure controlled by the behavior provider.
DATA           crosses into the provider's authority domain.
POLICY         is reported and enforced by that provider.
PROOF          supplies useful receipts, but ultimately rests on operator trust.
```

Causal and endpoint records remain valuable here, but they are not proof that the infrastructure
operator had no other access. The honest claim is evidence plus contractual and operational trust,
not cryptographic non-custody.

---

## What remains owned by whom

```text
PRINCIPAL OWNS
    namespace
    observations and accepted history
    credentials and endpoint policy
    grants and admission policy
    declarative workflows authored by the principal
    the decision to retain or reject a result

BEHAVIOR AUTHOR OWNS
    proprietary source
    compiler and optimizer
    compiled engram artifacts
    private workflow and domain implementation

OPEN PROTOCOL DEFINES
    how decisions bind inputs and authority
    how results are remembered
    how causal cones are represented
    how artifact identity is carried
    how exits and effect dispositions are recorded
    how another implementation verifies the same record
```

The protocol must not require memory custody by the protocol author. A runtime must not require a
central service to interpret a principal's own durable record.

---

## The commercial model

The protocol and its conformance suite may be open and free while proprietary value accumulates in
behavior:

```text
OPEN FOUNDATION
    interoperable runtimes
    portable memory
    replaceable providers
    independently verifiable seams

PROPRIETARY VALUE
    better compiled behavior
    better domain workflows
    better translation and recognition
    better selection and routing strategies
    certified artifacts for specific uses
```

Portability does not introduce another synchronization or transport mechanism. Moving memory to
another runtime is itself an explicit `Reach`; Reach is the only primitive permitted to cross that
boundary.

This is not security by obscurity. The security boundary is open, testable protocol and runtime
enforcement. What remains secret is the implementation of behavior inside that boundary.

The resulting promise is narrow and defensible:

> **You own your memory. You can verify which behavior ran, what it received, what you retained,
> and every exit the runtime made available. The behavior author may keep how that behavior works
> proprietary.**

---

## Requirements this creates

The model is only real if the implementation supplies:

```text
1  a canonical, versioned artifact identity and signature format.
2  an open causal-envelope and accepted-cone representation.
3  complete mediation of network, storage, execution, credentials, and effects.
4  no ambient authority by default.
5  explicit recording of denied, failed, discarded, and landed effects.
6  customer-verifiable endpoint and entity identity.
7  conformance fixtures for all three proofs.
8  deployment receipts that state their trust model rather than merely saying "verified."
9  a precise account of hidden continuity for every entity class.
10 semantic portability: another conforming runtime reads the same record the same way.
```

Until complete mediation exists, the confinement proof is a design target rather than a property of
the system. Until artifact measurement exists, a signature proves publication and not execution.
Until outer remembrance projects the retained causal cone, the causal proof describes the older
per-commit model rather than the intended transaction.
