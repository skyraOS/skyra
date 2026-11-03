# 53 — Capabilities, information flow, and need-to-know

Date: 2026-11-03

Status: **draft, for discussion.** Nothing is built. Settles the authority
domains before the apprentice is split into modules.

**Scope:** what authority is, where each kind lives, and who enforces it.
Builds on [52](52-kernel-provisions-apprentice-spawns-proposal.md) (the
kernel provisions, `skyra-spawn` controls, the apprentice means) and
[48](48-sql-schemas-proposal.md) (contracts in a c-list). Contests the
position-free handle, the use of `system.narrow` to scope what a model is
shown, and the trace syscall as a special walk (see "What this contests").

## Summary

1. **One state, with views.** The tree (records, plans, meaning entries,
   procedure state) is one store. The kernel reads all of it. Every grant
   reads and writes it only through a view keyed by its position, like
   row-level security.
2. **Two authority domains, kept apart.**
   - **Capabilities** (the c-list): what a position may *do*. Per grant,
     bound to a position, narrowed by the kernel, enforced by the OS.
   - **Information flow control** (IFC): where a value may *go*
     (confidentiality) and what may *shape* what a node accepts (integrity).
     Global, attached to contracts and sources, derived from lineage,
     enforced at intake.
3. **Lineage is the observation's type.** Every observation carries the hash
   chain of what produced it. Flow labels are derived from that chain, never
   stored and never claimed. The model never sees a lineage hash.
4. **Both flow rules are allowlists, opt-in, and fail closed.** A sink lists
   the sources it accepts. A source lists the hashes that may read it.
5. **Need-to-know (focus) is not authority.** It is what the apprentice puts
   in one model step's context. It is recorded, because a model step's output
   has everything in its context in its lineage.
6. **No release privileges for now.** Nothing declassifies or endorses at
   runtime. Only the operator edits an allowlist, outside the runtime.

## Why two domains

Access control decides who may touch what. It cannot follow a value once it
has been read: that is the confinement problem (Lampson, 1973). A grocery
procedure that may call `bank.login` may also call `web.post`, and with
access control alone, nothing stops the password going to the second.
Information flow control follows the value.

The two answer different questions and are attached to different things:

| | Capabilities | Information flow |
|---|---|---|
| Question | what may this position do | where may this value go; what may enter this node |
| Attached to | a position in the tree | a value's lineage, checked against a contract or source |
| Scope | per grant | global |
| Granted by | the kernel (`narrow`) | the operator, at registration |
| Enforced by | the OS (sandbox, held fds) and the view over state | the kernel / `skyra-spawn` at intake |
| Narrows | going down the tree | never relaxes at runtime |
| Literature | KeyKOS, EROS, seL4 | Denning's lattice; Bell–LaPadula, Biba; DIFC (Flume, HiStar) |

Routing a reference needs neither. A procedure, the apprentice and a focus
move references they cannot read, so composing them needs no new grant.
Authority is checked at the edges: at what a position may invoke, and at who
may read a value.

## One state, with views

| Row-level security | Skyra |
|---|---|
| the table | the tree |
| the `position` column | the plan stamp a row was written at |
| the session role | the grant bound to a socket or fd |
| `USING` (read) | rows at the grant's position, and the invoke and return rows of its direct children |
| `WITH CHECK` (write) | append-only, at the grant's own position; the position is stamped by the kernel, never supplied |
| `BYPASSRLS` | the kernel: provisioning, replay, recovery |

- **A narrowed child stays opaque.** Its parent sees one call and its return,
  as the trace syscall shows it today. Otherwise narrowing hides nothing, and
  `remember` would compile across grants.
- **No read by raw hash.** Objects are content-addressed and shared. A grant
  that can `get(hash)` can test whether another position holds given bytes.
  Reads go through what is reachable from the grant's position; sharing stays
  a storage detail.
- **The apprentice's state is rows.** `Position`, `Remembering` and the
  handle table become rows written at the apprentice's position. Recovery is
  reading them back.
- **`trace` becomes a read.** Reading the subtree under a position you hold
  replaces the walk in `replay.ReadTrace` that climbs and refuses at
  `origin`; the refusal becomes the read policy.
