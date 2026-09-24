# The three phases of an instance

Date: 2026-09-23  
Status: agreed 2026-09-23 (D4). Cleans up the discussion after [02](02-custody-and-the-stamp.md)
and [03](03-what-the-stamp-proves.md).

## Two kinds of code in one instance

An instance is one `skyra` process, together with the program it runs.

- **Trusted:** the `skyra` entrypoint. It interprets, reads and writes the tree,
  spawns, relays, and takes part in the handshake.
- **Untrusted:** the target executable T that the lineage asked to run. It runs
  under Seatbelt, cannot touch `.skyra/`, and never takes part in the handshake.
  It only produces bytes.

Custody passes between the trusted halves of consecutive instances. T is never
a party to it.

## The phases

```text
          instance N-1                instance N                          instance N+1
          ────────────               ──────────                          ────────────
               │  bytes ───────────▶ 1 INTAKE
               │                        read Move, read tree, authorize
               │                        commit R_N (fsync)
               │ ◀──────── stamp ────── stamp = hash(R_N)
               │   verify R_N
               │ ───────── release ──▶  (release = the stamp echoed back)
             exits                      N is now live
                                     2 EXECUTE
                                        spawn T under Seatbelt
                                        supervise T, relay T's stdout
                                        (frame + hash each emission)
                                     3 HANDOFF (per emission)
                                        spawn N+1, forward bytes ─────────▶ 1 INTAKE
                                                                              commit R_{N+1}
                                        ◀────────────────────────── stamp ─── 
                                        read R_{N+1}, verify 3 checks
                                        release ──────────────────────────▶  N+1 is live
                                        T still running? → back to 2
                                        otherwise exit
```

So there are three phases, and consecutive instances overlap: **N's HANDOFF is
N+1's INTAKE.** The stamp and the release are the two halves of one handshake
at that overlap.

### 1. Intake (trusted)

Read the Move from stdin. Read the tree to authorize it. Commit `R_N` and fsync.
Send `stamp = hash(R_N)` to the parent, then wait for the release. Nothing
executes before the release arrives.

### 2. Execute (untrusted runs, trusted supervises)

Spawn T with the surface named in `R_N` (binary, args, Seatbelt profile). Hold
T's stdin and stdout. Frame and hash each emission. Never interpret it.

### 3. Handoff (trusted, once per emission)

Spawn N+1 and forward the emission. Wait for N+1's stamp. Read `R_{N+1}` from
the tree and check that:

1. it hashes to the stamp,
2. its `input` equals the hash of what T emitted, and
3. its `predecessor` equals `hash(R_N)`.

Then send the release. If T is still running (streaming), return to phase 2. If
T has exited and its last emission is released, exit. The release is the last
write N makes.

The normal path of phase 3 writes nothing to the tree, because the successor's
commit is the write. N commits only on the failure path (below), which is the
"maybe commit".

## Why the release (the parent's counter-stamp) is needed

The stamp alone moves custody durably: once `R_{N+1}` exists and links, restore
treats it as N+1's. But without a release, N+1 could start executing while N is
still verifying, and if verification fails, N would be committing a fault at the
same moment N+1 was acting. That is two actors for one Move.

The release means **N+1 never executes until N has stopped acting.** Custody
moves at the commit, and the right to execute moves at the release. It costs one
pipe round trip, measured in microseconds.

**The release carries the stamp back; there is no new handle ID.** `hash(R_{N+1})`
already names the new live frame uniquely and unforgeably. N+1 checks that the
release equals its own stamp, so a release meant for another instance cannot
start it.

## Failure table

| What fails | Who holds custody | What happens |
|---|---|---|
| N-1 dies before release | N (committed) | N sees EOF without a release. Proposed: N faults its own frame, which is strict. Alternative: proceed, since custody is already durable |
| T dies without emitting | N | N commits a fault on its frame and exits |
| N+1 dies before stamping | N | N commits a fault on its frame and exits |
| Verification of `R_{N+1}` fails | Disputed: `R_{N+1}` exists, but no release was sent | N withholds the release, kills N+1 and commits a fault. The fault invalidates descendants, so `R_{N+1}` can never settle anything. Should not happen, since N+1 is trusted code |
| N dies mid-execute | Nobody live; last durable fact is `R_N` | Restore faults `R_N` (an intent with no result) |
| N dies after N+1 stamps, before release | N+1 (committed) | Same as the first row, from N+1's side |

## Open

1. **EOF without a release: strict or lenient?** Strict (fault) is simpler to
   reason about. Lenient avoids faulting work that was durably committed when
   only the verifier died.
2. **Streaming parents stay alive as long as T does.** N is the holder from
   proposal §4, so no separate holder role exists. Delivering a Return into a
   live T happens here too: N holds T's stdin.
3. **Fan-out:** one emission carrying several Invokes. Does N spawn one N+1 per
   Invoke, each with its own stamp and release, or one N+1 that commits all of
   them in one record?
