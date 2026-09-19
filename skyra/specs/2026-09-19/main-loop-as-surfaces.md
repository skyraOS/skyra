# Architectural thought: the main loop as Surface composition

Date: 2026-09-19  
Status: Note to revisit after dependency mapping. Not an implementation decision.

## Observation

The Surface abstraction is starting to resemble the runtime's `step()` function.
Both need to read and write runtime state, with a common phase structure:

```text
interpret → commit → execute → return
```

Surface ingress and egress profiles already describe state transitions using
interpret and commit. The main loop coordinates transitions with the same shape.

## Idea to explore later

The main loop might itself be expressed as Surfaces calling Surfaces, with the
**primordial space being the space every Move returns to**.

Rather than assuming `step()` must remain a separate architectural mechanism,
explore whether its behavior is a composition of the same primitives used by
other Surfaces. The role of the primordial space in receiving and interpreting
returning Moves needs to be defined.

## Questions for the later discussion

- Which parts of `step()` can be represented as ordinary Surface composition?
- Who owns scheduling, locking, and commit authority in that model?
- What minimal mechanism invokes the first Surface and drives subsequent work?
- What does returning to the primordial space mean for an individual Move versus
  satisfying a parent's input or return contract?
- Does the model preserve explicit interpret/commit phases, no canonical state
  reads past interpret, and execution outside the state locks?

## Revisit point

Finish mapping the current process and transport dependencies first. Then use
those concrete examples to assess whether the main loop needs a special mechanism
or can be expressed through Surface composition.

No runtime changes are requested by this note.

Related: [Surface abstraction](surface.md) and [phase rules](improvements.md).
