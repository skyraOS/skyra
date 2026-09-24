# Custody and the stamp

Date: 2026-09-23  
Status: agreed 2026-09-23 (D2). Details are resolved in 03–06.

## The idea

A step does not exit when its child starts. It stays alive as the supervisor of
the Move it is responsible for until the successor takes custody. The successor
takes custody by committing the Move at the commit boundary and returning the
commit's hash, the **stamp**. Receiving the stamp releases the parent, and only
then does it exit.

```text
skyra(N)                          target T                 skyra(N+1)
  interpret M, commit ─ stamp ──▶ (to N-1, which may now exit)
  spawn T ───────────────────────▶ runs
  supervise                        emits bytes ───────────▶ interpret
     │                                                     commit record R
     │ ◀──────────────────────────────── stamp = hash(R) ── 
  verify stamp, release custody, exit                      spawn next …
```

This replaces "the parent dies once the child is running" with **"the parent
dies once the successor has stamped"**.

## Invariant

**Every Move is in the custody of exactly one party at every instant: a live
supervisor, or a durable commit.** Custody moves at one point, the commit
(rename plus directory fsync). The stamp is the receipt for that commit.

## What it buys

- **No unowned gap.** Previously, between "N spawned and exited" and "N+1
  committed", nobody owned the Move. Now N owns it until N+1's commit is durable.
- **Failure has an owner.**
  - T dies without emitting: N still has custody and commits a fault.
  - N+1 dies before stamping: N still has custody and commits a fault.
  - N+1 rejects the emission: it still commits and stamps, but what it commits
    is a fault. **A stamp means "I have custody", not "I approve".** The checker
    records its verdict either way.
  - N is killed while supervising: nothing live holds custody, so the last
    durable commit (N's own) is an intent with no successor, and restore faults
    it. This is the same restore rule as before.
- **The lineage becomes a hash chain.** If every record includes its
  predecessor's stamp, the tree is content-addressed and tamper-evident, and
  replay can verify each link. This mostly answers open question 5 (content
  addressing) in favour of hashes.

## Open

1. **What does the parent verify in a stamp?** At minimum, that a record with
   that hash exists and names N's commit as its predecessor. Stronger: that it
   also commits exactly the bytes T emitted. The stronger check is possible only
   if N sees those bytes (see 2).
2. **Does T's stdout go straight to N+1, or through N?** If N relays (frame, hash,
   forward), it can check that the stamp covers exactly what was emitted. Framing
   and hashing are not interpretation, so the emitter is still never the checker,
   and N never branches on content. If the stdout goes straight to N+1, N can
   verify only the predecessor link.
3. **The stamp's return path.** Probably an inherited fd from N to N+1 (for
   example fd 4, write end), so only the successor N spawned can stamp to it.
4. **Streaming children.** If N supervises T for T's whole life, then each
   emission gets its own successor and its own stamp. N is then the holder from
   proposal §4: a `skyra` in its supervising phase, not a separate role. This
   needs agreement.
5. **Where does a Return into a live T go?** If N is the supervisor, it holds T's
   stdin and is the natural party to deliver a committed Return into it.
6. **The hash.** SHA-256 over the canonical bytes of the record. Is the store a
   plain content-addressed object tree, or git itself? Git gives objects, hashes
   and tooling for free, but brings its own locking and write path.
