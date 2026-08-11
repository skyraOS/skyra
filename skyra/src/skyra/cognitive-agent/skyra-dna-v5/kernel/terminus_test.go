package kernel_test

// questions.md §4 — VARIATION HAS ONE SOURCE, AND THE TERMINUS HAS NONE.
//
//	run-tests invoked directly · returns "480 passed" · cited · accepted by a human
//	reading the same output. short range positive. long range positive. both wrong,
//	and wrong TOGETHER because they read one instrument.
//
// §1 says this one is worth building deliberately: "Demonstrating a known limit
// is stronger evidence the model behaves as understood than any passing case."
//
// So this file asserts the FAILURE. If any of it starts passing in the other
// direction, something changed and it was not this file.

import (
	"strings"
	"testing"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity/scripted"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

func TestTerminusIsUnfalsifiableFromInside(t *testing.T) {
	// `run-tests` LIES. It always reports 480 passing, and the suite is broken.
	// Nothing in the chain can tell, because nothing reads a second instrument.
	table := scripted.Table{
		"verify the build is green": selectOne("run-tests"),
		"tests": leaf(content.Content{
			"passed": content.Num(480),
			"failed": content.Num(0),
		}),
	}
	r := newRig(t, table, map[string]string{"run-tests": "tests"})

	for i := 0; i < 6; i++ {
		r.ask(t, "verify the build is green", content.Content{"need": content.Str("run the tests")})
	}

	liar := r.kids["run-tests"]

	// Every signal the design has is POSITIVE, at every range.
	citedAll := true
	for _, inv := range liar.History {
		if !inv.Cited {
			citedAll = false
		}
	}
	prof := liar.History[len(liar.History)-1].HopProfile()

	t.Logf("the liar: %d invocations, cited every time = %v, hop profile %v",
		len(liar.History), citedAll, prof)

	if !citedAll {
		t.Fatal("setup wrong: the terminus case requires it to be cited every time")
	}
	if prof[1] == 0 {
		t.Fatal("setup wrong: it should be locally validated")
	}

	// THE LIMIT, ASSERTED. There is exactly one partition — the shape never
	// changed, the value never moved, and a stable history is INDISTINGUISHABLE
	// from a correct one.
	parts := liar.Partitions(r.G.Deref())
	if len(parts) != 1 {
		t.Fatalf("want 1 partition, got %d", len(parts))
	}
	if d := parts[0].Diffs(); len(d) != 0 {
		t.Fatalf("want no diffs — the instrument never moves — got %v", d)
	}

	// This is Toy 1 arriving at the credit layer: STABILITY IS NOT CORRECTNESS.
	// The design's own summary claims Toy 1 PASSES; what passes is that the model
	// does not CONFUSE the two. It still cannot detect this case, and nothing in
	// the graph distinguishes this engram from a genuinely good one.
	t.Log("CONFIRMED LIMIT: a perfect record and a perfect lie are the same record.")
	t.Log("  correction here is EXTERNAL AND DISCRETE — a later event from the world.")
	t.Log("  use does not kill a bad implementation; USE IS THE POSITIVE SIGNAL, so use protects it.")

	// And the second half of §4: variation has one source, and this has none.
	// Nothing above it can reject, and nothing fails to place — so the population
	// stays at one, and competition cannot run on a population of one.
	if len(r.kids) != 1 {
		t.Fatal("setup wrong: the terminus case requires exactly one option")
	}
	t.Log("  STARVATION REQUIRES ALTERNATIVES: a parent with one option for a need")
	t.Log("  is not selecting, it is CALLING. No pressure exists to apply.")
}

func TestHistoryMustNotBecomeACache(t *testing.T) {
	// questions.md §6 — the rule is written and A STRUCTURAL FORM OF IT HAS NOT
	// BEEN FOUND. This asserts the hole is still open rather than pretending it is
	// closed.
	//
	//	A stored return may be read as WHAT B SAID BEFORE.
	//	It may never substitute for INVOKING B NOW.
	calls := 0
	table := scripted.Table{
		"is the door open": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				return grammar.Select{Members: []grammar.Member{{
					Engram: "door", Cite: in.ID, Entity: ent,
					Observations: []ids.ObservationID{in.ID},
				}}}, nil
			}
			return commitNaming(f)
		},
		"d": func(f frame.Frame, step int) (grammar.Emission, error) {
			calls++
			return grammar.Commit{Content: content.Content{"open": content.Bool(calls%2 == 0)}}, nil
		},
	}
	r := newRig(t, table, map[string]string{"door": "d"})
	r.ask(t, "is the door open", content.Content{"need": content.Str("check the door")})

	door := r.kids["door"]
	stored := door.History[0].Return

	// The hole, demonstrated: a parent can see this and there is nothing in the
	// grammar, the frame, or the authorization arithmetic that stops it from
	// treating it as the current answer. Every OTHER constraint in the set is
	// structural — this one is a sentence in a document.
	if stored == nil {
		t.Fatal("no stored return")
	}
	t.Logf("OPEN HOLE (§6): the stored return %s is visible in frame as %q,",
		stored.ID, stored.Content.Render())
	t.Log("  and nothing makes substituting it for a fresh invocation UNEXPRESSIBLE.")
	t.Log("  `has the door usually been open` is cheap; `is the door open right now`")
	t.Log("  requires looking — and only a convention currently separates them.")

	// What the design DOES have: the stored return and a fresh one are
	// distinguishable, because a fresh invocation appends a record and reading
	// does not. That is detection after the fact, not prevention.
	before := len(door.History)
	r.ask(t, "is the door open", content.Content{"need": content.Str("check the door")})
	if len(door.History) != before+1 {
		t.Error("a second ask did not append a record")
	}
	t.Logf("  detectable after the fact: %d records for %d real invocations", len(door.History), calls)
}

