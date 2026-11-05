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
- `spec.json`: version 3, goal, segments, coverage, input observation
  references, proposed capability references and the profile. This is the
  Plan result's value schema, not another executable.

Every Spec states its `profile`, required: the constants Launch's tree runs
under, which its root sets for every agent in it. `act` is the most contracts
one node holds and still invokes itself, past which it delegates; `turns`,
`depth`, `fanout` (at most 32, the kernel's) and `budget` bound each agent's
turns, the tree's depth, the instructions in one call and the turns of the
whole tree. `DecodeSpec` refuses a Spec
without one or with one that cannot run a tree, and `Resolve` takes Launch's
root kind from the act bound (`Profile.KindOf`) and passes the profile into
the tree's input. A changed profile is a changed task (`SameTask`).
`DefaultProfile`, `{act 1, turns 8, depth 3, fanout 32, budget 64}`, is the one-contract
tree; the proposer must state a profile too, and a human save keeps the last
revision's unless it sets one.

A Spec's goal is made of segments, each with its source: `cited`, a quote
of at least `MinQuote` (3) words from an observation, by its reference;
`plan`, words Plan added; or `human`, words the human wrote. `coverage` counts
the goal's words by source. Both are derived: `DecodeSpec` refuses a Spec
whose goal or coverage is not what its segments assemble, and `Resolve` (so
`AuthorSpec`, `StartLaunch` and `ValidateHandoff`) refuses a cited segment the
named held observation does not contain, verbatim but for whitespace.
`Propose` builds segments from the proposer's `goal`: `{"quote", "of"}` or
`{"text"}`. A shorter quote counts as Plan's, so scattered words cannot pad
coverage, and the line-number gutter a passed observation is shown with is
removed from quotes. Plan can always add words; each one lowers its coverage.

Revisions build on the last Spec. `Revise` turns the human's edited text into
segments by a word diff against the previous Spec: kept words keep their
segment's source, inserted or changed words are `human`, and a cited run the
edit broke or cut below `MinQuote` becomes `human`; inputs carry over. So an
edit never relabels Plan's words as the human's. A Spec may itself be quoted
(Plan refining one): `CheckCites` accepts a quote of a Spec only within one of
its `human` or `cited` segments, never its `plan` ones, so Plan must add its
own earlier words again as its own.
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
