# Unification, Truth Maintenance, and Field Combination

**Status:** RESEARCH NOTE (2026-07-18). Strong candidate decomposition; no representation or
implementation decision.
**Companions:** [traversal-as-excitation.md](traversal-as-excitation.md) ·
[relational-region.md](relational-region.md) ·
[observation-side-learning.md](observation-side-learning.md)

This note investigates whether the field representation's unresolved combination rule:

```text
Combine(A, B)
    -> reinforce | compose | conflict | independent | unresolved
```

is substantially an existing problem from logic programming, constraint propagation, truth
maintenance, and graphical-model message passing.

## Finding

The correspondence is strong, but the five outcomes do not all come from ordinary unification.
They divide naturally across several mechanisms:

```text
UNIFICATION
    Can these relational terms share one binding?

CONSTRAINT PROPAGATION
    Which bindings remain possible after local constraints interact?

ATMS / TRUTH MAINTENANCE
    Under which assumptions and traversal paths is a binding supported or inconsistent?

FACTOR-GRAPH-LIKE MESSAGE PASSING
    How do local path contributions aggregate across the graph?

COMPRESSION
    Which repeated environments, justifications, and relational patterns deserve one bounded
    reusable representation?
```

Do not import Prolog, an ATMS, or a factor graph wholesale. Treat these as candidate mechanisms with
well-understood boundaries.

## Unification is the core compatibility operation

Unification solves equations between symbolic terms by finding substitutions that make the terms
equal. When successful, it returns a unifier—ideally a most general unifier that commits only the
bindings required by compatibility.

For relational structures:

```text
request(requester=X, resource=water)
request(requester=Kathy, resource=Y)
```

can unify as:

```text
X = Kathy
Y = water
```

This maps well onto two field outcomes:

```text
reinforce
    the structures already have the same bindings

compose
    the structures are compatible and their substitutions merge

conflict
    no substitution can make the structures compatible
```

Example conflict:

```text
water.carbonation = still
water.carbonation = sparkling
```

under a representation where the relational position accepts one value in the current environment.

### The five-way mapping is suggestive, not exact

The proposed direct mapping needs two corrections:

```text
independent
    is not ordinarily a unification result
    it means the constraints have disjoint scopes or no relation requiring combination

unresolved
    is not unification failure
    it means variables or domains remain partially bound after compatible information is applied
```

Ordinary unification more closely returns:

```text
success with substitution
failure
```

The substitution may retain variables, which helps represent unresolved structure, but whether two
structures should be treated as independent is a property of the surrounding constraint/field
graph.

## Constraint propagation supplies independence and unresolved domains

A constraint network represents variables with possible domains and relations restricting
compatible combinations. Local consistency algorithms prune values that cannot participate in a
solution before committing to one complete assignment.

This is close to the active field:

```text
relational position     variable
possible bindings       domain
path contribution       constraint
field combination       domain reduction / relation composition
conflict                empty domain or inconsistent relation
unresolved              several values remain possible
independent             constraints do not share variables
```

Mackworth's network-consistency formulation explicitly uses intersection and composition of
relations. That is directly relevant to traversal paths meeting at a relational position:

```text
intersection
    both constraints must hold over the same positions

composition
    a relation is induced through an intermediate variable/path
```

This may provide a classical, inspectable version of the interference combination rule without
inventing literal quantum phase.

## Truth maintenance preserves why a binding is present

Unification and constraint propagation answer whether bindings can coexist. They do not, by
themselves, preserve all of the reasons a datum is supported or the alternative contexts in which
it remains valid.

A truth maintenance system records justifications for beliefs and revises belief status when
discoveries contradict assumptions. This aligns with Skyra's requirements that:

- every active structure retain traversal-path provenance;
- contradiction not destroy the original observation;
- alternative interpretations remain simultaneously representable;
- later observations revise significance without rewriting durable content;
- action selection remain explainable from supporting paths.

## Why ATMS is especially relevant

An assumption-based truth maintenance system separates:

```text
assumption
    a hypothesis or contextual commitment

environment
    a set of assumptions

context
    what is derivable in that environment

justification
    why one datum follows from others

label
    the minimal environments under which a datum is derivable

nogood
    an inconsistent environment
```

This resembles Skyra's multi-root traversal:

```text
ATMS concept              possible Skyra analogue

assumption                root/path-specific interpretation or binding
environment               compatible set of traversal contributions
context                   relational morphology supported by that set
justification             recorded traversal path / derivation
label                     minimal supporting root/path sets for a field structure
nogood                    incompatible binding environment
```

