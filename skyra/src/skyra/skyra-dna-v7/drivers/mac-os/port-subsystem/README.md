# port-subsystem

**Date:** 2026-08-20
**Status: THROWAWAY PARTIAL.** `fabric.go` implements synchronous APPROACH and normalized local
ingress. External sockets, acknowledgment, and retry remain open.

> **One logical egress: the APPROACH socket. Two ingress paths: local and external. Both ingress paths
> become the same arrival before the CTU sees them.**

`APPROACH` is the provisional name of the egress socket, not an opcode and not ACTIVATE. A local CTU
submits every outbound MOVE to it with the live connection MOVE requires. The connection contains
the destination address; the host network stack selects loopback or an external interface.

```text
                                      ┌─→ LOCAL INGRESS ────┐
CTU ─→ APPROACH SOCKET ─→ HOST STACK ─┤                      ├─→ NORMALIZE ─→ CTU
                                      └─→ remote host ─→ EXTERNAL INGRESS
```

For a local connection, the host routes into local ingress. For a remote connection, it routes to
the destination runtime's external ingress. A CTU has one submission operation either way;
locality is connection and host-network state, not execution semantics.

```text
APPROACH EGRESS         local-only submission socket. accepts outbound MOVEs from CTUs.
HOST STACK         routes the connection's destination through loopback or an external interface.
LOCAL INGRESS      bound to the local host path. never exposed to another host.
EXTERNAL INGRESS   network-facing. carries the stronger authentication and refusal boundary.
NORMALIZATION      removes transport differences and produces one arrival envelope.
DISPATCH           REQUEST → GREET. RESPONSE → RECOGNIZE. CTU semantics begin here.
```

The split is by trust and transport, not by header. Requests and responses may arrive through
either ingress path, and both carry the same envelope shape. GREET and RECOGNIZE therefore stay
above this subsystem: the former checks request authority; the latter checks the correlated
parent's frame-resident contract.

## Custody

Socket submission is not delivery. Three events must not be collapsed:

```text
APPROACH accepted bytes                 local egress has the envelope.
destination ingress accepted       destination runtime has the envelope.
GREET / RECOGNIZE admitted      destination CTU accepted the causal transition.
```

Which event completes MOVE is still open. The working direction is destination-ingress
acknowledgment: APPROACH accepting bytes is too early to claim that custody reached the destination.

## Open

```text
SOCKET SHAPE       TCP loopback vs Unix-domain local ingress; connection-per-host vs pooled.
ACKNOWLEDGMENT     exact custody-transfer point and retry behavior after ambiguous disconnect.
CONNECTIONS        how THINK supplies and refreshes the live connection MOVE requires.
BACKPRESSURE       whether local and external queues expose one combined or separate pressure.
```

**Design:** `skyra/specs/2026-08-18/network/traffic-routing.md`.
