# Conversation reconstruction experiment

This bundle implements the reconstruction foundation of proposal 33. It
does not replace production conversation lifecycle or sidebar routing.

`reconstruct-conversation` reads the owning user adapter's checkpoint,
invokes the Horizon's `system.strand` with its root, accepted leaves and
selected contract, then calls the user adapter's pure projection operation.
`reconstruct-state` applies that procedure to every supplied conversation
reference and assembles one full app snapshot. Empty collections work.
Neither procedure starts provider work. The reference list and live-custody
availability must come from an authorized caller in the conversation's
Horizon; cross-root management discovery is not implemented here.

Register the ordinary interpreter declaration together with this bundle:

```json
{
  "interpreter": {
    "path": "/absolute/path/to/interp",
    "contract": "/absolute/path/to/controller/interp/contract.json",
    "state": true,
    "bundle": "/absolute/path/to/controller/interp/workflows/conversations/bundle.json"
  }
}
```

Registry loading derives additional contracts from the bundle and pins its
content. Boot installs all procedures before committing genesis. The
installer validates every source before publishing any new representation;
an identical installation can be repeated after interrupted provisioning.
A conflicting representation refuses startup. `contract.json` in this
directory documents the bundle's entry contracts; do not register those
contracts separately when using automatic bundle registration.

The user adapter owns `checkpoints/<reference>.json` under `SKYRA_STATE`.
Only the root invocation, selected contract and accepted frontier persist.
Checkpoint advancement uses a locked compare-and-update and fresh syscall
evidence, retaining unfinished siblings and removing covered endpoints.
Views are reconstructed each time; no transcript or provider mapping store
is added. Runtime acceptance receipts are published after release validation
and before acknowledgment or target effects. Winning resolve claims and
frame results certify joins and supervisor discharges.

`runtime/replay.RecoverTree` is an operator recovery primitive, separate
from the bounded syscall. Given an authorized root and requester context,
it reconciles acceptance receipts and winning claims and returns fresh
evidence for the user adapter to publish. Proposed releases, orphan records,
losing claims and foreign roots do not advance its frontier. Invalid evidence
returns unresolved recovery; it never licenses an external-effect retry or
revives descriptors. Production acceptance notification, checkpoint
publication and management recovery routing still need wiring.

The tree result is version 1: root, normalized frontier, visible records,
causal order, visible Horizon metadata, opacity boundaries and supplied-leaf
coverage. Visible intake faults also carry their observed input as bytes
(base64 in JSON), including malformed emissions. Narrowed internals stay
opaque; their final boundary results remain visible. Child completion is
displayed independently of whether the parent has consumed it. Live phase
comes from explicit custody availability, never from history alone.

Validation:

```sh
go test ./controller/...
go test -race ./controller/runtime/replay ./controller/targets/user/checkpoint
cd controller/interp
go test ./...
```

`TestReconstructionProcedureThroughRealRuntime` provisions the bundle and
runs checkpoint → syscall → projection → interpreter snapshot through real
processes. Set `SKYRA_RECONSTRUCTION_FIXTURE` to export its snapshot for UI
decoder checks. Tests cover partial fan-out, winning joins, stale writers,
crash-gap reconciliation, retained-frontier stability, foreign context,
opacity, malformed/refused emissions and unavailable custody.

Remaining proposal work includes production actor progress delivery and
checkpoint integration, scoped list/load discovery and dispatch, sidebar
and command-line procedure entry points, create/send/close/delete workflows,
provider operation/session extraction, validated resource ownership,
provider-home/catalog recovery and same-root continuation/rebind. The
production launcher and both clients still use their existing routes.
