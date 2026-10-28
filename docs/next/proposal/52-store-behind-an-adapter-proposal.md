# 52 — The store behind an adapter

Date: 2026-10-28

Status: **proposed, for review.** Nothing is built. Open questions are listed
with the design, each with a recommendation where there is one.

**Scope:** the line between Skyra and the storage that holds its record. What
Skyra asks of storage, in its own terms; where the database lives (behind an
adapter, on the store's side of the line); SQLite as the first database; and
what has to hold for the store to become distributed later. Builds on
[48](48-sql-schemas-proposal.md) (the schemas), [49](49-services-behind-adapters-proposal.md)
(services behind adapters), [50](50-write-policies-proposal.md) (the fold)
and 40 (a flush before each release point).

## Summary

1. **Skyra speaks a store protocol, never SQL.** Two kinds of thing are
   stored: **objects** and **claims**. Skyra never names a backend, opens a
   database or imports a driver.
2. **Objects are bytes named by their hash**: records, inputs, Horizons,
   blobs. They never change. A record names other objects by hash
   (predecessor, deps, emission, input), so the objects form a Merkle DAG.
3. **Claims are write-once names.** A name is set to a value exactly once:
   the first writer wins, and every later writer is told the winner. A
   frame's result, a slot, an emission seal and a fan-in claim are all claims
   today.
4. **Everything else is derived**: listings, indexes, 48's tables. An adapter
   keeps them however it likes and can rebuild them from objects and claims.
5. **The database is the adapter's business.** A filesystem adapter (today's
   `runtime/tree`, extracted), SQLite next, Postgres for 48, a distributed
   store later. The SQL and the driver live in the adapter's package and
   nowhere else.
6. **Swappable is shown, not promised.** One conformance suite every adapter
   passes, and a cross-adapter check: export a DAG from one adapter, import it
   into another, and replay byte-identically.

## Decisions

Recorded 2026-10-28.

1. **SQLite is the first database.**
2. **The backend is swappable, and the adapter lives on the store's side.**
   Skyra never speaks SQL. Changing databases is writing an adapter, not
   changing Skyra.
3. **Skyra will become distributed.** Nothing in the protocol may assume one
   machine, one disk or one process.
4. **The record stays a Merkle DAG.** Objects are named by their hash, and
   Skyra checks every name against its bytes.

## Why objects and claims

This is what `runtime/tree` already does. Nothing in `.skyra/` is updated in
place:

| Today (`runtime/tree`) | What it is |
|---|---|
| `Put`, `Store`, `Get`, `Commit`, `Load`, `PutHorizon`, `LoadHorizon` | objects: stored under their hash, refused on read unless they hash to it |
| `MakeFrame`, `Slot`, `Emit`, `Discharge`, `Claim` | claims: each "reports whether this write won" |
| `Result`, `Resolved`, `Emission`, `Children` | reading claims |
| `Stamps`, `HasSuccessor` | derived: a listing, and a scan of every record |

So the protocol is not a new model. It names the one the code already uses,
so that something other than files can provide it.

**Why write-once, and not a head that moves.** A distributed log usually keeps
a mutable head and moves it with compare-and-swap. A claim is simpler to
distribute:

- **Once set, a claim never changes.** Any replica that has seen it can serve
  it forever and never invalidate it. Only the first write needs agreement;
  every read after it is local.
- **A head changes**, so every read of it needs agreement or risks being
  stale.
- **The head is derivable.** The latest record of a lineage is the one with
  no successor. Nothing needs to store it.

**One step, one claim.** The README describes a step as read, authorize, one
atomic write, execute. Under this protocol, the atomic write is a claim. The
objects it refers to are stored first. A crash before the claim leaves objects
nothing refers to, which are harmless. A crash after it leaves a complete
step. 50's fold fits this too: the checkpoint and the diff are one record,
stored as one object and then claimed once.

## The protocol