func TestSiblingBlindnessAcrossACompositionRemains(t *testing.T) {
	// questions.md §3 — exact for selections, UNSOLVED for compositions.
	// A rejected chain does not say which of A+B+C caused it.
	//
	// What the design claims is narrower and is what this checks: storing EVENTS
	// rather than counts keeps it APPROACHABLE, because set membership survives.
	table := scripted.Table{
		"compose three": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				var ms []grammar.Member
				for _, n := range []string{"a", "b", "c"} {
					ms = append(ms, grammar.Member{
						Engram: n, Cite: in.ID, Entity: ent,
						Observations: []ids.ObservationID{in.ID},
					})
				}
				return grammar.Select{Members: ms}, nil
			}
			// Compose ALL THREE into one observation. Many causes, one commit.
			rets := f.Returned()
			return grammar.Commit{Content: content.Content{
				"a": content.Ref(rets[0].ID),
				"b": content.Ref(rets[1].ID),
				"c": content.Ref(rets[2].ID),
			}}, nil
		},
	}
	kids := map[string]string{}
	for _, n := range []string{"a", "b", "c"} {
		expr := "leaf-" + n
		table[expr] = leaf(content.Content{"from": content.Str(n)})
		kids[n] = expr
	}
	r := newRig(t, table, kids)

	// NOTE: `top` selects and then ORIGINATES a composition referencing its
	// children. That is legal only because the composer is authoring references,
	// not inventing content — but the kernel refuses it, because the forfeit
	// fires at the first select and does not care what the content is.
	obs := r.K.Ingress(content.Content{"need": content.Str("all three")})
	_, _, err := r.K.Ask("compose three", ent, []*graph.Observation{obs})

	if err == nil {
		t.Fatal("expected the origination forfeit to fire")
	}
	if !strings.Contains(err.Error(), "origination forfeited") {
		t.Fatalf("unexpected: %v", err)
	}

	// THE COST, NAMEABLE RATHER THAN NECESSARY (possibilities.md). An
	// orchestrator cannot say "here are my children's results, plus this note
	// that is mine." It must SPAWN A LEAF to say it.
	//
	// This is the "spawn a leaf to add a sentence" tax the spec predicted "will
	// feel absurd the first time someone hits it." It was hit here, on the third
	// test written, doing the most ordinary thing an orchestrator can do:
	// combining three children into one answer.
	t.Log("CONFIRMED COST: a multi-child composition is UNEXPRESSIBLE by an orchestrator.")
	t.Logf("  refused with: %v", err)
	t.Log("  to combine A, B and C the parent must invoke a fourth possibility whose")
	t.Log("  job is to hold the references — which is exactly the plumbing the")
	t.Log("  singular-commit rule was meant to prevent.")
}
