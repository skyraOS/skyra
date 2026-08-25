// Package drivers is the HOST TABLE — trusted hardware, system space
// (the-quadrant.md, left column). A driver ships with the OS, so its output is
// BELIEVED (never stamped). Beneath the kernel it persists: the kernel decides WHAT
// (the cone), a driver decides HOW.
//
// The distributed model leaves this boundary almost untouched — the one consequence
// is that the Cone a driver persists is now a set of content-addressed, signed
// segments (causal.Cone), so a later gather can Fetch them by address.
// AUTHORITY: skyra/specs/2026-08-18/kernel-ctu-boundary/durable-storage.md;
// skyra/specs/2026-08-22/drivers/README.md (no new interface).
package drivers

import c "skyra-v1/src/skyra/skyra-dna-v8/causal"

// Driver is a host-table entry. Its output is trusted because of WHERE IT CAME FROM
// (provenance: ships with the OS), not a runtime claim. dispatch.md.
type Driver interface {
	// Activate reads durable storage and returns the engram. The kernel makes it a
	// possibility; the driver only fetches. durable-storage.md.
	Activate(name c.Name) (c.Engram, c.Outcome)

	// Remember persists the cone and reports when durable. The kernel handed down
	// WHAT (the gathered, content-addressed segments); this is HOW. durable-storage.md.
	//
	// SEAM: teardown must not race the flush — Success must mean the cone is durable.
	Remember(cone c.Cone) c.Outcome
}