| Operation | What it does |
|---|---|
| `put(kind, bytes, links) → hash` | Store bytes under their hash. Storing the same bytes again is a no-op. `links` are the hashes the object names, each with a label (see "Links"). Returns once durable. |
| `get(kind, hash) → bytes` or none | Read an object. Skyra checks the bytes against the hash before using them. |
| `claim(name, value) → (won, winner)` | Set a name once. If it is already set, report the value that won. Returns once durable. |
| `lookup(name) → value` or none | Read a claim. |
| `list(prefix) → [(name, value)]` | Every claim under a prefix, in bytewise name order. Used for a round's slots. |
| `refs(hash, label) → [hash]` | Every object that links to `hash` under `label`. Replaces `HasSuccessor`'s scan. |
| `export() → stream`, `import(stream)` | Every object and claim, in an order an importer can apply: objects before the claims that name them. |

**The store is not trusted to check.** Skyra re-hashes everything it reads,
as `tree.Get` does today. The store is one more source of bytes, and the
emitter is never the checker. A distributed store holds bytes on machines
Skyra does not control, so this check is what makes it safe to read from
them.

**Names are Skyra's, and opaque to the adapter.** A name is a sequence of
plain components, each refused if it is empty, `.`, `..` or contains `/` (the
rule `component` applies today). The adapter treats names only as prefixes to
list. Today's layout maps over directly:

```text
frame/<F>                          the frame exists        (MakeFrame)
frame/<F>/result                   the record that discharged F's slot
frame/<F>/deps/<round>/<child>     a slot, holding its position
frame/<F>/emitted/<round>          the sealed emission
frame/<F>/resolved/<round>         the fan-in claim
```

**A frame handle keeps today's scoping.** Every write inside a frame now goes
through its directory fd, never a path. The protocol keeps that rule:
`frame(F)` returns a handle whose claims can only name things under
`frame/<F>/`. Code holding one frame's handle cannot reach another frame.
Without paths, the symlink and path-resolution hazards the `*at` calls guard
against do not arise.

## Links

The adapter knows the shape of the DAG but not its meaning. Skyra owns the
record format, so Skyra extracts each object's links when it puts it:

```text
record r   → (predecessor, r.Predecessor), (input, r.Input),
             (emission, r.Emission), (dep, d) for each d in r.Deps
```

The adapter stores the edges and answers `refs`. It never parses a record.
Skyra checks on read that the links it gets back match the ones it derives
from the bytes, so a wrong edge is caught, not trusted.

What this gives:

- **`HasSuccessor` stops scanning every record.** It becomes
  `refs(stamp, predecessor)`, which is an index lookup in SQLite.
- **Reachability is a store query,** which forgetting
  ([`ideas/forgetting.md`](../../../ideas/forgetting.md)) and any future
  garbage collection will need.
- **A distributed store can replicate a subgraph** by following links,
  without understanding records.

## The SQLite adapter

```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = FULL;
PRAGMA fullfsync = ON;             -- macOS: F_FULLFSYNC, as 40 requires
PRAGMA checkpoint_fullfsync = ON;
PRAGMA busy_timeout = 5000;        -- concurrent steps wait for the writer lock

CREATE TABLE objects (
    kind  TEXT NOT NULL,
    hash  TEXT NOT NULL CHECK (length(hash) = 64),
    bytes BLOB NOT NULL,
    PRIMARY KEY (kind, hash)
) STRICT, WITHOUT ROWID;

CREATE TABLE links (
    source TEXT NOT NULL,          -- the object that names...
    label  TEXT NOT NULL,
    target TEXT NOT NULL,          -- ...this hash
    PRIMARY KEY (target, label, source)
) STRICT, WITHOUT ROWID;

CREATE TABLE claims (
    name  TEXT NOT NULL PRIMARY KEY,
    value TEXT NOT NULL
) STRICT, WITHOUT ROWID;

-- Append-only. The adapter never issues these. The triggers are a backstop
-- against a bug, not the protection: that is the OS denying .skyra/ to every
-- target, so the step binary is the database's only possible client.
CREATE TRIGGER objects_immutable BEFORE UPDATE ON objects
    BEGIN SELECT RAISE(ABORT, 'objects are immutable'); END;
CREATE TRIGGER objects_kept BEFORE DELETE ON objects
    BEGIN SELECT RAISE(ABORT, 'objects are never deleted'); END;
CREATE TRIGGER links_immutable BEFORE UPDATE ON links
    BEGIN SELECT RAISE(ABORT, 'links are immutable'); END;
CREATE TRIGGER links_kept BEFORE DELETE ON links
    BEGIN SELECT RAISE(ABORT, 'links are never deleted'); END;
CREATE TRIGGER claims_immutable BEFORE UPDATE ON claims
    BEGIN SELECT RAISE(ABORT, 'a claim is set once'); END;
CREATE TRIGGER claims_kept BEFORE DELETE ON claims
    BEGIN SELECT RAISE(ABORT, 'claims are never deleted'); END;
```

