// Package contract owns input contracts: JSON-types schemas and the fit rule.
// It is pure and imports nothing from this module.
//
// Every capability has an input contract. The reply capability's input
// contract is the return contract of whatever is running, so one check covers
// both invoking and returning: does this payload fit the contract of the
// capability it was sent to?
package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"unicode/utf8"
)

// Schema is a JSON type. Fit is structural and strict: an object's unknown
// fields do not fit. Its title names it where a name is shown (a tool
// offered to a model, a procedure's header); fit ignores it.
type Schema struct {
	Title    string             `json:"title,omitempty"`    // what it is called, never part of fit
	Type     string             `json:"type"`               // object | array | string | number | bool | null | any
	Fields   map[string]*Schema `json:"fields,omitempty"`   // object: its known fields
	Required []string           `json:"required,omitempty"` // object: fields that must be present
	Items    *Schema            `json:"items,omitempty"`    // array: every element
}

// Check reports whether a schema is well formed.
func (s *Schema) Check() error {
	if s == nil {
		return errors.New("no schema")
	}
	switch s.Type {
	case "object":
		for _, r := range s.Required {
			if s.Fields[r] == nil {
				return fmt.Errorf("required field %q is not declared", r)
			}
		}
		for k, f := range s.Fields {
			if err := f.Check(); err != nil {
				return fmt.Errorf("%s: %w", k, err)
			}
		}
	case "array":
		if err := s.Items.Check(); err != nil {
			return fmt.Errorf("items: %w", err)
		}
	case "string", "number", "bool", "null", "any":
	default:
		return fmt.Errorf("unknown type %q", s.Type)
	}
	return nil
}

// Decode reads exactly one JSON value, keeping numbers as written.
func Decode(b []byte) (any, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, errors.New("trailing data after one JSON value")
	}
	return v, nil
}

// Canonical re-encodes a decoded value: object keys sorted, numbers as
// written. Records store payloads in this form, so equal payloads hash equal.
func Canonical(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// Normal is a copy of s in its normal form: every Required, at every depth,
// sorted and deduplicated (42). Two contracts that differ only in the order
// of their required fields are one contract. The apprentice's
// proc.Schema.Canonical computes the same form; testdata/normal.json holds
// the cases both must reproduce.
func Normal(s *Schema) *Schema {
	if s == nil {
		return nil
	}
	n := &Schema{Title: s.Title, Type: s.Type, Items: Normal(s.Items)}
	if len(s.Required) > 0 {
		n.Required = slices.Clone(s.Required)
		slices.Sort(n.Required)
		n.Required = slices.Compact(n.Required)
	}
	if len(s.Fields) > 0 {
		n.Fields = make(map[string]*Schema, len(s.Fields))
		for k, f := range s.Fields {
			n.Fields[k] = Normal(f)
		}
	}
	return n
}

// Same reports whether two contracts are one: the same normal form. Every
// comparison of contracts goes through it, so a contract restated in another
// order is recognized wherever it is compared, in old trees as in new.
func Same(a, b *Schema) bool {
	if a == nil || b == nil {
		return a == b
	}
	return string(Canonical(Normal(a))) == string(Canonical(Normal(b)))
}

// Fits checks a decoded value against a schema.
func Fits(s *Schema, v any) error {
	if s == nil {
		return errors.New("no contract")
	}
	switch s.Type {
	case "object":
		o, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("want an object, got %s", kind(v))
		}
		keys := make([]string, 0, len(o))
		for k := range o {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			f := s.Fields[k]
			if f == nil {
				return fmt.Errorf("unknown field %q", k)
			}
			if err := Fits(f, o[k]); err != nil {
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
			return fmt.Errorf("want an array, got %s", kind(v))
		}
		for i, e := range a {
			if err := Fits(s.Items, e); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
	case "string":
		str, ok := v.(string)
		if !ok {
			return fmt.Errorf("want a string, got %s", kind(v))
		}
		if !utf8.ValidString(str) {
			return errors.New("string is not valid UTF-8")
		}
	case "number":
		if _, ok := v.(json.Number); !ok {
			return fmt.Errorf("want a number, got %s", kind(v))
		}
	case "bool":
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("want a bool, got %s", kind(v))
		}
	case "null":
		if v != nil {
			return fmt.Errorf("want null, got %s", kind(v))
		}
	case "any":
		// Any JSON value: a syscall that takes a schema, or a value it passes on.
	default:
		return fmt.Errorf("unknown type %q", s.Type)
	}
	return nil
}

func kind(v any) string {
	switch v.(type) {
	case map[string]any:
		return "an object"
	case []any:
		return "an array"
	case string:
		return "a string"
	case json.Number:
		return "a number"
	case bool:
		return "a bool"
	case nil:
		return "null"
	}
	return fmt.Sprintf("%T", v)
}
