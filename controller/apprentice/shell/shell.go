// Package shell is the interim shell (56, step 5): it runs a tree's Step
// inside the apprentice, one kernel resume at a time. Step is pure and holds
// nothing between resumes, so every resume replays the frame's log, every
// event Step was ever handed, in order, then hands it what the kernel
// delivered since, answering every read and commit from the apprentice's
// state, and emits every action as one round.
//
// The state, under SKYRA_STATE/tree:
//
//	history.json        the input base types seen, a JSON array of ids
//	chains/<base>.json  a return base type's guards, oldest first
//	guards/<ref>        a guard, write-once under its refinement id
//	frames/<frame>.json a frame's log, its round, and what it has consumed
//	records/<frame>     record/1, one JSON object per line
//
// Identity, for now: a target's binary is sha256 of its name, since the
// envelope names targets, never hashes. One binary under two names is two.
package shell

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"skyra/apprentice/proc"
	"skyra/apprentice/tree"
	"skyra/apprentice/types"
)

// Name is the apprentice's own name in the trees it runs.
const Name = "apprentice"

// Ask is the tree contract's input.
type Ask struct {
	Goal  string   `json:"goal"`
	Kind  string   `json:"kind"`  // "act" or "delegate": the root's kind is given
	Actor string   `json:"actor"` // the target that runs every turn
	Reach []string `json:"reach"` // the targets the root holds
	Depth int      `json:"depth"`
}

// Contract is the tree contract.
func Contract() *proc.Schema {
	str := &proc.Schema{Type: "string"}
	return &proc.Schema{Title: "tree", Type: "object", Required: []string{"tree"}, Fields: map[string]*proc.Schema{
		"tree": {Type: "object", Required: []string{"actor", "goal", "kind", "reach"}, Fields: map[string]*proc.Schema{
			"goal": str, "kind": str, "actor": str, "reach": {Type: "array", Items: str}, "depth": {Type: "number"},
		}},
	}}
}

// Step is one slot as the kernel delivers it.
type Step struct {
	To    string          `json:"to"`
	State string          `json:"state"`
	Value json.RawMessage `json:"value,omitempty"`
	Fault *Fault          `json:"fault,omitempty"`
}

// Fault is a step's fault, as the kernel derived it.
type Fault struct {
	Kind  string `json:"kind"`
	Cause string `json:"cause"`
}