The important property is not merely contradiction detection. An ATMS preserves several possible
contexts without destructively choosing one global belief state. That is close to the intended
"superposition" of candidate structures and to reconstructing an old observation's significance
under a new interaction.

### Minimal labels are already a form of compression

ATMS labels retain minimal supporting environments and discard inconsistent or subsumed
environments. This resembles the compression requirement:

```text
ten redundant supersets of the same support
    -> one minimal supporting environment
```

The analogy suggests that compression may begin with provenance minimization rather than semantic
summarization.

### ATMS is not the whole field

An ATMS is a propositional support/context mechanism. It does not decide:

- what relational vocabulary Skyra should use;
- which observations are semantically equivalent;
- how continuous confidence or activation should be represented;
- which action has the best observed consequence;
- how closure duration changes policy;
- which patterns deserve durable regions;
- how an LLM translates an unresolved interaction.

It may be the provenance/contradiction layer beneath the field representation, not the full
representation.

## Factor graphs explain local aggregation, with limits

A factor graph represents a global function as local functions over subsets of variables.
Sum-product and related algorithms pass local messages to compute global or marginal summaries.

This resembles:

```text
variables/factors            relational positions and compatibility constraints
messages                     path contributions during traversal
local combination            cell-owned overlap behavior
global marginal/summary      completed activation morphology
```

The connection is strongest if Skyra eventually attaches probabilities or weights to compatible
environments.

But factor-graph inference is not automatically a truth-maintenance system:

- numerical messages may lose explicit derivational provenance;
- probabilistic independence assumptions must be justified;
- cycles generally make ordinary belief propagation iterative and approximate;
- local marginal probability does not distinguish dropped evidence from resolved evidence;
- a factor graph does not by itself preserve immutable observation lineage.

Factor graphs are therefore a useful model for aggregation, not a replacement for unification and
truth maintenance.

## Proposed responsibility split

The strongest candidate architecture from this research is:

```text
KERNEL
    builds traversal paths
    records occurrences and convergence
    knows no relational semantics

FIELD TERMS / VARIABLES
    represent relational positions, values, free variables, modality, and scope

UNIFIER
    finds compatible substitutions
    reports binding failure

CONSTRAINT PROPAGATOR
    intersects and composes path constraints
    preserves unresolved domains
    detects local inconsistency and independence

ATMS-LIKE SUPPORT LAYER
    records minimal root/path environments
    preserves justifications and alternative contexts
    records nogoods without erasing observations

WEIGHT / LEARNING LAYER
    ranks compatible environments and actions from observed outcomes
    retains probability, recurrence, closure duration, and significance

COMPRESSION
    crystallizes recurring field patterns into bounded relational regions
    minimizes redundant support while retaining residuals and contradiction

COLLAPSE
    carries structures and may emit expression
    cannot write world state
```

This decomposition keeps relational semantics out of the kernel while avoiding the need to invent
every layer of `Combine` from scratch.

## A more precise candidate for Combine

`Combine` may be a pipeline rather than one primitive:

```text
Combine(A, B):
    1. Scope:
       Do A and B constrain any shared relational positions?
       no  -> independent

    2. Unify:
       Is there a compatible substitution?
       no  -> conflict in this environment

    3. Propagate:
       Merge substitutions and restrict remaining domains.
       fully same     -> reinforce
       adds bindings  -> compose
       domains remain -> unresolved

    4. Maintain:
       Record the supporting root/path environment.
       If inconsistent, add the environment to nogoods.

    5. Compress:
       Remove subsumed support environments without losing provenance.
```

This should be treated as a hypothesis to test with hand-worked traversals, not an API specification.

## Relationship to the recurrent traversal

At traversal `Tt`, several roots may contribute bindings to one field position:

```text
root O1 -> path P1 -> binding contribution A
root O2 -> path P2 -> binding contribution B
```

The field representation can:

1. unify or reject the bindings;
2. preserve remaining uncertainty as a domain;
3. associate the result with the minimal supporting root/path environments;
4. retain incompatible environments as explicit nogoods rather than deleting either observation;
5. weight the surviving alternatives from observed history;
6. compress repeated support patterns into a region occurrence;
7. carry selected patterns into `T(t+1)`.

This provides a concrete route from traversal convergence to significance, contradiction, and
context reconstruction.

## Relationship to episodic segmentation

An ATMS-like label also suggests a sharper significance delta:

```text
Significance(O, T)
    = the environments and relational roles through which O supports the active field at T
```

An episode boundary may involve a material change in:

- minimal support environments;
- which environments became nogood;
- which bindings remain unresolved;
- which region patterns survive compression;
- which structures collapse or carry forward.

