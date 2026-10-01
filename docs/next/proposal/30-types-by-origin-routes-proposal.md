# 30 — Types by origin, and routes: proposal

Status (2026-10-01): **Not built.** Section F of the [list](README.md).

The implementation plan for [30](../30-types-by-origin-routes.md). Read 30
first; this page says how to build it, in what order, and what to ask before
building. Ids refer to section F of the list.

## Ground rules

- **The controller stays standard-library Go.** The root `go.mod` (module
  `skyra`) has no requirements, and keeps none. `go.starlark.net` lives only
  in the route runner, which is its own module (F4).
- **No record encoding change.** Improvements item 9 (E) is not built, and
  every stamp depends on `record.Encode`. Types are **derived**, never stored
  in a `record.Record`: anything that needs a return's type recomputes it from
  the record's payload and the child's surface. New data (route sources,
  route tables) goes in the tree under new kinds, through `tree.Put`.
- **Hashes never reach a model.** Anything added to `Continuation` or
  `StepView` (what an LLM continuation reads) is a presentation, never a hash.
- **Replay stays byte-identical.** Every `run.sh` scenario ends in a replay
  check; it must still pass after every step.
- **Tests:** `go test ./controller/... ./tui/...` from the repo root. Not
  `./...`: the git-ignored `worktrees/` holds an archive that does not build.
  `gofmt` and `go vet` clean.
- **Commits** follow the repo: `controller: ...`, `docs: ...`. After each
  step, update the Status column of section F in the [list](README.md) and the
  status line of [30](../30-types-by-origin-routes.md), with the commit.

## Decide first

These change what gets built. Each has a proposed default; **confirm with the
human before the step that needs it**, and do not guess past one.

