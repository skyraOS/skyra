# 41 + 42 — What gets built: one loop with an untrusted read path, and contracts equal by meaning

Date: 2026-10-16

Status: **proposed.** This is the build plan for two proposals:

- **41 — Read, commit, execute: one loop, an untrusted shell** (changes
  `controller/core/machine` and `controller/runtime/instance`).
- **42 — Equivalence by meaning: canonical contracts and behavioral
  signatures.** It was drafted as 40, but 40 is already
  [the durability policy](40-durability-policy-proposal.md), so it is
  renumbered here.

For each proposal this note says what gets built, what does not, and why.
The two land together because the kernel half of 42 changes the same
comparisons that 41 moves into the core. They land as separate commits, in
the order under [Landing order](#landing-order).

## Decisions at a glance

| | Build | Don't build |
|---|---|---|
| **42** | Required order ignored in proc (**done**, `324486b`); the same in the kernel, at entry and at every comparison; the 13 registered contracts sorted on disk | Behavioral signatures, tail classification and keeping each tail's reasoning (they move to the distribution procedure's own proposal); Title out of identity; number normalization; a store migration |
| **41** | Verification in the core; three effect kinds with outcomes; a handle table; the four true caches removed; one full flush per cycle; mutable queries as a fixed enum; replay checks one resume per resolve | Deriving Phase from disk; making replay a trivial shell in this change; a loser-verifies-winner claim; "the trusted base is `Step`" as written |

---

## 42 — Equivalence by meaning

### Build

**E1. Proc ignores Required order.** Done in `324486b`. `Schema.Canonical`
sorts and dedupes `Required` at every depth before marshalling, so `Key`,
`Same`, the store and the input type hashes all agree with the user.

**E2. The kernel ignores Required order.** The same gap exists in six
places, all comparing `contract.Canonical(schema)` bytes:

- `core/horizon/tools.go:70` (`PermitsTool`)
- `core/machine/room.go:188` and `:352`
- `runtime/registry/registry.go:120–131`
- `runtime/metric/metric.go:148`
- `runtime/replay/tree.go:182`

What gets built:

1. `contract.Normal(s *Schema) *Schema`, a copy with every `Required` sorted
   and deduplicated, and `contract.Same(a, b)`, which compares normal forms.
2. **Every comparison site calls `contract.Same`.** Old and new contracts
   then compare correctly whatever order they were stored in.
3. **Contracts are normalized where they enter:** at registry load, and
   where an actor adds a contract to itself (`remember`, `extend-reach`).
   Horizons committed from then on carry normal contracts, so a Horizon's id
   does not depend on how a contract was typed.
4. **`record.Encode` and the Horizon encoding do not change.** Old Horizons
   keep their bytes, still hash to their ids, and still pass 41's
   canonical-encoding check (V1). Normalizing inside `Encode` would make
   every Horizon recorded with unsorted `Required` fail that check, and would
   break replay of every published bundle.
5. **The 13 registered contracts with unsorted `Required` are sorted on
   disk** (`targets/read`, `targets/user`, `interp`, `interp/workflows/*`,
   `hosts/repo`). Normalizing at entry makes this cosmetic, but the source
   then reads as what the kernel holds.
6. **One set of test vectors for both normalizers.** The interpreter is its
   own module and does not import the kernel. A shared
   `testdata/contracts-normal.json` lists inputs and their normal bytes, and
   both `proc` and `core/contract` must reproduce it, so the two
   implementations cannot drift.

**Tests.**

- `PermitsTool`, `room` narrowing and registry lookup accept a contract whose
  `Required` is reordered.
- A tree recorded before E2, whose Horizons hold unsorted `Required`, still
  replays. Check in a small fixture tree for this.
- A newly committed Horizon holds sorted `Required`.

**Why.** It is a correctness bug in exactly the place Skyra claims rigor.
A contract restated in another order can be refused as a tool
(`PermitsTool`) or as a narrowing (`room.go`), even though it means the same
thing. The fix is mechanical, and normalizing at entry while leaving
`Encode` alone keeps every existing tree valid.

### Don't build

