package shell

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"skyra/apprentice/proc"
	"skyra/typesys-prototype/tree"
)

var str = &proc.Schema{Type: "string"}

var reach = map[string][]*proc.Schema{
	"model": {{Type: "object", Required: []string{"text"}, Fields: map[string]*proc.Schema{"text": str, "fresh": {Type: "bool"}, "tools": {Type: "array", Items: str}, "instructions": {Type: "any"}}}},
	"echo":  {{Type: "object", Required: []string{"text"}, Fields: map[string]*proc.Schema{"text": str}}},
}

var line = regexp.MustCompile(`(?m)^  (\d+): .*\((.*)\)$`)

// model is a scripted model: it reads its prompt as a model would, and
// answers by its kind and what has come back.
func model(prompt string) string {
	for _, m := range line.FindAllStringSubmatch(prompt, -1) {
		if m[2] == "what came back" || strings.HasSuffix(m[2], " returned") {
			return `[{"return":{"members":[` + m[1] + `]}}]`
		}
	}
	if strings.Contains(prompt, "of a tree, an act node") {
		return "Here:\n```json\n" + `[{"create_input":"{\"text\":\"hello\"}"},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]` + "\n```"
	}
	return `[{"goal":{"of":0,"lines":[1,1]}},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`
}

// kernel plays the kernel: it resumes the frame until it returns, running
// every instruction it emits, and returns the answer.
func kernel(t *testing.T, dir, frame string, ask Ask, answer func(to string, payload json.RawMessage) Step) (json.RawMessage, int) {
	t.Helper()
	input, _ := json.Marshal(map[string]any{"tree": ask})
	var steps []Step
	for resumes := 1; resumes < 50; resumes++ {
		out, err := Resume(dir, frame, input, reach, steps)
		if err != nil {
			t.Fatalf("resume %d: %v", resumes, err)
		}
		var ins []struct {
			To      string          `json:"to"`
			Payload json.RawMessage `json:"payload"`
		}
		if out[0] == '[' {
			if err := json.Unmarshal(out, &ins); err != nil {
				t.Fatal(err)
			}
		} else {
			ins = make([]struct {
				To      string          `json:"to"`
				Payload json.RawMessage `json:"payload"`
			}, 1)
			if err := json.Unmarshal(out, &ins[0]); err != nil {
				t.Fatal(err)
			}
		}
		if len(ins) == 1 && ins[0].To == frame {
			return ins[0].Payload, resumes
		}
		for _, in := range ins {
			steps = append(steps, answer(in.To, in.Payload))
		}
	}
	t.Fatal("the tree never returned")
	return nil, 0
}

func run(to string, payload json.RawMessage) Step {
	switch to {
	case "model":
		var p struct{ Text string }
		json.Unmarshal(payload, &p)
		v, _ := json.Marshal(map[string]string{"text": model(p.Text)})
		return Step{To: to, State: "return", Value: v}
	case "echo":
		var p struct{ Text string }
		json.Unmarshal(payload, &p)
		v, _ := json.Marshal(p.Text)
		return Step{To: to, State: "return", Value: v}
	}
	return Step{To: to, State: "fault", Fault: &Fault{"boundary", to + " is not in reach"}}
}

// A delegate root delegates to an act child, which invokes echo; what echo
// brought back comes up to the root and out of the tree.
func TestATreeThroughTheKernel(t *testing.T) {
	dir := t.TempDir()
	ask := Ask{Goal: "say hello", Kind: "delegate", Actor: "model", Reach: []string{"echo"}}
	got, resumes := kernel(t, dir, "frame-1", ask, run)
	if string(got) != `{"returned":["hello"]}` {
		t.Fatalf("the tree returned %s", got)
	}
	// root turn, child turn, echo, child turn, root turn: five rounds, and the return.
	if resumes != 6 {
		t.Fatalf("%d resumes", resumes)
	}

	// A second run over the same records: the root's input is seen.
	if got, _ := kernel(t, dir, "frame-2", ask, run); string(got) != `{"returned":["hello"]}` {
		t.Fatalf("the second run returned %s", got)
	}
	recs := records(t, dir, "frame-2")
	if !strings.Contains(recs, `"do":"seen"`) || !strings.Contains(recs, `"see":"seen"`) {
		t.Fatalf("the second run did not see what the first wrote:\n%s", recs)
	}
}

