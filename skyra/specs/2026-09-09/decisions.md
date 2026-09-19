# Decisions — 2026-09-09 driver registration

**Status: DECIDED.** The running record of what is settled for the driver-registration
thread. Terse by design: each entry is a decision plus its reason. The *shape* is
decided; field contents remain deferrable (see D7). This is the authoritative doc for
this thread; where the narrower docs use a shorthand, this one wins.

## D1 — The registry is a passive top-level struct
The thing that holds the drivers is its own top-level struct, not a map on the CTU. It is
memory-with-a-lock: it stores driver presence and serializes access; it holds no policy
and runs no loop. It is the **single writer of driver-presence state** — one owner, the
same discipline as the scheduler owning continuation state.

## D2 — Name: `DriverRegistry` (kept for now)
Carried from v14. Considered `InstructionStore` (dataflow: scheduler = matching store,
this = instruction store) and `DriverInstruction`/`InstructionDriver`, but rejected a
rename for now. It is genuinely a bit of both, and the reason is the interesting part: in
a classical dataflow machine the instruction is **already known and addressed**; here the
instruction is **not** known ahead of time — it is *found via execution* (resolve an
address to a driver, then run it). So it is part registry (names bound at runtime) and
part instruction store (indexed on the hot path). `DriverRegistry` is close enough;
revisit only if a better name earns it.

## D3 — Two interfaces over one struct (the two doors)
Read and write are separated structurally by interface segregation, not convention.
`*DriverRegistry` implements both; whoever is handed which interface at wiring time is the
enforcement — the compiler keeps the doors apart.

## D4 — Who holds which door
- CTU / controller are handed a **`Resolver`** — read-only, on the dispatch hot path.
- The kernel is handed the **`Registrar`** — the only holder of the write door. The
  registry exposes these mutating methods for the kernel to call.

## D5 — Register traps; Resolve does not
`Register` is privileged. A possibility cannot call it (it holds no `Registrar`); it emits
a Move to the kernel and faults, and the kernel fulfils it (trap-and-emulate). `Resolve`
is direct, ungated, and never traps — the CTU indexes the store itself. Register is a
page-table edit; Resolve is an MMU lookup.

## D6 — Register rules for this pass
A duplicate address is an error — a name is a namespace claim; no silent overwrite
(`Replace` deferred). Registration is a runtime operation — loadable after CTU
construction, no restart. The registry stores; it does not decide. Authority lives on the
trap path, not in the store.

## D7 — Stored value: v14 `DriverRegistration` retained
The map value stays as v14's `DriverRegistration` for now; its field contents (whether
`InputValidator` stays here, moves to the authority/contract layer, or is joined by a
descriptor/adapter) are deferred. Shape decided; fields not. (`Resolve` supersedes v14's
`Lookup`; `Register` is unchanged from v14.)

## The struct and its interfaces

```go
// Driver is unchanged from v14 this pass.
type Driver interface {
    Ingress(ctx context.Context, move protocol.Move, output func(string, error) error) error
    Egress(raw string) ([]protocol.Move, error)
}

// DriverRegistration is the stored value (v14; field contents deferred — see D7).
type DriverRegistration struct {
    Driver Driver
    Input  InputValidator
}

// DriverRegistry is the passive top-level store: the single writer of driver-presence
// state. Custody only — no policy, no loop.
type DriverRegistry struct {
    mu      sync.RWMutex
    drivers map[string]DriverRegistration
}

func NewDriverRegistry() *DriverRegistry

// Resolve is the READ door — handed to the CTU/controller as a Resolver. Direct,
// ungated, on the hot path; never traps. (Supersedes v14's Lookup.)
func (r *DriverRegistry) Resolve(address string) (DriverRegistration, error)

// Register is the WRITE door — handed only to the kernel as a Registrar. Privileged;
// reached by faulting to the kernel. A duplicate address is an error.
func (r *DriverRegistry) Register(address string, reg DriverRegistration) error

// The two doors, as interfaces. Interface segregation is the enforcement.
type Resolver  interface { Resolve(address string) (DriverRegistration, error) }
type Registrar interface { Register(address string, reg DriverRegistration) error }
```

## Deferred (recorded, not this pass)
Replace / Drain / Unload / Fault; descriptor; the adapter (Encode/Decode) split;
generation / lease / in-flight count; the async lease boundary; health / persistence /
driver kinds. See the deferred sections of `driver-registration.md` and
`ingress-egress-adapters.md`.
