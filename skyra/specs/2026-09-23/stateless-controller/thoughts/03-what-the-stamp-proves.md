# What the stamp proves

Date: 2026-09-23  
Status: agreed 2026-09-23 (D3). Follows [02](02-custody-and-the-stamp.md). Not git; a plain
content-addressed store.

## What the parent needs to know

Before it exits, the parent must know two things:

1. **Read:** the successor received exactly the bytes that were emitted.
2. **Write:** the successor durably committed a record of them.

## Candidates

| Stamp | Proves read | Proves write | Forgeable by |
|---|---|---|---|
| Echo the message back | Yes, if the parent compares | No. A successor can echo without writing anything | Anyone who saw the message |
| `hash(message)` | Yes, the compact form of the echo | No | Anyone who saw the message |
| A signed or HMAC'd "read and write happened" | Only as a claim | Only as a claim | Anyone with the key. It also needs a secret, which the rules forbid in the tree and in Moves |
| **`hash(record)`, verified against the tree** | Yes | Yes | Nobody who cannot write `.skyra/` |

## Proposal: the stamp is a pointer, not a claim

The successor commits a record R whose content includes:

```text
R = {
  predecessor: <the parent's own stamp>,
  input:       sha256(the exact emission bytes received),
  verdict:     accepted | fault,
  outcome:     <intents, result, or fault detail>,
}
```

R is stored under its own hash. The successor fsyncs, and only then writes
`stamp = sha256(R)` back to the parent.

The parent never trusts the stamp. It **reads R from the tree by that hash** and
checks three things:

1. R exists and hashes to the stamp. **The write happened.**
2. `R.input == sha256(bytes that were emitted)`. **The read happened, exactly.**
3. `R.predecessor == the parent's own stamp`. **It links into this lineage.**

Then it exits. Checking the verdict is not the parent's job.

## Why this cannot be forged

- A made-up stamp names nothing in the tree. Check 1 fails.
- A stamp copied from another lineage names a real record with the wrong
  predecessor and input. Checks 2 and 3 fail.
- To forge a passing stamp you must write a matching record into `.skyra/`.
  Only `skyra` steps can do that, because Seatbelt denies the target both read
  and write.
- The target never holds the stamp return fd, so it cannot even send a stamp.
  Topology prevents that, not cryptography.

So unforgeability comes from where the truth lives (the tree) and who can write
there (only `skyra`), not from a secret. The echo is kept in hashed form as
`R.input`.

## Consequences

- **Check 2 needs the parent to see the emitted bytes.** So the target's stdout
  goes through the parent: frame, hash, forward to the successor. This resolves
  open point 2 in [02](02-custody-and-the-stamp.md) in favour of relaying.
  Framing and hashing are not interpreting. The parent never branches on
  content, so the emitter is still never the checker.
- **Durability is ordered, not proven.** A read of R cannot tell a flushed file
  from one still in the page cache. The successor stamps only after fsync. That
  is a property of the `skyra` binary, which is trusted code on both ends.
- **Cost:** one small file read and one hash per handoff. This is negligible
  next to the fsync.
- **Content addressing (open question 5)** is now effectively decided for
  records: stored by hash, and chained through `predecessor`. Named slots remain
  only for once-claimed pointers (`result.json`, registry entries, fan-in claims).
