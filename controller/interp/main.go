// Command interp is the interpreter (next/32): an actor that runs procedures
// from its own state, as an interpreter process sits above a CPU. The kernel
// never sees a procedure. It sees an invoke of one of the interpreter's
// contracts, the instructions the interpreter emits, and a return; every
// instruction is checked at intake under the Boundary it runs in, like any
// other target's. It shares no code with skyra.
//
// Each run is one resume. The kernel execs it against the contract that was
// invoked, with that contract, the input and every step in the envelope:
//
//	read     the procedure for that contract, and where this frame is, from SKYRA_STATE
//	advance  over the steps the envelope delivers
//	emit     every call now ready as one emission, or the return to its slot;
//	         where the frame is is written before the emission
//
// Its registered contract is remember, whose body is embedded in this binary
// (H6). "interp install DIR SOURCE" compiles a procedure source into a
// state directory, for a procedure installed by hand.
package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"

	"skyra/interp/proc"
	"skyra/interp/refine"
	"skyra/interp/remember"
	"skyra/interp/types"
)

//go:embed contract.json
var declared []byte

// envelope is what the kernel hands a continuation on stdin.
// entry is one actor in reach: every contract it declares, the registered
// one first, and the indexes of those held.
type entry struct {
	Contracts []*proc.Schema `json:"contracts"`
	Held      []int          `json:"held"`
}

type envelope struct {
	Frame    string           `json:"frame"`
	Input    json.RawMessage  `json:"input"`
	Contract *proc.Schema     `json:"contract"`
	Reach    map[string]entry `json:"reach"`
	Steps    []proc.Step      `json:"steps"`
}

func main() {
	var err error
	if len(os.Args) == 4 && os.Args[1] == "install-bundle" {
		err = installBundle(os.Args[2], os.Args[3])
	} else if len(os.Args) == 4 && os.Args[1] == "install" {
		err = install(os.Args[2], os.Args[3])
	} else {
		var out []byte
		if out, err = run(); err == nil {
			fmt.Println(string(out))
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "interp:", err)
		os.Exit(1)
	}
}

// remember is the registered contract, the one the kernel names no other.
func rememberContract() *proc.Schema {
	var d struct {
		Input *proc.Schema `json:"input"`
	}
	_ = json.Unmarshal(declared, &d)
	return d.Input
}

func run() ([]byte, error) {
	var e envelope
	if err := json.NewDecoder(os.Stdin).Decode(&e); err != nil {
		return nil, err
	}
	store := proc.Store{Dir: os.Getenv("SKYRA_STATE")}
	if store.Dir == "" {
		return nil, errors.New("no state of its own (SKYRA_STATE is empty)")
	}
	c := e.Contract
	if c == nil || proc.Same(c, rememberContract()) {
		return runRemember(store, e)
	}
	if proc.Same(c, delegateContract()) {
		return runDelegate(e)
	}
	return runProcedure(store, e, c)
}

