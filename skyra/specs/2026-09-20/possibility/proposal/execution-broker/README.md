# Execution broker

Date: 2026-09-20

Status: Proposed composition to test. No standalone broker existed in the
proposal directories; this note gathers its responsibilities without inventing
final service interfaces.

Power: realize committed execution bindings, prepare restricted resources,
launch work, and bind its responses to the assigned execution identity.

"Resolve an admitted request" must mean realizing the binding already selected
through Horizon. It must not mean a fresh unrestricted lookup in SurfaceRegistry
after commit. Canonical authority and binding decisions stay in interpretation.

## Composition

The broker brings together the execution-side responsibilities already described
in the [separate-pools proposal](../../../../2026-09-19/persistent-parent-and-separate-pools.md):

| Responsibility | Boundary |
|---|---|
| Execution queue/scheduler | Run admitted jobs; bring state-dependent admission or retry decisions back through interpretation. |
| Resource preparation and launcher | Realize the approved resource constraints and spawn disposable workers. |
| Receive pump | Read/frame output and bind origin to the worker's assigned channel/execution. |
| Lifecycle observer and cleanup | Observe launch/exit, tear down processes/resources, report outcomes. |

The execution pool is separate from StateOwner's initially single-worker state
pool. Sharing a parent process remains possible; this diagram does not establish
OS isolation between components.

```text
authority owner: interpret -> commit exact binding
                         |
                         v
broker: schedule -> prepare resources -> launch worker
                         ^                  |
                         |            output / exit
                         +-- receive / observe
                                  |
                                  v
                    authority owner: interpret outcome
```

## Trust obligation

This broker is trusted: even without permission to edit Horizon, a launcher could
violate an admission decision by choosing different code or providing excessive
credentials/resources. That is why this role is separate from the untrusted
worker, not a claim that the broker is harmless or outside the trusted boundary.

Worker payloads cannot choose their canonical origin. The broker binds origin
from runtime-assigned execution resources; interpretation validates the claimed
result and its correlation. It does not trust a worker's assertion of success.

## Unresolved contracts

1. The committed job/binding and how it pins registration and authority versions.
2. Which resource rights the broker possesses and how their delegation is bounded.
3. Worker execution identity, continuation/attempt distinction, and retirement.
4. Provider-specific adapters: which portions run here, and which run isolated?
5. Failure between state commit and an external effect, including ambiguous results.
6. Persistent provider sessions and their isolation across Possibilities.

The [SurfaceRegistry](../authority-owner/surface-registry/README.md) still stores
a Driver in its registered Surface. That existing shape may need a definition/
binding split before these three roles can be implemented cleanly.
