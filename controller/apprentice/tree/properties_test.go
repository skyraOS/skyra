// One property test per Core.lean theorem not already tested in
// kinds_test.go, named for it, over small worlds enumerated in full.
package tree

import (
	"slices"
	"testing"
)

func TestIdentifyParts(t *testing.T) {
	for _, c := range pool {
		if got, ok := Identify(c.Binary(), c.Shape()); !ok || got != c {
			t.Errorf("identify_parts: %s", name(c))
		}
	}
	for _, b := range []Binary{"", "sha256:", "md5:" + hash('a')[7:], Binary("cdhash:" + string(hash('a')[7:47]) + "A"), hash('g')} {
		if _, ok := Identify(b, shapeA); ok {
			t.Errorf("identify: %q is not a known scheme and digest", b)
		}
	}
	if _, ok := Identify(binX, ""); ok {
		t.Error("identify: an empty shape")
	}
}

func TestObservationTypeOfParts(t *testing.T) {
	o := typeOf(x, shapeS)
	if got, ok := TypeOf(o.Contract(), o.Output()); !ok || got != o {
		t.Error("ObservationType.of_parts")
	}
	if _, ok := TypeOf(x, ""); ok {
		t.Error("ObservationType.of: an empty output shape")
	}
}

func TestApprenticeContracts(t *testing.T) {
	cs := app.Contracts()
	for i, c := range cs {
		if !app.Mine(c) {
			t.Errorf("Apprentice.contracts_mine: %s", name(c))
		}
		if slices.Contains(cs[:i], c) {
			t.Errorf("Apprentice.contracts_nodup: %s twice", name(c))
		}
	}
	c := func(title string) Contract { return contract(binApp, Schema(`{"title":"`+title+`"}`)) }
	if _, ok := ApprenticeOf(binApp, c("o"), c("r"), c("e"), x, c("w"), c("i"), c("d"), c("g"), c("cd"), c("pr")); ok {
		t.Error("Apprentice.of: a contract of another binary")
	}
	if _, ok := ApprenticeOf(binApp, c("o"), c("r"), c("e"), c("e"), c("w"), c("i"), c("d"), c("g"), c("cd"), c("pr")); ok {
		t.Error("Apprentice.of: one contract in two places")
	}
}

// Every observation the world makes, effects and computations, two deep.
func observations(t *testing.T) []Observation {
	t.Helper()
	w := newReturnWorld(t)
	effects := []Observation{w.goal, w.wrote, w.came, w.child}
	out := slices.Clone(effects)
	td := typeOf(app.Delegate(), shapeA)
	for _, ms := range observationLists(effects, 2) {
		if o, ok := Observe(app, KindDelegate, td, ms); ok {
			out = append(out, o)
		}
	}
	k := out[len(out)-1]
	out = append(out, must(Observe(app, KindDelegate, td, []Observation{k, w.other})))
	return out
}

func TestObserve(t *testing.T) {
	w := newReturnWorld(t)
	members := observationLists([]Observation{w.goal, w.came}, 2)
	for _, k := range kinds {
		for _, c := range append(app.Contracts(), x, y) {
			for _, s := range []Schema{shapeA, shapeS} {
				ty := typeOf(c, s)
				for _, ms := range members {
					o, ok := Observe(app, k, ty, ms)
					// observe_eq_some
					if ok != Admits(app, k, ty, ms) || ok && (o.Type() != ty || !slices.EqualFunc(o.Members(), ms, Observation.Equal)) {
						t.Fatalf("observe_eq_some: %v through %s, %d members", k, name(c), len(ms))
					}
					if !ok {
						continue
					}
					// Observation.effect_xor_computation
					if o.IsEffect() == o.IsComputation() {
						t.Fatalf("Observation.effect_xor_computation")
					}
					// observe_computation, observe_effect_or_invoke
					if o.IsComputation() && (k != KindDelegate || c != app.Delegate()) {
						t.Fatalf("observe_computation: %v through %s", k, name(c))
					}
					// observe_not_through
					if len(ms) == 0 && slices.Contains([]Contract{app.Observe(), app.Root(), app.Return(), app.Invoke(), app.Delegate()}, c) {
						t.Fatalf("observe_not_through: %v through %s", k, name(c))
					}
					// observe_composed
					if len(ms) == 0 && (c == app.Description() || c == app.Reach()) && k != KindDelegate {
						t.Fatalf("observe_composed: %v through %s", k, name(c))
					}
				}
			}
		}
	}
	// No member twice.
	if _, ok := Observe(app, KindDelegate, typeOf(app.Delegate(), shapeA), []Observation{w.goal, w.goal}); ok {
		t.Error("admits: a member twice")
	}
}

