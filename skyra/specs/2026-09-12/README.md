# September 12 contract set

Status: active reconciliation set. Proposals are reviewed one at a time.

## Purpose

September 11 established a working DNA-16 vertical slice and then exposed a
cleaner recursive boundary. This directory restates only the decisions that
survive that abstraction break.

September 11 remains design history. Its documents are not moved or rewritten
into an apparently continuous contract. When a September 12 proposal is
accepted, it takes precedence for its named owner.

## Ownership set

```text
2026-09-12/
├── surface/       universal addressable execution boundary
├── move/          canonical execution and continuation language
├── possibility/   visibility, authority, and state references
├── ctu/           interpretation and transition ownership
├── kernel/        root surface and privileged system behavior
└── boot/          construction, assumption, mounting, and validation
```

## Review order

```text
surface
   ↓
move
   ↓
possibility
   ↓
ctu
   ↓
kernel
   ↓
boot
```

`surface/proposal.md`, `possibility/proposal.md`, and the working
`ctu/minimal-loop.md` are now drafted. Each later proposal must use the accepted
contracts above it rather than silently redefining them.

## Proposal rule

An additional [CTU ownership exploration](<../2026-09-16/ownership-boundries of ctu/README.md>)
records the initial-demo direction: queue → sequential interpretation → dispatch
→ supervised concurrent execution, with dependency checks inside opcode handlers.
It remains an exploration requiring contract reconciliation, not an implemented
replacement for the working CTU loop.

Each owner proposal contains:

- purpose;
- contracts;
- ownership;
- recursive or transition law;
- compile-time checks;
- explicit open questions;
- DNA-16 migration pressure.

Implementation begins only after the six proposals agree. DNA-16 remains the
executable reference until a new work order explicitly replaces it.

## September 11 migration map

| Surviving material | September 12 owner |
|---|---|
| `host-driver/universal-surface.md` | `surface` |
| `host-driver/observe-express-realize.md` | `surface`, `kernel` |
| `dispatch-operand-ownership.md` | `move` |
| `continuation-language-operands.md` | `move` |
| `horizon.md` | `possibility` |
| `host-driver/ctu-kernel-proposal/possibility-boundary.md` | `possibility` |
| `system-boundary-cleanup-priorities.md` | `ctu`, `kernel` |
| `host-driver/ctu-kernel-proposal/kernel-assumption.md` | `ctu`, `kernel`, `boot` |
| `seatbelt-enforcement.md` | `kernel`, `boot` |

The Codex driver set remains a pressure test. It will be reconciled after the
six runtime contracts settle.