- **Publishing is the one cross-position write.** A remembered procedure is
  learned at one position and used from many. `propose` moves it from the
  apprentice's position into the scope of every position that holds the
  contract. It is the only syscall that writes outside the caller's own
  position.

## Lineage

Every observation's lineage is the hash chain of the records that produced it.
The records already exist and are already chained, so lineage costs nothing
more to keep.

- **Leaves are sources.** A source is an observation that brings bytes in
  from outside the tree: user input, a file read, a network response, a
  model's output. Everything between the leaves and the value (procedure
  nodes, the apprentice, `skyra-spawn`) is a transformer and contributes no
  leaf of its own.
- **A model step's leaves are everything it was shown.** That is everything in
  its focus, and everything its session saw before. A long-lived session's
  lineage only grows; a fresh session (`fresh: true`, which `delegate` already
  sets) starts empty. This makes fresh sessions a security property.
- **A compiled procedure keeps exact lineage.** `remember` binds every
  argument by provenance: the input, an earlier call's return, or a
  constant. A mechanical step's leaves are exactly the sources its bindings
  name. A habit is therefore more trustworthy than the model run it was
  learned from: it carries the account number only from the input, even if
  the same run read a web page.
- **The model never sees lineage.** A model could forge a hash it was shown.
  It sees references and presentations.

## Information flow

Both rules are allowlists over the leaves of a value's lineage. Both are
checked by the kernel, or by `skyra-spawn` under 52, at intake: on the edge
into a node. Neither is checked by the node itself, which cannot see lineage
and, if compromised, would skip the check (invariant 1).

| | Owned by | Allowlist | Check |
|---|---|---|---|
| **integrity** | the sink, on its contract | the sources whose values may enter it | every leaf of the value's lineage is listed |
| **confidentiality** | the source | the hashes that may read anything derived from it | the reader is on the allowlist of every source among the leaves (the intersection) |

### Integrity: what may enter me

`bank.transfer` lists the sources it accepts: the user's own input, and the
vault's account book. A transfer whose amount or payee has a web page among
its leaves is refused, whether or not anyone labeled the web page. Sources
nobody thought of are denied by default, which is why the rule is an
allowlist.

The common case this stops is prompt injection. Both calls below fit the
contract; only lineage tells them apart:

| Payload | Leaves | Result |
|---|---|---|
| `{"amount": 500, "to": "acct-8812"}` | the user's message, the account book | accepted |
| `{"amount": 500, "to": "acct-8812"}` | the user's message, a product page | refused |

### Confidentiality: where I may go

The vault lists who may read the bank password: the `bank.login` binary, by
hash, with its contract, and its egress pinned to the bank's domain. Holding
a reference to the password is not reading it.

- **Sealed references.** A focus shows `password: ⟨sealed⟩`. A procedure
  binds it by provenance. The apprentice routes it. Only `skyra-spawn`
  dereferences it, when it writes the input memfd of a node on the
  password's allowlist (52, `input`).
- **A rebuilt binary is not on the list.** The password stops flowing until
  the operator lists the new hash, as binary pins work today.
- **Mixed values get the intersection.** A value derived from the password and
  the statement may be read only where both may be, which is usually nowhere.

### Opt-in

A sink with no integrity allowlist accepts anything. A source with no
confidentiality allowlist flows freely. Flow checks cost nothing until a
policy exists: only an edge into a sink with a policy, or a value with a
listed source among its leaves, walks a lineage. Policies can be added later
without migrating anything, because the lineage is already recorded.

The intended users are sensitive sources (credentials, finance, health) and
high-stakes sinks (moving money, sending mail, deleting, publishing).

### Who edits an allowlist

The operator only, at registration or in configuration, like pinning a
binary. No syscall edits one, and no grant can hold the right to. Without
release privileges, editing an allowlist is the only way to relax a rule, so
it must not be reachable from inside the runtime.

A narrowing may **strengthen** a sink's integrity allowlist for its child
(fewer sources), never weaken it.

### No release privileges

Declassification (a value derived from a source may be read by one more
reader) and endorsement (a value with an unlisted leaf may enter a sink) are
left out. Each would be a capability in a c-list, scoped and attenuable, and
each is an attack surface. A flow violation is a refusal: the node faults
and the procedure falls back like any other refusal. Nothing asks the human
mid-run.

## Need-to-know (focus)

