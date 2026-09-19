# 2026-09-18 — the CU/CTU collapse (survivors of 2026-09-16)

Date: 2026-09-18.

Status: curated snapshot. This directory ports forward the parts of
[`../2026-09-16`](../2026-09-16) that survive the 09-18 collapse — one smart
component, one writer, one queue, fd-style identifiers minted at commit. The
per-claim rationale is in [`reconciliation.md`](reconciliation.md). Every file
here remains a proposal (`//go:build ignore`), not an implemented runtime.

## What was ported (survives)

| Path | Why it survives |
|---|---|
| [`cu-cycle.md`](cu-cycle.md) | The selected shape: read → commit → execute; identifier as file descriptor. (Draft B) |
| [`ctu/minimal-loop-v2.md`](ctu/minimal-loop-v2.md) | The selected loop. Supersedes v1. (Draft B) |
| [`MessageBroker/`](MessageBroker) | StateOwner as sole writer behind restricted typed profiles. (Draft B README) |
| [`tu/`](tu) | Stateless `Transition` profile; per-stage atomic commit; operations ≠ Moves. (Draft B) |
| [`possibility/`](possibility) | Identity/authority record + Horizon; unchanged by the collapse. |
| [`move/proposal.go`](move/proposal.go) | Data-only Move/Invoke/Return contract; opcode/provider neutral. |
| [`surface/proposal.go`](surface/proposal.go) | Surface / contract-pair / Driver data shapes; recursive capability identity. |

## What was left in 2026-09-16 (superseded — not ported)

Per [`reconciliation.md`](reconciliation.md):

- `ctu/minimal-loop.md` (v1) — superseded by v2.
- `interpretation/` — Draft A two-smart-component service. Its `Run` /
  `interpretAndTransfer`, `CustodyState`, two-mailbox handoff, and stale-revision
  re-interpret are gone; only the "one consumer, one Move at a time" idea survives,
  and it lives in `ctu/minimal-loop-v2.md`.
- `supervisor/` — Draft A interpreting service. Only the *worker machinery*
  (bounded pool, capacity-before-take, register-before-launch, independent
  monitoring) survives; its *decisions* fold into the single interpreter.
- `ownership-boundries of ctu/` — Draft A. Its ownership carve-up survives and is
  restated by `MessageBroker` + `possibility`; its loop mechanics (two mailboxes,
  custody bounce, global-revision OCC on the interpret path) are superseded.

## Known inconsistency still to resolve

The ported `.go` contracts predate the collapse and still encode the **two-component,
two-mailbox** world:

- [`MessageBroker/proposal.go`](MessageBroker/proposal.go): `systemState` still holds
  separate `interpretation` and `supervisor` mailboxes plus a `custody` map, and
  `SystemProfiles` still names a `Supervisor` profile.
- [`tu/proposal.go`](tu/proposal.go): `Component` is still `{Interpretation, Supervisor}`
  with a `CustodyRecord` and `from`/`to` routing.

`ctu/minimal-loop-v2.md` says the opposite: **one inbound queue, no custody bounce,
no `from`/`to`**. These proposals are ported verbatim as the agreed struct record;
collapsing them to the single-queue model is a design edit, deferred here rather than
made silently. This is the first thing to settle before implementation.

## Cross-references

Links inside the ported docs that point at non-ported siblings (v1 minimal-loop,
interpretation, supervisor, ownership-boundries) resolve against
[`../2026-09-16`](../2026-09-16), which is retained as the historical draft. Bodies
were copied unedited to preserve the contract text.
