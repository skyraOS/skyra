package proceduretools

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
)

type Tool struct {
	Name, Description, Target, Kind string
	Schema                          *Schema   // what the model is shown
	Contracts                       []*Schema // the contracts of Target a call to it may select
}

func Interpreter(reach map[string]*Schema) string {
	for _, name := range slices.Sorted(maps.Keys(reach)) {
		if s := reach[name]; s != nil && s.Title == "remember" {
			return name
		}
	}
	return ""
}

func ContractsOf(reach map[string]*Schema, contracts map[string][]*Schema, name string) []*Schema {
	if all, ok := contracts[name]; ok {
		return all
	}
	if s := reach[name]; s != nil {
		return []*Schema{s}
	}
	return nil
}

// Procedures adds the same deterministic procedure bindings to either
// transport's capability catalog. Reserved names keep the direct capability
// binding when a learned title collides with it.
func Procedures(reach map[string]*Schema, contracts map[string][]*Schema, reserved []string, rememberDescription string) []Tool {
	target := Interpreter(reach)
	if target == "" {
		return nil
	}
	all := ContractsOf(reach, contracts, target)
	ts := []Tool{
		{Name: "remember", Target: target, Kind: "remember", Description: rememberDescription,
			Schema: &Schema{Type: "object", Required: []string{"call"}, Fields: map[string]*Schema{"call": {Type: "number"}}}, Contracts: all[:1]},
		{Name: "procedure", Target: target, Kind: "procedure", Description: "Run a remembered procedure on new input: input is what you would give the call it was remembered from.",
			Schema: &Schema{Type: "object", Required: []string{"input"}, Fields: map[string]*Schema{"input": {Type: "any"}}}, Contracts: all[1:]},
	}
	reserved = append(slices.Clone(reserved), "remember", "procedure")
	for _, s := range all {
		if s == nil || s.Title == "remember" {
			continue
		}
		name := ToolName(s.Title)
		if name == "" || slices.Contains(reserved, name) {
			continue
		}
		reserved = append(reserved, name)
		description, kind := "A remembered procedure: "+s.Title, ""
		if s.Title == "delegate" {
			description, kind = "Delegate a goal to a fresh child agent and receive its findings. Input: {delegate:{goal,actor?,tools?}}. Omit actor to inherit your actor and model; explicit actor uses its configured default. Omit tools to inherit available work tools; tools narrows them, [] allows reasoning only. Only the goal is passed as task context. Check any returned failure before continuing.", "delegate"
		}
		shown := s
		if kind == "delegate" {
			shown = withoutModel(s)
		}
		ts = append(ts, Tool{Name: name, Target: target, Kind: kind, Schema: shown, Description: description, Contracts: []*Schema{s}})
	}
	return ts
}

// withoutModel is the delegate contract as a model is shown it: model and
// contracts are the adapter's to fill, never the model's to choose.
func withoutModel(s *Schema) *Schema {
	inner := s.Fields["delegate"]
	if inner == nil {
		return s
	}
	shownInner := *inner
	shownInner.Fields = maps.Clone(inner.Fields)
	delete(shownInner.Fields, "model")
	delete(shownInner.Fields, "contracts")
	shown := *s
	shown.Fields = maps.Clone(s.Fields)
	shown.Fields["delegate"] = &shownInner
	return &shown
}

type Call struct {
	Target, Observation string
	Payload             json.RawMessage
}

// Caller is the adapter a delegate is made from: its own name and the model
// it resolved, which a delegate naming no actor inherits, and the tools it
// offers, by which a delegate's tool names become contracts. The adapter
// supplies it from its recorded steps; the model's arguments never do.
type Caller struct {
	Actor, Model string
	Tools        []Tool
}

// Delegable is every actor in reach a delegate can hand a goal to: one with
// a contract its child input fits. Nothing marks an actor as an agent; its
// contract says whether it takes a goal.
func Delegable(reach map[string]*Schema, contracts map[string][]*Schema) []string {
	goal := map[string]any{"text": "", "header": "delegate", "fresh": true, "model": ""}
	var out []string
	for _, name := range slices.Sorted(maps.Keys(reach)) {
		if slices.ContainsFunc(ContractsOf(reach, contracts, name), func(c *Schema) bool { return Fits(c, goal) }) {
			out = append(out, name)
		}
	}
	return out
}