- **`put`** runs `INSERT ... ON CONFLICT DO NOTHING` for the object and its
  links in one transaction. Before inserting, the adapter hashes the bytes
  and refuses a mismatch, so corruption is caught at write time as well as at
  Skyra's read.
- **`claim`** runs `BEGIN IMMEDIATE`, then `INSERT ... ON CONFLICT DO NOTHING`,
  then reads the winner, then commits. The writer lock makes it one decision
  across every concurrent step process.
- **`list(prefix)`** is a range scan on `claims`' primary key.
- **No server.** Each step opens the file, runs its operations and exits,
  so the README's crash argument holds unchanged. WAL lets readers run while
  a step writes, and the database, its `-wal` and its `-shm` files live in
  `.skyra/`.
- **Driver: `modernc.org/sqlite`** (pure Go, no cgo), imported only by the
  adapter's package. `core` stays standard-library Go, and cross-compiling
  for the Linux port and Homebrew stays a plain `go build`.

## Where 48 fits

48's schema is written for Postgres. Under this proposal it is the schema of
the Postgres adapter's derived tables, and later of the SQLite adapter's.
Its rows already fit the protocol: a blob is an object, and an emission or
decision keyed by `(horizon_root, path, n)` is a claim under that name. Its
primary keys are already the write-once rule.

Porting to SQLite:

| Postgres (48) | SQLite |
|---|---|
| `INTEGER[]` paths | a BLOB of big-endian `uint32`s: byte order is tree order, and a parent's path is a prefix |
| `trim_array(path, 1)` | a stored `parent_path` column the adapter fills |
| `BYTEA`, `CHECK (sha256 = ...)` | `BLOB`; the adapter hashes on `put` |
| `COMMENT ON` | SQL comments in the schema file |
| `CREATE ROLE`, `GRANT`, `REVOKE` | the OS: `.skyra/` is denied to every target |
| PL/pgSQL functions (`narrow`, `emit`, `decide_*`) | see open question 1 |

## Toward distributed

Nothing is chosen here. What the protocol keeps open:

- **Objects replicate freely.** They are immutable and name themselves, so
  any replica can hold any object, and Skyra's read check makes every
  replica as good as any other.
- **Claims need agreement only on their first write.** After that they are
  as immutable as objects. Claims under different frames never compete, so
  claims can be split across machines by frame.
