package kernel_test

// The learning substrate, with everything else stubbed. questions.md §1:
//
//	FIRST BUILD   hand-seeded possibilities, deterministic bodies, one entity, no
//	              model, no create, no ingress. it tests the claim everything rests
//	              on and dodges every open item below.
//
//	WHAT IT TESTS does credit accumulate as described · does selection measurably
//	              improve · does the hop profile appear on a self-sustaining
//	              subtree · does the record channel stay bounded · does the
//	              TERMINUS case behave as predicted
//
// Each test below is one of those claims. They are claims, not coverage.

import (
	"fmt"
	"strings"
	"testing"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity/scripted"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/host"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/kernel"
)

const ent ids.EntityID = "scripted"

// rig builds a kernel over a table of deterministic bodies, plus named engrams
// hand-seeded as children of the root.
type rig struct {
	K    *kernel.Kernel
	G    *graph.Graph
	Root *graph.Engram
	kids map[string]*graph.Engram
}

func newRig(t *testing.T, table scripted.Table, kids map[string]string) *rig {
	t.Helper()
	g := graph.New()
	reg := entity.NewRegistry()
	reg.Register(ent, scripted.New(table))
	k := kernel.New(g, reg, &host.Local{Root: t.TempDir()}, nil)

	r := &rig{K: k, G: g, kids: map[string]*graph.Engram{}}

	// HAND-SEEDED. No discover, no create — this build dodges both.
	for name, expr := range kids {
		e := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: expr})
		g.Mint(e)
		k.Builtin[name] = e
		r.kids[name] = e
	}
	return r
}

func (r *rig) ask(t *testing.T, expr string, in content.Content) (*kernel.Root, *graph.Observation) {
	t.Helper()
	obs := r.K.Ingress(in)
	root, out, err := r.K.Ask(expr, ent, []*graph.Observation{obs})
	if err != nil {
		t.Fatalf("ask %q: %v", expr, err)
	}
	return root, out
}

// commitNaming is the orchestrator form: name what came back.
func commitNaming(f frame.Frame) (grammar.Emission, error) {
	rets := f.Returned()
	return grammar.Commit{Observed: rets[len(rets)-1].ID}, nil
}

// selectOne emits a one-member group citing the first thing in frame.
func selectOne(name string) func(frame.Frame, int) (grammar.Emission, error) {
	return func(f frame.Frame, step int) (grammar.Emission, error) {
		if step == 0 {
			in := f.Observations()[0]
			return grammar.Select{Members: []grammar.Member{{
				Engram: name, Cite: in.ID, Entity: ent,
				Observations: []ids.ObservationID{in.ID},
			}}}, nil
		}
		return commitNaming(f)
	}
}

// leaf returns fixed content. A leaf originates; it never selects.
func leaf(c content.Content) func(frame.Frame, int) (grammar.Emission, error) {
	return func(f frame.Frame, step int) (grammar.Emission, error) {
		return grammar.Commit{Content: c}, nil
	}
}

// ---------------------------------------------------------------------------
// 1 · Does credit accumulate as described?
// ---------------------------------------------------------------------------

