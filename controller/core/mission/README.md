# Mission contracts

A Mission is owned by a human root. Plan returns a Spec observation to that
root. The human explicitly passes the exact Spec, its cited observations and
its proposed capability grant to Launch. A Spec alone never grants authority.

This package validates that handoff without filesystem access or execution.
`Contract` uses the worker protocol's pinned binary and canonical input schema;
`InvokeType` and `ReturnType` use its existing hash domains. A catalogue reference
additionally binds routing, program configuration and resource scope. A tool
name or model-authored schema cannot substitute for a held capability.

The public schemas are in `contracts/`:

- `plan.json`: a request observation and explicit context observations.
- `spec.json`: version, goal, input observation references and proposed capability
  references. This is the Plan result's value schema, not another executable.
- `launch.json`: the exact approved Spec observation.

Transported observations carry their reference, producer binary/input contract,
value and original source. The trusted adapter constructs these envelopes from
the root's held observations. Model-supplied envelopes are not evidence of origin.

`Resolve` counts distinct callable contracts: one produces an `act` worker;
multiple produce a `delegate` worker. Resource scope is independently bounded
by the human's held catalogue. No tool count, read/write policy or additional
worker kind is attached to the Plan/Launch product names.

`ValidateHandoff` requires the recorded root invocation, exact Spec and cited
observations, preserved sources, and an effective grant matching the Spec.
Runtime preparation rechecks this rather than trusting a cached resolution.

The pure worker remains responsible for worker transitions, unchanged. The
Spec and cited observations reach it as the text of its ordinary root
endowment, made by `rootObservation`; there is no other way in. Nothing in the
worker or its proof establishes that an external receipt or transport envelope
is authentic. That is the trusted adapter's obligation.
