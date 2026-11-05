package shell

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"skyra/apprentice/proc"
	"skyra/apprentice/tree"
)

// A Run is one tree as it formed: its log replayed through Step, and the
// tree after each event that is not inside a cycle. The view is Step's own
// interpretation of the log, not a second one.
type Run struct {
	Frame  string   `json:"frame"`
	Goal   string   `json:"goal"`
	Kind   string   `json:"kind"`
	Reach  []string `json:"reach"`
	Frames []Frame  `json:"frames"`
}

// A Frame is the tree after one event.
type Frame struct {
	Event      string     `json:"event"` // started, replied, returned or stopped
	Agent      string     `json:"agent,omitempty"`
	Detail     string     `json:"detail,omitempty"`
	Nodes      []Node     `json:"nodes"`
	Activities []Activity `json:"activities,omitempty"`
}

// Activity describes committed work, rather than interpreting the model's
// requested instructions. Names are local to each endpoint; previews are bounded.
type Activity struct {
	Kind    string   `json:"kind"`
	From    string   `json:"from"`
	To      string   `json:"to"`
	Label   string   `json:"label"`
	Preview string   `json:"preview,omitempty"`
	Count   int      `json:"count"`
	Sources []string `json:"sources,omitempty"`
}

// A Node is an agent as the tree holds it.
type Node struct {
	ID       string   `json:"id"`
	Parent   string   `json:"parent,omitempty"`
	Kind     string   `json:"kind"`
	Phase    string   `json:"phase"`
	Depth    int      `json:"depth"`
	Turns    int      `json:"turns"`
	Holds    []string `json:"holds"` // the targets it may invoke or pass
	Names    int      `json:"names"`
	Returned int      `json:"returned"`
	Why      string   `json:"why,omitempty"`
	Lost     []string `json:"lost,omitempty"`
}

// View replays every tree under the apprentice's state dir.
func View(dir string) ([]Run, error) {
	fs, _ := filepath.Glob(filepath.Join(dir, "tree", "frames", "*.json"))
	if len(fs) == 0 {
		return nil, errors.New("no tree under " + dir)
	}
	var runs []Run
	for _, f := range fs {
		s := &shell{dir: filepath.Join(dir, "tree"), frame: strings.TrimSuffix(filepath.Base(f), ".json")}
		if err := s.load(); err != nil {
			return nil, err
		}
		if s.state.Ask == nil {
			return nil, fmt.Errorf("%s: the frame does not keep what it was asked (it predates the view)", f)
		}
		s.reach = s.state.Reach
		s.pins = s.state.Binaries
		if err := s.setup(*s.state.Ask); err != nil {
			return nil, err
		}
		r := Run{Frame: s.frame, Goal: s.goal, Kind: s.state.Ask.Kind, Reach: s.state.Ask.Reach}
		s.tree = tree.NewTree(s.app, s.depth)
		var last Frame
		var before tree.Tree
		for _, l := range s.state.Log {
			e, err := s.decode(l)
			if err != nil {
				return nil, err
			}
			if l.Type != "read" && l.Type != "committed" {
				last = Frame{Event: l.Type, Agent: string(l.Agent), Detail: detail(l)}
				before = s.tree
			}
			var es []tree.Effect
			s.tree, es = tree.Step(s.tree, e)
			s.names(es)
			last.Activities = append(last.Activities, s.activities(es)...)
			if !s.tree.Awaiting() {
				last.Nodes = s.nodes()
				last.Activities = append(last.Activities, s.transitions(before)...)
				last.Activities = summarizeActivities(last.Activities)
				r.Frames = append(r.Frames, last)
			}
		}
		runs = append(runs, r)
	}
	return runs, nil
}

// Join only when the parent really resumes after every child in this round
// has returned successfully. An earlier sibling return or a fault is not a join.
func (s *shell) transitions(before tree.Tree) []Activity {
	var out []Activity
	for _, id := range slices.Sorted(maps.Keys(before.Agents)) {
		old := before.Agents[id]
		now := s.tree.Agents[id]
		if now == nil || old.Phase != tree.PhaseDelegating || now.Turns <= old.Turns {
			continue
		}
		var sources []string
		success := true
		for _, pending := range old.Waiting {
			child := s.tree.Agents[pending.Child]
			if child == nil || child.Phase != tree.PhaseDone {
				success = false
				break
			}
			// An exit gives up rather than supplying a successful result.
			for _, o := range child.Returned {
				if o.Type().Contract() == s.app.Exit() {
					success = false
				}
			}
			sources = append(sources, string(pending.Child))
		}
		if success && len(sources) > 1 {
			out = append(out, Activity{Kind: "join", From: string(id), To: string(id), Label: "results joined", Count: len(sources), Sources: sources, Preview: fmt.Sprintf("Children %s returned; node %s resumed.", strings.Join(sources, ", "), id)})
		}
	}
	root, ok := s.tree.Root()
	old, was := before.Root()
	if ok && was && root.Phase == tree.PhaseDone && old.Phase != tree.PhaseDone {
		success := true
		for _, o := range root.Returned {
			if o.Type().Contract() == s.app.Exit() {
				success = false
			}
		}
		if success {
			out = append(out, Activity{Kind: "complete", From: string(root.ID), To: string(root.ID), Label: "complete", Count: 1, Preview: fmt.Sprintf("Root returned %d observations.", len(root.Returned))})
		}
	}
	return out
}