func TestCreditAccumulates(t *testing.T) {
	// root → mid → leaf. Three levels, so the leaf can be accepted AT RANGE.
	table := scripted.Table{
		"top":  selectOne("mid"),
		"mid":  selectOne("bottom"),
		"body": nil,
	}
	table["leafwork"] = leaf(content.Content{"answer": content.Str("42")})

	r := newRig(t, table, map[string]string{
		"mid":    "mid",
		"bottom": "leafwork",
	})
	table["top"] = selectOne("mid")

	_, out := r.ask(t, "top", content.Content{"need": content.Str("a number")})

	// The chain composed: root named mid's return, which named the leaf's.
	if len(out.Cites()) == 0 {
		t.Fatal("root's commit cites nothing — credit cannot follow data")
	}

	// FORMED ON CITATION: the child exists because the commit cited it.
	mid := r.kids["mid"]
	if len(mid.History) != 1 {
		t.Fatalf("mid: want 1 invocation, got %d", len(mid.History))
	}
	if !mid.History[0].Cited {
		t.Fatal("mid was cited by the root's commit but Cited is false — the local verdict is wrong")
	}

	// THE HOP PROFILE. The leaf sits two levels down, so it should carry an
	// acceptance at hop 1 (its own parent, mid) and one at hop 2 (the root).
	bottom := r.kids["bottom"]
	if len(bottom.History) != 1 {
		t.Fatalf("bottom: want 1 invocation, got %d", len(bottom.History))
	}
	prof := bottom.History[0].HopProfile()
	t.Logf("leaf hop profile: %v", prof)
	if prof[1] == 0 {
		t.Error("leaf has no hop-1 acceptance — its own parent's commit did not disperse to it")
	}
	if prof[2] == 0 {
		t.Error("leaf has no hop-2 acceptance — ACCEPTANCE DID NOT WALK DOWN BY HOPS, " +
			"which is the design's central claim")
	}
}

// ---------------------------------------------------------------------------
// 2 · Records are events, not counters — so trajectory survives
// ---------------------------------------------------------------------------

func TestSelectionImproves(t *testing.T) {
	// THE CENTRAL CLAIM. Selection learns; synthesis does not. A parent with two
	// alternatives must measurably move between them as one degrades — with no
	// weights changing, only what the parent SEES on its children.
	//
	// STARVATION REQUIRES ALTERNATIVES: a parent with exactly one option for a
	// need is not selecting, it is calling. So there are two.
	n := 0
	degrade := 3 // `veteran` works this many times, then breaks

	// The parent's policy is the whole experiment: read the child summaries in
	// frame, prefer whoever has been cited most RECENTLY, and never carry an
	// error forward. It reads only what a possibility is allowed to see.
	pick := func(f frame.Frame) string {
		v := f.Engrams()
		best, bestScore := "", -1
		for _, name := range []string{"veteran", "rookie"} {
			s, ok := v.Summary(name)
			if !ok {
				// Never tried. Unproven material has to be reachable or nothing
				// new is ever selected.
				if bestScore < 0 {
					best, bestScore = name, 0
				}
				continue
			}
			// Trajectory: weight the RECENT window, which is only possible
			// because records are events and the log is ordered.
			score := 0
			h := s.Engram.History
			from := len(h) - 3
			if from < 0 {
				from = 0
			}
			for _, inv := range h[from:] {
				if inv.Cited {
					score++
				}
			}
			if score > bestScore {
				best, bestScore = name, score
			}
		}
		return best
	}

	var chose []string
	table := scripted.Table{
		"top": func(f frame.Frame, step int) (grammar.Emission, error) {
			in := f.Observations()[0]
			// Already got something usable? Name it.
			for _, r := range f.Returned() {
				if _, bad := r.Content["error"]; !bad {
					return grammar.Commit{Observed: r.ID}, nil
				}
			}
			if step >= 2 {
				// Both failed. Name the last one anyway — a rejection means
				// "somewhere along this chain it failed," not "E is bad."
				rets := f.Returned()
				return grammar.Commit{Observed: rets[len(rets)-1].ID}, nil
			}
			name := pick(f)
			if step == 1 {
				// The first choice failed; take the other one.
				if name == "veteran" {
					name = "rookie"
				} else {
					name = "veteran"
				}
			} else {
				chose = append(chose, name)
			}
			return grammar.Select{Members: []grammar.Member{{
				Engram: name, Cite: in.ID, Entity: ent,
				Observations: []ids.ObservationID{in.ID},
			}}}, nil
		},
		"vet": func(f frame.Frame, step int) (grammar.Emission, error) {
			n++
			if n <= degrade {
				return grammar.Commit{Content: content.Content{"ok": content.Bool(true)}}, nil
			}
			return grammar.Commit{Content: content.Content{"error": content.Str("broke")}}, nil
		},
		"rook": leaf(content.Content{"ok": content.Bool(true)}),
	}

	r := newRig(t, table, map[string]string{"veteran": "vet", "rookie": "rook"})
	for i := 0; i < 8; i++ {
		r.ask(t, "top", content.Content{"need": content.Str("work")})
	}

	t.Logf("chose, in order: %v", chose)

	vet := r.kids["veteran"]
	rook := r.kids["rookie"]

	// THE HISTORY PARTITIONS. Successes and errors partition away from each other
	// WITHOUT ANYONE MODELLING ERRORS, because their shapes differ.
	parts := vet.Partitions(r.G.Deref())
	if len(parts) != 2 {
		t.Fatalf("veteran: want 2 partitions (ok vs error), got %d: %v", len(parts), shapes(parts))
	}
	t.Logf("veteran partitions: %v", shapes(parts))

	// The measurable claim: the parent stopped reaching for the veteran first.
	early := strings.Join(chose[:3], ",")
	late := strings.Join(chose[len(chose)-3:], ",")
	t.Logf("first three choices: %s", early)
	t.Logf("last three choices:  %s", late)

	if !strings.Contains(early, "veteran") {
		t.Error("the veteran was never preferred early — the experiment did not run")
	}
	if strings.Contains(late, "veteran") {
		t.Errorf("still reaching for the veteran after it broke: %s — SELECTION DID NOT IMPROVE", late)
	}
	if rook.History[len(rook.History)-1].Cited != true {
		t.Error("the rookie's last return was not cited")
	}

	// PRUNING IS BEHAVIOURAL, NOT STRUCTURAL. Nothing was deleted: the veteran is
	// still a child, still selectable, still carrying its whole history. The
	// selection simply stopped being made.
	if _, ok := r.K.Builtin["veteran"]; !ok {
		t.Error("the veteran was removed — pruning must be behavioural, never structural")
	}
	t.Logf("veteran still present with %d records; it is starved, not deleted", len(vet.History))
}