// Contract is one contract of one actor, as system.narrow keeps it.
type Contract struct {
	Target string  `json:"target"`
	Schema *Schema `json:"schema"`
}

// Kept is the contracts the named tools stand for, by the tools offered:
// a tool is named as offered, or with dots for its underscores (repo.write
// for repo_write).
// It reports false when a name is not one of them, so a delegate cannot name
// what its caller was not offered.
func Kept(offered []Tool, names []string) ([]Contract, bool) {
	kept := []Contract{}
	for _, name := range names {
		i := slices.IndexFunc(offered, func(t Tool) bool { return t.Name == strings.ReplaceAll(name, ".", "_") })
		if i < 0 {
			return nil, false
		}
		for _, c := range offered[i].Contracts {
			kept = append(kept, Contract{offered[i].Target, c})
		}
	}
	return kept, true
}

// Payload translates model arguments using only model calls in their original
// order. Provider transport steps never enter this index. Invalid remember
// references deliberately become an incomplete request refused by intake.
func Payload(kind string, args json.RawMessage, calls []Call, reach map[string]*Schema, contracts map[string][]*Schema, caller Caller) any {
	switch kind {
	case "delegate":
		var a struct {
			Delegate map[string]json.RawMessage `json:"delegate"`
		}
		if json.Unmarshal(args, &a) != nil || a.Delegate == nil {
			return args
		}
		delete(a.Delegate, "model")
		delete(a.Delegate, "contracts")
		var names *[]string
		if json.Unmarshal(a.Delegate["tools"], &names) == nil && names != nil {
			if kept, ok := Kept(caller.Tools, *names); ok {
				a.Delegate["contracts"], _ = json.Marshal(kept)
			}
		}
		if _, named := a.Delegate["actor"]; !named && caller.Actor != "" {
			a.Delegate["actor"], _ = json.Marshal(caller.Actor)
			if caller.Model != "" {
				a.Delegate["model"], _ = json.Marshal(caller.Model)
			}
		}
		return map[string]any{"delegate": a.Delegate}
	case "remember":
		var a struct {
			Call *int `json:"call"`
		}
		if json.Unmarshal(args, &a) != nil || a.Call == nil || *a.Call < 0 || *a.Call >= len(calls) {
			return map[string]any{"remember": map[string]any{}}
		}
		call := calls[*a.Call]
		return map[string]any{"remember": map[string]any{"observation": call.Observation,
			"contract": map[string]any{"target": call.Target, "schema": Selected(ContractsOf(reach, contracts, call.Target), call.Payload)}}}
	case "procedure":
		var a struct {
			Input json.RawMessage `json:"input"`
		}
		_ = json.Unmarshal(args, &a)
		return a.Input
	default:
		return args
	}
}

// Offered reports whether a tool is offered under a continuation's c-list:
// every contract it stands for is held, as Kept keeps them. nil holds
// everything the continuation is shown. The kernel separately checks the
// selected contract at intake.
func Offered(held map[string][]*Schema, t Tool) bool {
	if held == nil {
		return true
	}
	return len(t.Contracts) > 0 && !slices.ContainsFunc(t.Contracts, func(c *Schema) bool {
		return !slices.ContainsFunc(held[t.Target], func(h *Schema) bool { return same(c, h) })
	})
}

// Holds reports whether any contract of target is held.
func Holds(held map[string][]*Schema, target string) bool {
	return held == nil || len(held[target]) > 0
}

func same(a, b *Schema) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}

// Failure is a procedure failure outcome, not discovered observations.
func Failure(value json.RawMessage) string {
	var outcome struct {
		Failed *struct{ Kind, Cause string } `json:"failed"`
	}
	if json.Unmarshal(value, &outcome) == nil && outcome.Failed != nil {
		return outcome.Failed.Kind + ": " + outcome.Failed.Cause
	}
	return ""
}
