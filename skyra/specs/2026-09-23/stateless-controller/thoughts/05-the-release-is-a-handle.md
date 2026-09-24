# The release is a handle

Date: 2026-09-23  
Status: agreed 2026-09-23 (D5, D6). Refines the release in [04](04-the-three-phases.md).

## Agreed

- **No second stamp and no second commit.** Custody moves at one durable point,
  the successor's commit. The release is not written to disk.
- **The release is a handle the child needs in order to execute**, not just a
  go-signal. Execution without it is impossible, not merely against the rules.

## What the handle must be

1. **Only the parent can give it.** Otherwise it enforces nothing.
2. **Bound to exactly one commit.** It unlocks the execution recorded in
   `R_{N+1}` and nothing else.
3. **No secret.** Nothing goes in the tree or in Moves.

A hash passed into a function inside N+1 meets 2 but not 1. N+1 would be
checking a token it could have computed itself, which is a convention, not
enforcement. What meets 1 without a secret is an **OS handle**: a file
descriptor that only the parent holds and passes over the handshake socket
(`SCM_RIGHTS`).

## Proposal

The release carries two things:

| Part | What | Why |
|---|---|---|
| `stamp` | `hash(R_{N+1})`, echoed back | Binds the execution to the commit. Execute reads the surface from `R` by this hash |
| `node` | fd for N+1's own frame node, created by the parent | The capability. Execute cannot record supervision state, child slots or faults without it, and only the parent can hand it over |

During intake, N+1 holds only what it needs to commit `R_{N+1}`. It gets its
frame node, and so the ability to act, only at the release.

## Signature

Executable and args are not parameters. They were committed in `R_{N+1}`, so
`ex` reads them from there. If they were passed separately, `ex` would have to
check them against `R` anyway, and any mismatch would be a gap between what was
committed and what runs.

```go
// Live can only be constructed from a verified release.
type Live struct {
	stamp Hash     // hash(R), echoed by the parent
	node  *os.File // this frame's node, received over SCM_RIGHTS
}

// Execute reads R by l.stamp and spawns exactly R's surface
// (binary, args, Seatbelt profile). Nothing else can be run.
func Execute(l Live) (*Target, error)
```

Compared with `ex(executable, args, hash)`:

- `executable`, `args` → come from `R`, so they cannot diverge from what was
  committed.
- `hash` → becomes `stamp` plus `node`. The stamp binds, and the fd enforces.

## Open

1. What does intake write through, before it has its node? Probably a
   write-once record slot the parent creates alongside the node. This matches
   the "at most two handles" note in proposal §2, but they are handed over at
   different times.
2. Does one fd suffice as the capability, or should the release also carry T's
   pre-opened pipes, so the parent sets up all of T's wiring?
