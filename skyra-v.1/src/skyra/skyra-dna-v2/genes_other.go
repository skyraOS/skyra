package dnav2

import "skyra-v1/src/skyra"

// other:grade — grades the prior other-distribution against the incoming external
// evidence. Phase 1 (cold start / no prior bet): reports "no-prior" (no inference).
// Phase 2 ports the adversarial judge → "held" / "all-preds<0".
func otherGrade(root *EphemeralRoot, origin skyra.Space, impulse any) {
	// no grouping / no prior distribution yet → nothing to grade
	root.Condition = "no-prior"
}
