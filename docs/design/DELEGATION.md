# Delegation

Codex and Claude can call the interpreter's `delegate` procedure to give a goal
to a fresh child agent and receive its observations. One invocation starts one
child and waits for its result.

```json
{"delegate":{"goal":"Investigate how tool failures reach the caller."}}
```

Omitting `actor` preserves the caller's actor and resolved model. An explicit
actor selects that actor's configured provider default, even when it names the
same actor. The tool description lists operator-configured available actors.
Unknown actors and missing provider operations are rejected before child execution.

```json
{"delegate":{"goal":"Review this change for bugs.","actor":"claude","tools":["read","find"]}}
```

The child receives the goal as its only task context. Parent history, images,
thread IDs, and saved sessions are not forwarded. Execution instructions and
tool contracts still come from the runtime. Claude children neither resume nor
replace the parent's saved session. Each child gets its own provider edge and
sandbox, including when its parent also runs through the interpreter.

## Tools and authority

- Omit `tools` to inherit the caller's effective authority and available work tools.
- Supply `tools` to permit only those available work tools.
- Supply `tools: []` for a child that reasons without invoking work tools.
- Use canonical identifiers such as `repo.write`, or their transport labels such
  as `repo_write`. Unknown and ambiguous labels are rejected.

The runtime records a ceiling of concrete `(actor, contract)` bindings and checks
each invocation against it. Calling another procedure, adding a contract, or
delegating again cannot expand that ceiling. A procedure that needs excluded
tools will fail when it attempts to invoke them; its name alone grants none of
its dependencies. An explicit subset must include `delegate` for further
model-facing delegation to be offered.

Agent transport operations remain available to their adapters and are not work
tools. They cannot be requested in `tools`. Delegated Codex threads disable native
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
ordinary idle reclamation resumes. A completed delegated Codex turn records a
provider stop before returning; Claude uses a one-shot process that exits after
its result. The adapters retain their 2,000-step limit;
delegation also refuses to launch when the runtime stack already exceeds 64 frames.
Batch dispatch, detached children, and concurrent-write coordination are outside
this version's scope.

## Operator configuration

The `skyra` launcher configures these fields automatically. Custom registries
must mark agent adapters and selectable work tools explicitly:

```json
{
  "codex": {
    "path": "codexresident",
    "contract": "controller/codex/resident/contract.json",
    "agent": {
      "operations": ["codex.start", "codex.exchange", "codex.inspect", "codex.stop"]
    }
  },
  "read": {
    "path": "read",
    "contract": "controller/targets/read/contract.json",
    "work_tool": true
  }
}
```

This excerpt omits the transport registrations and OS authority grants; keep the
usual registrations and boundaries. An agent must be a continuation, and its
listed operations must be registered resident or connection surfaces. Authored
procedure bundles are marked as work tools automatically. The interpreter's
contract declares `delegate` alongside `remember`, and `think` retains the
interpreter so its worker can delegate.

The agent adapter records its resolved model through `system.agent-context`.
That metadata is derived from the registered emitter and carried through the
Boundary. It is execution metadata, not parent conversation context or a public
`delegate.model` argument.

## Verification

```sh
go test -race ./controller/core/... ./controller/codex/... ./controller/claude/... ./controller/runtime/... ./controller/skyra
(cd controller/interp && go test -race ./...)
python3 controller/experiments/appserver/tool_surface.py --delegate
```

The macOS runtime acceptance test uses deterministic subscription transport
fixtures while exercising real process boundaries, Seatbelt sandboxes, custody,
parent resumption, and byte-identical replay. It covers both caller actors,
cross-actor selection, goal-only context, empty and inherited tools, nested
delegation, widening attempts, and child exits. The installed-Codex capture probe
uses a local test server to verify the actual inference tool surface without
making an external model call. Live subscription inference remains separate.
