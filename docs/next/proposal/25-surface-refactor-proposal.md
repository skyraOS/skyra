# 25 — Surface refactor: proposal

Status (2026-09-30): **Implemented in part.** C1 and C2 built (`a275748`). **Corrected and deferred:** C3, since the launcher passes the real app-server's `--disable` flags through `Args`. **Not built:** C4.

The implementation plan for [25](../25-surface-refactor.md). Ids refer to the
[list](README.md). C1–C3 land with 24's step 1.

## C1: delete `Reach`, `Authority` and `Delegate` from the surface

- `core/horizon/horizon.go:85`: remove the three fields.
- `runtime/registry/registry.go`:
  - `:73` and `:148`: stop copying `d.Reach`, `authority` and `delegates`
    onto the surface.
  - `:116` and `:151`: the checks that a surface's reach and delegations name
    registered entries go; the registry table (the room's reach) is simply
    every registered name.
  - The per-entry authority (`e.Authority`, `e.Connect`, `e.Delegate`) moves
    to the room's genesis Horizon: one set of grants for the room, not one per
    surface. What the operator grants per agent comes back with facets (D4).
- `Declared.Reach` and the `"reach"` line in `targets/user`, `targets/agent`,
  `codex/resident` and `codex/adapter` `contract.json`: removed.
- `machine.Narrow`: no intersection (24, step 1).
- Tests: `registry_test`, `machine_test`.

## C2: delete `Leaf()`

- `horizon.go:111` and its one caller, `machine.go:359`.
- The stdin format becomes an explicit surface property, since it describes
  what the program is: a `continuation` flag in the registry entry (default
  false: bare payload). `user`, `agent`, `codex/resident`, `codex/adapter` set
  it.
- Decide first: a flag, or read from the input contract. The flag is the
  smaller change.

## C3: delete `Args`

- `horizon.go` (`Args`), `registry.go:70,79,148`, `instance.go:361`,
  `runtime/instance/appserver.go:178`. No registry sets it; nothing else
  changes.

## C4: later, with the loop operator

The resident and network fields (`Resident`, `Connection`,
`CloseConnection`, `ResourceOwnerDepth`, `ResidentProfile`, `ProxyHosts`,
`ProxyPath`, `ProxySHA256`): network grants move to the Horizon; the
lifecycle moves to control flow. Taken up with improvements item 4 and D6.

## What remains on the surface

```go
type Surface struct {
	Path, SHA256   string           // identity
	Input, Output  *contract.Schema // the contract
	Continuation   bool             // how it reads its input (C2)
	Host           string           // async: parked with this host (until 16's host is an agent)
	// resident fields, until C4
}
```

## Tests that prove it

- A registry entry with `"reach"` or `"args"` is refused (strict JSON).
- `echo`, `read`, `agent`, `user` and the Codex path run unchanged under a
  shared Horizon; `skyra demo` replays byte-identically.
