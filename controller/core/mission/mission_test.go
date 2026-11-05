package mission

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	root "skyra/controller/core/humanroot"
)

func TestPublishedContractsMatchProtocol(t *testing.T) {
	for path, schema := range map[string]*contract.Schema{
		"contracts/spec.json":                             SpecSchema(),
		"contracts/plan.json":                             PlanSchema(),
		"contracts/launch.json":                           LaunchSchema(),
		"../../targets/propose-spec/contract.json":        ProposeSchema(),
		"../../targets/view-plan-contracts/contract.json": ViewSchema(),
	} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var declared struct {
			Input *contract.Schema `json:"input"`
		}
		if err = json.Unmarshal(raw, &declared); err != nil {
			t.Fatal(err)
		}
		if !contract.Same(schema, declared.Input) {
			t.Fatalf("%s differs from the typed protocol", path)
		}
	}
}

func fixture(t *testing.T) (Catalogue, root.State, []Descriptor) {
	t.Helper()
	scope := boundary.Authority{Read: []boundary.PathGrant{{Path: "/workspace", Scope: "subpath"}}}
	first, _ := Identify("sha256:"+strings.Repeat("1", 64), &contract.Schema{Title: "one", Type: "string"})
	second, _ := Identify(first.Binary, &contract.Schema{Title: "two", Type: "number"})
	c, err := NewCatalogue(scope, []Descriptor{{Target: "tool", Contract: first, Pin: "pin", Authority: scope}, {Target: "tool", Contract: second, Pin: "pin", Authority: scope}})
	if err != nil {
		t.Fatal(err)
	}
	entries := c.Entries()
	refs := []root.Capability{entries[0].Ref, entries[1].Ref}
	s, err := root.New(root.Config{Mission: "mission", Human: "human", Entrypoints: []root.Capability{"launch"}, Ceiling: refs})
	if err != nil {
		t.Fatal(err)
	}
	return c, s, entries
}
func commit(t *testing.T, s root.State, event root.Event) (root.State, root.Effect) {
	t.Helper()
	s, e := root.Step(s, root.HumanIngress, event)
	if e.Kind != root.Commit {
		t.Fatal(e.Reason)
	}
	return root.Step(s, root.StorageIngress, root.Committed{Stamp: e.Stamp})
}
func TestContractTypesMatchWorkerKnownAnswers(t *testing.T) {
	c, err := Identify("sha256:"+strings.Repeat("1", 64), &contract.Schema{Type: "object"})
	if err != nil {
		t.Fatal(err)
	}
	invoke, err := InvokeType(c, []string{"sha256:" + strings.Repeat("2", 64)})
	if err != nil {
		t.Fatal(err)
	}
	if invoke != "sha256:eeed6b565a9469dc163b3757e1f49b23f795b1db9aed94b5609f92696f56ee3a" {
		t.Fatal("worker input type differs", invoke)
	}
	ret, err := ReturnType(invoke, `{"type":"string"}`)
	if err != nil {
		t.Fatal(err)
	}
	if ret != "sha256:283d0f465c4bf6e0017b53a7f678da10ee450cfd9c561770a6fd9d1b0e0bf3ad" {
		t.Fatal("worker return type differs", ret)
	}
}
func TestTwoContractsOnOneTargetClassifyAsDelegate(t *testing.T) {
	c, s, entries := fixture(t)
	spec := HumanSpec("do the work", []root.ID{}, []root.Capability{entries[0].Ref, entries[1].Ref}, DefaultProfile)
	resolved, err := c.Resolve(spec, s)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Kind != "delegate" || len(resolved.Targets) != 1 {
		t.Fatal("classification counted target names rather than contracts")
	}
	spec.Capabilities = spec.Capabilities[:1]
	resolved, err = c.Resolve(spec, s)
	if err != nil || resolved.Kind != "act" {
		t.Fatal("one contract is not act")
	}
}
func TestSpecReferencesBindScopesAndDoNotGrantByName(t *testing.T) {
	c, s, entries := fixture(t)
	narrow := entries[0]
	narrow.Ref = ""
	narrow.Authority.Read[0].Path = "/workspace/sub"
	changed, err := NewCatalogue(c.Ceiling(), []Descriptor{narrow})
	if err != nil {
		t.Fatal(err)
	}
	if changed.Entries()[0].Ref == entries[0].Ref {
		t.Fatal("resource scope did not change capability identity")
	}
	wide := entries[0]
	wide.Ref = ""
	wide.Authority.Write = []boundary.PathGrant{{Path: "/workspace", Scope: "subpath"}}
	if _, err = NewCatalogue(c.Ceiling(), []Descriptor{wide}); err == nil {
		t.Fatal("catalogue widened root authority")
	}
	if _, err = c.Resolve(HumanSpec("work", nil, []root.Capability{"read"}, DefaultProfile), s); err == nil {
		t.Fatal("plain tool name granted authority")
	}
	entries[0].Ref = "forged"
	if c.Entries()[0].Ref == "forged" {
		t.Fatal("catalogue aliases caller entries")
	}
}
func TestLaunchChecksRecordedInvocationAndExactGrant(t *testing.T) {
	c, s, entries := fixture(t)
	s, input := commit(t, s, root.Authored{RequestID: "input", Value: json.RawMessage(`"evidence"`)})
	spec := HumanSpec("work", input.Observations, []root.Capability{entries[0].Ref}, DefaultProfile)
	raw, _ := json.Marshal(spec)
	s, result := commit(t, s, root.Authored{RequestID: "spec", Value: raw})
	specID := result.Observations[0]
	s, e := commit(t, s, root.Invoke{RequestID: "launch", Target: "launch", Inputs: append([]root.ID{specID}, spec.Inputs...), Grant: spec.Capabilities})
	if _, err := c.ValidateHandoff(e, s); err != nil {
		t.Fatal(err)
	}
	forged := e
	inv := *e.Invocation
	inv.ID = "unknown"
	forged.Invocation = &inv
	if _, err := c.ValidateHandoff(forged, s); err == nil {
		t.Fatal("fabricated dispatch accepted")
	}
	// The human root can choose held capabilities; the Launch contract additionally
	// requires them to be the exact proposal this Spec observation contains.
	s2, other := commit(t, s, root.Invoke{RequestID: "other", Target: "launch", Inputs: append([]root.ID{specID}, spec.Inputs...), Grant: []root.Capability{entries[1].Ref}})
	if _, err := c.ValidateHandoff(other, s2); err == nil {
		t.Fatal("grant changed independently of spec")
	}
}
func TestMalformedSpecsAndForeignReferencesFail(t *testing.T) {
	c, s, entries := fixture(t)
	cov := `"segments":[{"text":"x","source":"human"}],"coverage":{"words":1,"cited":0,"plan":0,"human":1}`
	for _, raw := range []string{
		`{"version":1,"goal":"x",` + cov + `,"inputs":[],"capabilities":[]}`,
		`{"version":2,"goal":"x","inputs":[],"capabilities":[]}`,
		`{"version":2,"goal":"x",` + cov + `,"inputs":[],"capabilities":[],"shell":true}`,
		`{"version":2,"goal":"y",` + cov + `,"inputs":[],"capabilities":[]}`,
		`{"version":2,"goal":"x","segments":[{"text":"x","source":"human"}],"coverage":{"words":1,"cited":1,"plan":0,"human":0},"inputs":[],"capabilities":[]}`,
		`{"version":2,"goal":"x","segments":[{"text":"x","source":"cited","of":"` + strings.Repeat("a", 64) + `"}],"coverage":{"words":1,"cited":1,"plan":0,"human":0},"inputs":[],"capabilities":[]}`,
		`{"version":2,"goal":"a b c","segments":[{"text":"a b c","source":"plan","of":"` + strings.Repeat("a", 64) + `"}],"coverage":{"words":3,"cited":0,"plan":3,"human":0},"inputs":[],"capabilities":[]}`,
	} {
		if _, err := DecodeSpec(json.RawMessage(raw)); err == nil {
			t.Fatal("malformed spec accepted", raw)
		}
	}
	spec := HumanSpec("work", []root.ID{root.ID(strings.Repeat("a", 64))}, []root.Capability{entries[0].Ref}, DefaultProfile)
	if _, err := c.Resolve(spec, s); err == nil {
		t.Fatal("foreign observation accepted")
	}
}

