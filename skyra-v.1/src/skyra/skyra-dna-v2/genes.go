package dnav2

// WireGenes registers the ported competences on the DNA at boot. Genes are added
// here as they are ported (phase 1: cold-start turn 1; phase 2: grading + reframe).
func WireGenes(d *DNA) {
	d.RegisterObserve(string(StepSensoryIntake), sensoryIntake)
	d.RegisterObserve(string(StepOtherGrade), otherGrade)
	d.RegisterObserve(string(StepMidlevelFrame), midlevelFrame)
	// other:predict, self:predict, motor:collapse — porting next.
}
