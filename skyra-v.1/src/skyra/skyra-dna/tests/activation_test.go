package tests

import (
	"testing"

	dna "skyra-v1/src/skyra/skyra-dna"
)

// The activation record bumps two EMAs from the composite sample (no +1 constant)
// and decays toward 0 when absent. Verified through the public interface (inputs
// default neutral, so a visited cell bumps toward 1).
func TestActivationBumpAndDecay(t *testing.T) {
	rec := dna.NewMemoryCell("c", dna.KindSetpoint, "x", nil).NewActivation()

	// First visit: Familiarity 0.1·1 + Identity 0.02·1 = 0.12.
	v1 := rec.Activate(true, 1)
	if v1 < 0.11 || v1 > 0.13 {
		t.Fatalf("first visit activation = %v, want ~0.12", v1)
	}

	// Second visit: both EMAs climb toward the sample (1.0).
	v2 := rec.Activate(true, 2)
	if v2 <= v1 {
		t.Fatalf("second visit = %v, want > first %v (EMAs rising)", v2, v1)
	}

	// Absent: sample 0 → both EMAs decay toward 0.
	v3 := rec.Activate(false, 3)
	if v3 >= v2 {
		t.Fatalf("decay = %v, want < %v (EMAs falling toward 0)", v3, v2)
	}
}
