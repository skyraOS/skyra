# 24 — The room: proposal

Status (2026-09-30): **Implemented in part.** Steps 1 (`a275748`), 2 (`dad1bf0`), 4 and 5's resume (`dad1bf0`) are built; B1 and B2 in part. **Not built:** step 3 (the stack without the `Resume` copy), step 5's recovery by reading state, step 6 (marks backed by records).

The implementation plan for [24](../24-the-room.md). Ids refer to the
[list](README.md). 24 depends on improvements items 1 and 2 (A), which land
first.

## Step 1: the Boundary is authority, shared, rendered once (B1, B6, B7)

Lands together with 25's C1–C3.

- `core/boundary/boundary.go:126`: `Boundary` keeps `Reach` and gains the grants
  that leave `Surface` (`Authority`, and later the network grants). `Stamp`,
  `Frame` and `Resume` move to `machine.Instance` for now (a stepping stone
  to B5).
- `core/machine/machine.go:456` (`Narrow`): an invoke returns **the caller's
  Boundary unchanged**. The intersection with `Surface.Reach` (`:469`) and the
  `Delegate` check (`:471`) go. Returns and faults stop building an empty
  Boundary (`:459`): they discharge under the Boundary they ran under.
- `boundary.Render` (`:157`): renders from the Boundary's grants; `process-exec`
  only for the running program's path (already `h.Surface.Path`, which moves
  to the step). The floor and the loader reads stay as the exclusions.
  `RenderResident` (`resident.go:21`) follows.
- Tests: `boundary_test`, `boundary_os_test` (one profile per Boundary; exec of
  another registered binary denied), `machine_test` (`Narrow` shares).

## Step 2: committed, looked up, re-derived (B2, B3, B4)

- `record`/`boundary`: a Boundary id = hash(authority, creating invoke stamp,
  input contract). Until E2, a plain sha256 of that canonical form.
- `runtime/tree`: `boundaries/<id>`, write-once, hash-checked on read like
  `records/`.
- `machine.handoff` (`:413`): instead of `SendBoundary{Child}` carrying the
  struct, a new effect `CommitBoundary` when one is created, then
  `SendBoundary{ID}`.
- `instance` (`instance.go:177`) and `handshake`: send the id; the child
  loads it (`machine.accept`, `:330`, gains the lookup result as part of
  `Released`) and re-derives it: the id must equal what the state and its own
  record derive.

## Step 3: the stack is durable (B5)

- `Frame` and `Resume` leave the handoff. `gather` (`machine.go:509`) and
  `claim` (`:532`) read the waiting frame from the tree (they already read its
  rounds through `instance.gather`, `instance.go:242`).
- `accept`'s frame-count check (`Depth()`) is replaced by the lookup.
- Every write stays write-once (`deps/`, `result`, `resolved-<round>`).

## Step 4: the room at genesis (B8)

- `runtime/boot/boot.go:47`: genesis records the operator as author and a
  nonce; the root Boundary is the room's, its reach the registry table.
- The room is a position: whichever instance is woken at the root derives
  and exits. No long-lived process.

## Step 5: crash recovery and resume (B9, B10)

- `boot` and the webhost find work by reading state: orphans (no accepting
  derivation) dropped, accepted steps with no successor run again. The
  webhost's sibling rule (`533d98d`) goes.
- `runtime/park` carries a stamp; `skyra resume` (`skyra/main.go:107`)
  rebuilds the instance from the tree and refuses a stamp that is not a
  parked, unresumed step.
- The docs say effects are at least once.

## Step 6: marks backed by records (B11)

Improvements item 8, direction 1: replay checks every mark against a record.
Then frame fds can become hashes.

## Faults climb, and carry their observation (A6)

No new state: a fault already discharges the frame. What changes is that the
fault carries its observation (A6) and that a parent holding the authority
can invoke again with less narrowing (after D2).

## Tests that prove it

- A leaf under a shared Boundary can do what its caller can, and nothing more;
  exec of any other binary is denied at the OS.
- A forged Boundary id handed to a child fails the re-derivation.
- Kill an instance after its Boundary is committed and before its target
  emits; recovery runs it again; replay is byte-identical.
- Resume with a stamp that is not parked, or twice: refused.
- Replay of every existing demo scenario, under the new rules or the old
  version (E7).

## Open, from 24

The shape of `system.narrow`'s input contract; which binary and sandbox run
the room's derivation; truncation; registration's bytes in the tree.
