# 26 — Agents and facets: proposal

Status (2026-09-30): **Implemented in part.** D1 (`system.narrow` only), D2 and D3 (for `system.narrow`'s reference only) built (`8ba6f4c`). The c-list decision was taken as proposed, but a reference is written as a name, `system.ref.<n>` (it was the narrow invoke's stamp until next/30's F1: hashes never reach a model), not `<horizon id>/<facet>/<contract>`. **Not built:** `system.register`, the admin facet, D4, D5, D6.

The implementation plan for [26](../26-agents-and-facets.md). Ids refer to
the [list](README.md). After 24 and 25's C1–C3.

## Decide first: the c-list (D3)

Everything here depends on whether reach can grow by receiving references.
Proposal: yes.

- A return (or fault observation) may carry references. The room records
  each one, and the receiver's reach gains it, within its Horizon's
  authority.
- A reference in `to` is written as `<horizon id>/<facet>/<contract>`; a
  plain name is the shorthand for an entry of your own Horizon's reach.
- Replay derives every reach from genesis plus the references recorded on
  the way.

## D1: the room's facets

- `system.narrow` and `system.register`, implemented in `machine` (they
  create Horizons, a runtime operation). The registry refuses any entry named
  `system.*`, the same way it refuses an exclusion.
- An admin facet (widen, change the table), in the human's reach only.

## D2: narrowing in two steps

```text
{"to": "system.narrow", "payload": {"actor": "<name in reach>", "narrow": {...}}}
  → return: a reference to a new Horizon (B2: authority, this invoke, the actor's input contract)
{"to": <that reference>, "payload": <the actor's input>}
  → an ordinary invoke under the narrowed Horizon
```

Intake checks: actor not in reach → horizon fault; narrowing names anything
not held → horizon fault; payload does not fit → contract fault.
`system.narrow` keeps a frame (plain frames first; tail calls later).

## D4: facets at registration

- A registry entry (and later `system.register`'s input) declares facets:
  `{"facets": {"creator": [...contracts], "public": [...]}}`. A single
  contract agent may omit it (one default facet).
- The registrant receives the facets as references and decides who gets the
  public one.

## D5: calling another agent

`machine.Narrow` (by then, the derivation): a call to a facet of another
agent runs under that agent's Horizon; a call to your own runs under yours.

## D6: Codex as the first agent

- Registry: `codex` with a creator facet (`codex.start`, `codex.exchange`,
  `codex.inspect`, `codex.stop`) and a public facet (the turn).
- `ResourceOwnerDepth` goes: the owner is the holder of the creator facet.
- `codex/resident`'s `modelTools` list becomes the facet handed to the
  model's side. With improvements item 4 and 25's C4.

## Tests that prove it

- `system.narrow` then invoke: the child's profile is the narrowed one; the
  caller's is unchanged.
- Narrowing to something not held: horizon fault, recorded.
- A reference received in a return can be invoked; a reference never
  received cannot (horizon fault), even if its id is known.
- A private facet (never handed out) is invokable by its owner only.
- Replay re-derives every reach from the recorded references.