func shapes(ps []*graph.Partition) []string {
	var out []string
	for _, p := range ps {
		out = append(out, fmt.Sprintf("%s×%d", p.Shape, len(p.Invs)))
	}
	return out
}

// ---------------------------------------------------------------------------
// 3 · The self-sustaining subtree reads off the hop profile
// ---------------------------------------------------------------------------

func TestSelfSustainingSubtreeIsVisible(t *testing.T) {
	// credit.md: a subtree deep enough that no boundary signal reached it would
	// sustain itself on being followed — internally coherent and undetectable
	// from inside. The claim is that LONG-RANGE RECORDS BREAK IT and the hop
	// profile is what makes the absence visible.
	//
	// Here: root cites `useful` but NOT `parasite`, though parasite's own parent
	// cites it happily. Parasite should show {hop1} and nothing beyond.
	table := scripted.Table{
		"top": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				return grammar.Select{Members: []grammar.Member{{
					Engram: "branch", Cite: in.ID, Entity: ent,
					Observations: []ids.ObservationID{in.ID},
				}}}, nil
			}
			return commitNaming(f)
		},
		// branch invokes BOTH, and cites only the useful one.
		"branch": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				return grammar.Select{Members: []grammar.Member{
					{Engram: "useful", Cite: in.ID, Entity: ent, Observations: []ids.ObservationID{in.ID}},
					{Engram: "parasite", Cite: in.ID, Entity: ent, Observations: []ids.ObservationID{in.ID}},
				}}, nil
			}
			// Name the FIRST return only. The second is invoked, returns fine, and
			// is never carried forward.
			return grammar.Commit{Observed: f.Returned()[0].ID}, nil
		},
		"u": leaf(content.Content{"useful": content.Bool(true)}),
		"p": leaf(content.Content{"parasite": content.Bool(true)}),
	}

	r := newRig(t, table, map[string]string{
		"branch":   "branch",
		"useful":   "u",
		"parasite": "p",
	})

	r.ask(t, "top", content.Content{"need": content.Str("x")})

	useful := r.kids["useful"].History[0]
	para := r.kids["parasite"].History[0]

	t.Logf("useful   cited=%v hops=%v", useful.Cited, useful.HopProfile())
	t.Logf("parasite cited=%v hops=%v", para.Cited, para.HopProfile())

	if !useful.Cited {
		t.Error("useful was carried forward but reads as uncited")
	}
	if para.Cited {
		t.Error("parasite was never carried forward but reads as cited — " +
			"citation is not being derived from content")
	}
	if len(para.Acceptances) != 0 {
		t.Errorf("parasite earned %d acceptances without being cited — "+
			"credit is not following data", len(para.Acceptances))
	}
	if len(useful.Acceptances) == 0 {
		t.Error("useful earned nothing")
	}
}

