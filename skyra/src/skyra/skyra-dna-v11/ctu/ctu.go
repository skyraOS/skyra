// Package ctu is the execution core — the HOT path (ACT).
//
// FRESH BUILD. The CTU is nearly-stateless. Move is its single ENTRY (ingress calls it; local THINK
// calls it). It CARRIES OS-specific driver mechanisms but has NO ambient authority — it fires nothing
// without a horizon handed to it (CHERI-shaped). It holds only VIEWS/HANDLES: the carried horizon and
// cut-IDs as fds. The authoritative state (table, socket table, mint) is the kernel's.
//
// AUTHORITY: skyra/specs/2026-08-24/skyra-dna-v11/research/move-and-the-kernel-barrier.md,
// think-act-paths.md, capability-table-and-horizon.md, bootstrap/ctu.md.
package ctu

import c "skyra-v1/src/skyra/skyra-dna-v11/causal"

// CTU is the execution endpoint. Move is the door; ACT (Drive) lives behind it.
type CTU interface {
	// Move is the CTU's single entry, called by ingress (or local THINK). It dispatches on Header:
	//   Request  → VERIFY the carried Capability offline (macaroon), MATERIALIZE a Context from the
	//              (cached) Body + the carried horizon, and Drive it (ACT). the HOT path needs no kernel.
	//   Response → land the observation on its fd (InFlight[cut-ID]), re-drive.
	// It never mints and never widens a horizon. (move-and-the-kernel-barrier; think-act-paths)
	Move(t c.Transfer) c.Outcome
}

// Drive is the reservation-station cycle — ACT. It runs a resident Context within its CARRIED horizon,
// turning landings into ready moves, dispatching locally or issuing a cut, and returning on quiescence.
// A move to a target NOT in the horizon FAULTs (a page fault → the kernel). Never blocks, never recurses.
type Drive interface {
	Drive(ctx *c.Context) c.Outcome
}

// Materialize is the LOADER: Body + Horizon → a fresh Context. On the hot path the CTU does this
// locally (no kernel) from a cached Body; on the cold path the kernel does it. (cpu-execution-model)
type Materialize interface {
	Materialize(b *c.Body, h c.Horizon) *c.Context
}

// Verify is the offline capability check (macaroon signature + caveats). This is what makes the hot
// path kernel-free: the user PRESENTS the capability, the CTU verifies it, no round-trip.
type Verify interface {
	Verify(cap c.Capability, t c.Target) bool
}

// Interpret is block selection over the CONTEXT's HORIZON (its provisioned view). A name that resolves
// in the horizon → dispatch; a name that MISSES because it is unprovisioned → FAULT (page fault). A
// name that resolves to a driver on THIS ctu runs it; the ctu never enumerates its own drivers.
type Interpret interface {
	Interpret(ctx *c.Context, mv *c.Move) (EligibleSet, c.Outcome)
}

type EligibleSet struct{ Targets []c.Target }

// Drivers are the CTU's OS-specific mechanisms — inert, present, unknown to the CTU. It ADVERTISES them
// to the kernel at handshake (enumeration); it never fires one except via a horizon. (bootstrap/ctu.md)
type Drivers interface {
	// Advertise exposes the driver set for the kernel to READ and register into the sysblock. The CTU
	// does not use this inventory itself.
	Advertise() []c.Target
}
