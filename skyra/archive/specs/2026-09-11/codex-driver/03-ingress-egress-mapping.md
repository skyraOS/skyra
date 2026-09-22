# Codex ingress and egress mapping

Direction is named from Skyra's perspective:

```text
Ingress: Skyra → Codex
Egress:  Codex → Skyra
```

## Ingress

| Skyra input | Codex operation | Driver responsibility |
| --- | --- | --- |
| first admitted Invoke | `thread/start` | compile Focus into a complete `dynamicTools` snapshot |
| thread created | `turn/start` | translate ordered operands into Codex input |
| child Return | response to pending `item/tool/call` | correlate child identity to the exact server request |
| cancellation | `turn/interrupt` | interrupt only the stored active turn |

`Operands[0]` is the method header and `Operands[1:]` are parameters because the
scheduler has admitted that structure. The Codex driver decides how their content
becomes Codex input.

Focus projection is discoverability, not authority. The driver keeps a private
mapping:

```text
Codex tool name → entity/address/capability
```

The kernel still checks the resulting child invocation against Horizon.

## Egress

| Codex event | Skyra effect | Notes |
| --- | --- | --- |
| `thread/start` response | replace opaque state | records `ThreadID` |
| `turn/started` | replace opaque state | records `TurnID` |
| `item/tool/call` | child Invoke | arguments become ordered observations |
| final agent-message item | replace opaque state | retains candidate return value |
| successful `turn/completed` | Return | emits the retained observation |
| failed `turn/completed` | driver failure signal | preserves provider error details in driver-owned form |
| process/transport loss | driver failure signal | may trigger later reconciliation policy |
| progress and deltas | telemetry or ignored | never mutate continuation laws |

Codex sends `item/tool/call` as a server request, so the driver must retain its
JSON-RPC request id until the child returns. Sending an ordinary notification is
not enough to resume Codex.

## Unsupported server requests

The initial driver supports dynamic tool requests only. Interactive approval,
permission, elicitation, and user-input requests are deferred. Unknown or
unsupported server requests fail closed and become explicit driver failures; they
must never receive an invented approval.

The process should start with kernel-derived restrictions that minimize the need
for interactive approval, but Codex policy is not a substitute for kernel
authority.

## Focus lifetime

Dynamic tools are experimental App Server fields. The first implementation fixes
the tool registry when it calls `thread/start`. A changed Focus requires an
explicit later design for a safe thread or turn boundary; it is not silently
patched into an active turn.

Reference: [official Codex App Server documentation](https://learn.chatgpt.com/docs/app-server).