// An act root invokes echo itself.
func TestAnActRoot(t *testing.T) {
	got, _ := kernel(t, t.TempDir(), "f", Ask{Goal: "say hello", Kind: "act", Actor: "model", Reach: []string{"echo"}}, run)
	if string(got) != `{"returned":["hello"]}` {
		t.Fatalf("the tree returned %s", got)
	}
}

// A model that writes nonsense faults its node, and the tree says why.
func TestAFaultComesOut(t *testing.T) {
	bad := func(to string, payload json.RawMessage) Step {
		if to == "model" {
			return Step{To: to, State: "return", Value: json.RawMessage(`{"text":"I cannot"}`)}
		}
		return run(to, payload)
	}
	got, _ := kernel(t, t.TempDir(), "f", Ask{Goal: "g", Kind: "delegate", Actor: "model", Reach: []string{"echo"}}, bad)
	if !strings.Contains(string(got), "faulted") {
		t.Fatalf("the tree returned %s", got)
	}
}

// An ask that names what is not in reach is refused before anything runs.
func TestRefused(t *testing.T) {
	input, _ := json.Marshal(map[string]any{"tree": Ask{Goal: "g", Kind: "act", Actor: "model", Reach: []string{"shell"}}})
	out, err := Resume(t.TempDir(), "f", input, reach, nil)
	if err != nil || !strings.Contains(string(out), "refused") {
		t.Fatalf("%s %v", out, err)
	}
}

func records(t *testing.T, dir, frame string) string {
	s := &shell{dir: dir + "/tree", frame: name(frame)}
	b, err := os.ReadFile(s.path("records", s.frame))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Trees run at once over one state lose no write: every input either seen
// is in the history.
func TestConcurrentTreesLoseNothing(t *testing.T) {
	dir := t.TempDir()
	done := make(chan string)
	for i := range 8 {
		go func() {
			got, _ := kernel(t, dir, "frame-"+string(rune('a'+i)), Ask{Goal: "say hello", Kind: "delegate", Actor: "model", Reach: []string{"echo"}}, run)
			done <- string(got)
		}()
	}
	for range 8 {
		if got := <-done; got != `{"returned":["hello"]}` {
			t.Errorf("a tree returned %s", got)
		}
	}
	b, _ := os.ReadFile(dir + "/tree/history.json")
	var ids []string
	json.Unmarshal(b, &ids)
	if len(ids) == 0 {
		t.Fatal("no history")
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("an input appended twice: %s", id)
		}
		seen[id] = true
	}
}

// The actor that runs the tree is never in its reach: a root that holds its
// own actor faults before any turn, of either kind.
func TestTheActorNeverInvokesItself(t *testing.T) {
	for _, kind := range []string{"act", "delegate"} {
		calls := 0
		count := func(to string, payload json.RawMessage) Step {
			calls++
			return run(to, payload)
		}
		got, _ := kernel(t, t.TempDir(), "f", Ask{Goal: "g", Kind: kind, Actor: "model", Reach: []string{"echo", "model"}}, count)
		if !strings.Contains(string(got), "faulted") || calls != 0 {
			t.Errorf("%s: a root holding its own actor: %s, %d instructions run", kind, got, calls)
		}
	}
}

