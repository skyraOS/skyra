# Durable storage

Date: 2026-09-23  
Status: agreed in outline. Holds the contracts that [`faults.md`](faults.md)'s
interrupts key on. Builds on
[`thoughts/13-dependencies.md`](stateless-controller/thoughts/13-dependencies.md).

## Where it sits

There are three places of state. "Durable" covers two of them, which behave
differently:

| Place | Holds | Scope | Changes by |
|---|---|---|---|
| **Event log** | what happened: the exact bytes each step received | one lineage | append only |
| **Stack** | open dependencies: who waits on whom | one lineage | claims (a slot fills once) |
| **Durable storage (memory)** | what the system **knows**: input contracts, return contracts, procedures | **all lineages** | grows; entries added by claim, never edited |

Records sit between the event log and memory: they are one lineage's
conclusions. Contracts and procedures are cross-lineage knowledge. A contract
learned in one run is known in every later run.

Memory is stored next to the lineages, inside `.skyra/`.

## Layout

```text
.skyra/
  records/  inputs/  frames/      lineages (as now)
  memory/
    objects/<hash>                input contracts, return contracts, procedures
    roots/<root>                  manifests: one per version of memory (snapshots)
    head                          the current root
    next/<previous-root>          claims that advance head
```

Everything in `objects/` and `roots/` is content-addressed: its hash is its
name and its identity, and it is never edited.

## Objects

### Schema: JSON types

A schema describes shape, never values. It is canonical, so equal schemas hash
equally.

```json
{ "type": "object",
  "fields": { "text": { "type": "string" } },
  "required": ["text"] }
```

Types: `object` (with `fields` and `required`), `array` (with `items`),
`string`, `number`, `bool`, `null`.

**Fit** is structural, pure and deterministic, so replay reproduces every
verdict:

- every required field is present, with the right type;
- optional fields, if present, have the right type;
- **unknown fields do not fit.** This is strict, matching how intake already
  rejects unknown fields in a Move. A new field is a new shape, which is a
  novelty interrupt.

### Input contract: a root of knowledge

```json
// memory/objects/ic1
{ "kind": "input",
  "potential": "echo",
  "args": { "type": "object", "fields": { "text": { "type": "string" } }, "required": ["text"] } }
```

What a Potential accepts. Input contracts are the roots: everything else memory
knows hangs off one.

### Return contract: parent is an input contract

```json
// memory/objects/rc1
{ "kind": "return",
  "parent": "ic1",
  "shape": { "type": "object", "fields": { "value": { "type": "string" } }, "required": ["value"] } }
```

A return shape accepted for one input contract. `parent` is the hash of that
input contract. One input contract can have many accepted return contracts.

### Procedure: parent is an input contract

```json
// memory/objects/pr1
{ "kind": "procedure",
  "parent": "ic7",
  "steps": [ { "invoke": { "target": "…", "args": { "…": "hole:0" } } } ] }
```

The series of invocations that implements an input contract, possibly with
holes. `parent` is the hash of the input contract it implements.

```text
ic1 (echo: {text: string})
 ├── rc1  return {value: string}
 └── rc2  return {value: string, extra: string}     ← added by an accepted novelty interrupt
ic7 (…)
 └── pr1  procedure
```

## Snapshots

A **snapshot** is one hash that names the whole state of memory at a moment. A
**manifest** lists what memory knows, by hash. Its own hash is the **root**:

```json
// memory/roots/r0      (r0 = sha256 of this file)
{ "previous": "",
  "inputs": {
    "ic1": { "returns": ["rc1"], "procedure": "" }
  } }
```

Memory grows by writing a **new** manifest, never by editing one:

```json
// memory/roots/r1
{ "previous": "r0",
  "inputs": {
    "ic1": { "returns": ["rc1", "rc2"], "procedure": "" }
  } }
```

`r0` still exists, unchanged. `head` points at the current root.

The manifest is organised by input contract, matching the objects: each entry is
one input contract with its accepted returns and its procedure.

**Starts flat.** One manifest file per version. It becomes a real Merkle tree
only when memory is big enough that rewriting the whole manifest is expensive.

**Not the lineage's hash tree.** Memory roots version *knowledge*: a new root
appears whenever something is learned, and nothing waits on children. The
lineage has a different hash tree, where a parent's completion hash can only be
formed once every child has discharged (the join rule, enforced by hashing; see
[`thoughts/13`](stateless-controller/thoughts/13-dependencies.md)). Do not blur
the two.

### Advancing head is a claim

Two lineages can learn something at the same moment. Each writes its new
manifest naming its `previous`, then claims `memory/next/<previous-root>` with
`O_EXCL`. The first claim wins and moves `head`. The loser rebuilds its manifest
on top of the winner's root and claims again. No knowledge is lost, and nothing
is edited.

## Every record names the root it read

```json
{ "predecessor": "…", "input": "…", "state": "return", "memory": "r0" }
```

`step` is pure, so it cannot read memory. Before intake, `drive` reads `head`,
loads the entries the step needs from that root, and passes them in as data
(with the root). The step's verdict (fits, novel, procedure found) depends only
on its inputs, and the record commits which root it used.

**Why it matters for replay:**

- Echo returned `{value, extra}` while memory was `r0`, so intake judged it
  **novel**.
- Later, memory is `r1`, where that shape is known.
- With snapshots, replay reads `r0` because the record says so: novel again, and
  the verdicts match.
- With live reads, replay would read `r1` and call it known, which contradicts
  what happened.

A released child sees a fixed snapshot. A contract learned mid-lineage is
visible to later steps, never to one already running.

## How memory grows

| Event | Writes | New root because |
|---|---|---|
| **Genesis / registration** | the declared input contracts and their return contracts | seeds memory, so first returns are not all novel |
| **Novelty interrupt, accepted** | a return contract, `parent` = the input contract it answers | a new accepted shape |
| **Procedure interrupt, resolved** | the missing input contract (if new) and a procedure, `parent` = that input contract | a new way to call a reachable Potential |

Each write is done by the step that intakes the LLM's answer: it commits the
objects, writes the next manifest, and claims `next/<previous-root>`.

## Open

1. **What `potential` names** in an input contract: the surface's registered
   name, or its content hash (so a changed binary starts with no contracts)?
2. **Optional fields and nullability** in the schema: is `required` enough, or
   is an explicit optional marker needed?
3. **Garbage collection** of unreachable manifests and objects, once lineages
   that read them are settled and collected.
4. **Seeding format:** do surfaces declare their contracts in the registration
   (like DNA-19's `profiles.json`), written into memory at genesis as `r0`?
