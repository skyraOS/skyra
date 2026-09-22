package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
)

const Codex ExecutableID = 19

type ParentRef struct {
	Namespace  string    `json:"namespace"`
	Owner      ContextID `json:"owner"`
	Sequence   FrameID   `json:"sequence"`
	Generation uint64    `json:"generation"`
	Token      string    `json:"token"`
}
type Field struct {
	Key      string   `json:"key"`
	Encoding string   `json:"encoding"`
	Required bool     `json:"required"`
	MaxBytes int      `json:"max_bytes"`
	Values   []string `json:"values,omitempty"`
}
type Profile struct {
	Operation    string   `json:"operation"`
	Fields       []Field  `json:"fields"`
	Results      []string `json:"results"`
	Failures     []string `json:"failures"`
	Continuation bool     `json:"continuation,omitempty"`
	Lifecycle    string   `json:"lifecycle"`
}
type Descriptor struct {
	Protocol int       `json:"protocol"`
	Version  int       `json:"version"`
	Profiles []Profile `json:"profiles"`
}

func StrictJSON(raw []byte, value any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return err
	}
	if d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("trailing JSON")
	}
	return nil
}
func LoadDescriptor(raw []byte) (Descriptor, error) {
	var d Descriptor
	if err := StrictJSON(raw, &d); err != nil {
		return d, err
	}
	if d.Protocol != 1 || d.Version != 1 || len(d.Profiles) == 0 {
		return d, fmt.Errorf("incompatible descriptor")
	}
	seen := map[string]bool{}
	for _, p := range d.Profiles {
		if p.Operation == "" || seen[p.Operation] || len(p.Results) == 0 || len(p.Failures) == 0 || p.Lifecycle == "" {
			return d, fmt.Errorf("malformed profile")
		}
		seen[p.Operation] = true
		fields := map[string]bool{}
		for _, f := range p.Fields {
			if f.Key == "" || f.Key == "operation" || fields[f.Key] || f.MaxBytes < 1 || f.MaxBytes > 65536 {
				return d, fmt.Errorf("malformed field")
			}
			fields[f.Key] = true
			switch f.Encoding {
			case "utf8", "parent", "uint", "tools":
			default:
				return d, fmt.Errorf("unknown encoding")
			}
		}
	}
	return d, nil
}
func (d Descriptor) Select(args []Observation, continuation bool) (Profile, map[string]string, error) {
	values := map[string]string{}
	for _, a := range args {
		if _, ok := values[a.Key]; ok {
			return Profile{}, nil, fmt.Errorf("duplicate observation")
		}
		values[a.Key] = string(a.Bytes)
	}
	for _, p := range d.Profiles {
		if p.Operation == values["operation"] {
			err := p.Check(args, continuation)
			return p, values, err
		}
	}
	return Profile{}, nil, fmt.Errorf("unknown operation")
}
func (p Profile) Check(args []Observation, continuation bool) error {
	if p.Continuation != continuation {
		return fmt.Errorf("profile lifecycle violation")
	}
	fields := map[string]Field{"operation": {Key: "operation", Encoding: "utf8", Required: true, MaxBytes: 64, Values: []string{p.Operation}}}
	for _, f := range p.Fields {
		fields[f.Key] = f
	}
	seen := map[string]bool{}
	for _, a := range args {
		f, ok := fields[a.Key]
		if !ok || seen[a.Key] || len(a.Bytes) > f.MaxBytes || !utf8.Valid(a.Bytes) {
			return fmt.Errorf("invalid observation %q", a.Key)
		}
		seen[a.Key] = true
		s := string(a.Bytes)
		if len(f.Values) > 0 {
			ok = false
			for _, v := range f.Values {
				if v == s {
					ok = true
				}
			}
			if !ok {
				return fmt.Errorf("constraint rejected %s", a.Key)
			}
		}
		switch f.Encoding {
		case "parent":
			var r ParentRef
			if StrictJSON(a.Bytes, &r) != nil || r.Namespace == "" || r.Owner == "" || r.Sequence == 0 || r.Generation == 0 || r.Token == "" {
				return fmt.Errorf("invalid parent reference")
			}
		case "uint":
			n, e := strconv.ParseUint(s, 10, 32)
			if e != nil || n < 1 || n > 30000 {
				return fmt.Errorf("deadline must be 1..30000 ms")
			}
		case "tools":
			var names []string
			if StrictJSON(a.Bytes, &names) != nil || names == nil {
				return fmt.Errorf("invalid tools")
			}
			ts := map[string]bool{}
			for _, n := range names {
				if n != "echo" || ts[n] {
					return fmt.Errorf("tool outside grant")
				}
				ts[n] = true
			}
		}
	}
	for k, f := range fields {
		if f.Required && !seen[k] {
			return fmt.Errorf("missing %s", k)
		}
	}
	return nil
}
func Args(operation string, pairs ...string) []Observation {
	a := []Observation{{Key: "operation", Bytes: []byte(operation)}}
	for i := 0; i+1 < len(pairs); i += 2 {
		a = append(a, Observation{Key: pairs[i], Bytes: []byte(pairs[i+1])})
	}
	return a
}
