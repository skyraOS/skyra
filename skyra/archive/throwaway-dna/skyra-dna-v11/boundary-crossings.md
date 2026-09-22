# Boundary crossings — v11 CTU hot path

**Date:** 2026-08-26  
**Scope:** `ctu` HOT path from the `overnight.md` handoff.

## Where the frozen surface did not hold

The first implementation pass stopped before editing because the `Transfer → Move` fold had removed the
request seed. `Move` could name an operation, but it could not carry the `BodyID`, horizon, seed register
values, or correlated return route needed to create the first resident `Context`. The same surface exposed
driver declarations but no callable local mechanism, gave `Core` neither an interpreter nor a self target,
and retained only a return target rather than the parent's fd.

`resolution.md` returned the design to the propose lane and re-froze it with `Move.Seed`, `Route`, `Driver`,
and `Core.{Self,Interp,Local}`. The implementation applies those changes rather than hiding the missing state
in package globals or reconstructing it from names.

`Materialize(BodyID, Horizon)` remains narrower than the whole request seed. The local materializer therefore
owns body resolution and register-file allocation; `Transfer` overlays the seed's possibility ID, input
values, horizon, and return route. That keeps the frozen loader signature while making the ownership split
explicit.

## What the fd path forced

The parent move's `MoveID` is the fd and `Work[Issued]` is its table. Issuing a cut copies that fd into
`Seed.ReturnTo.Fd` and parks the resident move in `Issued`. The child response echoes the fd as its own
`Move.ID`; `Transfer(Response)` can then land directly in
`Cache[id.Possibility].Registers[id.Writes]`, pop `Issued`, and wake dependents. No second attention map or
kernel lookup is needed inside the CTU.

The cut seed carries only capabilities matching the selected `(Target, Action)`. This makes attenuation a
literal removal operation and avoids giving the child unrelated entries from the parent's horizon.

## The concept that needed a home

Driver advertisement and driver execution are different objects. `DriverDecl` remains inert metadata read
by the kernel at handshake; `Driver` is the installed mechanism, held in `Core.Local`. A horizon never makes
a target local by itself, and a local installation never grants authority by itself. Both checks must pass.

`Core.Self` is likewise endpoint identity, not authority. It supplies egress's `from` value and the return
route for a cut; permission still comes exclusively from the resident context's horizon.

## Calls forced on the open threads

- **Interpret local vs cut:** `Interpret` answers only whether the target is in the context's provisioned
  horizon. `Core.Local` independently decides LOCAL versus CUT for the selected target.
- **Done detection:** writing the body's return register sets `Context.Result`. `Run` emits/evicts only when
  that result exists and the possibility has no Ready or Issued move, matching the handoff's definition.
- **Dependent-flip indexing:** the first build scans `Work[Waiting]` within the same possibility after a
  landing. This is intentionally simple and correct; a reverse dependency index remains an optimization,
  not another source of truth.
- **Run scheduling:** `Transfer` never drives inline. It signals a condition variable; the immortal `Run`
  loop parks with an empty Ready bucket and drains the queue after wake-up. A fault leaves its move Ready and
  parks until a later transfer represents out-of-scope kernel progress.
