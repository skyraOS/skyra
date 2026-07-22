package dnav3

import "math"

// InterpretationRoute is the express-time reading of Phase 1's counted topology.
// Maturity is deliberately separate: increasing evidence may strengthen a route, but
// cannot turn a balanced fork into confidence or an unresolved bridge into coldness.
type InterpretationRoute string

const (
	InterpretationCold       InterpretationRoute = "cold"
	InterpretationConfident  InterpretationRoute = "confident"
	InterpretationAmbiguous  InterpretationRoute = "ambiguous"
	InterpretationUnresolved InterpretationRoute = "unresolved"
)

// InterpretationHandoff carries counted interpretation into the existing inference
// context. Available distinguishes a cold handoff produced by ObserveImpulse from the
// zero value on a Being that has not observed an impulse yet.
type InterpretationHandoff struct {
	Available  bool
	Route      InterpretationRoute
	EffectiveN int
	Maturity   float64
	Groups     []TargetGroup
}

func buildInterpretationHandoff(groups []TargetGroup) InterpretationHandoff {
	h := InterpretationHandoff{Available: true, Route: InterpretationCold, Groups: groups}
	if len(groups) == 0 {
		return h
	}

	h.EffectiveN = math.MaxInt
	h.Route = InterpretationConfident
	for _, group := range groups {
		if n := groupEffectiveTrials(group); n < h.EffectiveN {
			h.EffectiveN = n
		}
		// An unresolved component vetoes every clean or dominant branch. Ambiguity
		// similarly prevents a global commit, but loses precedence to unresolved.
		switch group.State {
		case SourceUnresolved:
			h.Route = InterpretationUnresolved
		case SourceAmbiguous:
			if h.Route != InterpretationUnresolved {
				h.Route = InterpretationAmbiguous
			}
		}
	}
	if h.EffectiveN == math.MaxInt {
		h.EffectiveN = 0
	}
	if h.EffectiveN > 0 {
		n := float64(h.EffectiveN)
		h.Maturity = n / (n + confHalf)
	}
	return h
}

// groupEffectiveTrials is the support of the least-settled candidate involved in
// the topology. This prevents a dense branch from lending maturity to a barely
// witnessed alternative or conjunct.
func groupEffectiveTrials(group TargetGroup) int {
	n := math.MaxInt
	for _, branch := range group.Branches {
		for _, completion := range branch.Completions {
			if completion.Trials < n {
				n = completion.Trials
			}
		}
	}
	for _, completion := range group.Unresolved {
		if completion.Trials < n {
			n = completion.Trials
		}
	}
	if n == math.MaxInt {
		return 0
	}
	return n
}
