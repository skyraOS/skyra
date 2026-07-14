package dnav2

// The STATIC workflow definition — the routing DAG `(step, condition) → next`.
// Config, not runtime data (the dynamic per-impulse state is the Board). Small and
// concrete for these steps; the orchestrator reads it. See
// specs/2026-07-03/slice-based-dispatch.md ("The workflow definition").

// Step is a workflow node — a competence namespace + gene, "namespace:gene".
type Step string

const (
	StepSensoryIntake    Step = "sensory:intake"
	StepSensoryIngest    Step = "sensory:ingest" // deposit a perceive-return's surroundings on the spine
	StepFinalGrade       Step = "final:grade"
	StepEfficientGrade   Step = "efficient:grade"
	StepMidlevelFrame    Step = "midlevel:frame"
	StepFinalPredict     Step = "final:predict"
	StepEfficientPredict Step = "efficient:predict"
	StepMotorCollapse    Step = "motor:collapse"
)

// Route is what a step's condition dispatches to: either a set of next steps to
// run (one = sequential, two = fan-out), or a terminal (done / reframe / join-wait).
type Route struct {
	Next    []Step // steps to spawn (fan-out if >1)
	Reframe bool   // → midlevel with the failure (subject to the one-reframe rail)
	Done    bool   // terminal — flush
	Wait    bool   // this branch marked the board; wait for the join (no spawn)
}

// transition is the DAG: (step, condition) → Route. A missing entry is inert.
var transition = map[Step]map[string]Route{
	StepSensoryIntake: {
		"external": {Next: []Step{StepFinalGrade}},
		// the reafference lane forks by primitive: an action-return grades my
		// execution; a perceive-return deposits what came back (a reafference too —
		// NOT internal; internal is interrupts). perceive-build-uncertainties.md §1.
		"reafference":     {Next: []Step{StepEfficientGrade}},
		"perceive-return": {Next: []Step{StepSensoryIngest}},
		"internal":        {Next: []Step{StepMidlevelFrame}},
	},
	StepSensoryIngest: {
		// continue into framing for now (decision 1b — revisit with motor independence)
		"deposited": {Next: []Step{StepMidlevelFrame}},
	},
	StepFinalGrade: {
		"all-preds<0": {Reframe: true},
		"held":        {Next: []Step{StepMidlevelFrame}},
		"no-prior":    {Next: []Step{StepMidlevelFrame}},
	},
	StepEfficientGrade: {
		"exec=no": {Reframe: true},
		"exec-ok": {Next: []Step{StepMidlevelFrame}},
	},
	StepMidlevelFrame: {
		"framed":              {Next: []Step{StepFinalPredict, StepEfficientPredict}}, // fan-out + declare expects
		"empty":               {Done: true},
		"query-format-error":  {Done: true}, // correction exhausted; active state retained
		"causal-format-error": {Done: true}, // either causal worker failed; active state retained
	},
	StepFinalPredict: {
		"predicted": {Wait: true}, // marks final done; the join fires motor
	},
	StepEfficientPredict: {
		"predicted": {Wait: true}, // marks efficient done; the join fires motor
	},
	StepMotorCollapse: {
		"emitted": {Done: true},
	},
}

// routeFor looks up (step, condition); ok=false if the DAG has no such edge.
func routeFor(step Step, condition string) (Route, bool) {
	byCond, ok := transition[step]
	if !ok {
		return Route{}, false
	}
	r, ok := byCond[condition]
	return r, ok
}

// namespaceOf splits "namespace:gene" → "namespace" (the active-state slice key).
func namespaceOf(step Step) string {
	s := string(step)
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return s[:i]
		}
	}
	return s
}

// causeOf returns the causal lane a predict step marks ("final" / "efficient"), or "".
func causeOf(step Step) string {
	switch step {
	case StepFinalPredict:
		return "final"
	case StepEfficientPredict:
		return "efficient"
	}
	return ""
}