// A cited segment holds only if the root holds the observation it names and
// that observation says it, verbatim but for whitespace.
func TestCitedSegmentsAreCheckedAgainstTheHeldObservation(t *testing.T) {
	c, s, entries := fixture(t)
	s, said := commit(t, s, root.Authored{RequestID: "said", Value: json.RawMessage(`"please   echo the request\nback to me"`)})
	quote := func(text string, of root.ID) Spec {
		segments := []Segment{{Text: "Do this:", Source: ByPlan}, {Text: text, Source: Cited, Of: of}}
		goal, coverage := Assemble(segments)
		return Spec{Version: SpecVersion, Goal: goal, Segments: segments, Coverage: coverage, Inputs: []root.ID{}, Capabilities: []root.Capability{entries[0].Ref}, Profile: DefaultProfile}
	}
	ok := quote("echo the request back", said.Observations[0])
	if _, err := c.Resolve(ok, s); err != nil {
		t.Fatal(err)
	}
	if ok.Coverage != (Coverage{Words: 6, Cited: 4, Plan: 2}) {
		t.Fatalf("coverage %+v", ok.Coverage)
	}
	for name, spec := range map[string]Spec{
		"words never said": quote("echo the request twice", said.Observations[0]),
		"unheld source":    quote("echo the request back", root.ID(strings.Repeat("e", 64))),
	} {
		if _, err := c.Resolve(spec, s); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if h := HumanSpec("all mine here", nil, nil, DefaultProfile); h.Coverage != (Coverage{Words: 3, Human: 3}) {
		t.Fatalf("human coverage %+v", h.Coverage)
	}
}

// A human edit is a revision of the last Spec: what they kept keeps its
// source, what they wrote is theirs, so an edit never launders Plan's words.
func TestARevisionKeepsTheSourcesOfWhatTheHumanKept(t *testing.T) {
	c, s, entries := fixture(t)
	s, said := commit(t, s, root.Authored{RequestID: "said", Value: json.RawMessage(`"please echo the whole request back to me"`)})
	of := said.Observations[0]
	segments := []Segment{{Text: "Plan says:", Source: ByPlan}, {Text: "echo the whole request back", Source: Cited, Of: of}, {Text: "then stop.", Source: ByPlan}}
	goal, coverage := Assemble(segments)
	prev := Spec{Version: SpecVersion, Goal: goal, Segments: segments, Coverage: coverage, Inputs: []root.ID{of}, Capabilities: []root.Capability{entries[0].Ref}, Profile: DefaultProfile}
	caps := []root.Capability{entries[0].Ref}

	same := Revise(prev, goal, caps, prev.Profile)
	if same.Coverage != prev.Coverage || same.Goal != goal || !slices.Equal(same.Inputs, prev.Inputs) {
		t.Fatalf("an unchanged revision changed: %+v", same)
	}
	added := Revise(prev, "  Plan says: echo the whole request back then stop. And log it.", caps, prev.Profile)
	if added.Goal != "  Plan says: echo the whole request back then stop. And log it." || added.Coverage != (Coverage{Words: 12, Cited: 5, Plan: 4, Human: 3}) {
		t.Fatalf("appending: %+v", added)
	}
	if _, err := c.Resolve(added, s); err != nil {
		t.Fatalf("a kept quote no longer holds: %v", err)
	}
	broken := Revise(prev, "Plan says: echo the request back then stop.", caps, prev.Profile)
	if broken.Coverage.Cited != 0 || broken.Coverage.Human != 4 || broken.Coverage.Plan != 4 {
		t.Fatalf("a quote the edit broke stayed cited: %+v", broken.Coverage)
	}
	if _, err := c.Resolve(broken, s); err != nil {
		t.Fatal(err)
	}
	trimmed := Revise(prev, "echo the then stop.", caps, prev.Profile)
	if trimmed.Coverage != (Coverage{Words: 4, Human: 2, Plan: 2}) {
		t.Fatalf("a quote cut below three words stayed cited: %+v", trimmed.Coverage)
	}
	rewritten := Revise(prev, "Something else entirely.", caps, prev.Profile)
	if rewritten.Coverage != (Coverage{Words: 3, Human: 3}) {
		t.Fatalf("a rewrite kept sources: %+v", rewritten.Coverage)
	}
	for _, r := range []Spec{same, added, broken, trimmed, rewritten} {
		raw, _ := json.Marshal(r)
		if _, err := DecodeSpec(raw); err != nil {
			t.Fatalf("revision %q does not decode: %v", r.Goal, err)
		}
	}
}

// A Spec may be quoted, but only where a human wrote or cited it.
func TestQuotingASpecCountsOnlyWhatAHumanWroteOrCited(t *testing.T) {
	c, s, entries := fixture(t)
	prev := Spec{Version: SpecVersion, Segments: []Segment{{Text: "Plan wrote these words.", Source: ByPlan}, {Text: "And the human added these.", Source: ByHuman}}, Inputs: []root.ID{}, Capabilities: []root.Capability{entries[0].Ref}, Profile: DefaultProfile}
	prev.Goal, prev.Coverage = Assemble(prev.Segments)
	raw, _ := json.Marshal(prev)
	s, held := commit(t, s, root.Authored{RequestID: "prev", Value: raw})
	quote := func(text string) Spec {
		segments := []Segment{{Text: text, Source: Cited, Of: held.Observations[0]}}
		goal, coverage := Assemble(segments)
		return Spec{Version: SpecVersion, Goal: goal, Segments: segments, Coverage: coverage, Inputs: []root.ID{}, Capabilities: prev.Capabilities, Profile: DefaultProfile}
	}
	if _, err := c.Resolve(quote("the human added these"), s); err != nil {
		t.Fatal(err)
	}
	for _, laundered := range []string{"Plan wrote these words", "these words. And the human"} {
		if _, err := c.Resolve(quote(laundered), s); err == nil {
			t.Errorf("quoting %q of a Spec counted as cited", laundered)
		}
	}
}

// Plan's reply travels with its Spec as a note, outside the task; a Spec may
// have no goal yet, but only one with a goal resolves for Launch.
func TestANoteIsNotTheTaskAndAGoallessSpecCannotLaunch(t *testing.T) {
	c, s, entries := fixture(t)
	empty := Spec{Version: SpecVersion, Segments: []Segment{}, Inputs: []root.ID{}, Capabilities: []root.Capability{}, Profile: DefaultProfile, Note: "What should Launch do?"}
	raw, _ := json.Marshal(empty)
	if _, err := DecodeSpec(raw); err != nil {
		t.Fatalf("Plan's goal-less first reply does not decode: %v", err)
	}
	if _, err := c.Resolve(empty, s); err == nil {
		t.Fatal("a goal-less Spec resolved for Launch")
	}
	a := HumanSpec("do the work now", []root.ID{}, []root.Capability{entries[0].Ref}, DefaultProfile)
	b := a
	b.Note = "Nothing changed."
	if !a.SameTask(b) {
		t.Fatal("a note changed the task")
	}
	b.Capabilities = []root.Capability{entries[1].Ref}
	if a.SameTask(b) {
		t.Fatal("different tools are the same task")
	}
}

// A plan may attach only what the human holds: the view is the catalogue's
// entries among the held capabilities, never one the human does not hold.
func TestThePlanViewIsWhatTheHumanHolds(t *testing.T) {
	c, _, entries := fixture(t)
	v := c.ViewOf([]root.Capability{entries[1].Ref, root.Capability("sha256:" + strings.Repeat("f", 64))})
	if len(v.Contracts) != 1 || v.Contracts[0].Ref != entries[1].Ref {
		t.Fatalf("view of what is held: %+v", v.Contracts)
	}
	if len(c.ViewOf(nil).Contracts) != 0 || len(c.View().Contracts) != len(c.Entries()) {
		t.Fatal("a human holding nothing may attach something, or the whole catalogue view changed")
	}
}

// A Launch's root kind follows its profile's act bound: under the default,
// one contract acts and two delegate; under a bound of 3, three act.
func TestTheRootKindFollowsTheActBound(t *testing.T) {
	three := DefaultProfile
	three.Act = 3
	if DefaultProfile.KindOf(1) != "act" || DefaultProfile.KindOf(2) != "delegate" || three.KindOf(3) != "act" || three.KindOf(4) != "delegate" {
		t.Fatal("the root's kind does not follow the act bound")
	}
}

// Every Spec states the profile its Launch's tree runs under: one without
// one, or with one that cannot run a tree, is not a Spec. Launch's root kind
// and profile are the Spec's, and a changed profile is a changed task.
func TestASpecStatesItsProfile(t *testing.T) {
	c, s, entries := fixture(t)
	spec := HumanSpec("do the work", []root.ID{}, []root.Capability{entries[0].Ref, entries[1].Ref}, DefaultProfile)
	raw, _ := json.Marshal(spec)
	var loose map[string]any
	json.Unmarshal(raw, &loose)
	delete(loose, "profile")
	if _, err := DecodeSpec(mustJSON(loose)); err == nil {
		t.Fatal("a Spec without a profile decoded")
	}
	for _, p := range []Profile{{0, 8, 3, 32, 64}, {1, 0, 3, 32, 64}, {1, 8, -1, 32, 64}, {1, 8, 3, 33, 64}, {1, 8, 3, 32, 0}} {
		bad := spec
		bad.Profile = p
		if _, err := DecodeSpec(mustJSON(bad)); err == nil {
			t.Fatalf("a Spec with profile %+v decoded", p)
		}
	}
	wide := spec
	wide.Profile = Profile{Act: 2, Turns: 4, Depth: 1, FanOut: 8, Budget: 20}
	resolved, err := c.Resolve(wide, s)
	if err != nil || resolved.Kind != "act" || resolved.Profile != wide.Profile {
		t.Fatalf("two contracts under an act bound of 2: %v %+v", err, resolved)
	}
	if spec.SameTask(wide) {
		t.Fatal("a changed profile is the same task")
	}
}