// Resume is one resume of a tree frame: what to print.
func Resume(dir, frame string, input json.RawMessage, reach map[string][]*proc.Schema, steps []Step) ([]byte, error) {
	var in struct {
		Tree Ask `json:"tree"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, err
	}
	s := &shell{dir: filepath.Join(dir, "tree"), frame: name(frame), reach: reach}
	unlock, err := s.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	if err := s.load(); err != nil {
		return nil, err
	}
	if err := s.setup(in.Tree); err != nil {
		return addressed(frame, map[string]any{"refused": err.Error()}), nil
	}
	// Replay: the same events give the same tree (check 6).
	s.tree = tree.NewTree(s.app, s.depth)
	for _, l := range s.state.Log {
		e, err := s.decode(l)
		if err != nil {
			return nil, err
		}
		var es []tree.Effect
		s.tree, es = tree.Step(s.tree, e)
		s.names(es)
	}
	if s.tree.Awaiting() {
		return nil, errors.New("the log ends inside a cycle")
	}

	var acts []tree.Action
	if len(s.state.Log) == 0 {
		s.keep(in.Tree)
		started, err := s.started(in.Tree)
		if err != nil {
			return addressed(frame, map[string]any{"refused": err.Error()}), nil
		}
		if acts, err = s.drive(started, entry{}); err != nil {
			return nil, err
		}
	} else {
		fresh := steps[min(s.state.Consumed, len(steps)):]
		if len(fresh) != len(s.state.Round) {
			return nil, fmt.Errorf("the frame emitted %d instructions but the envelope delivers %d new steps", len(s.state.Round), len(fresh))
		}
		for i, st := range fresh {
			e, l := s.back(s.state.Round[i], st)
			more, err := s.drive(e, l)
			if err != nil {
				return nil, err
			}
			acts = append(acts, more...)
		}
	}
	s.state.Consumed = len(steps)

	if len(acts) == 0 {
		if !s.tree.Done() {
			return nil, errors.New("nothing to do and the tree is not done")
		}
		s.state.Round = nil
		if err := s.save(); err != nil {
			return nil, err
		}
		return addressed(frame, s.answer()), nil
	}
	var emit [][]byte
	s.state.Round = nil
	seen := map[string]int{}
	for _, a := range acts {
		to, payload, p, err := s.out(a)
		if err != nil {
			return nil, err
		}
		b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
		// The kernel keeps repeats apart by their bytes.
		k := string(b)
		b = append(b[:len(b)-1], append(bytes.Repeat([]byte(" "), seen[k]), '}')...)
		seen[k]++
		emit = append(emit, b)
		s.state.Round = append(s.state.Round, p)
	}
	if err := s.save(); err != nil { // where the frame is, before the emission
		return nil, err
	}
	if len(emit) == 1 {
		return emit[0], nil
	}
	return append(append([]byte("["), bytes.Join(emit, []byte(","))...), ']'), nil
}

// pending is one emitted instruction: whose, and what it was.
type pending struct {
	Agent   tree.AgentID `json:"agent"`
	Turn    bool         `json:"turn,omitempty"`
	Element int          `json:"element"`
}

// entry is one event of the log, as stored.
type entry struct {
	Type     string          `json:"type"`
	Agent    tree.AgentID    `json:"agent,omitempty"`
	Out      string          `json:"out,omitempty"`
	Element  int             `json:"element,omitempty"`
	Shape    string          `json:"shape,omitempty"`
	Err      string          `json:"err,omitempty"`
	Value    json.RawMessage `json:"value,omitempty"`
	Found    [][]byte        `json:"found,omitempty"`
	Outcomes []tree.Outcome  `json:"outcomes,omitempty"`
}

type state struct {
	Log      []entry   `json:"log"`
	Round    []pending `json:"round"`
	Consumed int       `json:"consumed"`
	// What it was asked and the reach it was set up with, so the view can
	// replay it alone.
	Ask   *Ask                      `json:"ask,omitempty"`
	Reach map[string][]*proc.Schema `json:"reach,omitempty"`
}

type shell struct {
	dir, frame string
	reach      map[string][]*proc.Schema
	state      state

	app       tree.Apprentice
	depth     int
	kind      tree.Kind
	self      tree.Contract
	rootReach []tree.Contract
	goal      string
	targets   map[tree.Contract]target // every contract, to its target
	names_    map[tree.Binary]string

	tree    tree.Tree
	values  map[tree.AgentID]map[int]value
	results map[string]json.RawMessage // agent/element → what came back, while replaying
	records []json.RawMessage          // this resume's, appended at commit
	faults  []string
}

type target struct {
	name   string
	schema *proc.Schema
}

// value is a name's value, and how a model is told where it came from.
type value struct {
	V    any    `json:"v"`
	From string `json:"from"`
}

func name(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func binaryOf(target string) tree.Binary {
	return tree.Binary("sha256:" + name("skyra/target/v1|"+target))
}

func (s *shell) contract(t string, c *proc.Schema) (tree.Contract, error) {
	k, ok := tree.Identify(binaryOf(t), tree.Schema(c.Canonical()))
	if !ok {
		return k, fmt.Errorf("%s: not a contract", t)
	}
	s.targets[k] = target{t, c}
	s.names_[k.Binary()] = t
	return k, nil
}

// setup is the apprentice and the root's reach, from the ask. It is the same
// on every resume.
func (s *shell) setup(a Ask) error {
	s.targets, s.names_, s.values = map[tree.Contract]target{}, map[tree.Binary]string{}, map[tree.AgentID]map[int]value{}
	s.results = map[string]json.RawMessage{}
	bin := binaryOf(Name)
	var cs []tree.Contract
	for _, t := range []string{"observe", "root", "exit", "return", "perform_effect", "create_input", "delegate", "goal", "create_delegation", "create_act", "package_reach"} {
		c, _ := tree.Identify(bin, tree.Schema(`{"title":"`+t+`","type":"object"}`))
		cs = append(cs, c)
	}
	app, ok := tree.ApprenticeOf(bin, cs[0], cs[1], cs[2], cs[3], cs[4], cs[5], cs[6], cs[7], cs[8], cs[9], cs[10])
	if !ok {
		return errors.New("the apprentice's contracts")
	}
	s.app, s.goal = app, a.Goal
	if s.depth = a.Depth; s.depth <= 0 {
		s.depth = 3
	}
	switch a.Kind {
	case "act":
		s.kind = tree.KindAct
	case "delegate":
		s.kind = tree.KindDelegate
	default:
		return fmt.Errorf("the root's kind is act or delegate, not %q", a.Kind)
	}
	actor := s.reach[a.Actor]
	if len(actor) == 0 {
		return fmt.Errorf("the actor %q is not in reach", a.Actor)
	}
	var err error
	if s.self, err = s.contract(a.Actor, actor[0]); err != nil {
		return err
	}
	for _, t := range a.Reach {
		if len(s.reach[t]) == 0 {
			return fmt.Errorf("%q is not in reach", t)
		}
		for _, c := range s.reach[t] {
			k, err := s.contract(t, c)
			if err != nil {
				return err
			}
			s.rootReach = append(s.rootReach, k)
		}
	}
	return nil
}

func (s *shell) started(a Ask) (tree.Event, error) {
	return tree.Started{Self: s.self, Reach: s.rootReach, Kind: s.kind, Actor: s.self, Endowment: tree.Schema(`{"type":"string"}`)}, nil
}

// back is the event for what came back for one emitted instruction.
func (s *shell) back(p pending, st Step) (tree.Event, entry) {
	why := st.State
	if st.Fault != nil {
		why = st.Fault.Kind + ": " + st.Fault.Cause
	}
	if p.Turn {
		if st.State != "return" {
			return tree.Stopped{Agent: p.Agent, Reason: why}, entry{}
		}
		out := reply(st.Value)
		if why := s.badCut(p.Agent, out); why != "" {
			return tree.Stopped{Agent: p.Agent, Reason: why}, entry{}
		}
		return tree.Replied{Agent: p.Agent, Out: out}, entry{}
	}
	if st.State != "return" {
		return tree.Returned{Agent: p.Agent, Element: p.Element, Shape: tree.Schema("-"), Err: why}, entry{}
	}
	v, err := types.Decode(st.Value)
	if err != nil {
		return tree.Returned{Agent: p.Agent, Element: p.Element, Shape: tree.Schema("-"), Err: err.Error()}, entry{}
	}
	return tree.Returned{Agent: p.Agent, Element: p.Element, Shape: tree.Schema(types.Shape(v))}, entry{Value: st.Value}
}

// badCut is why a goal in a reply cuts what it cannot: Step never sees a
// value, so the shell checks each cut against the text it cuts, before the
// reply reaches Step. A name the agent has no value for is left to Step.
func (s *shell) badCut(agent tree.AgentID, out []byte) string {
	var call []map[string]json.RawMessage
	if json.Unmarshal(out, &call) != nil {
		return ""
	}
	for _, ins := range call {
		var g struct {
			Of    *int  `json:"of"`
			Lines []int `json:"lines"`
		}
		if ins["goal"] == nil || json.Unmarshal(ins["goal"], &g) != nil || g.Of == nil || len(g.Lines) != 2 {
			continue
		}
		v, ok := s.values[agent][*g.Of]
		if !ok {
			continue
		}
		text, ok := v.V.(string)
		if !ok {
			return fmt.Sprintf("a goal is cut from text; name %d is not text", *g.Of)
		}
		if n := len(lines(text)); g.Lines[1] > n {
			return fmt.Sprintf("a goal cut lines %d to %d of name %d, which has %d", g.Lines[0], g.Lines[1], *g.Of, n)
		}
	}
	return ""
}

// lines is a text's lines, a trailing newline ending the last.
func lines(text string) []string {
	return strings.Split(strings.TrimRight(text, "\n"), "\n")
}

// cut is lines from to to of a text, from 1, both included.
func cut(text string, from, to int) string {
	ls := lines(text)
	if from < 1 || to > len(ls) || from > to {
		return ""
	}
	return strings.Join(ls[from-1:to], "\n")
}

// reply is the call in what a model's actor returned: the instructions it
// called, in order; or, from an actor that writes them, the instruction
// array in its text, with any fence or prose around it left out.
func reply(v json.RawMessage) []byte {
	var o struct {
		Text  *string `json:"text"`
		Calls []struct {
			Name string          `json:"name"`
			Args json.RawMessage `json:"args"`
		} `json:"calls"`
	}
	text := string(v)
	if json.Unmarshal(v, &o) == nil && len(o.Calls) > 0 {
		var call []json.RawMessage
		for _, c := range o.Calls {
			// The instruction is the one called: its input is {<name>: v},
			// and v is the instruction's. Any other input is v itself, which
			// Step interprets, and faults if it is not one.
			v := c.Args
			var args map[string]json.RawMessage
			if json.Unmarshal(c.Args, &args) == nil && len(args) == 1 && args[c.Name] != nil {
				v = args[c.Name]
			}
			b, _ := json.Marshal(map[string]json.RawMessage{c.Name: v})
			call = append(call, b)
		}
		b, _ := json.Marshal(call)
		return b
	}
	if json.Unmarshal(v, &o) == nil && o.Text != nil {
		text = *o.Text
	} else {
		var str string
		if json.Unmarshal(v, &str) == nil {
			text = str
		}
	}
	i, j := strings.Index(text, "["), strings.LastIndex(text, "]")
	if i < 0 || j < i {
		return []byte(text)
	}
	return []byte(text[i : j+1])
}

// drive hands Step one event, and every read and commit it leads to,
// answered and logged; it returns the actions it decided.
func (s *shell) drive(e tree.Event, l entry) ([]tree.Action, error) {
	var acts []tree.Action
	for {
		if r, ok := e.(tree.Returned); ok && l.Value != nil {
			s.results[fmt.Sprintf("%s/%d", r.Agent, r.Element)] = l.Value
		}
		s.state.Log = append(s.state.Log, s.encode(e, l))
		var es []tree.Effect
		s.tree, es = tree.Step(s.tree, e)
		s.names(es)
		if len(es) == 0 {
			return acts, nil
		}
		var err error
		switch ef := es[0].(type) {
		case tree.Read:
			e, err = s.read(ef)
		case tree.Commit:
			e, err = s.commit(ef)
		case tree.Execute:
			return append(acts, ef.Acts...), nil
		}
		if err != nil {
			return nil, err
		}
		l = entry{}
	}
}

func (s *shell) encode(e tree.Event, l entry) entry {
	switch e := e.(type) {
	case tree.Started:
		l.Type = "started"
	case tree.Replied:
		l.Type, l.Agent, l.Out = "replied", e.Agent, string(e.Out)
	case tree.Returned:
		l.Type, l.Agent, l.Element, l.Shape, l.Err = "returned", e.Agent, e.Element, string(e.Shape), e.Err
	case tree.Stopped:
		l.Type, l.Agent, l.Err = "stopped", e.Agent, e.Reason
	case tree.ReadResult:
		l.Type, l.Found = "read", e.Found
	case tree.Committed:
		l.Type, l.Outcomes = "committed", e.Outcomes
	}
	return l
}

func (s *shell) decode(l entry) (tree.Event, error) {
	switch l.Type {
	case "started":
		return s.started(Ask{})
	case "replied":
		return tree.Replied{Agent: l.Agent, Out: []byte(l.Out)}, nil
	case "returned":
		if l.Value != nil {
			s.results[fmt.Sprintf("%s/%d", l.Agent, l.Element)] = l.Value
		}
		return tree.Returned{Agent: l.Agent, Element: l.Element, Shape: tree.Schema(l.Shape), Err: l.Err}, nil
	case "stopped":
		return tree.Stopped{Agent: l.Agent, Reason: l.Err}, nil
	case "read":
		return tree.ReadResult{Found: l.Found}, nil
	case "committed":
		return tree.Committed{Outcomes: l.Outcomes}, nil
	}
	return nil, fmt.Errorf("a log entry of type %q", l.Type)
}

// --- the records ---

func (s *shell) path(parts ...string) string {
	return filepath.Join(append([]string{s.dir}, parts...)...)
}

func (s *shell) read(r tree.Read) (tree.Event, error) {
	var found [][]byte
	for _, k := range r.Keys {
		var b []byte
		var err error
		switch k := k.(type) {
		case tree.HistoryKey:
			b, err = os.ReadFile(s.path("history.json"))
		case tree.ChainKey:
			b, err = os.ReadFile(s.path("chains", name(k.Base.String())+".json"))
		case tree.CELKey:
			// Only the default guard is ever written (56, open item 15): it
			// holds, the same base type as here. A written guard is not run
			// yet, so it never holds.
			b, err = []byte(fmt.Sprint(k.CEL == "")), nil
		}
		if errors.Is(err, os.ErrNotExist) {
			b, err = nil, nil
		}
		if err != nil {
			return nil, err
		}
		found = append(found, b)
	}
	return tree.ReadResult{Found: found}, nil
}

func (s *shell) commit(c tree.Commit) (tree.Event, error) {
	var out []tree.Outcome
	for _, w := range c.Writes {
		o := tree.OutcomeCreated
		var err error
		switch w := w.(type) {
		case tree.RecordWrite:
			b, _ := json.Marshal(map[string]any{"kind": w.Kind, "agent": w.Agent, "body": w.Body})
			s.records = append(s.records, b)
		case tree.HistoryWrite:
			o, err = s.history(w)
		case tree.GuardWrite:
			o, err = s.guard(w)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := s.flushRecords(); err != nil {
		return nil, err
	}
	return tree.Committed{Outcomes: out}, nil
}

// guard writes a guard once under its refinement id: the same bytes again
// are Identical, other bytes a Conflict.
func (s *shell) guard(w tree.GuardWrite) (tree.Outcome, error) {
	p := s.path("guards", name(string(w.ID)))
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return 0, err
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		old, err := os.ReadFile(p)
		if err != nil {
			return 0, err
		}
		if bytes.Equal(old, w.Bytes) {
			return tree.OutcomeIdentical, nil
		}
		return tree.OutcomeConflict, nil
	}
	if err != nil {
		return 0, err
	}
	if _, err := f.Write(w.Bytes); err != nil {
		f.Close()
		return 0, err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return 0, err
	}
	if err := f.Close(); err != nil {
		return 0, err
	}
	return tree.OutcomeCreated, s.appendJSON(s.path("chains", name(w.Base.String())+".json"), json.RawMessage(w.Bytes))
}

// history appends an input base type once: one another run appended since
// this one read the history is Identical.
func (s *shell) history(w tree.HistoryWrite) (tree.Outcome, error) {
	var id string
	if err := json.Unmarshal(w.Bytes, &id); err != nil {
		return 0, err
	}
	var ids []string
	b, err := os.ReadFile(s.path("history.json"))
	if err == nil {
		if err := json.Unmarshal(b, &ids); err != nil {
			return 0, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}
	if slices.Contains(ids, id) {
		return tree.OutcomeIdentical, nil
	}
	return tree.OutcomeCreated, s.appendJSON(s.path("history.json"), id)
}

// appendJSON appends to the JSON array in a file, replacing it whole.
func (s *shell) appendJSON(p string, v any) error {
	var xs []json.RawMessage
	b, err := os.ReadFile(p)
	if err == nil {
		if err := json.Unmarshal(b, &xs); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	x, _ := json.Marshal(v)
	b, _ = json.Marshal(append(xs, x))
	return publish(p, b)
}

func publish(p string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(p), ".tmp-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), p)
}

func (s *shell) flushRecords() error {
	if len(s.records) == 0 {
		return nil
	}
	p := s.path("records", s.frame)
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for _, r := range s.records {
		w.Write(append(r, '\n'))
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	s.records = nil
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// lock holds the apprentice's tree state for one resume, exclusively: every
// read and write of a resume is one, and no two frames interleave theirs.
// A resume never waits on a model: those run between resumes.
func (s *shell) lock() (func(), error) {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(s.path("lock"), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}

func (s *shell) load() error {
	b, err := os.ReadFile(s.path("frames", s.frame+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &s.state)
}

func (s *shell) save() error {
	b, _ := json.Marshal(s.state)
	return publish(s.path("frames", s.frame+".json"), b)
}

// --- the values ---

// names keeps every name's value, from the name records Step commits: the
// values never enter the tree (55, invariant 3).
func (s *shell) names(es []tree.Effect) {
	for _, e := range es {
		c, ok := e.(tree.Commit)
		if !ok {
			continue
		}
		for _, w := range c.Writes {
			if r, ok := w.(tree.RecordWrite); ok {
				switch r.Kind {
				case "name":
					s.name(r)
				case "fault":
					s.faults = append(s.faults, fmt.Sprintf("%s: %v", r.Agent, r.Body["why"]))
				}
			}
		}
	}
}

func (s *shell) name(r tree.RecordWrite) {
	n, _ := r.Body["name"].(int)
	if s.values[r.Agent] == nil {
		s.values[r.Agent] = map[int]value{}
	}
	var v value
	switch r.Body["from"] {
	case "input":
		if t, ok := r.Body["text"].(string); ok {
			v = value{t, "an input you made"}
		} else if f, ok := r.Body["filled"].(string); ok {
			v = value{s.filled(r.Agent, f), "an input you made, filled from names"}
		} else {
			v = value{s.goal, "what you were asked to do; a child's goal is lines you cut from it"}
		}
	case "goal":
		ls := r.Body["lines"].([]int)
		text, _ := s.values[r.Agent][r.Body["of"].(int)].V.(string)
		v = value{cut(text, ls[0], ls[1]), fmt.Sprintf("a goal you cut: lines %d to %d of name %d", ls[0], ls[1], r.Body["of"].(int))}
	case "exit":
		v = value{r.Body["text"], "an exit"}
	case "part":
		if t, ok := r.Body["text"]; ok {
			v = value{t, "what you are asked to do"}
		} else {
			var ts []string
			for _, b := range r.Body["contracts"].([]string) {
				ts = append(ts, s.names_[tree.Binary(b)])
			}
			v = value{ts, "the targets you hold"}
		}
	case "compose":
		var vs []any
		for _, m := range r.Body["names"].([]int) {
			vs = append(vs, s.values[r.Agent][m].V)
		}
		v = value{vs, "what you were delegated"}
	case "copy":
		from := tree.AgentID(r.Body["agent"].(string))
		v = s.values[from][r.Body["of"].(int)]
		switch {
		case strings.HasPrefix(string(from), string(r.Agent)+"."):
			v.From = "what node " + string(from) + " returned"
		case strings.HasPrefix(v.From, "a goal you cut"):
			v.From = "your goal"
		default:
			v.From = "passed to you"
		}
	case "result":
		raw := s.results[fmt.Sprintf("%s/%d", r.Agent, r.Body["element"].(int))]
		var x any
		_ = json.Unmarshal(raw, &x)
		v = value{x, "what came back"}
	}
	s.values[r.Agent][n] = v
}

// filled is an input's value: its JSON, with every {"name": n} replaced
// by name n's value. Step checked every name is in scope.
func (s *shell) filled(agent tree.AgentID, f string) any {
	var x any
	d := json.NewDecoder(strings.NewReader(f))
	d.UseNumber()
	_ = d.Decode(&x)
	var fill func(any) any
	fill = func(x any) any {
		switch x := x.(type) {
		case map[string]any:
			if raw, ok := x["name"]; ok && len(x) == 1 {
				if num, ok := raw.(json.Number); ok {
					if n, err := num.Int64(); err == nil {
						return s.values[agent][int(n)].V
					}
				}
			}
			out := make(map[string]any, len(x))
			for k, v := range x {
				out[k] = fill(v)
			}
			return out
		case []any:
			out := make([]any, len(x))
			for i, v := range x {
				out[i] = fill(v)
			}
			return out
		}
		return x
	}
	return fill(x)
}

// --- out ---

// out is one action as a kernel instruction.
func (s *shell) out(a tree.Action) (string, any, pending, error) {
	switch a := a.(type) {
	case tree.Turn:
		actor := s.names_[a.Binary]
		// A turn is granted nothing: its instructions are its reply, and
		// every effect is a perform_effect, typed by the tree.
		return actor, map[string]any{"text": s.prompt(a), "fresh": true, "tools": []string{}, "instructions": instructions(s.act(a))},
			pending{Agent: a.Agent, Turn: true}, nil
	case tree.InvokeContract:
		t, ok := s.targets[a.Contract]
		if !ok {
			return "", nil, pending{}, errors.New("a dispatch of a contract no target holds")
		}
		v := s.values[a.Agent][a.Input].V
		// An input is text; a model writes an object as JSON.
		if str, ok := v.(string); ok && t.schema.Type != "string" {
			var x any
			if json.Unmarshal([]byte(str), &x) == nil {
				v = x
			}
		}
		return t.name, v, pending{Agent: a.Agent, Element: a.Element}, nil
	}
	return "", nil, pending{}, fmt.Errorf("an action %T", a)
}

// instructions are what a node's kind loads (55, loaded): the tools its
// model's thread is granted. Each one's input is the instruction itself.
func instructions(act bool) []map[string]any {
	ref := map[string]any{"anyOf": []any{
		map[string]any{"type": "integer", "minimum": 0, "description": "a name, by its number"},
		map[string]any{"type": "object", "properties": map[string]any{"earlier": map[string]any{"type": "integer", "minimum": 0}}, "required": []string{"earlier"},
			"description": "what an earlier instruction of this turn made, by its position"},
	}}
	refs := map[string]any{"type": "array", "items": ref}
	text := map[string]any{"type": "string"}
	input := map[string]any{"anyOf": []any{
		map[string]any{"type": "string", "description": "the input as text; for a contract that takes an object, its JSON"},
		map[string]any{"type": "object", "description": "the input as JSON, in which {\"name\": n} stands for name n's value"},
	}}
	one := func(name, description string, v any) map[string]any {
		return map[string]any{"name": name, "description": description,
			"schema": map[string]any{"type": "object", "properties": map[string]any{name: v}, "required": []string{name}, "additionalProperties": false}}
	}
	obj := func(props map[string]any, required ...string) map[string]any {
		return map[string]any{"type": "object", "properties": props, "required": required, "additionalProperties": false}
	}
	var out []map[string]any
	if act {
		out = append(out,
			one("create_input", "Make an input: text, or for a contract that takes an object, its JSON. Where the input holds something you already have, write {\"name\": n} and name n's value is filled in: you never copy what you were given.", input),
			one("perform_effect", "Invoke a contract you hold, by its number, with an input you made. What comes back is a new name when you are asked again.",
				obj(map[string]any{"contract": map[string]any{"type": "integer", "minimum": 0}, "input": ref}, "contract", "input")))
	} else {
		out = append(out,
			one("goal", "Cut a goal for a child: lines from to to, counted from 1, of a name that is text. You never write a goal; you cut it from what you were asked to do.",
				obj(map[string]any{"of": map[string]any{"type": "integer", "minimum": 0, "description": "a name that is text, by its number"},
					"lines": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "minItems": 2, "maxItems": 2, "description": "[from, to]"}}, "of", "lines")),
			one("delegate", "Start a child with observations and at least one contract you hold, by number. A goal must be among the observations: one you cut with goal, as {\"earlier\": k}, or a goal you were given. Pass exactly one contract and the child is an act node, which invokes it; pass two or more and the child is a delegate node, which can only delegate again. What it returns is a new name when you are asked again.",
				obj(map[string]any{"observations": refs, "contracts": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 0}}}, "observations", "contracts")))
	}
	return append(out,
		one("return", "Return, alone: what came back to you, never what you were passed or what you wrote yourself. Your turn ends.", obj(map[string]any{"members": refs}, "members")),
		one("exit", "Give up, alone, with a reason. Your turn ends.", text))
}

// act is whether a turn is an act node's: what its offer loads, never what
// it holds.
func (s *shell) act(t tree.Turn) bool {
	return slices.Contains(t.Offer.Apprentice, s.app.CreateInput())
}

// prompt is what a turn shows its model: its names, what it holds, and the
// instructions it may write.
func (s *shell) prompt(t tree.Turn) string {
	var b strings.Builder
	act := s.act(t)
	kind := "a delegate"
	if act {
		kind = "an act"
	}
	fmt.Fprintf(&b, "You are node %s of a tree, %s node. Act only by calling your instructions; each call is one instruction of this turn, and they run, in order, when your turn ends. If you cannot call them, reply with one JSON array of instructions and nothing else.\n\n", t.Agent, kind)
	b.WriteString("Your names, which you refer to by number:\n")
	for _, n := range t.Names {
		v := s.values[t.Agent][n]
		// Text of many lines is shown by line, numbered, so a model can refer
		// to its lines.
		if str, ok := v.V.(string); ok && len(lines(str)) > 1 {
			fmt.Fprintf(&b, "  %d: text of lines (%s):\n", n, v.From)
			for i, l := range lines(str) {
				fmt.Fprintf(&b, "      %d| %s\n", i+1, l)
			}
			continue
		}
		j, _ := json.Marshal(v.V)
		fmt.Fprintf(&b, "  %d: %s (%s)\n", n, j, v.From)
	}
	held := t.Offer.PassOnly
	verb := "you may pass to a child"
	if act {
		held, verb = t.Offer.Invoke, "you may invoke"
	}
	if len(t.Faults) > 0 {
		b.WriteString("\nChildren of yours that faulted, and gave you nothing:\n")
		for _, f := range t.Faults {
			fmt.Fprintf(&b, "  - node %s: %s\n", f.Child, f.Why)
		}
	}
	fmt.Fprintf(&b, "\nThe contracts %s, by number:\n", verb)
	for i, c := range held {
		tg := s.targets[c]
		fmt.Fprintf(&b, "  %d: %s, input %s\n", i, tg.name, c.Shape())
	}
	b.WriteString("\nA reference is a name's number, or {\"earlier\": k} for what instruction k of this reply made.\n")
	b.WriteString("Instructions:\n")
	if act {
		b.WriteString(`  {"create_input": "<text>"}  make an input; for an object input, write the JSON as the text` + "\n")
		b.WriteString(`  {"create_input": {...}}  make an object input, in which {"name": n} is name n's value, filled in for you` + "\n")
		b.WriteString(`  {"perform_effect": {"contract": <number>, "input": <reference>}}  invoke a contract with an input you made` + "\n")
	} else {
		b.WriteString(`  {"goal": {"of": <name>, "lines": [<from>, <to>]}}  cut a goal for a child: those lines of a text name; text of one line is line 1` + "\n")
		b.WriteString(`  {"delegate": {"observations": [<reference>...], "contracts": [<number>...]}}  start a child with a goal and at least one contract` + "\n")
	}
	b.WriteString(`  {"return": {"members": [<reference>...]}}  return, alone: never what you were passed or what you wrote yourself` + "\n")
	b.WriteString(`  {"exit": "<reason>"}  give up, alone` + "\n")
	if act {
		b.WriteString("\nWhat you invoke comes back as a new name, then you are asked again. Return what came back.\n")
	} else {
		b.WriteString("\nWhat a child returns comes back as a new name, then you are asked again. Return what came back.\n")
		b.WriteString("Only an act node invokes a contract: to have one invoked, delegate a goal with exactly that one contract.\n")
	}
	fmt.Fprintf(&b, "You have %d turns at most; this is turn %d.\n", tree.MaxTurns, s.tree.Agents[t.Agent].Turns)
	return b.String()
}

// answer is the tree's return: the values the root returned.
func (s *shell) answer() any {
	root, _ := s.tree.Root()
	if root.Phase != tree.PhaseDone {
		return map[string]any{"faulted": s.faults}
	}
	var vs []any
	for _, n := range root.Gave {
		vs = append(vs, s.values[root.ID][n].V)
	}
	return map[string]any{"returned": vs}
}

func addressed(frame string, payload any) []byte {
	b, _ := json.Marshal(map[string]any{"to": frame, "payload": payload})
	return b
}
