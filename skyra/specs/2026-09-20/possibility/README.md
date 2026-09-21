# What is a Possibility?

Date: 2026-09-20

The purpose of this is to trace a Possibility's lifespan and establish what it owns.

Latest refinement: [Engrams, Surfaces, and the pipe protocol](engrams-surfaces-and-pipe-protocol.md).
Engrams are serialized compositions; Surfaces are addressable executable endpoints.
Executables parse Moves and speak the Skyra protocol over pipes. The note also
records the bounded-unit read-before-write-before-return profile.

The [consistency review](consistency-review.md) records conflicting descriptions,
stale declarations, and missing contracts across these notes and the September 19 specs.

Start with the [code-on-paper proposal](proposal/README.md).

The [discussion takeaways](takeaways.md) record the pre-invocation state-operation
invariant and distinguish established rules from proposed ownership and open questions.

The next [authority namespace and execution-context refinement](authority-namespace-and-execution-context.md)
separates the invocation stack from a minimal execution context: serialized Moves
and their handlers, ideally a single executable handler list.

[Loop-owned authority and Move serialization](loop-owned-authority-and-move-serialization.md)
records the next refinement: the loop owner authorizes preparation work, services
declare their required operations, and everything leaving the loop for execution
is an AuthorizedMove.