- **Candidates for the agreement**: Postgres with one primary (the simplest,
  and 48's schema already targets it), or etcd or FoundationDB if one
  primary is not enough.
- **A store service runs out of process**, as one of 49's hosts. The
  protocol's operations are its messages. In-process (SQLite today) and
  out-of-process are two transports for the same messages.

## What makes it swappable: the conformance suite

Every adapter runs one suite. A new adapter is done when it passes, and
claims are raced across separate step processes, not just goroutines:

1. **Objects:** putting the same bytes twice gives the same hash and no
   error. An object corrupted on disk is refused on `get` by Skyra's check.
   `put` returns only once the object is durable.
2. **Claim race:** N processes claim one name at once. Exactly one wins, and
   all N are told the same winner.
3. **Crash between put and claim** (`kill -9`): after reopening, the object
   may exist and the claim does not. Nothing refers to the object.
4. **Crash after a claim is won:** after reopening, `lookup` returns the
   winner.
5. **Names:** a frame handle cannot claim outside its frame, and
   components `""`, `.`, `..` and anything containing `/` are refused.
6. **Listing:** `list` returns exactly the names under a prefix, in bytewise
   order, and never a name under a sibling prefix that merely starts the same
   way (`frame/ab` vs `frame/abc`).
7. **Links:** `refs` returns every source that linked to a target under a
   label, and nothing else.
8. **Round trip:** export from this adapter, import into the filesystem
   adapter, and the reverse. Replay of the imported store is byte-identical
   to replay of the original.

Item 8 runs in CI for every pair of adapters. Replay already exists, so it is
a cheap way to show that two backends hold the same truth.

**Which adapter runs is configuration, never part of the record.** A record
that named its adapter would get a different stamp in each one, and the
round trip could not be byte-identical.

## Open questions

1. **Where do 48's checks live?** 48 puts `narrow`, `emit` and the `decide_*`
   checks in PL/pgSQL with `SECURITY DEFINER`, so a database role keeps the
   runtime from widening a Horizon. SQLite has no roles and no stored
   functions.
   - **In the store:** every adapter implements the checks again, in its own
     language, and replay never runs those copies.
   - **In `core`:** the checks are written once, in the pure function replay
     already re-derives. The store keeps only storage rules (hashes,
     write-once, links).

   Recommendation: `core`. Today the runtime and the kernel are the same
   binary, so a role mostly guards that binary against itself, and replay
   catches a widening after the fact. When the store becomes a separate
   service, it can link `core` and run the same function as an admission
   check: one implementation, run in two places.

2. **Batched claims.** Today each claim stands alone, ordered by flushes, and
   the code is correct with that. 48's `decide_invoke` opens several children
   in one transaction. Recommendation: no batch operation now. If 48's shape
   is adopted, write the decision as one record object and claim it once,
   with the children derived from it.

3. **Does the adapter also verify hashes on `put`?** Skyra verifies on every
   read regardless. Recommendation: yes, in every adapter, so corruption is
   refused where it enters, not discovered later.

4. **Keep the filesystem adapter once SQLite is the default?**
   Recommendation: keep it. It already exists, it can be read with `ls`, and
   it is the second adapter the round trip in item 8 needs.

5. **Large objects.** SQLite reads small blobs faster than separate files,
   and large ones slower. Recommendation: store everything inline to start
   and measure. If needed, the adapter moves large objects to side files,
   written and flushed before the row that names them commits. Skyra never
   sees the difference.

6. **Objects nothing refers to** (a crash between put and claim).
   Recommendation: keep them. They are harmless, and removing them belongs to
   forgetting, which `refs` makes possible.

7. **Claims that point at objects.** Most claim values are stamps, but a
   slot holds a position, so the store cannot check in general that a claim
   names an object it holds. Recommendation: Skyra's ordering rule (put
   before claim) keeps it true, and conformance item 3 tests it. Revisit if
   claims get a type.

8. **`hosts/repo` and `hosts/web` read `runtime/tree` directly.** They move
   to the protocol in the same pass as everything else, or they get a
   read-only view. Recommendation: the protocol, read-only, so nothing
   outside the runtime keeps a path into `.skyra/`.

## Migration

1. **Extract the protocol** into its own package, and make `runtime/tree` the
   filesystem adapter behind it. No behavior change: the same files, the
   same flushes, and replay byte-identical. Every caller moves to the
   protocol (`runtime/boot`, `runtime/instance`, `runtime/replay`,
   `runtime/metric`, `skyra`, `hosts/repo`, `hosts/web`).
2. **Add links.** Skyra passes them on `put`, and the filesystem adapter keeps
   them in an index file per target. `HasSuccessor` becomes `refs`.
3. **The conformance suite**, run against the filesystem adapter.
4. **The SQLite adapter**, behind configuration. It passes the suite, and the
   round trip with the filesystem adapter runs in CI.
5. **SQLite becomes the default.** Existing `.skyra/` trees open through the
   filesystem adapter, or are imported once.
6. **48's tables** as the SQLite adapter's derived index, when 48 is picked
   up. Postgres and a distributed store when a second machine needs one.
