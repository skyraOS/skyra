package tests

import (
	"testing"
)

// The cognitive chain forwards stage to stage, each spawning the next on Express,
// each wave owned by the stage it lands on:
//
//	other → self → voting → motor
//
// (Behavior is stubbed; these assert the spawn wiring only.)

func TestOtherFiresSelfOnExpress(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)
	other := m.Sensory().Global().Other()
	assertSpawn(t, spy, other.Express, "skyra:self")
}

func TestSelfFiresVotingOnExpress(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)
	self := m.Sensory().Global().Other().Self()
	assertSpawn(t, spy, self.Express, "skyra:voting")
}

func TestVotingFiresMotorOnExpress(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)
	voting := m.Sensory().Global().Other().Self().Voting()
	assertSpawn(t, spy, voting.Express, "skyra:motor")
}

// The terminus: voting's motor is the SAME instance as the membrane's out-face,
// not a second motor.
func TestVotingMotorIsMembraneMotor(t *testing.T) {
	m := newBeingWithKernel("skyra", newSpyKernel())
	votingMotor := m.Sensory().Global().Other().Self().Voting().Motor()
	if votingMotor == nil || votingMotor != m.Motor() {
		t.Fatalf("voting's motor (%p) is not the membrane's out-face (%p)", votingMotor, m.Motor())
	}
}