// ---------------------------------------------------------------------------
// 4 · Does the record channel stay bounded?
// ---------------------------------------------------------------------------

func TestExplorationCostsNothingPermanent(t *testing.T) {
	// "The durable graph records what WORKED, not what was tried." A possibility
	// may try five things, use one, and only the one attaches.
	table := scripted.Table{
		"top": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				var ms []grammar.Member
				for _, n := range []string{"a", "b", "c", "d", "e"} {
					ms = append(ms, grammar.Member{
						Engram: n, Cite: in.ID, Entity: ent,
						Observations: []ids.ObservationID{in.ID},
					})
				}
				return grammar.Select{Members: ms}, nil
			}
			return grammar.Commit{Observed: f.Returned()[2].ID}, nil // only "c"
		},
	}
	kids := map[string]string{}
	for _, n := range []string{"a", "b", "c", "d", "e"} {
		expr := "leaf-" + n
		table[expr] = leaf(content.Content{"from": content.Str(n)})
		kids[n] = expr
	}

	r := newRig(t, table, kids)
	root, _ := r.ask(t, "top", content.Content{"need": content.Str("one of five")})

	// FIVE were invoked. ONE attached to the root's children.
	if got := len(root.Engram.Engrams); got != 1 {
		t.Errorf("root attached %d children, want 1 — exploration is costing structure", got)
	}
	if _, ok := root.Engram.Engrams["c"]; !ok {
		t.Error("the cited child is not the one that attached")
	}

	// But all five have HISTORY, on the engram the parent owns. Discovery sees
	// attempts; selection sees successes.
	for n, e := range r.kids {
		if len(e.History) != 1 {
			t.Errorf("%s: want 1 history record, got %d", n, len(e.History))
		}
	}
}

// ---------------------------------------------------------------------------
// 5 · Eviction safety, and it must hold in the WORLD as well as the graph
// ---------------------------------------------------------------------------

func TestEvictedPossibilityLeavesNoTrace(t *testing.T) {
	// A possibility cut off mid-exploration has written nothing, so nothing
	// happened. The graph half is easy. The world half is what the chain-as-
	// transaction exists for.
	table := scripted.Table{
		// Never commits. Selects forever until the step bound evicts it.
		"spin": func(f frame.Frame, step int) (grammar.Emission, error) {
			in := f.Observations()[0]
			return grammar.Select{Members: []grammar.Member{{
				Engram: "noop", Cite: in.ID, Entity: ent,
				Observations: []ids.ObservationID{in.ID},
			}}}, nil
		},
		"n": leaf(content.Content{"noop": content.Bool(true)}),
	}
	r := newRig(t, table, map[string]string{"noop": "n"})

	obs := r.K.Ingress(content.Content{"need": content.Str("spin")})
	_, _, err := r.K.Ask("spin", ent, []*graph.Observation{obs})
	if err == nil {
		t.Fatal("a possibility that never commits should be evicted")
	}
	if !strings.Contains(err.Error(), "evicted") {
		t.Fatalf("want eviction, got %v", err)
	}

	// The evicted possibility invoked `noop` repeatedly and wrote NOTHING,
	// because every durable write is caused by a commit and it never committed.
	if n := len(r.kids["noop"].History); n != 0 {
		t.Errorf("evicted possibility left %d records — eviction safety is broken", n)
	}
}