// A model's actor that calls its instructions as tools: the calls, in
// order, are the reply; the tool called decides the instruction.
func TestCallsAreTheReply(t *testing.T) {
	asTools := func(to string, payload json.RawMessage) Step {
		if to != "model" {
			return run(to, payload)
		}
		var p struct {
			Text         string
			Instructions []struct{ Name string }
		}
		json.Unmarshal(payload, &p)
		var granted []string
		for _, in := range p.Instructions {
			granted = append(granted, in.Name)
		}
		var call []map[string]json.RawMessage
		json.Unmarshal(reply(json.RawMessage(mustText(model(p.Text)))), &call)
		var calls []map[string]any
		for _, in := range call {
			for name, v := range in {
				if !slices.Contains(granted, name) {
					t.Errorf("%s was called but not granted: %v", name, granted)
				}
				calls = append(calls, map[string]any{"name": name, "args": map[string]json.RawMessage{name: v}})
			}
		}
		b, _ := json.Marshal(map[string]any{"text": "", "calls": calls})
		return Step{To: to, State: "return", Value: b}
	}
	for _, kind := range []string{"act", "delegate"} {
		got, _ := kernel(t, t.TempDir(), "f", Ask{Goal: "say hello", Kind: kind, Actor: "model", Reach: []string{"echo"}}, asTools)
		if string(got) != `{"returned":["hello"]}` {
			t.Errorf("%s: the tree returned %s", kind, got)
		}
	}
	// A call of goal whose input says delegate is a goal, and faults as one.
	if r := string(reply(json.RawMessage(`{"calls":[{"name":"goal","args":{"delegate":{}}}]}`))); r != `[{"goal":{"delegate":{}}}]` {
		t.Errorf("the tool called does not decide: %s", r)
	}
}

func mustText(s string) []byte {
	b, _ := json.Marshal(s)
	return b
}

// What each kind is granted is what it loads, and nothing else.
func TestEachKindIsGrantedItsInstructions(t *testing.T) {
	names := func(act bool) (out []string) {
		for _, in := range instructions(act, 1) {
			out = append(out, in["name"].(string))
		}
		return
	}
	if got := names(true); !slices.Equal(got, []string{"create_input", "perform_effect", "return", "exit"}) {
		t.Errorf("act: %v", got)
	}
	if got := names(false); !slices.Equal(got, []string{"goal", "delegate", "return", "exit"}) {
		t.Errorf("delegate: %v", got)
	}
}

// An act root that holds nothing is still an act node: its kind is what
// its offer loads.
func TestAnActRootHoldingNothing(t *testing.T) {
	var granted []string
	spy := func(to string, payload json.RawMessage) Step {
		var p struct{ Instructions []struct{ Name string } }
		json.Unmarshal(payload, &p)
		for _, in := range p.Instructions {
			granted = append(granted, in.Name)
		}
		return Step{To: to, State: "return", Value: json.RawMessage(`{"text":"[{\"exit\":\"nothing to invoke\"}]"}`)}
	}
	kernel(t, t.TempDir(), "f", Ask{Goal: "g", Kind: "act", Actor: "model", Reach: []string{}}, spy)
	if !slices.Equal(granted, []string{"create_input", "perform_effect", "return", "exit"}) {
		t.Fatalf("granted %v", granted)
	}
}

