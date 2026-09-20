# Architectural thought: the main loop as Surface composition

Date: 2026-09-19  
Status: **Parked.** The phase shape below is real; the collapse it suggests is not
pursued. Not an implementation decision.

## Disposition (2026-09-19)

The shape is genuine — every turn obeys `interpret → commit → execute → return`,
the main loop included. Keep that as shared *discipline*. But **do not make
`step()` a composable Surface.** The loop is where authority lives: it owns
scheduling, locking, and commit authority. A Surface *consumes* authority; the
loop *holds* it. Collapsing the two erases the authority-holder / authority-
consumer boundary (D3) at the most load-bearing point in the system, and invites
turtles-all-the-way-down (to run a Surface you need the loop; if the loop is a
Surface, you need a loop to run it — there is always a bottom turtle).

The honest model is a sandwich with two things application code cannot author:
the **main loop** on top (drives turns, is not a turn) and **syscalls** on the
bottom (invoked, not authored — see `claude-syscall/syscall-boundary.md`), with
composable Surfaces in between. This mirrors the receive-loop call in
`claude-syscall/the-handoff-set.md`: the loop *drives* turns, it is not one.

Two halves of the note below, disposed:

- **"loop as Surface composition"** — parked. Revisit only as shared phase
  *vocabulary* (a thin privileged driver that calls interpret/commit), never as a
  collapsed boundary.
- **"primordial space"** — this is the concrete **root return destination** hole
  (cf. DNA-17 `rootOnce` and "the root still needs an agreed result
  destination"). Resolve it as plumbing in return-correlation, not as cosmology
  here.

The original note is retained below as the record of the observation.

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