// ---------------------------------------------------------------------------
// 6 · The chain is the transaction
// ---------------------------------------------------------------------------

func TestUncitedEffectIsDiscardedAndRecorded(t *testing.T) {
	// A wrote a file, B failed, the root accepted anyway. Which staged effects
	// land? The ones the credit walk reaches — and A's, if uncited, must be
	// DISCARDED AND THE DISCARD RECORDED. Never a silent no-op.
	dir := t.TempDir()
	g := graph.New()
	reg := entity.NewRegistry()
	h := &host.Local{Root: dir}

	var k *kernel.Kernel
	var root *kernel.Root

	table := scripted.Table{
		"top": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				return grammar.Select{Members: []grammar.Member{
					{Engram: "writes-kept", Cite: in.ID, Entity: ent, Observations: []ids.ObservationID{in.ID}},
					{Engram: "writes-dropped", Cite: in.ID, Entity: ent, Observations: []ids.ObservationID{in.ID}},
				}}, nil
			}
			return grammar.Commit{Observed: f.Returned()[0].ID}, nil // cite only the first
		},
	}

	stage := func(path, body string) func(frame.Frame, int) (grammar.Emission, error) {
		return func(f frame.Frame, step int) (grammar.Emission, error) {
			h := root.Stage(&kernel.Effect{
				Engram: f.Self(), Class: "storage", Path: path, Body: []byte(body),
			})
			return grammar.Commit{Content: content.Content{
				"wrote": content.WorldRef{Host: "here", Handle: h},
				"path":  content.Str(path),
			}}, nil
		}
	}
	table["kept"] = stage("kept.txt", "this one was cited")
	table["dropped"] = stage("dropped.txt", "this one was not")

	reg.Register(ent, scripted.New(table))
	k = kernel.New(g, reg, h, nil)
	for name, expr := range map[string]string{"writes-kept": "kept", "writes-dropped": "dropped"} {
		e := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: expr})
		g.Mint(e)
		k.Builtin[name] = e
	}

	obs := k.Ingress(content.Content{"need": content.Str("write two files")})

	// The root has to exist before the leaves stage into it, which is what makes
	// the chain the transaction boundary rather than the possibility.
	rootEngram := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: "top"})
	root = &kernel.Root{Span: "span-test", Engram: rootEngram}

	out, err := k.Drive(root, ent, []*graph.Observation{obs})
	if err != nil {
		t.Fatalf("ask: %v", err)
	}

	if err := k.Accept(root, out); err != nil {
		t.Fatalf("accept: %v", err)
	}

	// The cited one landed.
	if b, err := h.Storage().Read("kept.txt"); err != nil || string(b) != "this one was cited" {
		t.Errorf("cited effect did not materialize: %v %q", err, b)
	}
	// The uncited one did NOT.
	if _, err := h.Storage().Read("dropped.txt"); err == nil {
		t.Error("an uncited effect materialized — materialization is not following citation")
	}
	// And the discard is a RECORDED EVENT.
	ds := g.Discards()
	if len(ds) != 1 {
		t.Fatalf("want 1 recorded discard, got %d — a discard must never be a silent no-op", len(ds))
	}
	t.Logf("recorded discard: %s", ds[0].Detail)
}

// ---------------------------------------------------------------------------
// 7 · The origination forfeit
// ---------------------------------------------------------------------------

func TestOrchestratorCannotOriginate(t *testing.T) {
	// Originate only where it could not be mistaken for composition. The forfeit
	// fires at the FIRST select and is PERMANENT.
	table := scripted.Table{
		"top": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				return grammar.Select{Members: []grammar.Member{{
					Engram: "child", Cite: in.ID, Entity: ent,
					Observations: []ids.ObservationID{in.ID},
				}}}, nil
			}
			// Having selected, try to author anyway.
			return grammar.Commit{Content: content.Content{
				"invented": content.Str("a field none of my children produced"),
			}}, nil
		},
		"c": leaf(content.Content{"real": content.Bool(true)}),
	}
	r := newRig(t, table, map[string]string{"child": "c"})

	obs := r.K.Ingress(content.Content{"need": content.Str("x")})
	_, _, err := r.K.Ask("top", ent, []*graph.Observation{obs})
	if err == nil {
		t.Fatal("an orchestrator authored content and the kernel allowed it")
	}
	if !strings.Contains(err.Error(), "origination forfeited") {
		t.Fatalf("want the forfeit, got: %v", err)
	}
	t.Logf("refused: %v", err)
}