func TestRootObservation(t *testing.T) {
	for _, c := range append(app.Contracts(), x) {
		o, ok := RootObservation(app, typeOf(c, shapeS))
		if ok != (c == app.Root()) || ok && !o.IsEffect() {
			t.Errorf("rootObservation through %s = %v", name(c), ok)
		}
	}
}

func subset(xs, ys []Observation) bool {
	for _, x := range xs {
		if !slices.ContainsFunc(ys, x.Equal) {
			return false
		}
	}
	return true
}

func TestMayPassIff(t *testing.T) {
	obs := observations(t)[:4]
	for _, sender := range observationLists(obs, 2) {
		for _, passed := range observationLists(obs, 2) {
			if MayPass(sender, passed) != subset(passed, sender) {
				t.Fatalf("mayPass_iff")
			}
		}
	}
}

func TestReceive(t *testing.T) {
	obs := observations(t)[:4]
	for _, r := range observationLists(obs, 2) {
		for _, p := range observationLists(obs, 2) {
			if got := Receive(DirInvoke, r, p); !slices.EqualFunc(got, p, Observation.Equal) {
				t.Fatal("receive: a child starts with only what it is passed")
			}
			if got := Receive(DirRet, r, p); !slices.EqualFunc(got, append(slices.Clone(r), p...), Observation.Equal) {
				t.Fatal("receive: a parent keeps its scope and adds what is returned")
			}
		}
	}
}

func TestNodeMayReturnIff(t *testing.T) {
	obs := observations(t)
	for _, received := range observationLists(obs, 1) {
		for _, returned := range observationLists(obs, 2) {
			want := true
			for _, o := range returned {
				for _, x := range Lineage(o) {
					for _, r := range received {
						if slices.ContainsFunc(Lineage(r), x.Equal) {
							want = false
						}
					}
				}
			}
			if NodeMayReturn(received, returned) != want {
				t.Fatal("nodeMayReturn_iff")
			}
		}
	}
}

// --- the store ---

var bases = domain

// Every store reachable from the empty one in at most three operations.
func stores() []Store {
	out := []Store{EmptyStore()}
	frontier := out
	for range 3 {
		var next []Store
		for _, s := range frontier {
			for _, b := range bases {
				_, t := See(s, b)
				next = append(next, t)
				for _, cel := range []string{"cel-1", "cel-2"} {
					if t, ok := Rewrite(s, b, cel); ok {
						next = append(next, t)
					}
				}
			}
		}
		out = append(out, next...)
		frontier = next
	}
	return out
}

// Store.WellFormed: a slot exactly for every known base type.
func wellFormed(s Store) bool {
	for _, b := range bases {
		if _, ok := s.Guard(b); ok != slices.Contains(s.Known(), b) {
			return false
		}
	}
	return len(s.guards) == len(s.known)
}

func TestStore(t *testing.T) {
	if !wellFormed(EmptyStore()) {
		t.Fatal("Store.empty_wellFormed")
	}
	for _, s := range stores() {
		if !wellFormed(s) {
			t.Fatal("a reachable store is not well formed")
		}
		for _, b := range bases {
			n, s2 := See(s, b)
			if (n == SeenUnseen) != !slices.Contains(s.Known(), b) {
				t.Fatal("see_unseen_iff")
			}
			if !slices.Contains(s2.Known(), b) {
				t.Fatal("see_known")
			}
			for _, a := range s.Known() {
				if !slices.Contains(s2.Known(), a) {
					t.Fatal("see_grows")
				}
			}
			for _, a := range bases {
				g1, ok1 := s.Guard(a)
				g2, ok2 := s2.Guard(a)
				if a != b && (g1 != g2 || ok1 != ok2) {
					t.Fatal("see_other")
				}
			}
			if !wellFormed(s2) {
				t.Fatal("see_wellFormed")
			}
			if s3, ok := Rewrite(s, b, "cel-9"); ok {
				if !slices.Equal(s3.Known(), s.Known()) {
					t.Fatal("rewrite_known")
				}
				if !wellFormed(s3) {
					t.Fatal("rewrite_wellFormed")
				}
				if g, _ := s3.Guard(b); g != WrittenGuard("cel-9") {
					t.Fatal("rewrite: the one slot is overwritten")
				}
			} else if slices.Contains(s.Known(), b) {
				t.Fatal("rewrite: a known base type has a slot")
			}
			r, ok := Refinement(toy, s, b)
			if ok != slices.Contains(s.Known(), b) {
				t.Fatal("refinement_known")
			}
			if r2, _ := Refinement(toy, s, b); ok && r2 != r {
				t.Fatal("one_refinement")
			}
		}
	}
}

