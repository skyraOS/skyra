package tests

import (
	"testing"
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

func TestSelfFiresMotorOnExpress(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)
	self := m.Sensory().MidLevel().Other().Self()
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
