// Package drivers is the HOST TABLE — trusted hardware, system space
// (the-quadrant.md, left column; dispatch.md). A driver ships with the OS, so
// its output is BELIEVED (never stamped). Beneath the kernel it persists: the
// kernel decides WHAT (the cone), a driver decides HOW. durable-storage.md.
//
package drivers

import c "skyra-v1/src/skyra/skyra-dna-v7/causal"

// Driver is a host-table entry. Its output is trusted because of WHERE IT CAME
// FROM (provenance: ships with the OS), not a runtime claim. dispatch.md.
type Driver interface {
	// Activate reads durable storage and returns the engram. The kernel makes it
	// a possibility; the driver only fetches. durable-storage.md.
	Activate(name c.Name) (c.Engram, c.Outcome)

	// Remember persists the cone and reports when durable. The kernel handed down
	// WHAT; this is HOW. durable-storage.md.
	//
	// SEAM: teardown must not race the flush — REMEMBER's Success must mean the
	// cone is durable, or a crash loses the arena. See interface-rationale.md.
	Remember(cone c.Cone) c.Outcome
}
