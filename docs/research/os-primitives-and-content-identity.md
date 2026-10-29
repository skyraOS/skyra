# Content identity: the runtime speaks hashes

Date: 2026-10-10
Status: current direction. Decisions are settled here; nothing below is
implemented unless marked as existing. The earlier research (code-signing APIs,
audit tokens, Endpoint Security) is retained in the second half as the evidence
these decisions rest on.

## Decisions

1. **The runtime speaks only hashes.** An invocation names a binary by its
   SHA-256. Reach is a set of binary hashes, each with the schemas declared for
   it. The runtime never knows that a hash is "Claude" or "Codex".
2. **Actors resolve meaning to hashes.** Which hash `claude` currently means, or
   which hash a model's "call read" refers to, is decided by an actor. Each
   resolution is that actor's output and is therefore a recorded observation.
3. **Hashed objects are binaries, schemas, observations and records.** A hash
   covers only its own bytes. Every relationship between hashed objects is
   stated in a record.
4. **Hashes are identity, not authority.** A hash identifies, verifies integrity
   and enables search. It is public and derivable; naming one grants nothing.
   Authority is membership in the Boundary, enforced by Seatbelt.
5. **Records name both observations explicitly.** The current `Input` field
   (hash of the bytes a target emitted) becomes `output`; a new `input` field
   carries the hash of the invocation's observation. This is a new record
   version; old histories keep their bytes.

## The four hashed objects

| Object | Hash of | Role |
| --- | --- | --- |
| Binary | Exact file bytes | The implementation. Stored as `pinned/<sha256>`; the runtime executes these bytes |
| Schema | Canonical JSON schema | A type. Input and output schemas are hashed separately so composition can match an output schema to an input schema |
| Observation | Exact bytes observed | A contract filled in: an invocation's input or a target's output |
| Record | Canonical record | The only place relationships are stated |

An observation is a *filled-in contract* by virtue of the record, not its
identity. Its hash is its content alone. Two reasons:

- The emitter is never the checker. A target emits bytes; the runtime decides
  which declared schema they fit (structural fit selects exactly one). That
  judgement belongs in the record beside the observation.
- Identical outputs from different binaries must be one hash, or behavioural
  search cannot find them.

A contract (input schema plus output schema) does not get a runtime hash. The
runtime matches on binary hash plus schema fit. Actors that want a contract key
for search can derive one themselves.

## Record shape

```text
record {
  to:      <binary hash>       what ran
  input:   <observation hash>  the filled-in input
  output:  <observation hash>  the filled-in output (today's `Input`)
  fit:     <schema hash>       which declared schema the runtime matched
  boundary: <predecessor stamp> under what authority
}
```

Today a record carries `To` as a capability name, `Payload` inline and `Input`
as the hash of emitted bytes; the binary hash is reachable only by walking
`Predecessor` back to genesis `Reach`. The new version states each fact
directly. Existing `omitempty` fields let the change be versioned without
rewriting old Boundary or record bytes. Replay consumes recorded facts; it never
queries the filesystem, process table or Security framework to replay old work.

## What a binary implements is asserted, not known

A binary does not declare its APIs in a form the runtime can trust. The link
between a binary and its schemas is one more relationship, stated by whoever
provisioned it:

```text
binding { binary: <hash>, schemas: [<schema hash>...], asserted by: <actor/record> }
```

Boot asserts the contracts it knows for Claude, Codex and Skyra's own targets.
Observations then accumulate: each call either fits a declared schema or is a
recorded `contract` fault. Today these bindings are implicit in genesis `Reach`;
making them explicit records is the change.

## Updates are new binaries

A new Claude version is a new hash. It is not in reach until an actor with the
authority to add it does so, and that grant is a recorded event. No binary is
treated as "the same binary" because it occupies the same path or slot. A
persistent slot ("the Claude provider") is an actor concept, not a runtime one.

Paths are locators, never identity. `~/.local/bin/claude` is a symlink the
updater repoints across `~/.local/share/claude/versions/`; `/opt/homebrew/bin/codex`
is a Node script whose native binary changes on reinstall; `exec.LookPath`
depends on `PATH` order.

No UUID table maps names to hashes. A one-to-one UUID ↔ hash table is a second
name with a mutable source of truth. The Boundary already is the list of hashes
an execution may use, it only narrows, and it is recorded.

## Behavioural search

Hashes give a starting point for later search: matching what was produced, not
which binary produced it.

```text
(binary, input observation) → output observation
```

Two binaries agree on a schema when the same input yields the same output. That
is evidence accumulated per call, never proof; a new binary starts with none.

Whether outputs should match depends on the operation:

| Class | Key | Output equality means |
| --- | --- | --- |
| Pure | input observation | Behavioural agreement |
| Reads the environment | input observation + hashes of content read (`read` already returns one) | Agreement given the same world |
| Nondeterministic (model calls) | input observation | Nothing; searchable by input only |

The schema binding should declare its class.

The index is built from OS primitives and is always derived from records, never
authoritative:

| Need | Primitive |
| --- | --- |
| Index | Directory tree `index/<binary>/<input>/<output>`, leaves naming records |
| Publish a fact once | Exclusive link/create and directory sync, as `tree.Put` does |
| Search | `readdir` |
| Cheap duplicate content | APFS `clonefile` (optional) |

Extended attributes and Spotlight metadata are mutable, uncovered by any hash
and lost on copy; at most a hint layer.

## Consequences for current code

- `controller/runtime/instance/claude_resident.go`: `prepareClaude` resolves
  `claude` from `PATH` at every launch, follows the symlink and runs whatever it
  finds, unhashed. The registered `claude.start` surface is hashed; the program
  that actually runs is not. **This is the first change:** the Claude binary
  itself becomes the registered hash in reach, copied into `pinned/<sha256>`, and
  the runtime executes those bytes.
- `ResidentProfile` values `darwin-claude-v1` and `darwin-appserver-v1` select
  provider-specific branches inside the runtime. Under this direction, home
  directories, auth paths and profile fragments become configuration supplied by
  an actor and recorded with the call. This aligns with the
  [Boundary concern spec](../design/BOUNDARY-CONCERNS-SPEC.md).
- `controller/runtime/instance/appserver.go`: the Codex native binary is hashed
  in place under Homebrew rather than pinned, and `--version` runs outside the
  sandbox. It should be pinned like every other binary; version becomes a
  recorded fact from provisioning.
- `skyra` launcher, function `pinned`: copies only binaries under `build/`. It
  becomes the general hash → bytes store for every binary in reach.
- `controller/runtime/instance/instance.go` (`run`, `startResident`) and
  `controller/runtime/boot/boot.go`: already recheck SHA-256 before launch. They
  remain integrity checks, not authorization.

## Sequence

1. **Pin provider binaries.** Claude and Codex native binaries enter the pinned
   store by hash at registration; launch executes the pinned bytes; a changed or
   missing binary refuses with a re-provisioning message.
2. **Record version.** Add `input`, rename `Input` → `output`, add the binary
   hash and fitted schema hash. Old versions replay unchanged.
3. **Explicit bindings.** Provisioning emits binary → schemas assertions as
   records rather than only genesis `Reach`.
4. **Move provider knowledge out of the runtime.** Resident profiles become
   actor-supplied configuration, per the Boundary concern spec.
5. **Behavioural index.** Derive `index/` from existing histories, read-only,
   and test the three operation classes against real data before recording
   anything new.

Useful checks: changed binary bytes are refused; a provider update appears as a
new hash requiring a recorded grant; identical outputs from different binaries
share one observation hash; two schemas of one binary remain distinct fits; old
histories replay without reinterpretation; the runtime contains no provider
names after step 4.

---

# Evidence: OS mechanisms examined

The following research established what the OS provides and what it does not.
It informs the decisions above but does not change them.

## What the repository already does

- `controller/runtime/tree/tree.go`: records, inputs and Boundaries are addressed
  by SHA-256. `Put` writes and syncs a temporary file, publishes by exclusive
  hard link, then syncs the directory. Existing objects are reverified. `Get`
  rejects objects whose contents do not hash to their names.
- The same package creates frame results, dependencies and fan-in claims with
  write-once publication through directory descriptors and `os.Root`.
- `controller/runtime/handshake/handshake.go`: a private Unix datagram socketpair
  carries the released Boundary reference and custody descriptors via SCM_RIGHTS.
  Close-on-exec prevents accidental descriptor inheritance by targets.
- `controller/interp/proc/proc.go`: procedure bodies and nodes have content-derived
  identities.
- `controller/runtime/registry/registry.go`: registration reads each binary and
  records its SHA-256; `run` and `startResident` recheck bytes before launching
  through `sandbox-exec`.

This is already a content-addressed data model. Retain typed namespaces and
existing canonical encodings; a global object-store rewrite is not needed. Locks
and mutable pointers remain coordination state, not immutable objects.

## Code-signing identity