Focus is what one model step is shown: a projection of the state its position
can read, computed by the apprentice. It is not authority:

- it is free to create and nest, with no syscall;
- a model step emits within it, and the apprentice refuses an emission
  outside it, as a refusal of its own, distinct from a kernel fault;
- it sits inside the capabilities of the position it runs at, and cannot widen
  them.

It matters to flow because it decides lineage. What goes into a focus is
among the leaves of everything the step emits. So the apprentice records each
focus (in the meaning log under 52), and a narrow focus is what keeps a model
step's output clean: a grocery step that never had the statement in focus
emits values the statement's allowlist does not restrict.

## The layers

```
need-to-know   what one model step is shown; decides the leaves of its output    the apprentice
procedures     route references between contracts; need no authority             the apprentice
information    integrity allowlists on sinks, confidentiality on sources,         the kernel / skyra-spawn
flow           over the leaves of lineage; checked at intake
capabilities   what a position may do; narrowed by the kernel                     the kernel, the OS
```

A runaway procedure meets them in order:

1. **Refinements** (CEL edges): it notices it is off the path it was learned
   on and falls back to the model.
2. **Capabilities**: whatever it does, it cannot reach past its grant.
3. **Information flow**: inside its grant, a sensitive value cannot reach the
   wrong reader, and an unlisted source cannot shape a high-stakes sink.

## Under 52

52 lets bytes move between nodes on channels with no checker. Flow survives
that at the granularity of a process:

- **At `check`,** `skyra-spawn` computes each channel's leaves from the plan:
  a node's output has as leaves the union of everything it reads (its
  `input`, its upstream channels) and, for a model node, its session. The plan
  is a DAG, so this is static and conservative, as Flume labels processes.
- **A plan is refused whole** if any edge would deliver a value to a node
  whose integrity allowlist does not cover its leaves, or to a reader not on
  a confidentiality allowlist.
- **A sealed value is delivered only through `input`,** which `skyra-spawn`
  writes. It never goes on a channel, where nothing could check who reads it.

## What this contests

- **Position-free handles.** `reprovisions` lets any Boundary that could have
  provisioned a grant invoke its handle. Here a grant is bound to a position
  (52's socket), and revocation follows derivation. The apprentice's handle
  table goes.
- **Narrowing to scope a model.** `delegate` calls `system.narrow` mostly to
  cut down a model's tool list. That is focus, done with a syscall. A child
  is narrowed only when it must not be trusted to stay in focus.
- **The c-list as tool list (48).** What a position holds and what a model is
  shown are different sets. The c-list stays authority; the tool list is the
  focus.
- **`trace` as a special walk.** It becomes a read of a subtree the requester
  holds.

## A leak today

`remember` receives the whole trace, payloads included, and `refineAsk` shows
the model `command`, `rounds` (every payload and value) and `constants`. A
secret that passed through a remembered run would be compiled into the
procedure as a constant and shown back to the model when it writes its
refinements. Until sealed references exist, `remember` should refuse to
compile a trace that touched a source with a confidentiality allowlist.

## Decisions to make

1. **What counts as a source.** User input, file reads, network responses and
   model output, as above; or every target marked a source at registration,
   as `continuation` is marked today. Recommendation: marked at registration,
   with those four as the defaults.
2. **Lineage granularity.** Per process (a node's output has all its inputs as
   leaves), or per value (a compiled procedure's provenance bindings give each
   field its own leaves). Recommendation: per process under 52; per value
   inside compiled procedures, where it is free.
3. **What an allowlist entry names.** A binary hash; a binary hash and a
   contract; or those and an egress allowlist. Recommendation: all three for
   confidentiality, since the reader's egress is where a secret leaves;
   sources by registration name and contract for integrity.
4. **Focus enforcement.** Whether the apprentice refuses an emission outside
   the focus (a refusal) or only records it (the focus is a hint, the lineage
   still exact). Recommendation: refuse.

## Later

- **Release privileges:** declassify and endorse as scoped, single-use
  capabilities minted from a human approval. Where "ask the human" would
  live: not "may I use this tool", but "may this value cross this line".
- **Value-level labels for model steps:** a model whose output is split into
  fields with separate leaves, in the style of CaMeL's quarantined model
  (Debenedetti et al., 2025).