**Behavioral signatures, tail classification, keeping each tail's reasoning
(parts 2–4).** The multi-model distribution procedure is not in this repo,
so there is nothing to test these against. And the signature as specified,
`sha256([inv_1(answer), …, inv_n(answer)])` over pass/fail, is the same
all-pass vector for every answer that passes every invariant. The "genuinely
different design" row of the tail table is empty by construction. The
replacement belongs in that procedure's proposal:

- for contracts, an accept/reject vector over a fixed set of probe inputs;
- for designs, descriptive properties (batches writes, fails closed), kept
  separate from the correctness invariants.

Turning tails into candidate invariants becomes "add the probe that
separates this tail from the mode". That idea is worth keeping, and it
travels with them.

**Title out of identity.** The interpreter routes and names by `Title`
(`main.go:252`, `:400`, `:579`; `proc.go:340`). If Title left the key, a user
contract shaped like `remember` would pass `proc.Same(c, rememberContract())`.
Title can leave identity only after routing stops depending on it, which is
a separate change.

**Number normalization in `canon`.** A node's input is the bytes its target
receives, and invariant 3 stores observations as received, so `1` and `1.0`
stay different nodes. `324486b` writes this into `canon`'s comment.

**Rejecting unsatisfiable contracts in proc.** The kernel's `Schema.Check`
already refuses a `Required` field that is not declared, and the interpreter
only sees contracts that passed the kernel.

**A store migration.** A procedure stored before `324486b`, under a contract
with unsorted `Required`, sits under a key nothing looks up. The rule
proposed: an interpreter store from before `324486b` is re-remembered, not
re-keyed. If a store that matters exists, a one-shot re-key is small, but it
is not built speculatively. **Confirm before building.**

---

## 41 — Read, commit, execute

### Build

**V1. Verification moves into the core. Do this first.** The core gets raw
bytes for every read and checks them itself:

- **Records and Horizons:** the bytes hash to the key, *and* they are the
  canonical encoding (decode, `record.Encode`, require byte equality).
- **Inputs:** the hash only. Inputs are observations stored as received
  (invariant 3), so they have no canonical encoding to compare against.

Build this behind the existing effects before anything else moves. It is the
largest trust gain for the smallest change: today `tree.Get`
(`runtime/tree/tree.go:92`) does the hash check, so the core takes the
shell's word that a record is genuine.

**V2. Three kinds of effect, and their outcomes.** `Read`, `Commit` and
`Execute`, with the events `ReadResult`, `Committed` and `Executed`, as
41 specifies. `TryClaim` becomes a `Commit` that can return `Lost`.
**`Conflict` is an integrity fault that halts the instance, not an ordinary
outcome.** Different bytes under one content address means corruption or a
hash collision, never a race to resolve.

**V3. A handle table, and the handle-count check in the core.**
Descriptors received over SCM_RIGHTS, child processes and handshake sockets
stay in the shell's handle table. The core refers to them by index.
`Executed.Handles` reports what arrived, and the core compares it to the
Horizon's depth. The check moves out of `instance.go:143`.

**V4. Remove the four true caches.** `State` (it is `R.State`), `Stamp` (it
is `record.Stamp(R)`), `Child` (it is `Kids[0].Stamp`) and `Outs` (it is
`Instructions(Out)`) become functions. Release status leaves `Instance`
along with V3.

