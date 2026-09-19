# Authority, Focus, and dependencies

**Reviewed:** 2026-09-07  
**Status:** Accepted distinctions. Capability verification and durable choice encoding remain design work.

| Surface | Meaning |
|---|---|
| Horizon | Kernel-provisioned permissions for the workflow. |
| Focus | Available operations presented as choices to a particular Move. |
| Dependencies | Inputs or executable relationships the Move has committed to using. |

Focus is `[]c.Capability`, selected from `Horizon.Capabilities`. Use the existing capability vocabulary:
Target, Action, ExpiresAt, and Sig. No OperationDecl wrapper or parallel Attention field is needed.

Showing a capability does not invoke it or make it a required dependency. Selecting Focus does not
mint grants, widen Horizon, or revoke grants outside Focus. Focus is not an additional access gate.
The workflow Horizon remains the authority boundary; no per-child narrowing mechanism is selected
for this phase.

## Action and access

Action is the sole model-visible method name/header, with a design cap of 160 tokens. There are no
separate name, title, and description fields for that role. Tokenization and enforcement remain open.
Action stays on the invocation and capability; Contract contains only its schema.

Kernel provisions the Horizon. CTU checks the declared target/action when an operation accesses a
mechanism, including after waiting or recovery. Finding an executable or an address does not grant
access. Installed drivers are subject to the authority provided for their use.

The current `Touching` seam exposes target/action for access checks without making the MMU understand
each opcode's implementation. Signature verification, trust material, expiry enforcement, and denial
transport are not yet implemented as a complete boundary. This is not a demonstrated sandbox claim.

## Declaration and remembering

Available choices must be visible during declaration, before a remembered executable exists. They
cannot be reconstructed solely from work already invoked. Focus provides that live surface.

Remembered executable structure records committed relationships. It need not preserve the entire
catalog visible during declaration. If replay still delegates a choice to a model, the reusable choice
references it needs must have a durable home and be resolved against the new run's Horizon. A previous
run's expiring grants must not become newly minted authority merely because they were stored.

Durable choice-reference encoding and replay against historical versus current catalogs remain open.

## Reviewed sources

- [09-01 Focus and Move](../2026-09-01/focus-and-move.md): three surfaces, capability slice, Action, and replay distinction.
- [08-27 Horizon and Attention](../2026-08-27/horizon-and-attention.md): kernel-created workflow Horizon and provision-versus-check ownership.

The old Attention placement, Work/Issued catalog, kernel-free hot path, optional separate backing ID,
and proposed durable parentContract/childContracts fields are not imported.
