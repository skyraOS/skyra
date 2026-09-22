# Codex driver design

**Date:** 2026-09-11  
**Status: DESIGN DIRECTION.** These documents define the first Codex driver and
the runtime seams it pressures. They do not authorize implementation changes.

## Goal

Implement the stable Codex entity address through `codex app-server` without
allowing Codex concepts to enter scheduler or kernel semantics.

```text
Skyra Invoke / Return
        |
        v
Codex driver translation
        |
        v
Codex App Server protocol
        |
        v
possibility-scoped process and transport
```

## Documents

1. [Boundary and components](01-boundary-and-components.md)
2. [Invocation state machine](02-invocation-state-machine.md)
3. [Ingress and egress mapping](03-ingress-egress-mapping.md)
4. [Possible shared driver seams](04-shared-driver-seams.md)
5. [Executable prototype boundary](05-prototype-boundary.md)
6. [Codex native tools and containment](06-native-tools-and-containment.md)

## First implementation boundary

The first driver supports:

- one `codex app-server` process per possibility;
- one Codex thread per Codex-backed `InvocationID`;
- one active Codex turn per invocation;
- Focus projected into `dynamicTools` when the thread starts;
- Codex dynamic tool calls translated into child Skyra invocations;
- a child return translated into the corresponding Codex tool response; and
- final Codex output translated into a Skyra `Return`.

Steering, compaction, thread forking, interactive approvals, and changing Focus
inside an existing thread are deferred.

## Related decisions

- [Codex ingress boundary](../../2026-09-10/codex-ingress-boundary.md)
- [Driver composition and state](../../2026-09-10/driver-composition-and-state.md)
- [Dispatch and operand ownership](../dispatch-operand-ownership.md)
- [Entity address, driver, and transport](../../2026-09-10/entity-address-driver-transport.md)

Protocol claims in this set are based on the
[official Codex App Server documentation](https://learn.chatgpt.com/docs/app-server).