**V5. One full flush per cycle.** Within a cycle, commits are ordered with
`F_BARRIERFSYNC` and each directory is flushed once. One `F_FULLFSYNC`
happens at the cycle boundary, before that cycle's executes are released.
**A cycle may release several independent executes after that one flush**
(answering 41's third open question). Fan-out releases depend only on
intent that is already durable, and a full flush per release would bring
back the cost this removes.

**V6. Mutable queries are a fixed enum in the core** (answering the second
open question). There are few of them, each is a concurrency point, and an
enum keeps them named and auditable. A "mutable" flag on arbitrary keys
would invite more.

**V7. Replay checks that each resolve has exactly one resume.** The claim
still rests on the filesystem's exclusive create, so it is trusted for
prevention. Replay makes a double winner something it reports. 41's
trust-assumption table changes that row to "trusted for prevention, verified
after the fact."

**V8. Tests.**

- **A lying shell** in `machine_test.go`, in memory, with no OS. The core
  must fault on each of these:
  - flipped bytes in a record read by hash;
  - a non-canonical encoding of a valid record;
  - a valid record returned under another record's key;
  - a false discharge;
  - a handle count that does not match the Horizon's depth.

  For absence, it must exit cleanly instead.
- **Loop grammar:** a property test that every effect sequence `Step` emits
  is `Read* Commit? Execute*` per cycle, where the executes are independent
  ones (V5).
- **Equivalence:** every `skyra demo` scenario produces byte-identical
  records before and after each 41 commit, and every published website
  bundle still replays.
- **Durability:** with `SKYRA_SYNC_TRACE=1`, each cycle shows exactly one
  full flush, before its executes.
- **Measurement:** re-run the overhead sweep and the native comparison. 41's
  figures (225 → 80 ms per round trip, 250 → 110 ms per one-step run) stay
  labelled estimates until then.

**Why build these.** The diagnosis in 41 is right, and the code bears it
out: the shell makes decisions (`instance.go:143`) and verifies on the
core's behalf (`tree.go:92`). Sorting effects by what a crash leaves turns
recovery reasoning into a three-row table. One flush before execute is the
only place durability can be stated once. Together they bring back the
package comment's claim that the shell makes no decisions.

### Don't build

**Deriving Phase from disk.** Phase is not entirely a cache. "The target is
running and I am waiting for its output" is on disk nowhere, so no read can
recover it. Phase is the core's own program counter, and it is pure.
Deriving it every cycle also adds filesystem reads to every step, eating
into V5's savings. And the validity matrix in `Step` does not go away; it
becomes "what the reads found × which event arrived". **Phase stays as core
state; position is derived from disk only on recovery** (answering the
first open question).

**Making replay a trivial shell in this change.** Today replay already calls
`machine.Step` for intake (`replay.go:160`) but walks settle, gather and
claim itself. Folding it into "`Step` with a shell that only reads" is the
right end state. But during this migration the separate walker is the
**independent oracle** behind the byte-identical test. If both change at
once, a bug in the new core can agree with itself. Do it as a follow-up,
once V1–V5 have landed and the scenarios are green.

**A loser-verifies-winner claim.** A loser that reads the claim back and
records the winner still depends on the same exclusive create, so it
narrows little. V7 gives detection for less.

**"The trusted computing base shrinks to `Step`" as written.** The shell
still performs execute: it applies the sandbox the core rendered, runs the
target and reports `Executed.Obs`. A faulty shell can invent a target's
output, which the core commits as an observation. Shell and core share an
address space, so this is a discipline for containing bugs, not a security
boundary. The sentence 41 should carry is: *the shell cannot lie about
anything content-addressed; it is trusted to execute faithfully and to
report observations as received.*

**Changes to the record format, the tree layout, or what targets, adapters
and procedures see.** These were 41's non-goals, and they stay non-goals.

---

## Landing order

Each step is its own commit and keeps `scripts/check.sh` green.

1. **V1:** verification in the core, behind the existing effects.
2. **V2 and V3, one path at a time:** intake, then settle and gather, then
   handoff, then claim. Records stay byte-identical at every commit.
3. **V4:** remove the caches.
4. **V5 and V6:** the flush at the cycle boundary, and the mutable-query
   enum. Records stay byte-identical, and the sync trace shows one full
   flush per cycle.
5. **V7:** the one-resume-per-resolve replay check.
6. **E2:** kernel normalization, and the 13 contract files sorted. **This is
   the one commit where scenario records change.** Every new Horizon that
   carries one of those contracts gets a new id. The commit message lists
   the changed ids, and the pre-E2 fixture tree proves old trees still
   replay.
7. Delete the old effect and event types; re-run the overhead sweep and
   the native comparison; update `docs/performance.md`.

E2 comes last for a reason. Steps 1–5 promise byte-identical records, and
that promise is how a mistake in the loop gets caught. If E2 landed in the
middle, the equivalence test could not tell a normalization difference from
a regression.

## Open questions this leaves

- **Should `Title` leave contract identity once the interpreter stops
  routing by it?** That is worth a small proposal of its own.
- **Where should the shared normal-form test vectors live,** given two
  modules? `controller/core/contract/testdata` is suggested, with
  `controller/interp/proc` reading it by relative path.
