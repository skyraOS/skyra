package proceduretools

import (
	"encoding/json"
	"maps"
	"slices"
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
		ts = append(ts, Tool{Name: name, Target: target, Schema: s, Description: "A remembered procedure: " + s.Title})
	}
	return ts
}

type Call struct {
	Target, Observation string
	Payload             json.RawMessage
}

// Payload translates model arguments using only model calls in their original
// order. Provider transport steps never enter this index. Invalid remember
// references deliberately become an incomplete request refused by intake.
func Payload(kind string, args json.RawMessage, calls []Call, reach map[string]*Schema, contracts map[string][]*Schema) any {
	switch kind {
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
