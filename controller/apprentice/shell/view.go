package shell

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Event  string `json:"event"` // started, replied, returned or stopped
	Agent  string `json:"agent,omitempty"`
	Detail string `json:"detail,omitempty"`
	Nodes  []Node `json:"nodes"`
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
		if err := s.setup(*s.state.Ask); err != nil {
			return nil, err
		}
		r := Run{Frame: s.frame, Goal: s.goal, Kind: s.state.Ask.Kind, Reach: s.state.Ask.Reach}
		s.tree = tree.NewTree(s.app, s.depth)
		var last Frame
		for _, l := range s.state.Log {
			e, err := s.decode(l)
			if err != nil {
				return nil, err
			}
			if l.Type != "read" && l.Type != "committed" {
				last = Frame{Event: l.Type, Agent: string(l.Agent), Detail: detail(l)}
			}
			var es []tree.Effect
			s.tree, es = tree.Step(s.tree, e)
			s.names(es)
			if !s.tree.Awaiting() {
				last.Nodes = s.nodes()
				r.Frames = append(r.Frames, last)
			}
		}
		runs = append(runs, r)
	}
	return runs, nil
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
	s.state.Reach = map[string][]*proc.Schema{a.Actor: s.reach[a.Actor]}
	for _, t := range a.Reach {
		s.state.Reach[t] = s.reach[t]
	}
}
