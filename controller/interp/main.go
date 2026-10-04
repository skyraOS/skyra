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

	"skyra/interp/proc"
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
func remember() *proc.Schema {
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
	if c == nil || proc.Key(c) == proc.Key(remember()) {
		return addressed(e.Frame, map[string]string{"refused": "remember is not built yet"}), nil
	}
	return runProcedure(store, e, c)
}

// runProcedure is one resume of a remembered procedure.
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
	next, err := proc.Plan(p, nodes, e.Input, pos, e.Steps)
	if err != nil {
		return nil, err
	}
	switch {
	case next.Stuck != nil:
		st := next.Stuck.Step
		cause := st.State
		if st.Fault != nil {
			cause = fmt.Sprintf("%s (%s): %s", st.State, st.Fault.Kind, st.Fault.Cause)
		}
		return addressed(e.Frame, map[string]string{"refused": fmt.Sprintf("a call did not return: %s", cause)}), nil
	case next.Return != nil:
		return addressed(e.Frame, next.Return), nil
	}
	pos.Rounds = append(pos.Rounds, next.Round)
	if err := store.SetPosition(e.Frame, pos); err != nil {
		return nil, err
	}
	return next.Emission(), nil
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
