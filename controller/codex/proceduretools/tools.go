package proceduretools

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
)

type Tool struct {
	Name, Description, Target, Kind string
	Schema                          *Schema
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
	ts := []Tool{
		{Name: "remember", Target: target, Kind: "remember", Description: rememberDescription,
			Schema: &Schema{Type: "object", Required: []string{"call"}, Fields: map[string]*Schema{"call": {Type: "number"}}}},
		{Name: "procedure", Target: target, Kind: "procedure", Description: "Run a remembered procedure on new input: input is what you would give the call it was remembered from.",
			Schema: &Schema{Type: "object", Required: []string{"input"}, Fields: map[string]*Schema{"input": {Type: "any"}}}},
	}
	reserved = append(slices.Clone(reserved), "remember", "procedure")
	for _, s := range ContractsOf(reach, contracts, target) {
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
		ts = append(ts, Tool{Name: name, Target: target, Kind: kind, Schema: shown, Description: description})
	}
	return ts
}

// withoutModel is the delegate contract as a model is shown it: the model
// field is the adapter's to fill, never the model's to choose.
func withoutModel(s *Schema) *Schema {
	inner := s.Fields["delegate"]
	if inner == nil || inner.Fields["model"] == nil {
		return s
	}
	shownInner := *inner
	shownInner.Fields = maps.Clone(inner.Fields)
	delete(shownInner.Fields, "model")
	shown := *s
	shown.Fields = maps.Clone(s.Fields)
	shown.Fields["delegate"] = &shownInner
	return &shown
}

type Call struct {
	Target, Observation string
	Payload             json.RawMessage
}

// Caller is the adapter a delegate inherits from when the model names no
// actor: its own name and the model it resolved. The adapter supplies it from
// its recorded steps; the model's arguments never do.
type Caller struct{ Actor, Model string }

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

// Allows applies the runtime's tool ceiling to transport names. The kernel
// separately checks the concrete target and selected contract at intake.
func Allows(names *[]string, name string) bool {
	if names == nil {
		return true
	}
	for _, allowed := range *names {
		if name == allowed || strings.ReplaceAll(name, ".", "_") == strings.ReplaceAll(allowed, ".", "_") {
			return true
		}
	}
	return false
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
