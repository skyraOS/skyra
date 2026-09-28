# Profiles become input contracts

Date: 2026-09-26  
Status: to carry forward. Source: DNA-19's `codex/profiles.json` and
`protocol/profiles.go`, after the handle-authorization refactor (`e2db112`).

## Why

DNA-19's profiles are the part of its contract layer that worked. They are
declarative data, strictly parsed, and checked on both sides of the boundary.
Removing Codex knowledge from the runtime was possible only because the
requirements had somewhere to go: the profile. The controller's input contract
(`next/agenda.md`, section 1) is undefined; profiles are the working draft.

## What carries over

| DNA-19 profile | Input contract |
|---|---|
| `fields`: key, encoding, required, `max_bytes`, `values` | the input schema. Fit is structural and strict: unknown fields do not fit |
| `encoding: "parent"` plus `handle: fresh / started / end / borrowed` | a capability operand, and what the operation requires of it. State is derived from the log, never stored |
| `results` | the return contract (the `reply` capability's input contract) |
| `failures` | the fault kinds the operation may discharge its slot with |
| `retains` | which failures keep the operation's handles live: fault routing as data |
| checked by the host and again by the actor | emitter never checker, one step further: the contract checks both directions |
| strict JSON, digest compared at `connect` | the contract's hash; every record names the memory root it read |

## What to fix in the move

- **One definition.** DNA-19 defines the browser contract twice, in Go and in
  TypeScript, and they drifted (the `null`-list bug, `ae77e13`). An input
  contract lives once in memory, addressed by hash; everything else reads it.
- **Typed, not strings.** Failures are matched by message prefix
  (`"unsupported:"`) and refusal reasons are free text. Fault kinds should be a
  field, not a prefix.
- **`lifecycle` is prose.** It duplicates what `handle` now says per field.
  Drop it.
