# Delegation

Codex and Claude can call the apprentice's `delegate` procedure to give a goal
to a fresh child agent and receive its observations. One invocation starts one
child and waits for its result.

```json
{"delegate":{"goal":"Investigate how tool failures reach the caller."}}
```

Omitting `actor` preserves the caller's actor and resolved model: the adapter
fills both in, since it knows which actor it is and which model it resolved.
An explicit actor selects that actor's provider default, even when it names the
same actor. The tool description lists the actors in reach whose contract a
goal fits; nothing marks an actor as an agent. An actor outside reach is
rejected before child execution.

```json
{"delegate":{"goal":"Review this change for bugs.","actor":"claude","tools":["read","find"]}}
```

The child receives the goal as its only task context. Parent history, images,
thread IDs, and saved sessions are not forwarded. Execution instructions and
tool contracts still come from the runtime. The child's input carries
`fresh: true`, and its adapter starts a new session: Claude neither resumes nor
replaces the parent's saved session. Each child gets its own provider edge and
sandbox, including when its parent also runs through the apprentice.

## Tools and authority

- Omit `tools` to inherit the caller's effective authority and everything it holds.
- Supply `tools` to hold only the contracts behind those tools.
- Supply `tools: []` for a child that holds only `system`, its own actor and
  the actors that actor runs through.
- Use the names the caller was offered, such as `repo_write`, or canonical
  identifiers such as `repo.write`. Labels the caller was not offered are rejected.

The adapter translates the names into contracts, and the kernel narrows the
child's Boundary to hold exactly those (proposal 48's c-list): `system.narrow`
takes `contracts: [{target, schema}]`, never tool labels. Every invocation, by
name or through a reference, is checked at intake against what is held.
Calling another procedure, exporting a contract, or delegating again cannot
widen it. A procedure whose contract is held still holds none of its
dependencies, and fails when it invokes one that is not. An explicit subset must
include `delegate` for further delegation to be offered, and delegating to
another actor needs that actor held too.

The actors an adapter runs through (`codex.open`, `claude.exchange`, …) stay in
reach and held through every narrowing, as `system` does; they are not offered
to a model, and cannot be requested in `tools`. Fresh Codex threads disable native
image viewing and web search; their provider sandbox also denies direct workspace
and conversation-scratch image reads. Claude native tools remain disabled.

## Results and failures

A successful child returns its existing observation representation, usually a
payload containing `text`. The caller supplies no completion test or output
schema. Findings can include uncertainty or incomplete work.

Ordinary recorded child faults and provider refusals return a failure outcome:

```json
{"failed":{"kind":"custody","cause":"provider exited before completing"}}
```

Both transports present that outcome as an unsuccessful tool result and let the
caller continue. Intake refusals, such as an unavailable tool, can instead arrive
as a normal recorded tool fault. Integrity and persistence failures retain the
runtime's stop behavior; delegation does not synthesize observations for an
execution whose result cannot be established.

The parent's provider connection remains alive while its recorded agent frame
is unfinished, including while it waits for a child. Once the frame settles,
ordinary idle reclamation resumes. A completed fresh Codex turn records a
provider stop before returning; Claude uses a one-shot process that exits after
its result. The adapters retain their 2,000-step limit;
every narrowing, delegation included, is refused when the runtime stack already
holds 64 frames.
Batch dispatch, detached children, and concurrent-write coordination are outside
this version's scope.

## Operator configuration

The `skyra` launcher configures this automatically. A custom registry lists,
for each binary, the actors it runs through:

```json
{
  "codex": {
    "path": "codexresident",
    "contract": "controller/codex/resident/contract.json",
    "uses": ["codex.open", "codex.exchange", "codex.inspect", "codex.stop"]
  }
}
```

This excerpt omits the transport registrations and OS authority grants; keep the
usual registrations and boundaries. Each name in `uses` must be registered. No
entry is marked as an agent or a work tool: what a model is offered is its
adapter's catalog, read from the continuation's `reach` (every contract each
actor declares, and which are held). The apprentice's contract declares
`delegate` alongside `remember`, and `think` retains the apprentice so its
worker can delegate.

The model is the adapter's: it passes its resolved model in the `delegate`
payload, and the apprentice passes it to the child in its input as `model`.
The kernel carries no model, and a model being run cannot set the field: the
adapter hides it from the tool schema and overwrites it.

## Verification

```sh
go test -race ./controller/core/... ./controller/codex/... ./controller/claude/... ./controller/runtime/... ./controller/skyra
(cd controller/apprentice && go test -race ./...)
python3 controller/experiments/appserver/tool_surface.py --delegate
```

The macOS runtime acceptance test uses deterministic subscription transport
fixtures while exercising real process boundaries, Seatbelt sandboxes, custody,
parent resumption, and byte-identical replay. It covers both caller actors,
cross-actor selection, goal-only context, empty and inherited tools, nested
delegation, widening attempts, and child exits. The installed-Codex capture probe
uses a local test server to verify the actual inference tool surface without
making an external model call. Live subscription inference remains separate.
