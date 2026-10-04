// Command interp is the interpreter (next/32): an actor that runs procedures
// from its own state, as an interpreter process sits above a CPU. The kernel
// never sees a procedure. It sees an invoke of one of the interpreter's
// contracts, the instructions the interpreter emits, and a return; every
// instruction is checked at intake under the Horizon it runs in, like any
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
	"os"
	"slices"

	"skyra/interp/proc"
	"skyra/interp/remember"
	"skyra/interp/types"
)

//go:embed contract.json
var declared []byte

// envelope is what the kernel hands a continuation on stdin.
type envelope struct {
	Frame     string                    `json:"frame"`
	Input     json.RawMessage           `json:"input"`
	Contract  *proc.Schema              `json:"contract"`
	Tools     map[string]*proc.Schema   `json:"tools"`
	Contracts map[string][]*proc.Schema `json:"contracts"`
	Steps     []proc.Step               `json:"steps"`
}

func main() {
	var err error
	if len(os.Args) == 4 && os.Args[1] == "install" {
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
	return runProcedure(store, e, c)
}

// runRemember is remember's body (decision 8), embedded in this binary: a
// syscall for the strand, the compile, then one extend-reach per contract a
// procedure was stored for. The interpreter extends itself; an actor can
// only extend itself.
//
//	no steps        → system strand: the observation, and the contract that requested it
//	the strand      → compile it; refuse a contract that already has a procedure, before
//	                  anything is stored; store every procedure; extend-reach the first
//	each extension  → extend-reach the next, or return what was remembered
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
		return instruction("system", map[string]any{"strand": map[string]any{"observation": r.Observation,
			"contract": map[string]any{"target": r.Contract.Target, "schema": r.Contract.Schema}}}), nil
	}
	last := e.Steps[len(e.Steps)-1]
	if last.State != "return" {
		cause := last.State
		if last.Fault != nil {
			cause = last.Fault.Cause
		}
		if len(e.Steps) > 1 {
			cause = "extend-reach: " + cause
		}
		return refused(cause)
	}
	done, compiled, err := store.Remembering(e.Frame)
	if err != nil {
		return nil, err
	}
	if !compiled {
		if len(e.Steps) != 1 {
			return nil, errors.New("extensions before a compile")
		}
		var st remember.Strand
		if err := json.Unmarshal(last.Value, &st); err != nil {
			return nil, fmt.Errorf("strand: %w", err)
		}
		reach := reachOf(e)
		self := selfName(reach)
		if self == "" {
			return refused("the interpreter is not in its own reach")
		}
		srcs, err := remember.Compile(st, r.Observation, r.Contract.Target, r.Contract.Schema, reach, self)
		var why remember.Refused
		if errors.As(err, &why) {
			return refused(string(why))
		}
		if err != nil {
			return nil, err
		}
		// One representation per contract (decision 28), refused before
		// anything is stored, so no Horizon is given a procedure another
		// Horizon remembered.
		for i, src := range srcs {
			if store.Has(src.Contract) {
				return refused(fmt.Sprintf("%s (from %s) already has a procedure: one representation per contract", titleOf(src.Contract), src.Original))
			}
			for _, other := range srcs[:i] {
				if proc.Same(other.Contract, src.Contract) {
					return refused(fmt.Sprintf("%s is remembered twice in one strand", titleOf(src.Contract)))
				}
			}
		}
		for _, src := range srcs {
			p, nodes, err := proc.Compile(src)
			if err != nil {
				return nil, err
			}
			if err := store.Put(p, nodes); err != nil {
				return nil, err
			}
			done.Contracts = append(done.Contracts, src.Contract)
			done.Originals = append(done.Originals, src.Original)
		}
		if err := store.SetRemembering(e.Frame, done); err != nil {
			return nil, err
		}
	}
	if next := len(e.Steps) - 1; next < len(done.Contracts) {
		return instruction("system", map[string]any{"extend-reach": map[string]any{"schema": done.Contracts[next]}}), nil
	}
	var names []string
	for i, c := range done.Contracts {
		name := c.Title
		if name == "" && i < len(done.Originals) {
			name = done.Originals[i] // an untitled contract is named by the actor it was remembered from
		}
		names = append(names, name)
	}
	return addressed(e.Frame, map[string]any{"remembered": names}), nil
}

// reachOf is every contract each name in the envelope holds.
func reachOf(e envelope) remember.Reach {
	reach := remember.Reach{}
	for name, t := range e.Tools {
		if all, ok := e.Contracts[name]; ok {
			reach[name] = all
		} else if t != nil {
			reach[name] = []*proc.Schema{t}
		}
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
	if len(e.Steps) == 0 && ref != nil && ref.Input != "" && proc.InputType(c, in) != ref.Input {
		why := fmt.Sprintf("the input is %s; this procedure was compiled for %s", proc.InputShown(c, in), ref.InputShown)
		return fallback(store, e, pos, p, nodes, edge{at: "input", why: why})
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
	if refused := refusedEdge(ref, nodes, next.Values); refused != nil {
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

// refusedEdge checks the default refinement (decision 10) on every node that
// has returned: the base type of what it returned must be the strand's.
func refusedEdge(ref *proc.Refinement, nodes map[string]proc.Node, values map[string]any) *edge {
	if ref == nil {
		return nil
	}
	ids := make([]string, 0, len(ref.Edges))
	for id := range ref.Edges {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		v, done := values[id]
		if !done {
			continue
		}
		if call := nodes[id].Call; proc.EdgeType(call, v) != ref.Edges[id] {
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
