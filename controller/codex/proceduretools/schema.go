// Package proceduretools contains pure model-facing procedure bindings shared by Codex transports.
package proceduretools

import (
	"encoding/json"
	"strings"
)

type Schema struct {
	Title    string             `json:"title,omitempty"`
	Type     string             `json:"type"`
	Fields   map[string]*Schema `json:"fields,omitempty"`
	Required []string           `json:"required,omitempty"`
	Items    *Schema            `json:"items,omitempty"`
}

func Selected(contracts []*Schema, payload json.RawMessage) *Schema {
	var v any
	d := json.NewDecoder(strings.NewReader(string(payload)))
	d.UseNumber()
	if d.Decode(&v) != nil {
		return nil
	}
	var fit []*Schema
	for _, c := range contracts {
		if Fits(c, v) {
			fit = append(fit, c)
		}
	}
	if len(fit) != 1 {
		return nil
	}
	return fit[0]
}

func Fits(s *Schema, v any) bool {
	if s == nil {
		return false
	}
	switch s.Type {
	case "object":
		o, ok := v.(map[string]any)
		if !ok {
			return false
		}
		for k, e := range o {
			if f := s.Fields[k]; f == nil || !Fits(f, e) {
				return false
			}
		}
		for _, r := range s.Required {
			if _, ok := o[r]; !ok {
				return false
			}
		}
		return true
	case "array":
		a, ok := v.([]any)
		if !ok {
			return false
		}
		for _, e := range a {
			if !Fits(s.Items, e) {
				return false
			}
		}
		return true
	case "string":
		_, ok := v.(string)
		return ok
	case "number":
		_, ok := v.(json.Number)
		return ok
	case "bool":
		_, ok := v.(bool)
		return ok
	case "null":
		return v == nil
	case "any":
		return true
	}
	return false
}

func ToolName(title string) string {
	if title == "" || len(title) > 64 {
		return ""
	}
	for _, r := range title {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return ""
		}
	}
	return title
}

func JSONSchema(s *Schema) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	switch s.Type {
	case "object":
		p := map[string]any{}
		for k, v := range s.Fields {
			p[k] = JSONSchema(v)
		}
		r := s.Required
		if r == nil {
			r = []string{}
		}
		return map[string]any{"type": "object", "properties": p, "required": r, "additionalProperties": false}
	case "array":
		return map[string]any{"type": "array", "items": JSONSchema(s.Items)}
	case "bool":
		return map[string]any{"type": "boolean"}
	case "any":
		return map[string]any{}
	default:
		return map[string]any{"type": s.Type}
	}
}
