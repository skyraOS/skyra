// Package mission defines the typed Plan/Spec/Launch boundary. It performs
// only pure validation and resolution; it never grants OS access or runs a tree.
package mission

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	root "skyra/controller/core/humanroot"
)

// Contract is the worker protocol's identity: binary plus canonical input shape.
// Name, resource scope, and the human's authority are not its type.
type Contract struct {
	Binary string `json:"binary"`
	Shape  string `json:"shape"`
}

func validHash(id, prefix string, n int) bool {
	if !strings.HasPrefix(id, prefix) || len(id) != len(prefix)+n {
		return false
	}
	for _, c := range id[len(prefix):] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func Identify(binary string, schema *contract.Schema) (Contract, error) {
	if !(validHash(binary, "sha256:", 64) || validHash(binary, "cdhash:", 40)) {
		return Contract{}, errors.New("contract requires a pinned binary identity")
	}
	if err := schema.Check(); err != nil {
		return Contract{}, err
	}
	return Contract{binary, string(contract.Canonical(contract.Normal(schema)))}, nil
}
func (c Contract) Validate() error {
	var schema contract.Schema
	if err := strict([]byte(c.Shape), &schema); err != nil {
		return err
	}
	normal, err := Identify(c.Binary, &schema)
	if err != nil {
		return err
	}
	if normal != c {
		return errors.New("contract shape is not canonical")
	}
	return nil
}
func hash(value any) string {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(value); err != nil {
		return ""
	}
	sum := sha256.Sum256(bytes.TrimSuffix(b.Bytes(), []byte("\n")))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// InvokeType and ReturnType use the tree protocol's existing domain tags and
// canonical encoding. OutputShape is the verified runtime shape, not a claim
// supplied by the model. Known-answer tests pin these to the worker core.
func InvokeType(c Contract, passed []string) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	for _, id := range passed {
		if !validHash(id, "sha256:", 64) {
			return "", errors.New("invalid passed return type")
		}
	}
	return hash([]any{"skyra/input/v2", c.Binary, c.Shape, append([]string{}, passed...)}), nil
}
func ReturnType(invocation, outputShape string) (string, error) {
	if !validHash(invocation, "sha256:", 64) || outputShape == "" {
		return "", errors.New("a return type needs its invocation and verified output shape")
	}
	return hash([]any{"skyra/return/v1", invocation, outputShape}), nil
}

type Descriptor struct {
	Ref      root.Capability `json:"ref"`
	Target   string          `json:"target"`
	Contract Contract        `json:"contract"`
	// Pin binds routing, uses, and state to the registered program identity.
	Pin       string             `json:"pin"`
	Authority boundary.Authority `json:"authority"`
}

func copyDescriptor(d Descriptor) Descriptor {
	b, _ := json.Marshal(d)
	var out Descriptor
	_ = json.Unmarshal(b, &out)
	return out
}

type Catalogue struct {
	ceiling boundary.Authority
	entries map[root.Capability]Descriptor
}

func NewCatalogue(ceiling boundary.Authority, descriptors []Descriptor) (Catalogue, error) {
	b, _ := json.Marshal(ceiling)
	var held boundary.Authority
	_ = json.Unmarshal(b, &held)
	out := Catalogue{ceiling: held, entries: map[root.Capability]Descriptor{}}
	for _, d := range descriptors {
		d = copyDescriptor(d)
		if d.Target == "" || d.Pin == "" {
			return Catalogue{}, errors.New("capability needs a target and registered program pin")
		}
		if err := d.Contract.Validate(); err != nil {
			return Catalogue{}, err
		}
		if !held.Allows(d.Authority) {
			return Catalogue{}, errors.New("capability scope exceeds the human's authority ceiling")
		}
		expected := root.Capability(hash([]any{"skyra/mission-capability/v1", d.Target, d.Contract, d.Pin, d.Authority}))
		if d.Ref != "" && d.Ref != expected {
			return Catalogue{}, errors.New("capability reference does not match its contract and scope")
		}
		d.Ref = expected
		if _, ok := out.entries[d.Ref]; ok {
			return Catalogue{}, errors.New("duplicate catalogue capability")
		}
		out.entries[d.Ref] = d
	}
	return out, nil
}
func (c Catalogue) Entries() []Descriptor {
	out := []Descriptor{}
	for _, d := range c.entries {
		out = append(out, copyDescriptor(d))
	}
	slices.SortFunc(out, func(a, b Descriptor) int { return strings.Compare(string(a.Ref), string(b.Ref)) })
	return out
}
func (c Catalogue) Lookup(ref root.Capability) (Descriptor, bool) {
	d, ok := c.entries[ref]
	return copyDescriptor(d), ok
}
func (c Catalogue) Ceiling() boundary.Authority {
	b, _ := json.Marshal(c.ceiling)
	var a boundary.Authority
	_ = json.Unmarshal(b, &a)
	return a
}

// Spec is an immutable proposed task and grant. References are informational
// until the human's committed Invoke conveys the held capabilities to Launch.
type Spec struct {
	Version      int               `json:"version"`
	Goal         string            `json:"goal"`
	Inputs       []root.ID         `json:"inputs"`
	Capabilities []root.Capability `json:"capabilities"`
}

func SpecSchema() *contract.Schema {
	str := &contract.Schema{Type: "string"}
	return &contract.Schema{Title: "mission-spec/v1", Type: "object", Required: []string{"version", "goal", "inputs", "capabilities"}, Fields: map[string]*contract.Schema{
		"version": {Type: "number"}, "goal": str, "inputs": {Type: "array", Items: str}, "capabilities": {Type: "array", Items: str},
	}}
}
func endpoint(name string, fields map[string]*contract.Schema, required []string) *contract.Schema {
	return &contract.Schema{Title: name, Type: "object", Required: []string{name}, Fields: map[string]*contract.Schema{name: {Type: "object", Fields: fields, Required: required}}}
}

// ObservationSchema is the adapter's transported observation envelope. Models
// use local names; the adapter supplies the held reference, producer and bytes.
func ObservationSchema() *contract.Schema {
	str := &contract.Schema{Type: "string"}
	return &contract.Schema{Type: "object", Required: []string{"ref", "binary", "contract", "value", "source"}, Fields: map[string]*contract.Schema{"ref": str, "binary": str, "contract": str, "value": {Type: "any"}, "source": {Type: "any"}}}
}
func PlanSchema() *contract.Schema {
	return endpoint("plan", map[string]*contract.Schema{"request": ObservationSchema(), "context": {Type: "array", Items: ObservationSchema()}}, []string{"request", "context"})
}
func LaunchSchema() *contract.Schema {
	return endpoint("launch", map[string]*contract.Schema{"spec": ObservationSchema()}, []string{"spec"})
}
func ProposeSchema() *contract.Schema {
	return endpoint("propose_spec", SpecSchema().Fields, SpecSchema().Required)
}
func strict(raw []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if !json.Valid(raw) {
		return errors.New("expected one JSON value")
	}
	return nil
}
func DecodeSpec(raw json.RawMessage) (Spec, error) {
	v, err := contract.Decode(raw)
	if err != nil {
		return Spec{}, err
	}
	if err = contract.Fits(SpecSchema(), v); err != nil {
		return Spec{}, err
	}
	var spec Spec
	if err = strict(raw, &spec); err != nil {
		return Spec{}, err
	}
	if spec.Version != 1 || strings.TrimSpace(spec.Goal) == "" {
		return Spec{}, errors.New("spec needs version 1 and a goal")
	}
	if len(spec.Goal) > 256*1024 {
		return Spec{}, errors.New("spec goal exceeds 256 KiB")
	}
	for i, id := range spec.Inputs {
		if !validHash(string(id), "", 64) || slices.Contains(spec.Inputs[:i], id) {
			return Spec{}, errors.New("invalid or duplicate input observation reference")
		}
	}
	for i, ref := range spec.Capabilities {
		if !validHash(string(ref), "sha256:", 64) || slices.Contains(spec.Capabilities[:i], ref) {
			return Spec{}, errors.New("invalid or duplicate capability reference")
		}
	}
	spec.Inputs = append([]root.ID{}, spec.Inputs...)
	spec.Capabilities = append([]root.Capability{}, spec.Capabilities...)
	return spec, nil
}

type Resolved struct {
	Spec      Spec
	Contracts []Descriptor
	Inputs    []root.Observation
	Kind      string
	Targets   []string
}

func (c Catalogue) Resolve(spec Spec, state root.State) (Resolved, error) {
	raw, _ := json.Marshal(spec)
	spec, err := DecodeSpec(raw)
	if err != nil {
		return Resolved{}, err
	}
	out := Resolved{Spec: spec, Contracts: []Descriptor{}, Inputs: []root.Observation{}, Targets: []string{}}
	seen := map[Contract]bool{}
	for _, ref := range spec.Capabilities {
		d, ok := c.Lookup(ref)
		if !ok {
			return Resolved{}, errors.New("spec requests a capability absent from the human's catalogue")
		}
		if !root.MayGrant(state.Config().Ceiling, []root.Capability{ref}) {
			return Resolved{}, errors.New("root does not hold the requested capability")
		}
		// Same contract under two scopes is still one callable contract.
		seen[d.Contract] = true
		out.Contracts = append(out.Contracts, d)
		if !slices.Contains(out.Targets, d.Target) {
			out.Targets = append(out.Targets, d.Target)
		}
	}
	for _, id := range spec.Inputs {
		o, ok := state.Lookup(id)
		if !ok {
			return Resolved{}, errors.New("spec cites an observation outside the human's scope")
		}
		out.Inputs = append(out.Inputs, o)
	}
	slices.Sort(out.Targets)
	switch len(seen) {
	case 0:
		return Resolved{}, errors.New("a launch requires at least one callable contract")
	case 1:
		out.Kind = "act"
	default:
		out.Kind = "delegate"
	}
	return out, nil
}

// ValidateHandoff checks the exact observations and grant committed by the
// human root. The spec's value and source are retained, not authored anew.
func (c Catalogue) ValidateHandoff(effect root.Effect, state root.State) (Resolved, error) {
	if effect.Kind != root.Dispatch || effect.Invocation == nil || len(effect.Passed) == 0 {
		return Resolved{}, errors.New("Launch requires a committed root dispatch")
	}
	recorded, ok := state.Invocation(effect.Invocation.ID)
	if !ok || recorded.Target != effect.Invocation.Target || !slices.Equal(recorded.Inputs, effect.Invocation.Inputs) || !slices.Equal(recorded.Grant, effect.Invocation.Grant) {
		return Resolved{}, errors.New("Launch invocation is not committed in this root")
	}
	spec, err := DecodeSpec(effect.Passed[0].Value())
	if err != nil {
		return Resolved{}, err
	}
	resolved, err := c.Resolve(spec, state)
	if err != nil {
		return Resolved{}, err
	}
	expected := append([]root.ID{effect.Passed[0].ID()}, spec.Inputs...)
	if !slices.Equal(effect.Invocation.Inputs, expected) || len(effect.Passed) != len(expected) {
		return Resolved{}, errors.New("Launch did not receive precisely the spec and its cited inputs")
	}
	if !sameSet(effect.Invocation.Grant, spec.Capabilities) {
		return Resolved{}, errors.New("Launch grant differs from the approved spec")
	}
	for i, id := range expected {
		original, ok := state.Lookup(id)
		if !ok || effect.Passed[i].ID() != id || original.Source() != effect.Passed[i].Source() || !bytes.Equal(original.Value(), effect.Passed[i].Value()) {
			return Resolved{}, fmt.Errorf("passed observation %d changed identity or provenance", i)
		}
	}
	return resolved, nil
}
func sameSet(a, b []root.Capability) bool {
	a = slices.Clone(a)
	b = slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}