Apple's Security framework exposes `SecStaticCodeCreateWithPath`,
`SecStaticCodeCheckValidity`, `SecCodeCopyGuestWithAttributes`,
`SecCodeCheckValidity` and `SecCodeCopySigningInformation`. Signing information
must only be trusted after a successful validity check.
[Apple's API declarations](https://github.com/apple-oss-distributions/Security/blob/main/OSX/libsecurity_codesigning/lib/SecCode.h)

CDHash identifies a CodeDirectory, not the whole file; architecture and
hash-algorithm selection matter, and a script's content is not identified by its
interpreter's CDHash.
[TN3126](https://developer.apple.com/documentation/technotes/tn3126-inside-code-signing-hashes)

Skyra's Go targets are ad-hoc signed by the Go linker: a CodeDirectory with no
certificate (`codesign -dv build/controller/echo` reports no TeamIdentifier). For
these, CDHash restates "these bytes", which the whole-file SHA-256 already
covers more strongly. Vendor CLIs differ: the Claude binary carries a Developer
ID signature (`Identifier=com.anthropic.claude-code`, `TeamIdentifier=Q6L2SF6YDW`).

A code-signing requirement (publisher + identifier) is the OS's mechanism for
"accept any version from this signer". Under the decisions above it is not a
runtime identity: an actor may use it when deciding whether to grant a new
hash, and record the signing facts (Team ID, identifier, CDHash, architecture)
alongside the SHA-256 as provenance. The runtime still sees only the hash.

Implementation without cgo is sufficient for provisioning-time checks:
`codesign --verify --strict -R '<requirement>'` and `codesign -dvvv`. The
controller does not currently use cgo; a Security framework helper, if ever
needed on a hot path, belongs behind a build tag or in a separate binary.

### Local probe results

A temporary C executable linked against Security and CoreFoundation queried its
own running code (`SecCodeCopySelf`) and `build/controller/echo` as static code.
Both validity checks returned `errSecSuccess` without root or an ES entitlement.
Echo reported CDHash `a1ed99fa6daa907997ee96bd88bd52301bd570dc`, matching
`codesign -d --verbose=4`. This proves API availability for identity inspection,
not pre-execution verification or equivalence to AUTH_EXEC.

Probe sources were written to `/private/tmp/skyra-code-identity-research.c` and
`/private/tmp/skyra-peer-token-research.c`; copy them under
`docs/research/probes/` if the results need to be reproducible.

## Audit tokens: not a socketpair shortcut

macOS declares `LOCAL_PEERTOKEN` for a Unix socket peer's audit token, usable
with `kSecGuestAttributeAudit`.
[socket definitions](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/sys/un.h),
[audit-token attribute](https://developer.apple.com/documentation/security/ksecguestattributeaudit)

A probe reproducing Skyra's AF_UNIX/SOCK_DGRAM socketpair followed by fork read
`LOCAL_PEERTOKEN` at the parent's endpoint and got the parent's PID, not the
child's: the pair was created before inheritance. Adding `getsockopt` to the
existing handshake therefore would not authenticate the successor, and even a
correct identity would not prove which target it then launched. A socket
connected by the actual peer, or a Mach message audit trailer, is a candidate
for live identity and needs its own probe. Never accept a token supplied as
ordinary JSON.

## Reservation and custody mechanisms

| Need | Existing mechanism | Scope |
| --- | --- | --- |
| Reserve an endpoint | Bind the socket or listener and retain its descriptor | Atomic reservation rather than discover-then-lock |
| Coordinate owners | Advisory file lock where participants cooperate | Not a substitute for sandbox enforcement |
| Hand over custody | SCM_RIGHTS and close-on-exec | Already used for frame directories |
| Publish one winner | Exclusive link/create and directory sync | Already used by tree publication |
| Restrict target effects | Seatbelt profile derived from effective authority | Generic boundary; provider exceptions remain |

An endpoint address or a lock does not authorize invocation or identify code.

## The hash-check/exec race

`run` reads the file, hashes it, then executes by path. A same-uid process with
write access to the pinned path could swap bytes in between, but such a process
can already do worse. The useful defence is ensuring sandboxed targets cannot
write into `pinned/` or provider directories (the Seatbelt floor should cover
this; verify with a test), not additional identity machinery.

## Why AUTH_EXEC is not the first step

Endpoint Security offers an exec-time identity gate but requires a restricted
entitlement and operational permissions, and its clients must manage deadlines,
caching and event loss.
[Endpoint Security](https://developer.apple.com/videos/play/wwdc2020/10159/)
Launch constraints impose signing and launch policy and are not a drop-in
replacement for `exec.Command` through `sandbox-exec`.
[Launch constraints](https://developer.apple.com/documentation/security/defining-launch-environment-and-library-constraints)
Both remain optional enforcement experiments. Post-launch logging is evidence
collection, not prevention.