// activities uses Step's accepted records and effects. Refused instructions
// produce only a fault, never an invented transfer or tool invocation.
func (s *shell) activities(es []tree.Effect) []Activity {
	var out []Activity
	for _, effect := range es {
		switch e := effect.(type) {
		case tree.Commit:
			for _, w := range e.Writes {
				r, ok := w.(tree.RecordWrite)
				if !ok {
					continue
				}
				id := string(r.Agent)
				if r.Kind == "fault" {
					out = append(out, Activity{Kind: "fault", From: id, To: id, Label: "refused", Preview: preview(fmt.Sprint(r.Body["why"])), Count: 1})
					continue
				}
				if r.Kind != "name" {
					continue
				}
				n, _ := r.Body["name"].(int)
				switch r.Body["from"] {
				case "goal":
					out = append(out, Activity{Kind: "cut", From: id, To: id, Label: "spec cut", Preview: fmt.Sprintf("Spec cut from name %v, lines %v\n%s", r.Body["of"], r.Body["lines"], preview(s.values[r.Agent][n].V)), Count: 1})
				case "copy":
					source := r.Body["agent"].(string)
					of := r.Body["of"].(int)
					kind, label := "pass", "observation"
					if strings.HasPrefix(source, id+".") {
						kind, label = "return", "result"
					} else if a := s.tree.Agents[tree.AgentID(source)]; a != nil && of < len(a.Names) && a.Names[of].Obs.Type().Contract() == s.app.Goal() {
						label = "spec"
					}
					out = append(out, Activity{Kind: kind, From: source, To: id, Label: label, Preview: fmt.Sprintf("%s name %d → %s name %d\n%s", source, of, id, n, preview(s.values[r.Agent][n].V)), Count: 1})
				case "result":
					out = append(out, Activity{Kind: "received", From: id, To: id, Label: "tool result", Preview: preview(s.values[r.Agent][n].V), Count: 1})
				}
			}
		case tree.Execute:
			for _, act := range e.Acts {
				if a, ok := act.(tree.InvokeContract); ok {
					out = append(out, Activity{Kind: "tool", From: string(a.Agent), To: string(a.Agent), Label: s.targets[a.Contract].name, Preview: preview(s.values[a.Agent][a.Input].V), Count: 1})
				}
			}
		}
	}
	return out
}

// Coalesce observations on the same edge without changing their direction
// or meaning. Keep the reference list in the expandable payload preview.
func summarizeActivities(in []Activity) []Activity {
	var out []Activity
	for _, a := range in {
		i := slices.IndexFunc(out, func(b Activity) bool {
			return a.Kind == b.Kind && a.From == b.From && a.To == b.To && a.Label == b.Label
		})
		if i < 0 {
			out = append(out, a)
			continue
		}
		out[i].Count += a.Count
		out[i].Preview = preview(out[i].Preview + "\n\n" + a.Preview)
	}
	return out
}

func preview(v any) string {
	text, ok := v.(string)
	if !ok {
		b, _ := json.MarshalIndent(v, "", "  ")
		text = string(b)
	}
	chars := []rune(text)
	if len(chars) > 800 {
		return string(chars[:800]) + "…"
	}
	return text
}

func detail(l entry) string {
	switch l.Type {
	case "replied":
		var calls []map[string]json.RawMessage
		if json.Unmarshal([]byte(l.Out), &calls) != nil {
			return l.Out
		}
		var ps []string
		for _, c := range calls {
			for k, v := range c {
				ps = append(ps, k+" "+string(v))
			}
		}
		return strings.Join(ps, "; ")
	case "returned":
		if l.Err != "" {
			return fmt.Sprintf("element %d: %s", l.Element, l.Err)
		}
		return fmt.Sprintf("element %d came back", l.Element)
	case "stopped":
		return l.Err
	}
	return ""
}

var phases = map[tree.Phase]string{
	tree.PhaseThinking: "thinking", tree.PhaseInvoking: "invoking", tree.PhaseDelegating: "delegating",
	tree.PhaseDone: "done", tree.PhaseFaulted: "faulted",
}

func (s *shell) nodes() []Node {
	var ns []Node
	for _, a := range s.tree.Agents {
		n := Node{ID: string(a.ID), Parent: string(a.Parent), Kind: a.Kind.String(), Phase: phases[a.Phase],
			Depth: a.Depth, Turns: a.Turns, Names: len(a.Names), Returned: len(a.Returned), Why: a.Why}
		for _, c := range append(slices.Clone(a.Offer.Invoke), a.Offer.PassOnly...) {
			if t, ok := s.targets[c]; ok && !slices.Contains(n.Holds, t.name) {
				n.Holds = append(n.Holds, t.name)
			}
		}
		for _, f := range a.Lost {
			n.Lost = append(n.Lost, string(f.Child)+": "+f.Why)
		}
		ns = append(ns, n)
	}
	slices.SortFunc(ns, func(a, b Node) int { return order(a.ID, b.ID) })
	return ns
}

// order sorts agent ids as paths: 0, 0.0, 0.0.1, 0.1, 0.10.
func order(a, b string) int {
	x, y := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(x) && i < len(y); i++ {
		if len(x[i]) != len(y[i]) {
			return len(x[i]) - len(y[i])
		}
		if c := strings.Compare(x[i], y[i]); c != 0 {
			return c
		}
	}
	return len(x) - len(y)
}

// keep is what a frame keeps of what it was asked, for the view: the ask,
// and the reach it was set up with.
func (s *shell) keep(a Ask) {
	s.state.Ask = &a
	s.state.Binaries = s.pins
	s.state.Reach = map[string][]*proc.Schema{a.Actor: s.reach[a.Actor]}
	for _, t := range a.Reach {
		s.state.Reach[t] = s.reach[t]
	}
}