// What a node is and what its model receives are one to one: every turn's
// payload is its node, and nothing else. Its instructions are what the
// node's kind loads; its names are the node's names, in order, each with
// the value the shell keeps for it; its contracts are what the node holds,
// in its offer's order; it is granted no target.
func TestATurnIsItsNode(t *testing.T) {
	dir := t.TempDir()
	turns := 0
	var s *shell
	check := func(to string, payload json.RawMessage) Step {
		if to != "model" {
			return run(to, payload)
		}
		turns++
		var p struct {
			Text         string
			Fresh        bool
			Tools        *[]string
			Instructions []struct{ Name string }
		}
		if err := json.Unmarshal(payload, &p); err != nil {
			t.Fatal(err)
		}
		// The node, as the tree holds it now: replay the frame's log.
		s = replayed(t, dir, "f")
		var node tree.Agent
		for _, a := range s.tree.Agents {
			if a.Phase == tree.PhaseThinking && strings.Contains(p.Text, "You are node "+string(a.ID)+" of") {
				node = *a
			}
		}
		if node.ID == "" {
			t.Fatalf("a turn for no thinking node:\n%s", p.Text)
		}
		act := node.Kind == tree.KindAct
		var want, got []string
		for _, in := range instructions(act, 1) {
			want = append(want, in["name"].(string))
		}
		for _, in := range p.Instructions {
			got = append(got, in.Name)
		}
		if !slices.Equal(got, want) || p.Tools == nil || len(*p.Tools) != 0 || !p.Fresh {
			t.Errorf("node %s (%v): granted %v and %v", node.ID, node.Kind, got, p.Tools)
		}
		shown := line.FindAllStringSubmatch(p.Text, -1)
		held := node.Offer.PassOnly
		if act {
			held = node.Offer.Invoke
		}
		contracts := regexp.MustCompile(`(?m)^  \d+: [^ ]+, input .*$`).FindAllString(p.Text, -1)
		if len(shown) != len(node.Names) || len(contracts) != len(held) {
			t.Errorf("node %s: %d names and %d contracts, shown %d and %d", node.ID, len(node.Names), len(held), len(shown), len(contracts))
		}
		for n := range node.Names {
			v, _ := json.Marshal(s.values[node.ID][n].V)
			if !strings.Contains(p.Text, fmt.Sprintf("  %d: %s (%s)\n", n, v, s.values[node.ID][n].From)) {
				t.Errorf("node %s: name %d (%s) not shown as it is", node.ID, n, v)
			}
		}
		for _, f := range node.Lost {
			if !strings.Contains(p.Text, fmt.Sprintf("  - node %s: %s\n", f.Child, f.Why)) {
				t.Errorf("node %s: its child %s's fault is not shown", node.ID, f.Child)
			}
		}
		if len(node.Lost) == 0 && strings.Contains(p.Text, "faulted, and gave you nothing") {
			t.Errorf("node %s: faults shown that it does not have", node.ID)
		}
		for i, c := range held {
			if !strings.Contains(p.Text, fmt.Sprintf("  %d: %s, input %s\n", i, s.targets[c].name, c.Shape())) {
				t.Errorf("node %s: contract %d not shown", node.ID, i)
			}
		}
		return run(to, payload)
	}
	got, _ := kernel(t, dir, "f", Ask{Goal: "say hello", Kind: "delegate", Actor: "model", Reach: []string{"echo"}}, check)
	if string(got) != `{"returned":["hello"]}` || turns != 4 {
		t.Fatalf("%s after %d turns", got, turns)
	}
}

// replayed is a frame's tree, rebuilt from its log as a resume does.
func replayed(t *testing.T, dir, frame string) *shell {
	t.Helper()
	input, _ := json.Marshal(map[string]any{"tree": Ask{Goal: "say hello", Kind: "delegate", Actor: "model", Reach: []string{"echo"}}})
	var in struct{ Tree Ask }
	json.Unmarshal(input, &in)
	s := &shell{dir: dir + "/tree", frame: name(frame), reach: reach}
	if err := s.load(); err != nil {
		t.Fatal(err)
	}
	if err := s.setup(in.Tree); err != nil {
		t.Fatal(err)
	}
	s.tree = tree.NewTree(s.app, s.profile)
	for _, l := range s.state.Log {
		e, err := s.decode(l)
		if err != nil {
			t.Fatal(err)
		}
		var es []tree.Effect
		s.tree, es = tree.Step(s.tree, e)
		s.names(es)
	}
	return s
}

