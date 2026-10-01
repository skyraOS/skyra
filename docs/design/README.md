# Design: the stateless controller

The spec behind [`controller/`](../../controller/), and the reasoning that produced
it. Everything here is agreed and, except where noted, built.

## Read in this order

1. **[`proposal.md`](proposal.md)**: the spec. Decisions D1–D9: interpret
   before spawning, custody, the stamp, the four phases, the release, the
   rendered sandbox, leaves, states. It lags the later notes on details.
2. **[`thoughts/`](thoughts/)**: how each decision was reached, one short note
   per step, in order:

   | Note | Settles |
   |---|---|
   | [01](thoughts/01-where-the-step-boundary-falls.md) | one process per step, and where its boundary falls |
   | [02](thoughts/02-custody-and-the-stamp.md), [03](thoughts/03-what-the-stamp-proves.md) | custody, and why the stamp is just a hash |
   | [04](thoughts/04-the-three-phases.md) | the phases of an instance |
   | [05](thoughts/05-the-release-is-a-handle.md), [06](thoughts/06-the-release-is-the-horizon.md) | the release: a directory fd plus the child's Horizon |
   | [07](thoughts/07-the-sandbox-is-rendered-from-the-horizon.md) | the Seatbelt profile is generated from the Horizon |
   | [08](thoughts/08-leaves.md), [09](thoughts/09-leaf-as-capability-vs-state.md) | leaves: tools that can only return |
   | [10](thoughts/10-the-move-is-the-state.md), [11](thoughts/11-the-potential.md) | a Move is the state of a computation; the Potential |
   | [12](thoughts/12-stateless-adapters-stateful-actors.md) | long-running services (Codex) as actors behind stateless adapters. *Not built* |
   | [13](thoughts/13-dependencies.md) | dependencies, the join rule, and fan-in. *Not built* |

   Thought 14 moved to [`next/`](../next/): it starts the capability work.
3. **The pieces around a single call**, agreed in outline, not built:
   - [`faults.md`](faults.md): faults (the runtime proves something broke)
     versus interrupts (memory lacks a contract).
   - [`durable-storage.md`](durable-storage.md): the event log, the stack, and
     a memory of contracts.
   - [`io-buckets.md`](io-buckets.md), [`improvements.md`](improvements.md):
     proposals for separating custody from the effects it performs.
4. [`summary.md`](summary.md): the full map as of 2026-09-25. Long; the
   [root README](../../README.md) is the current overview.
