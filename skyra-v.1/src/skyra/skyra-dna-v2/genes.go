package dnav2

// WireGenes registers the competences on the DNA at boot. Genes are added
// here as they are ported (phase 1: cold-start turn 1; phase 2: grading + reframe).
func WireGenes(d *DNA) {
	d.RegisterObserve(string(StepSensoryIntake), sensoryIntake)
	d.RegisterObserve(string(StepSensoryIngest), sensoryIngest)
	d.RegisterObserve(string(StepFinalGrade), finalGrade)
	d.RegisterObserve(string(StepEfficientGrade), efficientGrade)
	d.RegisterObserve(string(StepMidlevelFrame), midlevelFrame)
	d.RegisterObserve(string(StepFinalPredict), finalPredict)
	d.RegisterObserve(string(StepEfficientPredict), efficientPredict)
	d.RegisterObserve(string(StepMotorCollapse), motorCollapse)
	// Phase 1 (cold-start turn 1) + phase 2 (final:grade, efficient:grade, reframe) ported.
}