func TestRetContinueIff(t *testing.T) {
	for _, n := range []Seen{SeenSeen, SeenUnseen} {
		for _, g := range []bool{true, false} {
			if (ReturnOf(n, g) == ReturnContinue) != (n == SeenSeen && g) {
				t.Errorf("ret_continue_iff: %v %v", n, g)
			}
		}
	}
}

func TestInvokeProcedureIff(t *testing.T) {
	for _, n := range []Seen{SeenSeen, SeenUnseen} {
		for _, g := range []bool{true, false} {
			if (InvokeOf(n, g) == InvokeProcedure) != (n == SeenSeen && g) {
				t.Errorf("invoke_procedure_iff: %v %v", n, g)
			}
		}
	}
}

// --- pass time and provision time ---

func TestPassableIff(t *testing.T) {
	for _, held := range reaches(others, 2) {
		for _, passed := range reaches(others, 2) {
			want := true
			for _, p := range passed {
				want = want && slices.Contains(held, p)
			}
			if Passable(held, passed) != want {
				t.Fatalf("passable_iff: held %s, passed %s", names(held), names(passed))
			}
		}
	}
}

func TestOffered(t *testing.T) {
	for _, k := range kinds {
		for _, p := range reaches(others, 2) {
			o := Offered(app, k, p)
			if len(o.Invoke) != 0 && k != KindAct {
				t.Errorf("offered_invoke: %v invokes", k)
			}
			if !slices.Equal(o.Apprentice, Loaded(app, k)) {
				t.Errorf("offered_apprentice: %v", k)
			}
			if k == KindAct && !slices.Equal(o.Invoke, p) || k == KindDelegate && !slices.Equal(o.PassOnly, p) {
				t.Errorf("offered: %v is not given what it was passed", k)
			}
		}
	}
}

func TestLoadedNeverPrivileged(t *testing.T) {
	for _, k := range kinds {
		l := Loaded(app, k)
		if slices.Contains(l, app.Observe()) || slices.Contains(l, app.Root()) ||
			slices.Contains(l, app.Description()) || slices.Contains(l, app.Reach()) {
			t.Errorf("loaded_never_privileged: %v", k)
		}
		for _, c := range l {
			if !app.Mine(c) {
				t.Errorf("loaded_mine: %v loads %s", k, name(c))
			}
		}
	}
}

// --- the two edges ---

// input_eq_iff, return_eq_iff, base_eq_iff: an edge's base type is one
// exactly when its parts are, over every binary, passed list and shape here.
func TestEdgesEqIff(t *testing.T) {
	type in struct {
		b Binary
		p []ReturnBase
	}
	var ins []in
	for _, b := range []Binary{hb, binX} {
		for _, p := range [][]ReturnBase{nil, domain[:1], domain[:2], {domain[1], domain[0]}} {
			ins = append(ins, in{b, p})
		}
	}
	shapes := []Schema{shapeA, shapeS}
	for _, x := range ins {
		for _, y := range ins {
			ix, iy := must(CreateInvoke(x.b, x.p)), must(CreateInvoke(y.b, y.p))
			same := x.b == y.b && slices.Equal(x.p, y.p)
			if (ix == iy) != same {
				t.Fatalf("input_eq_iff: %v %v", x, y)
			}
			for _, ox := range shapes {
				for _, oy := range shapes {
					if (must(CreateReturn(ix, ox)) == must(CreateReturn(iy, oy))) != (same && ox == oy) {
						t.Fatalf("base_eq_iff: %v %s, %v %s", x, ox, y, oy)
					}
				}
			}
		}
	}
}

func TestDoUnseenIff(t *testing.T) {
	is := []InputBase{must(CreateInvoke(hb, nil)), must(CreateInvoke(binX, nil)), must(CreateInvoke(hb, domain[:1]))}
	for _, history := range [][]InputBase{nil, is[:1], is[1:], is} {
		for _, i := range is {
			if (Do(history, i) == SeenUnseen) != !slices.Contains(history, i) {
				t.Fatalf("do_unseen_iff")
			}
		}
	}
}
