package proc

import (
	"encoding/json"
	"fmt"
	"slices"
	"unicode/utf8"
)

// Fits checks a decoded value against a contract, as the kernel does at
// intake: structural and strict. The interpreter uses it to know which of a
// name's contracts a recorded call went through (the kernel selected the one
// it fits), never to admit anything: intake is the kernel's.
func Fits(s *Schema, v any) error {
	if s == nil {
		return fmt.Errorf("no contract")
	}
	switch s.Type {
	case "object":
		o, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("want an object")
		}
		for k, e := range o {
			f := s.Fields[k]
			if f == nil {
				return fmt.Errorf("unknown field %q", k)
			}
			if err := Fits(f, e); err != nil {
				return fmt.Errorf("%s: %w", k, err)
			}
		}
		for _, r := range s.Required {
			if _, ok := o[r]; !ok {
				return fmt.Errorf("missing required field %q", r)
			}
		}
	case "array":
		a, ok := v.([]any)
		if !ok {
			return fmt.Errorf("want an array")
		}
		for i, e := range a {
			if err := Fits(s.Items, e); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
	case "string":
		str, ok := v.(string)
		if !ok || !utf8.ValidString(str) {
			return fmt.Errorf("want a string")
		}
	case "number":
		if _, ok := v.(json.Number); !ok {
			return fmt.Errorf("want a number")
		}
	case "bool":
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("want a bool")
		}
	case "null":
		if v != nil {
			return fmt.Errorf("want null")
		}
	case "any":
	default:
		return fmt.Errorf("unknown type %q", s.Type)
	}
	return nil
}

// Selected is the one contract among several that a payload fits, or nil.
func Selected(contracts []*Schema, payload json.RawMessage) *Schema {
	v, err := decode(payload)
	if err != nil {
		return nil
	}
	var fit []*Schema
	for _, c := range contracts {
		if Fits(c, v) == nil {
			fit = append(fit, c)
		}
	}
	if len(fit) != 1 {
		return nil
	}
	return fit[0]
}

// Same reports whether two contracts are one: the same canonical bytes.
func Same(a, b *Schema) bool {
	return a != nil && b != nil && slices.Equal(a.Canonical(), b.Canonical())
}

// CallName is the name a node's call stands for in its base types: the name
// it calls, or "reference" for a call through a binding, whose name is the
// run's own (a reference narrow hands back is named afresh every run).
func CallName(call json.RawMessage) string {
	var s string
	if json.Unmarshal(call, &s) == nil {
		return s
	}
	return "reference"
}