// runRemember is remember's body (decision 8), embedded in this binary: a
// syscall for the trace, the compile, the refinements, then one
// export per contract a procedure was stored for. The interpreter
// extends itself; an actor can only extend itself.
//
//	no steps        → system trace: the observation, and the contract that requested it
//	the trace      → compile it; refuse a contract that already has a procedure, before
//	                  anything is stored; store every procedure, under its base types
//	then, for each  → ask the actor it was remembered from to write its refinements, when
//	                  it holds a refine contract; check the answer; a refusal is shown, and it
//	                  may try again (MaxAttempts); a valid answer replaces the base types
//	then            → export each contract, and return what was remembered
func runRemember(store proc.Store, e envelope) ([]byte, error) {
	var ask struct {
		Remember struct {
			Observation string `json:"observation"`
			Contract    struct {
				Target string       `json:"target"`
				Schema *proc.Schema `json:"schema"`
			} `json:"contract"`
		} `json:"remember"`
	}
	if err := json.Unmarshal(e.Input, &ask); err != nil {
		return nil, err
	}
	r := ask.Remember
	refused := func(why string) ([]byte, error) {
		return addressed(e.Frame, map[string]string{"refused": why}), nil
	}
	if len(e.Steps) == 0 {
		return instruction("system", map[string]any{"trace": map[string]any{"observation": r.Observation,
			"contract": map[string]any{"target": r.Contract.Target, "schema": r.Contract.Schema}}}), nil
	}
	last := e.Steps[len(e.Steps)-1]
	cause := last.State
	if last.Fault != nil {
		cause = last.Fault.Cause
	}
	rec, compiled, err := store.Remembering(e.Frame)
	if err != nil {
		return nil, err
	}
	switch {
	case !compiled:
		if len(e.Steps) != 1 {
			return nil, errors.New("a remember frame went on without a compile")
		}
		if last.State != "return" {
			return refused(cause)
		}
		var st remember.Trace
		if err := json.Unmarshal(last.Value, &st); err != nil {
			return nil, fmt.Errorf("trace: %w", err)
		}
		reach := reachOf(e)
		self := selfName(reach)
		if self == "" {
			return refused("the interpreter is not in its own reach")
		}
		compiled, err := remember.Compile(st, r.Observation, r.Contract.Target, r.Contract.Schema, reach, self)
		var why remember.Refused
		if errors.As(err, &why) {
			return refused(string(why))
		}
		if err != nil {
			return nil, err
		}
		// One representation per contract (decision 28), refused before
		// anything is stored, so no Boundary is given a procedure another
		// Boundary remembered.
		for i, c := range compiled {
			if store.Has(c.Source.Contract) {
				return refused(fmt.Sprintf("%s (from %s) already has a procedure: one representation per contract", titleOf(c.Source.Contract), c.Source.Original))
			}
			for _, other := range compiled[:i] {
				if proc.Same(other.Source.Contract, c.Source.Contract) {
					return refused(fmt.Sprintf("%s is remembered twice in one trace", titleOf(c.Source.Contract)))
				}
			}
		}
		for _, c := range compiled {
			p, nodes, err := proc.Compile(c.Source)
			if err != nil {
				return nil, err
			}
			if err := store.Put(p, nodes); err != nil {
				return nil, err
			}
			ev, _ := json.Marshal(c.Evidence)
			rec.Contracts = append(rec.Contracts, c.Source.Contract)
			rec.Originals = append(rec.Originals, c.Source.Original)
			rec.Evidence = append(rec.Evidence, ev)
		}
	case rec.Asked == "refine":
		accepted := false
		if last.State == "return" {
			ref, err := refinements(store, rec, last.Value)
			if err == nil {
				if err := store.SetRefinement(rec.Contracts[rec.Refining], ref); err != nil {
					return nil, err
				}
				accepted = true
			} else {
				rec.Attempts++
				rec.Refused = err.Error()
			}
		}
		// Accepted, given up on, or unanswerable: the next one. A procedure
		// without refinements keeps its base types (decision 10).
		if accepted || last.State != "return" || rec.Attempts >= MaxAttempts {
			rec.Refining, rec.Attempts, rec.Refused = rec.Refining+1, 0, ""
		}
	case rec.Asked == "extend":
		if last.State != "return" {
			return refused("export: " + cause)
		}
		rec.Extended++
	}

	for rec.Refining < len(rec.Contracts) && !refinable(reachOf(e)[rec.Originals[rec.Refining]]) {
		rec.Refining++
	}
	var next []byte
	switch {
	case rec.Refining < len(rec.Contracts):
		rec.Asked = "refine"
		next = instruction(rec.Originals[rec.Refining], map[string]any{"refine": refineAsk(rec)})
	case rec.Extended < len(rec.Contracts):
		rec.Asked = "extend"
		next = instruction("system", map[string]any{"export": map[string]any{"schema": rec.Contracts[rec.Extended]}})
	default:
		var names []string
		for i, c := range rec.Contracts {
			name := c.Title
			if name == "" {
				name = rec.Originals[i] // an untitled contract is named by the actor it was remembered from
			}
			names = append(names, name)
		}
		return addressed(e.Frame, map[string]any{"remembered": names}), nil
	}
	if err := store.SetRemembering(e.Frame, rec); err != nil {
		return nil, err
	}
	return next, nil
}

// MaxAttempts is how often the model may write refinements for one procedure
// before it keeps its base types.
const MaxAttempts = 3