// ---------------------------------------------------------------------------
// 8 · Capability is name resolution
// ---------------------------------------------------------------------------

func TestCannotNameWhatIsNotInFrame(t *testing.T) {
	table := scripted.Table{
		"top": func(f frame.Frame, step int) (grammar.Emission, error) {
			in := f.Observations()[0]
			return grammar.Select{Members: []grammar.Member{{
				Engram: "something-nobody-gave-me", Cite: in.ID, Entity: ent,
				Observations: []ids.ObservationID{in.ID},
			}}}, nil
		},
	}
	r := newRig(t, table, nil)

	obs := r.K.Ingress(content.Content{"need": content.Str("x")})
	_, _, err := r.K.Ask("top", ent, []*graph.Observation{obs})
	if err == nil || !strings.Contains(err.Error(), "not in frame") {
		t.Fatalf("naming something out of frame should fail; got %v", err)
	}
}

// ---------------------------------------------------------------------------
// 9 · Monitor fires in EMITTED order, never completion order
// ---------------------------------------------------------------------------

func TestMonitorFiresInEmittedOrder(t *testing.T) {
	// Arrival order is a scheduling accident. If context is shaped by it, replay
	// diverges and partial completion arrives through the back door.
	//
	// `slow` is emitted FIRST and returns LAST. Monitor must still see it first.
	done := make(chan struct{})
	table := scripted.Table{
		"top": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				return grammar.Select{Members: []grammar.Member{
					{Engram: "slow", Cite: in.ID, Entity: ent, Observations: []ids.ObservationID{in.ID}},
					{Engram: "fast", Cite: in.ID, Entity: ent, Observations: []ids.ObservationID{in.ID}},
				}}, nil
			}
			return commitNaming(f)
		},
		"s": func(f frame.Frame, step int) (grammar.Emission, error) {
			<-done // wait for fast
			return grammar.Commit{Content: content.Content{"who": content.Str("slow")}}, nil
		},
		"fst": func(f frame.Frame, step int) (grammar.Emission, error) {
			close(done)
			return grammar.Commit{Content: content.Content{"who": content.Str("fast")}}, nil
		},
	}

	g := graph.New()
	reg := entity.NewRegistry()
	var captured *scripted.Interpreter
	first := true
	reg.Register(ent, func() entity.Interpreter {
		i := &scripted.Interpreter{T: table}
		if first { // the top-level possibility is opened first
			captured = i
			first = false
		}
		return i
	})
	k := kernel.New(g, reg, &host.Local{Root: t.TempDir()}, nil)
	for name, expr := range map[string]string{"slow": "s", "fast": "fst"} {
		e := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: expr})
		g.Mint(e)
		k.Builtin[name] = e
	}

	obs := k.Ingress(content.Content{"need": content.Str("order")})
	if _, _, err := k.Ask("top", ent, []*graph.Observation{obs}); err != nil {
		t.Fatalf("ask: %v", err)
	}

	if len(captured.Seen) != 2 {
		t.Fatalf("monitor saw %d returns, want 2", len(captured.Seen))
	}
	who := func(o *graph.Observation) string {
		if s, ok := o.Content["who"].(content.Str); ok {
			return string(s)
		}
		return "?"
	}
	if who(captured.Seen[0]) != "slow" || who(captured.Seen[1]) != "fast" {
		t.Errorf("monitor order was [%s %s], want [slow fast] — "+
			"members are being pushed in COMPLETION order, and replay will diverge",
			who(captured.Seen[0]), who(captured.Seen[1]))
	}
}