// A goal is cut, never written: the root cuts lines of its spec, and its
// child is shown exactly those lines as what it is asked to do. A cut past
// the end of the text faults the root, by the shell, before Step.
func TestAGoalIsCutFromTheSpec(t *testing.T) {
	spec := "Build the thing.\nFirst, say hello.\nThen say goodbye."
	one := func(cut string) string {
		return `[{"goal":{"of":0,"lines":` + cut + `}},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`
	}
	run := func(root string) (json.RawMessage, []string) {
		var prompts []string
		answer := func(to string, payload json.RawMessage) Step {
			if to == "echo" {
				var p struct{ Text string }
				json.Unmarshal(payload, &p)
				v, _ := json.Marshal(map[string]string{"text": p.Text})
				return Step{To: to, State: "return", Value: v}
			}
			var p struct{ Text string }
			json.Unmarshal(payload, &p)
			prompts = append(prompts, p.Text)
			out := root
			switch {
			case strings.Contains(p.Text, "(what came back)"), strings.Contains(p.Text, " returned)"):
				var m []string
				for _, l := range line.FindAllStringSubmatch(p.Text, -1) {
					if l[2] == "what came back" || strings.HasSuffix(l[2], " returned") {
						m = l
					}
				}
				out = `[{"return":{"members":[` + m[1] + `]}}]`
			case strings.Contains(p.Text, "of a tree, an act node"):
				out = `[{"create_input":"{\"text\":\"hello\"}"},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`
			}
			v, _ := json.Marshal(map[string]string{"text": out})
			return Step{To: to, State: "return", Value: v}
		}
		got, _ := kernel(t, t.TempDir(), "frame-1", Ask{Goal: spec, Kind: "delegate", Actor: "model", Reach: []string{"echo"}}, answer)
		return got, prompts
	}
	// Two cuts, the rules and a part, passed together: two observations.
	if got, prompts := run(`[{"goal":{"of":0,"lines":[1,1]}},{"goal":{"of":0,"lines":[3,3]}},{"delegate":{"observations":[{"earlier":0},{"earlier":1}],"contracts":[0]}}]`); string(got) != `{"returned":[{"text":"hello"}]}` ||
		!strings.Contains(prompts[1], `"Build the thing."`) || !strings.Contains(prompts[1], `"Then say goodbye."`) {
		t.Fatalf("two cuts passed together: %s\n%s", got, prompts[1])
	}
	got, prompts := run(one("[2,3]"))
	if string(got) != `{"returned":[{"text":"hello"}]}` {
		t.Fatalf("the tree returned %s", got)
	}
	child := prompts[1]
	if !strings.Contains(prompts[len(prompts)-1], "a goal you cut: lines 2 to 3 of name 0") || !strings.Contains(child, "(your goal):") || !strings.Contains(child, "      1| First, say hello.\n      2| Then say goodbye.\n") || strings.Contains(child, "Build the thing.") {
		t.Fatalf("the child's prompt:\n%s", child)
	}
	// The rules and a part, one goal of two runs.
	if _, prompts := run(one("[[1,1],[3,3]]")); !strings.Contains(prompts[1], "      1| Build the thing.\n      2| Then say goodbye.\n") {
		t.Fatalf("a goal of two runs, the child's prompt:\n%s", prompts[1])
	}
	if got, _ := run(one("[[1,1],[2,9]]")); !strings.Contains(string(got), "a goal cut lines 2 to 9 of name 0, which has 3") {
		t.Fatalf("a run past the end: %s", got)
	}
	if got, _ := run(one("[2,9]")); !strings.Contains(string(got), "a goal cut lines 2 to 9 of name 0, which has 3") {
		t.Fatalf("a cut past the end: %s", got)
	}
}

// The profile is set in the tree's input, at its root: its model is told the
// turn bound and, as a delegate, the act bound its children's kinds follow
// from. With none, the default profile; a profile that cannot run a tree is
// refused.
func TestTheRootSetsTheProfile(t *testing.T) {
	for _, tc := range []struct {
		profile *tree.Profile
		turns   string
		passing string
	}{
		{nil, "You have 8 turns at most", "Pass exactly one contract"},
		{&tree.Profile{Act: 3, Turns: 5, Depth: 2, FanOut: 4, Budget: 20}, "The whole tree has 19 turns left", "Pass at most 3 contracts"},
	} {
		var text, described string
		spy := func(to string, payload json.RawMessage) Step {
			var p struct {
				Text         string
				Instructions []struct{ Name, Description string }
			}
			json.Unmarshal(payload, &p)
			text = p.Text
			for _, in := range p.Instructions {
				if in.Name == "delegate" {
					described = in.Description
				}
			}
			return Step{To: to, State: "return", Value: json.RawMessage(`{"text":"[{\"exit\":\"looked\"}]"}`)}
		}
		kernel(t, t.TempDir(), "f", Ask{Goal: "g", Kind: "delegate", Actor: "model", Reach: []string{"echo"}, Profile: tc.profile}, spy)
		if !strings.Contains(text, tc.turns) || !strings.Contains(text, tc.passing) || !strings.Contains(described, tc.passing) {
			t.Fatalf("profile %+v: the model was told %q / %q", tc.profile, text, described)
		}
	}
	s := &shell{reach: reach}
	if err := s.setup(Ask{Goal: "g", Kind: "act", Actor: "model", Reach: []string{}, Profile: &tree.Profile{Act: 0, Turns: 8, Depth: 3, FanOut: 32, Budget: 64}}); err == nil || !strings.Contains(err.Error(), "act bound") {
		t.Fatalf("a profile with an act bound of 0: %v", err)
	}
}
