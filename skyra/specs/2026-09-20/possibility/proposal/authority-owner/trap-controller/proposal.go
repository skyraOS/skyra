//go:build ignore

// Package trapcontroller contains discussion declarations, not runtime code.
// Surface is the proposed struct in ../../agent-surface-worker/surface; package wiring is provisional.
package trapcontroller

// TrapController is a Surface with a registered identity, contracts,
// dependencies, driver, and ingress/egress profiles. Embedding carries that
// full shape without inventing a second controller execution protocol.
// Horizon.TrapControllerID addresses this Surface's Identity.
type TrapController struct {
	Surface
}

// TODO: define trap/fault input and output contracts, controller authority,
// original-invocation correlation, and controller-failure behavior.
// The gate prepares controller work; it does not invoke the Driver under a
// state lock. Controller emissions return through the common interpretation path.