This is richer than a raw node-set delta and preserves the reason an old observation changed role.

## Risks and cautions

### Environment explosion

For `n` assumptions there can be exponentially many environments. This is the primary danger of
adopting ATMS ideas literally.

Skyra will need:

- traversal-local scope;
- aggressive minimal-label/subsumption pruning;
- relevance and context budgets;
- region compression;
- possibly approximate top environments;
- an experiment showing when full provenance is actually required.

### Syntactic versus semantic compatibility

Unification only works after the field has a term language. It will not decide whether:

```text
"fine"
```

means agreement, concealment, or conflict avoidance. Translation or learned region structure must
first expose terms that can be unified.

### Time, modality, and provenance must be first-class

These should not unify accidentally:

```text
Michael said X
Michael believes X
X is observed true
X was true yesterday
X is desired
```

The term representation must include observation source, modality, temporal scope, and relational
role strongly enough that unification cannot collapse epistemically different claims.

### Contradiction is contextual

Two bindings may conflict in one environment while each remains valid under a different set of
assumptions or observation roles. Do not globally negate or erase either durable observation.

### Belief propagation on cycles is not exact by default

Skyra's relational field is likely cyclic. Factor-graph message passing on cyclic graphs needs an
explicit approximation, scheduling, or transformation decision; it cannot be assumed to converge
to an exact answer.

## Recommended research exercise

Before implementation, work four examples by hand:

1. **Identical support**
   - two roots produce the same binding;
   - verify minimal support environments and reinforcement.

2. **Compatible composition**
   - each root binds a different slot in one region;
   - unify them and preserve the joint provenance.

3. **Contextual contradiction**
   - two paths bind incompatible values;
   - retain both observations, create separate environments, and record the combined environment as
     nogood.

4. **Unresolved domain**
   - several bindings remain possible;
   - ensure collapse cannot silently convert uncertainty into one fact.

For each example, record:

```text
roots
paths
terms and variables
substitution/domain
minimal support environments
nogoods
region occurrence
what carries forward
what expression, if any, is permitted
```

Then compare the size of:

- the raw traversal tree;
- full provenance environments;
- minimal ATMS-like labels;
- a compressed region representation.

That experiment will show whether unification plus ATMS is the field representation's foundation,
one layer inside it, or too expensive for the active traversal.

## Research sources

- J. A. Robinson, “A Machine-Oriented Logic Based on the Resolution Principle,” *Journal of the
  ACM* 12(1), 1965. DOI: <https://doi.org/10.1145/321250.321253>
- Franz Baader and Wayne Snyder, “Unification Theory,” *Handbook of Automated Reasoning*, 2001:
  <https://iccl.inf.tu-dresden.de/web/LATPub217/en>
- Jon Doyle, “A Truth Maintenance System,” *Artificial Intelligence* 12(3), 1979. DOI:
  <https://doi.org/10.1016/0004-3702(79)90008-0>
- Johan de Kleer, “An Assumption-Based TMS,” *Artificial Intelligence* 28(2), 1986. DOI:
  <https://doi.org/10.1016/0004-3702(86)90080-9>
- Johan de Kleer, “Problem Solving with the ATMS,” *Artificial Intelligence* 28(2), 1986:
  <https://www.dekleer.org/Publications/Problem%20Solving%20with%20the%20ATMS.pdf>
- Alan K. Mackworth, “Consistency in Networks of Relations,” *Artificial Intelligence* 8(1), 1977:
  <https://www.cs.ubc.ca/~mack/Publications/AI77.pdf>
- Frank R. Kschischang, Brendan J. Frey, and Hans-Andrea Loeliger, “Factor Graphs and the Sum-Product
  Algorithm,” *IEEE Transactions on Information Theory* 47(2), 2001:
  <https://www.isiweb.ee.ethz.ch/papers/arch/aloe-2001-1.pdf>

## Bottom line

The five-way field combination is not one unexplored mystery:

```text
reinforce / compose / conflict
    strongly resemble unification plus constraint propagation

independent / unresolved
    are better expressed through constraint scope and remaining variable domains

provenance / alternative contexts / contradiction
    strongly resemble ATMS labels, environments, justifications, and nogoods

field-wide aggregation
    resembles local message passing on factor graphs

bounded reusable regions
    remain a compression problem
```

The most promising direction is not “replace the field with Prolog.” It is to test whether
unification, constraint propagation, and an ATMS-like support layer can provide the symbolic core
of the field representation while traversal, weighting, compression, and collapse remain Skyra's
own mechanisms.