| # | Decision | Proposed default | Needed by |
|---|---|---|---|
| 1 | Scalar kinds in a shape | JSON's four (string, number, bool, null), as 30's encoding says | F2 |
| 2 | An array's shape | the distinct element shapes, sorted; independent of length and order | F2 |
| 3 | The actor's identity | the hash of the child's canonical `horizon.Surface` encoding: covers binaries (`sha256`) and host surfaces alike, and any change to a registration is a new identity | F3 |
| 4 | When routes run | at gather, after the join rule is met: see F6 | F6 |
| 5 | What a send's return does | it is routed again, up to a fixed chain depth (8), then it wakes the parent | F6 |
| 6 | Which previous return | the last return of this (binding, type) in the same frame; across frames later | F6 |
| 7 | `for` in routes | allowed (Starlark's default; it terminates over a finite value) | F4 |

## F1: the hash audit

**Deliverable:** a table in [30](../30-types-by-origin-routes.md) ("Hashes never
reach a model") with every place a hash reaches a model or an external actor,
and a fix or an accepted exception for each. No runtime change unless one
leaks.

Where to look:
- `machine.runOf` (`core/machine/machine.go`): `SKYRA_FRAME` in the
  environment. Confirm no adapter forwards it into a prompt.
- `machine.Continuation.Frame` and `StepView`: what each continuation program
  does with them. `targets/agent`, `targets/user`, `codex/adapter`,
  `codex/resident`: grep each for where stdin fields reach model-facing text
  (`dynamicTools`, turn input, tool results).
- `hosts/web`: what the browser sees (frame ids in URLs and responses). The
  human is not a model, but the page is outside the boundary; record it.
- Observation hashes in payloads (doc 21): none are passed today; confirm.

## F2: shapes, contents, diffs, presentations

Pure functions over the value `contract.Decode` returns. New file
`core/contract/shape.go`. No machine change.

```go
func Shape(v any) string      // hex sha256 of the shape encoding
func Content(v any) string    // hex sha256 of the content encoding
func Present(v any) string    // the shape as text: {media: [{url: string}], text: string}
func Diff(a, b any) Delta

type Delta struct {
	Changed []string // paths whose content differs, sorted ("$.text", "$.media[0].url")
	Added   []string // paths in b and not in a
	Removed []string // paths in a and not in b
	Leaves  int      // leaves in b
	Ratio   float64  // len(Changed) / Leaves; 0 when Leaves is 0
}
```

The encoding, exactly (so it never drifts):
- every node hash is `sha256(tag ‖ body)`; tags are one byte: `s n b z o a`;
- every variable-length part is prefixed by its length as 8 bytes big-endian;
- a child is its 32 raw hash bytes, not hex;
- objects: the count, then for each key sorted bytewise: `len(k) ‖ k ‖ child`;
- shape arrays: the count of distinct element shapes, then those shapes sorted
  (decision 2); content arrays: the length, then each element's content in
  order;
- numbers: content is the `json.Number` text as written (`1` and `1.0`
  differ in content, not in shape);
- the root is domain-separated: `Shape` returns
  `sha256("skyra/shape/v1" ‖ root)`, `Content` returns
  `sha256("skyra/content/v1" ‖ root)`.

`Diff` walks both values together, skipping subtrees whose content hashes are
equal. Array elements are compared by index. Paths use `$`, `.key` and `[i]`;
a key that is not an identifier is written `["key"]`.

`Present` prints keys sorted, scalars by kind, arrays as `[T]`, and arrays of
several element shapes as `[T | U]`.

**Tests** (`shape_test.go`):
- injectivity: `{"ab":1}` vs `{"a":…,"b":…}`; `"1"` vs `1`; `[]` vs `{}`;
  `[[]]` vs `[]`; an empty key;
- key order does not change either hash;
- shape is independent of leaf values and array length; content is not;
- **golden hashes** for a handful of values, pinned in the test, so any
  change to the encoding is a test failure, not a silent drift;
- `Diff`: an unchanged value; one leaf changed; a key added and removed; a
  changed array element; `Ratio`.

## F3: the type of a return

```go
// core/contract
func Type(actor, contract, shape string) string // sha256("skyra/type/v1" ‖ len‖actor ‖ len‖contract ‖ len‖shape)
```

- `actor`: decision 3, the hash of the child's canonical `horizon.Surface`.
- `contract`: `record.Sum(contract.Canonical(surface.Input))`. Until facets
  (D4) a surface has one contract, so this is implied by the actor, but keep it
  in the hash so facets do not change the formula.
- `shape`: `Shape(payload)`.

Add a helper in `core/machine` that derives a discharge's type from what a
round already holds (`Settled.Rounds[…].Children[…]`: the invoke's `To`, the
discharge record, and the child's Horizon from the tree). It is derived on
demand, not stored (ground rules).

The presentation shown to a model is `<invoke To>.<contract name> → Present(payload)`;
until facets, `<invoke To> → Present(payload)`.

**Tests:** two surfaces with the same name and different `sha256` give
different types for the same payload; the same surface and two payloads with
one shape give one type.

## F4: the route runner

A runtime-owned surface, like `system.narrow`, but with a binary: it runs
Starlark, so it cannot live in the controller.

- **Module:** `controller/targets/route/` with its own `go.mod`
  (`module skyra/route`, requiring `go.starlark.net`). `run.sh` builds it with
  `(cd targets/route && go build -o ../../bin/route .)`.
- **Registered by the runtime, not the registry:** add it next to
  `NarrowFacet` in `core/machine/room.go`, with a pinned path and `sha256`. The
  registry refuses the name, as it refuses `system.*`.
- **Its Horizon is empty:** no reach, no grants. The floor and its own exec
  only (B7).
- **Input** (stdin):
  ```json
  {"source": "<starlark>", "return": <json>, "previous": <json or null>, "targets": ["name", ...]}
  ```
- **Output:** it returns to its slot (`SKYRA_FRAME`) with one decision:
  ```json
  {"wake": true} | {"send": {"to": "<name>", "payload": <json>}} | {"fault": "<cause>"}
  ```
  Its output contract says exactly that.
- **Starlark:** `syntax.FileOptions{}` (no `While`, no `Recursion`, no
  `TopLevelControl`); `thread.Load` is nil, so `load()` fails;
  `thread.SetMaxExecutionSteps` set to a named constant (start at 100000).
  The source must define `def route():`; the runner calls it once and requires
  one decision back. Decision 7 says whether `for` stays allowed.
- **The `skyra` module** (`starlarkstruct.Module`, predeclared): `get(path)`,
  `previous()`, `changed(path)`, `diff()`, `wake()`, `send(to, payload)`,
  `fault(cause)`. `send` refuses a `to` not in `targets`. JSON converts to
  frozen Starlark values (dict, list, string, int or float, bool, None) and
  back; `changed` and `diff` use a `Diff` with the same semantics as F2's
  (copy the function into this module, with a test that pins both against the
  same cases, since the two modules cannot import each other cleanly).
- **Tests:** each builtin; `while` and recursion rejected at compile; a
  runaway route stopped by the step cap and returned as a fault decision; the
  same input run twice gives byte-identical output; `load` fails; a route that
  returns nothing, or two decisions, is a fault decision.

## F5: routes declared at registration

For deterministic parents, which cannot answer an interrupt.

- A registry entry may carry `"routes": {"<child name>": "<path to .star>"}`.
  The loader reads each file, `tree.Put`s it under kind `routes`, and puts the
  hash in a new `horizon.Surface` field, `Routes map[string]string` (child name
  → source hash), `omitempty` so existing Horizon encodings do not change.
- A declared route covers **every type** at its binding: the parent cannot
  know types in advance, so the route branches on content with `if`.
- Lookup order at a discharge: a route for (binding, type) (F7), then the
  binding's declared route, then novelty.

## F6: routing at discharge

The step that changes `core/machine`. **Confirm decisions 4–6 first.**

Proposed (decision 4): routes run at **gather**, after the join rule is met,
so the join rule is unchanged.

1. `gather` has every discharge of the round. For each child that **returned**
   (faults are not routed in v1) and whose binding has a route, the runtime
   issues a new effect, `RunRoute{child, source, return, previous, targets}`,
   which runs the route runner under its empty Horizon. Its record is
   committed like any step, so replay re-derives it.
2. Decisions:
   - **wake**: the discharge is delivered to the parent as today.
   - **send**: the runtime emits the envelope as an invoke under the
     **parent's** Horizon (`interpret` with the parent's Horizon, so a `to`
     outside the parent's reach is a horizon fault). Its return is routed
     again, up to the chain depth (decision 5); at the depth it wakes the
     parent.
   - **fault**: the discharge is delivered to the parent as a fault carrying
     the return as its observation, and the route's cause.
3. The parent resumes when at least one discharge in the round wakes it, or
   every chain has ended. A round whose every discharge was sent on does not
   resume the parent until those sends come back.
4. A child with no route is delivered as today; F7 makes it novel.

`previous` (decision 6): the last earlier discharge in the same frame with the
same binding and type, read from the frame's rounds; `null` if none.

**Tests:** a declared route that wakes (behaviour identical to today); one that
sends to `echo` and wakes on echo's return; one that faults; a send outside
the parent's reach (horizon fault); a chain that hits the depth; replay of
each, byte-identical. Add a `run.sh` scenario for the send case.

## F7: the novelty interrupt and `system.route`

- **Novelty:** a returned discharge with no route for (binding, type) and no
  declared route wakes the parent, and its `StepView` gains two fields:
  `"novel": true` and `"type": "<presentation>"`. Never the hash.
- **The route contract** (30, "What the model sees"): when any step in a
  continuation's input is novel, the `Continuation` also carries a `route`
  field, built by the runtime, holding the five parts:
  ```json
  "route": {
    "conditions": {"get": "path -> value of that path's kind", "previous": "-> the previous return of this type, or None",
                   "changed": "path -> bool", "diff": "-> float, the ratio of changed leaves"},
    "decisions":  {"wake": "-> deliver the return to you", "send": "to, payload -> send to a name in your Horizon",
                   "fault": "cause -> refuse, carrying the return"},
    "horizon":    {"reach": {"<name>": <input contract>, ...}, "refs": ["<name>", ...],
                   "authority": {"read": [{"path", "scope"}], "write": [...], "mach": [...], "connect": [...]}},
    "faults":     ["malformed", "horizon", "contract", "custody", "leaf"],
    "interrupts": ["novelty"]
  }
  ```
  The return itself (part 1) is the novel step's `type`. Build `horizon` from
  `horizon.Horizon` by copying `Reach` (names and `Input` only), the names of
  `Refs`, and `Authority`; **strip every hash**: `Stamp`, `Frame`, `Resume`,
  each surface's `Path`, `SHA256` and proxy fields, each reference's stamp. A
  test asserts that no 64-hex string appears anywhere in a `route` field built
  from a real Horizon. Show the whole Horizon for now (decided 2026-10-01);
  narrowing what a model sees is later.
- **Adapters pass it on.** `targets/agent` and the Codex adapter put the route
  contract into what the model reads when a step is novel (for Codex: the turn
  input), and offer `system.route` as a tool. F1's audit covers both.
- **`system.route`:** a new room facet beside `system.narrow`
  (`core/machine/room.go`), unregistrable. Input:
  ```json
  {"step": <index into the continuation's steps>, "source": "<starlark>"}
  ```
  The runtime resolves `step` to (binding, type) from the frame's rounds, so
  the model names a step it was shown, never a hash. It stores the source
  under kind `routes`, records the route under the key
  `(parent actor identity, binding, type)` in the tree, and returns `{"ok": true}`.
  A route that fails to compile is caught on its first run (a fault decision),
  which leaves the type novel again.
- **Scope of a route:** keyed by the parent's actor identity, so it outlives
  the frame: "next time" means the next return of that type to any frame of
  that parent.
- **Exposure:** `system.route` is in a continuation's reach the way
  `system.narrow` is.
- **The demo agent** (`targets/agent`): a scenario where it receives a novel
  `echo` return, writes a route that wakes only when `$.text` changes, and a
  second call with the same text is not delivered.

## F8: the log as the test suite

When `system.route` is invoked for a key that already has a route, the runtime
runs the new source over every recorded route run for that key (same return,
same previous) before accepting it. If any decision differs, `system.route`
returns the flipped cases (as presentations and step indexes) and refuses,
unless the payload carries `"accept": [<case indexes>]` naming every flip.
Accepted flips are recorded with the parent as author.

## F9: the output contract goes

After F6 and F7: `"output"` becomes optional in `contract.json`, and
`interpret` stops requiring a return to fit it. A surface that still declares
one keeps the structural check, until every target has moved. Remove
`Surface.Output` last, in its own commit.

## Not in this plan

- Hashes in payloads as observation references (doc 21) and their reach.
- Across-frame `previous` (decision 6), which needs an index over the log.
- How much a string changed (edit distance).
- Input types computed instead of declared, and novelty at an input (29).
- Storing types in records: with E.
