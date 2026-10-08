package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"skyra/interp/proc"
)

func delegateContract() *proc.Schema {
	str := &proc.Schema{Type: "string"}
	return &proc.Schema{Title: "delegate", Type: "object", Required: []string{"delegate"}, Fields: map[string]*proc.Schema{
		"delegate": {Type: "object", Required: []string{"goal"}, Fields: map[string]*proc.Schema{
			"goal": str, "actor": str, "tools": {Type: "array", Items: str},
		}},
	}}
}

// Delegate is an interpreter procedure, like remember: its control is the
// recorded narrow/invoke/return composition, not a new actor or provider host.
// No parent transcript or provider session is part of the child input.
func runDelegate(e envelope) ([]byte, error) {
	var ask struct {
		Delegate struct {
			Goal  string    `json:"goal"`
			Actor *string   `json:"actor"`
			Tools *[]string `json:"tools"`
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
	actor := ""
	if ask.Delegate.Actor != nil {
		actor = *ask.Delegate.Actor
	} else if e.Agent != nil {
		actor = e.Agent.Actor
	}
	if _, ok := e.Agents[actor]; !ok || actor == "" {
		return failure("horizon", fmt.Sprintf("%q is not an available agent", actor))
	}
	if ask.Delegate.Actor == nil && (e.Agent == nil || e.Agent.Model == "") {
		return failure("horizon", "caller model is unavailable")
	}
	if len(e.Steps) == 0 {
		scope := map[string]any{"actor": actor, "delegated": true, "inherit_model": ask.Delegate.Actor == nil}
		if ask.Delegate.Tools != nil {
			scope["tools"] = *ask.Delegate.Tools
		}
		return instruction("system", map[string]any{"narrow": scope}), nil
	}
	if len(e.Steps) > 2 {
		return nil, errors.New("delegate has unexpected recorded steps")
	}
	last := e.Steps[len(e.Steps)-1]
	if last.State != "return" {
		kind, cause := "child", last.State
		if last.Fault != nil {
			kind, cause = last.Fault.Kind, last.Fault.Cause
		}
		return failure(kind, cause)
	}
	var scope struct {
		Reference string `json:"reference"`
	}
	if e.Steps[0].To != "system" || json.Unmarshal(e.Steps[0].Value, &scope) != nil || scope.Reference == "" {
		return nil, errors.New("delegate scope did not return a reference")
	}
	if len(e.Steps) == 1 {
		return instruction(scope.Reference, map[string]string{"text": ask.Delegate.Goal, "header": "delegate"}), nil
	}
	if last.To != scope.Reference {
		return nil, errors.New("delegate result is from another invocation")
	}
	var outcome struct {
		Refused string `json:"refused"`
	}
	if json.Unmarshal(last.Value, &outcome) == nil && outcome.Refused != "" {
		return failure("child", outcome.Refused)
	}
	return addressed(e.Frame, last.Value), nil
}
