package tests

import (
	"testing"
	"time"

	"skyra-v1/src/skyra"
)

// The cognitive chain forwards stage to stage, each spawning the next on Express:
//
//	other → self → motor   (voting parked for the MVP — self hands straight to motor)
//
// These assert the spawn wiring; the predict behavior is covered in other_test.go
// / self_test.go.

func TestOtherFiresSelfOnExpress(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)
	other := m.Sensory().MidLevel().Other()
	assertSpawn(t, spy, other.Express, "skyra:self")
}

// The C wiring rule: self fires motor only if it deposited a plan this wave.
// After a planning Observe → motor spawns. With no plan (bare Express) → no spawn
// (else it would fire a stale intent).
func TestSelfFiresMotorOnlyWhenPlanned(t *testing.T) {
	spy := newSpyKernel()
	mock := gradingMock(t, "<match>none</match>", "<grade>yes</grade>", nil)
	m := newBeingWith("skyra", spy, mock)
	sensory := m.Sensory()
	mid := sensory.MidLevel()
	self := mid.Other().Self()

	// no plan yet → Express must NOT spawn motor (the C rule)
	self.Express(nil, skyra.Impulse{Payload: "x"})
	select {
	case call := <-spy.calls:
		t.Fatalf("motor spawned with no plan: %v (C rule violated)", call.traversed)
	case <-time.After(100 * time.Millisecond):
	}

	// plan via a real cycle → Express spawns motor
	sensory.Slice().SetCount(1)
	mid.Slice().SetCount(1)
	imp := engageFrom("michael", "can you help me with the skyra project")
	sensory.Observe(nil, imp)
	mid.Observe(nil, imp)
	self.Observe(nil, imp) // deposits the proposal → plannedThisWave

	assertSpawn(t, spy, self.Express, "skyra:motor")
}

// The terminus: self's motor is the SAME instance as the membrane's out-face,
// not a second motor.
func TestSelfMotorIsMembraneMotor(t *testing.T) {
	m := newBeingWithKernel("skyra", newSpyKernel())
	selfMotor := m.Sensory().MidLevel().Other().Self().Motor()
	if selfMotor == nil || selfMotor != m.Motor() {
		t.Fatalf("self's motor (%p) is not the membrane's out-face (%p)", selfMotor, m.Motor())
	}
}