// refinable reports whether an actor holds a refine contract: one titled
// refine, taking {"refine": ...}.
func refinable(contracts []*proc.Schema) bool {
	for _, c := range contracts {
		if c.Title == "refine" && c.Fields["refine"] != nil {
			return true
		}
	}
	return false
}

// refineAsk is what the model that was there is shown (32, "The walk, and
// the refinement"): its own command, the rounds it ran, the result it
// returned, the base types of each and the constants, as names and
// presentations, never hashes; and why its last answer was refused.
func refineAsk(rec proc.Remembering) map[string]any {
	var ev remember.Evidence
	_ = json.Unmarshal(rec.Evidence[rec.Refining], &ev)
	contract := rec.Contracts[rec.Refining].Title
	if contract == "" {
		contract = "your contract, " + rec.Originals[rec.Refining]
	}
	ask := map[string]any{
		"contract":    contract,
		"command":     ev.Input,
		"input_type":  ev.InputType,
		"rounds":      ev.Rounds,
		"result":      ev.Result,
		"constants":   ev.Constants,
		"environment": refine.Environment,
		"write": "Write the refinements: the inputs and returns for which these decisions still hold. Anything outside them comes back to you. " +
			"Answer with {\"header\": <CEL over input>, \"edges\": {<node>: <CEL over the returns so far>}}. The constants are fixed on every run: " +
			"refine the edges where they should not apply. An edge you leave out keeps its default: the same base type as here.",
	}
	if rec.Refused != "" {
		ask["refused"] = rec.Refused
	}
	return ask
}

// refinements checks the model's answer for the procedure being refined, and
// returns the refinement it makes: each predicate type-checks, has a bounded
// cost, accepts its own trace, and hits no error on it.
func refinements(store proc.Store, rec proc.Remembering, answer json.RawMessage) (*proc.Refinement, error) {
	var a struct {
		Header string            `json:"header"`
		Edges  map[string]string `json:"edges"`
	}
	if err := json.Unmarshal(answer, &a); err != nil {
		var text string // an answer written as text that holds the JSON
		if json.Unmarshal(answer, &text) != nil || json.Unmarshal([]byte(text), &a) != nil {
			return nil, errors.New(`the answer is not {"header": ..., "edges": {...}}`)
		}
	}
	var ev remember.Evidence
	if err := json.Unmarshal(rec.Evidence[rec.Refining], &ev); err != nil {
		return nil, err
	}
	p, _, err := store.Procedure(rec.Contracts[rec.Refining])
	if err != nil {
		return nil, err
	}
	ids := map[string]string{}
	for id, name := range p.Names {
		ids[name] = id
	}
	ref := proc.Refinement{}
	if p.Header.Refinement != nil {
		ref = *p.Header.Refinement
	}
	ref.Header, ref.On = "", nil
	if a.Header != "" {
		if err := accepts("the header", a.Header, refine.Vars{Input: ev.Input, InputType: ev.InputType}); err != nil {
			return nil, err
		}
		ref.Header = a.Header
	}
	returns, shown := map[string]any{}, map[string]string{}
	for _, round := range ev.Rounds {
		for _, c := range round {
			returns[c.Node], shown[c.Node] = c.Value, c.Type
		}
		for _, c := range round {
			src, ok := a.Edges[c.Node]
			if !ok {
				continue
			}
			vars := refine.Vars{Input: ev.Input, InputType: ev.InputType, Returns: maps.Clone(returns), Types: maps.Clone(shown)}
			if err := accepts("the edge "+c.Node, src, vars); err != nil {
				return nil, err
			}
			if ref.On == nil {
				ref.On = map[string]string{}
			}
			ref.On[ids[c.Node]] = src
		}
	}
	for name := range a.Edges {
		if _, ok := ids[name]; !ok {
			return nil, fmt.Errorf("there is no node %q", name)
		}
	}
	return &ref, nil
}

// accepts checks one refinement against its own trace.
func accepts(what, src string, v refine.Vars) error {
	prg, err := refine.Compile(src)
	if err != nil {
		return fmt.Errorf("%s %s", what, err)
	}
	ok, err := prg.Eval(v)
	switch {
	case err != nil:
		return fmt.Errorf("%s hits an error on its own trace: %v", what, err)
	case !ok:
		return fmt.Errorf("%s refuses the trace it was compiled from", what)
	}
	return nil
}

