package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"skyra/interp/proc"
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

// machinery is what a child narrowed to some tools still holds, whatever
// they are: every available agent and the operations it runs through, so
// the child can run and delegate in turn. Neither is a tool the model calls.
func machinery(e envelope) []contract {
	var kept []contract
	for _, name := range slices.Sorted(maps.Keys(e.Agents)) {
		var agent struct {
			Operations []string `json:"operations"`
		}
		_ = json.Unmarshal(e.Agents[name], &agent)
		for _, actor := range append([]string{name}, agent.Operations...) {
			for _, c := range contractsOf(e, actor) {
				kept = append(kept, contract{actor, c})
			}
		}
	}
	return kept
}

func contractsOf(e envelope, name string) []*proc.Schema {
	if all, ok := e.Contracts[name]; ok {
		return all
	}
	if s := e.Tools[name]; s != nil {
		return []*proc.Schema{s}
	}
	return nil
}

// Delegate is an interpreter procedure, like remember: its control is the
// recorded narrow/invoke/return composition, not a new actor or provider host.
// No parent transcript or provider session is part of the child input.
func runDelegate(e envelope) ([]byte, error) {
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
	actor := ask.Delegate.Actor
	if _, ok := e.Agents[actor]; !ok || actor == "" {
		return failure("boundary", fmt.Sprintf("%q is not an available agent", actor))
	}
	if len(e.Steps) == 0 {
		scope := map[string]any{"actor": actor}
		if ask.Delegate.Tools != nil {
			// The calling adapter turned the tool names into the contracts
			// they stand for (proceduretools.Kept), or left them out when
			// it was not offered one of them.
			if ask.Delegate.Contracts == nil {
				return failure("boundary", fmt.Sprintf("tools %q are not available to the caller", *ask.Delegate.Tools))
			}
			scope["contracts"] = append(slices.Clone(*ask.Delegate.Contracts), machinery(e)...)
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
		// fresh: a one-shot session that loads and keeps nothing of the
		// caller's, and is stopped before the child returns.
		child := map[string]any{"text": ask.Delegate.Goal, "header": "delegate", "fresh": true}
		if ask.Delegate.Model != "" {
			child["model"] = ask.Delegate.Model
		}
		return instruction(scope.Reference, child), nil
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
