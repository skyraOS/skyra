package dnav3

import (
	"math"
	"strings"
	"testing"
)

func phase2Completion(source, object string, heat float64, trials int) Completion {
	c := phase1Completion(source, object, heat)
	c.Trials = trials
	return c
}

func phase2Branch(completion Completion) Branch {
	return Branch{Completions: []Completion{completion}, Heat: completion.Heat}
}

func TestToy7GraduationHandoffCoversFourRoutes(t *testing.T) {
	a := phase2Completion("S", "A", .90, 20)
	b := phase2Completion("S", "B", .20, 20)
	c := phase2Completion("S", "C", .15, 20)

	tests := []struct {
		name   string
		groups []TargetGroup
		want   InterpretationRoute
	}{
		{name: "cold", want: InterpretationCold},
		{name: "dense confident", groups: []TargetGroup{{
			Source: "S", State: SourceConfident, Committed: 0,
			Branches: []Branch{phase2Branch(a), phase2Branch(b)},
		}}, want: InterpretationConfident},
		{name: "dense ambiguous", groups: []TargetGroup{{
			Source: "S", State: SourceAmbiguous, Committed: -1,
			Branches: []Branch{phase2Branch(a), phase2Branch(b)},
		}}, want: InterpretationAmbiguous},
		{name: "dense unresolved", groups: []TargetGroup{{
			Source: "S", State: SourceUnresolved, Committed: -1,
			Unresolved: []Completion{a, b, c},
		}}, want: InterpretationUnresolved},
		{name: "unresolved vetoes confident branch", groups: []TargetGroup{
			{Source: "clean", State: SourceConfident, Committed: 0, Branches: []Branch{phase2Branch(a)}},
			{Source: "bridge", State: SourceUnresolved, Committed: -1, Unresolved: []Completion{b, c}},
		}, want: InterpretationUnresolved},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := buildInterpretationHandoff(tt.groups)
			if !h.Available || h.Route != tt.want {
				t.Fatalf("handoff=%+v, want route %s", h, tt.want)
			}
			if tt.want == InterpretationCold {
				if h.EffectiveN != 0 || h.Maturity != 0 {
					t.Fatalf("cold handoff carried support: %+v", h)
				}
				return
			}
			if h.EffectiveN != 20 || math.Abs(h.Maturity-.8) > 1e-12 {
				t.Fatalf("dense handoff support=(%d, %.3f), want (20, .800)", h.EffectiveN, h.Maturity)
			}
		})
	}
}

func TestToy7MaturityFadesInWithoutChangingTopology(t *testing.T) {
	last := -1.0
	for _, trials := range []int{1, 5, 20, 100} {
		a := phase2Completion("S", "A", .9, trials)
		b := phase2Completion("S", "B", .86, trials)
		h := buildInterpretationHandoff([]TargetGroup{{
			Source: "S", State: SourceAmbiguous, Committed: -1,
			Branches: []Branch{phase2Branch(a), phase2Branch(b)},
		}})
		if h.Route != InterpretationAmbiguous {
			t.Fatalf("N=%d changed irreducible topology to %s", trials, h.Route)
		}
		if h.Maturity <= last {
			t.Fatalf("maturity not monotone: N=%d got %.3f after %.3f", trials, h.Maturity, last)
		}
		last = h.Maturity
	}
}

func TestToy7EffectiveSupportUsesWeakestCandidate(t *testing.T) {
	dense := phase2Completion("S", "dense", .9, 100)
	thin := phase2Completion("S", "thin", .85, 1)
	h := buildInterpretationHandoff([]TargetGroup{{
		Source: "S", State: SourceAmbiguous, Committed: -1,
		Branches: []Branch{phase2Branch(dense), phase2Branch(thin)},
	}})
	if h.EffectiveN != 1 || math.Abs(h.Maturity-(1.0/6.0)) > 1e-12 {
		t.Fatalf("dense candidate inflated comparison maturity: %+v", h)
	}
}

func TestToy7HandoffRendersIntoExistingPresent(t *testing.T) {
	b := NewBeing(nil, nil)
	a := phase2Completion("SIG_S", "A", .9, 20)
	other := phase2Completion("SIG_S", "B", .86, 20)
	b.Interpretation = buildInterpretationHandoff([]TargetGroup{{
		Source: "SIG_S", State: SourceAmbiguous, Committed: -1,
		Branches: []Branch{phase2Branch(a), phase2Branch(other)},
	}})

	present := b.present(nil, nil)
	for _, want := range []string{
		"COUNTED INTERPRETATION — route=ambiguous effective-N=20 maturity=0.800",
		"Hold every branch",
		"[N=20] slot:test completion A",
		"[N=20] slot:test completion B",
	} {
		if !strings.Contains(present, want) {
			t.Fatalf("present missing %q:\n%s", want, present)
		}
	}
}

func TestToy7NovelImpulseInstallsColdHandoff(t *testing.T) {
	b := NewBeing(nil, nil)
	result := b.ObserveImpulse([]Triple{{
		Mode: ModeExact, Subject: "topic:new", Predicate: "mentions", Object: "unknown",
	}}, nil)
	if !result.Resolution.Novel || result.Interpretation.Route != InterpretationCold {
		t.Fatalf("novel impulse did not enter cold express route: %+v", result)
	}
	if b.Interpretation.Route != InterpretationCold || !b.Interpretation.Available {
		t.Fatalf("cold handoff not retained for next inference call: %+v", b.Interpretation)
	}
	if present := b.present(nil, nil); !strings.Contains(present, "route=cold") {
		t.Fatalf("cold route missing from next inference context:\n%s", present)
	}
}
