package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"skyra/apprentice/proc"
)

func delegateContract() *proc.Schema {
	str := &proc.Schema{Type: "string"}
	return &proc.Schema{Title: "delegate", Type: "object", Required: []string{"delegate"}, Fields: map[string]*proc.Schema{
		"delegate": {Type: "object", Required: []string{"goal"}, Fields: map[string]*proc.Schema{
			"goal": str, "actor": str, "model": str, "tools": {Type: "array", Items: str},
			"contracts": {Type: "array", Items: &proc.Schema{Type: "object", Required: []string{"target", "schema"}, Fields: map[string]*proc.Schema{
				"target": str, "schema": {Type: "any"},
			}}},
		}},
	}}
}

type contract struct {
	Target string       `json:"target"`
	Schema *proc.Schema `json:"schema"`
}

func contractsOf(e envelope, name string) []*proc.Schema {
	return e.Reach[name].Contracts
}

// Delegate is an apprentice procedure, like remember: its control is the
// recorded narrow/invoke/return composition, not a new actor or provider host.
// No parent transcript or provider session is part of the child input.
//
// The kernel provisions the child's Boundary once, through system.narrow, and
// hands back a handle: a hash, which is all the kernel knows it by. The
// apprentice keeps the handle in its table under a name for the scope, and
// later delegations to the same scope invoke it directly, with no narrowing.
// A handle the kernel refuses (this Boundary could not provision it) is
// dropped and provisioned again.
//
//	no steps          → invoke the remembered handle, or narrow
//	the narrowing     → remember its handle, and invoke it
//	a refused handle  → forget it, and narrow
//	the child         → its result
func runDelegate(store proc.Store, e envelope) ([]byte, error) {
	var ask struct {
		Delegate struct {
			Goal      string      `json:"goal"`
			Actor     string      `json:"actor"`
			Model     string      `json:"model"`
			Tools     *[]string   `json:"tools"`
			Contracts *[]contract `json:"contracts"`
		} `json:"delegate"`
	}
	if err := json.Unmarshal(e.Input, &ask); err != nil {
		return nil, err
	}
	failure := func(kind, cause string) ([]byte, error) {
		return addressed(e.Frame, map[string]any{"failed": map[string]string{"kind": kind, "cause": cause}}), nil
	}
	if strings.TrimSpace(ask.Delegate.Goal) == "" {
		return failure("contract", "delegate goal must be nonempty")
	}
	// The calling adapter fills in actor and model when the model names no
	// actor (proceduretools.Payload); a delegate with neither has no caller
	// to inherit from.
	// Any actor in reach can be delegated to: whether it takes a goal is its
	// contract's to say, at intake. What it runs through, narrow keeps.
	actor := ask.Delegate.Actor
	if actor == "" || contractsOf(e, actor) == nil {
		return failure("boundary", fmt.Sprintf("%q is not an actor in reach", actor))
	}
	scope := map[string]any{"actor": actor}
	if ask.Delegate.Tools != nil {
		// The calling adapter turned the tool names into the contracts
		// they stand for (proceduretools.Kept), or left them out when
		// it was not offered one of them.
		if ask.Delegate.Contracts == nil {
			return failure("boundary", fmt.Sprintf("tools %q are not available to the caller", *ask.Delegate.Tools))
		}
		scope["contracts"] = *ask.Delegate.Contracts
	}
	named, _ := json.Marshal(scope)
	name := "delegate " + string(named)
	narrow := instruction("system", map[string]any{"narrow": scope})
	// fresh: a one-shot session that loads and keeps nothing of the
	// caller's, and is stopped before the child returns.
	child := map[string]any{"text": ask.Delegate.Goal, "header": "delegate", "fresh": true}
	if ask.Delegate.Tools != nil {
		child["tools"] = *ask.Delegate.Tools // the child's grant, in its thread
	}
	if ask.Delegate.Model != "" {
		child["model"] = ask.Delegate.Model
	}

	if len(e.Steps) == 0 {
		handle, ok, err := store.Handle(name)
		if err != nil {
			return nil, err
		}
		if !ok {
			return narrow, nil
		}
		return instruction(handle, child), nil
	}
	if len(e.Steps) > 3 {
		return nil, errors.New("delegate has unexpected recorded steps")
	}
	last := e.Steps[len(e.Steps)-1]
	failed := func() ([]byte, error) {
		kind, cause := "child", string(last.State)
		if last.Fault != nil {
			kind, cause = last.Fault.Kind, last.Fault.Cause
		}
		return failure(kind, cause)
	}
	if last.To == "system" {
		if last.State != "return" {
			return failed()
		}
		var got struct {
			Handle string `json:"handle"`
		}
		if json.Unmarshal(last.Value, &got) != nil || got.Handle == "" {
			return nil, errors.New("delegate scope did not return a handle")
		}
		if err := store.SetHandle(name, got.Handle); err != nil {
			return nil, err
		}
		return instruction(got.Handle, child), nil
	}
	if len(e.Steps) == 1 {
		handle, _, err := store.Handle(name)
		if err != nil {
			return nil, err
		}
		if handle != "" && refused(last, handle) {
			// The remembered handle: the kernel would not run it under
			// this Boundary. Provision the scope again.
			if err := store.DropHandle(name); err != nil {
				return nil, err
			}
			return narrow, nil
		}
	}
	if last.State != "return" {
		return failed()
	}
	var outcome struct {
		Refused string `json:"refused"`
	}
	if json.Unmarshal(last.Value, &outcome) == nil && outcome.Refused != "" {
		return failure("child", outcome.Refused)
	}
	return addressed(e.Frame, last.Value), nil
}

// refused reports whether a step is intake refusing the invoke of handle
// itself: a boundary fault whose observation is that attempt, not a fault of
// the child, which discharges the same slot.
func refused(step proc.Step, handle string) bool {
	if step.State != "fault" || step.Fault == nil || step.Fault.Kind != "boundary" {
		return false
	}
	var attempt struct {
		To string `json:"to"`
	}
	return json.Unmarshal(step.Value, &attempt) == nil && attempt.To == handle
}