// reachOf is every contract each name in the envelope holds.
func reachOf(e envelope) remember.Reach {
	reach := remember.Reach{}
	for name, entry := range e.Reach {
		reach[name] = entry.Contracts
	}
	return reach
}

// selfName is the interpreter's own name in reach: the one holding remember.
func selfName(reach remember.Reach) string {
	var names []string
	for name, cs := range reach {
		for _, c := range cs {
			if proc.Same(c, rememberContract()) {
				names = append(names, name)
			}
		}
	}
	slices.Sort(names)
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

func titleOf(c *proc.Schema) string {
	if c.Title != "" {
		return c.Title
	}
	return "its contract"
}

func instruction(to string, payload any) []byte {
	b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	return b
}

// runProcedure is one resume of a remembered procedure. One representation
// per contract (decision 28): inside its refinement the compiled body runs;
// outside it, the original actor does, from the edge that refused.
func runProcedure(store proc.Store, e envelope, c *proc.Schema) ([]byte, error) {
	p, nodes, err := store.Procedure(c)
	if err != nil {
		return nil, err
	}
	pos, err := store.Position(e.Frame)
	if err != nil {
		return nil, err
	}
	switch key := proc.Key(c); {
	case pos.Procedure == "":
		pos.Procedure = key
	case pos.Procedure != key:
		return nil, errors.New("this frame runs another procedure")
	}
	if pos.Fallback {
		// The original actor answered, and its answer is the procedure's. A
		// fault would have discharged this frame without resuming it.
		last := e.Steps[len(e.Steps)-1]
		if last.State != "return" {
			return nil, errors.New("the fallback did not return")
		}
		return addressed(e.Frame, last.Value), nil
	}
	in, err := types.Decode(e.Input)
	if err != nil {
		return nil, err
	}
	ref := p.Header.Refinement
	if len(e.Steps) == 0 && ref != nil {
		if why := refusedInput(ref, c, in); why != "" {
			return fallback(store, e, pos, p, nodes, edge{at: "input", why: why})
		}
	}
	next, err := proc.Plan(p, nodes, e.Input, pos, e.Steps)
	if err != nil {
		return nil, err
	}
	if st := next.Stuck; st != nil {
		cause := st.Step.State
		if st.Step.Fault != nil {
			cause = fmt.Sprintf("a fault (%s): %s", st.Step.Fault.Kind, st.Step.Fault.Cause)
		}
		return fallback(store, e, pos, p, nodes, edge{at: proc.CallName(nodes[st.Node].Call), why: fmt.Sprintf("%s returned %s", st.Step.To, cause)})
	}
	if refused := refusedEdge(p, nodes, next.Values); refused != nil {
		return fallback(store, e, pos, p, nodes, *refused)
	}
	if next.Return != nil {
		return addressed(e.Frame, next.Return), nil
	}
	pos.Rounds = append(pos.Rounds, next.Round)
	if err := store.SetPosition(e.Frame, pos); err != nil {
		return nil, err
	}
	return next.Emission(), nil
}

// edge is where a refinement refused, and why, as names and presentations.
type edge struct{ at, why string }

// refusedInput checks the input: the header refinement the model wrote, or
// else its base type (decision 10). It says why it refused, or "".
func refusedInput(ref *proc.Refinement, c *proc.Schema, in any) string {
	if ref.Header != "" {
		prg, err := refine.Compile(ref.Header)
		if err == nil {
			var ok bool
			if ok, err = prg.Eval(refine.Vars{Input: in, InputType: proc.InputShown(c, in)}); err == nil && ok {
				return ""
			}
		}
		if err != nil {
			return fmt.Sprintf("the input refinement %s: %v", ref.Header, err)
		}
		return fmt.Sprintf("the input refinement %s is false for this input", ref.Header)
	}
	if ref.Input != "" && proc.InputType(c, in) != ref.Input {
		return fmt.Sprintf("the input is %s; this procedure was compiled for %s", proc.InputShown(c, in), ref.InputShown)
	}
	return ""
}

// refusedEdge checks every node that has returned: its refinement, where the
// model wrote one, else the base type of what it returned (decision 10),
// which must be the trace's.
func refusedEdge(p proc.Procedure, nodes map[string]proc.Node, values map[string]any) *edge {
	ref := p.Header.Refinement
	if ref == nil {
		return nil
	}
	returns, shown := map[string]any{}, map[string]string{}
	for id, v := range values {
		if name, ok := p.Names[id]; ok {
			returns[name], shown[name] = v, proc.Present(nodes[id].Call, v)
		}
	}
	ids := slices.Sorted(maps.Keys(values))
	for _, id := range ids {
		v, call := values[id], nodes[id].Call
		if src, ok := ref.On[id]; ok {
			prg, err := refine.Compile(src)
			ok := false
			if err == nil {
				ok, err = prg.Eval(refine.Vars{Returns: returns, Types: shown})
			}
			switch {
			case err != nil:
				return &edge{at: proc.CallName(call), why: fmt.Sprintf("the refinement on %s, %s: %v", p.Names[id], src, err)}
			case !ok:
				return &edge{at: proc.CallName(call), why: fmt.Sprintf("the refinement on %s, %s, is false: %s returned %s", p.Names[id], src, proc.CallName(call), proc.Present(call, v))}
			}
			continue
		}
		if want, ok := ref.Edges[id]; ok && proc.EdgeType(call, v) != want {
			return &edge{at: proc.CallName(call), why: fmt.Sprintf("%s returned %s; this procedure was compiled for %s",
				proc.CallName(call), proc.Present(call, v), ref.EdgesShown[id])}
		}
	}
	return nil
}

// fallback invokes the original actor, the one the procedure was remembered
// from (decision 15): an ordinary instruction, no interrupt. At an edge, it
// resumes there through the actor's resume contract when it holds one, shown
// the command, what already ran on this input, and which edge refused and
// why, as names and presentations; otherwise, or at entry, it is given the
// command and starts from it.
func fallback(store proc.Store, e envelope, pos proc.Position, p proc.Procedure, nodes map[string]proc.Node, at edge) ([]byte, error) {
	original := p.Header.Original
	if original == "" {
		return addressed(e.Frame, map[string]string{"refused": at.why}), nil
	}
	var payload any = e.Input
	if len(e.Steps) > 0 && resumable(reachOf(e)[original]) {
		var calls []proc.Call
		for _, r := range pos.Rounds {
			calls = append(calls, r...)
		}
		var ran []map[string]any
		for i, st := range e.Steps {
			step := map[string]any{"to": st.To, "payload": st.Payload, "state": st.State}
			if st.Fault != nil {
				step["fault"] = st.Fault
			} else if v, err := types.Decode(st.Value); err == nil && i < len(calls) {
				step["value"] = st.Value
				step["type"] = proc.Present(nodes[calls[i].Node].Call, v)
			}
			ran = append(ran, step)
		}
		payload = map[string]any{"resume": map[string]any{"command": e.Input, "contract": titleOf(p.Header.Contract),
			"ran": ran, "refused": map[string]string{"at": at.at, "why": at.why}}}
	}
	pos.Fallback = true
	pos.Rounds = append(pos.Rounds, []proc.Call{{Node: "fallback", Index: -1}})
	if err := store.SetPosition(e.Frame, pos); err != nil {
		return nil, err
	}
	return instruction(original, payload), nil
}

// resumable reports whether an actor holds a resume contract: one titled
// resume, taking {"resume": ...}.
func resumable(contracts []*proc.Schema) bool {
	for _, c := range contracts {
		if c.Title == "resume" && c.Fields["resume"] != nil {
			return true
		}
	}
	return false
}

// addressed is a return: the payload, addressed to the frame's slot.
func addressed(frame string, payload any) []byte {
	b, _ := json.Marshal(map[string]any{"to": frame, "payload": payload})
	return b
}

// install compiles a procedure source into a state directory.
func install(dir, file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var src proc.Source
	if err := json.Unmarshal(b, &src); err != nil {
		return err
	}
	p, nodes, err := proc.Compile(src)
	if err != nil {
		return err
	}
	if err := (proc.Store{Dir: dir}).Put(p, nodes); err != nil {
		return err
	}
	fmt.Printf("installed %s: %d nodes\n", proc.Key(src.Contract)[:8], len(nodes))
	return nil
}
