// Package humanroot is the pure state machine above a Mission's worker trees.
// A human root is not an act/delegate worker. Only the trusted human adapter
// supplies Authored/Invoke events; only the runtime supplies Returned events.
// It neither authenticates a caller nor runs a child. Those are shell duties.
package humanroot

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
)

// Capabilities are opaque references in the trusted adapter's pinned catalogue.
// A name in JSON grants nothing: only references in Config may be passed. The
// adapter resolves each reference to its exact contract and resource scope and
// still uses the runtime's normal narrowing and OS enforcement at dispatch.
type Capability string
type ID string

type Config struct {
	Mission     string       `json:"mission"`
	Human       string       `json:"human"`
	Entrypoints []Capability `json:"entrypoints"`
	Ceiling     []Capability `json:"ceiling"`
}

type Source struct {
	Human     string     `json:"human,omitempty"`
	RequestID string     `json:"requestId,omitempty"`
	Call      ID         `json:"call,omitempty"`
	Target    Capability `json:"target,omitempty"`
	Receipt   string     `json:"receipt,omitempty"`
	Position  int        `json:"position"`
}

// Observation is immutable outside this package. Its identity binds the
// Mission, source decision and output position, not just the JSON value.
type Observation struct {
	id     ID
	source Source
	value  json.RawMessage
}

func (o Observation) ID() ID                 { return o.id }
func (o Observation) Source() Source         { return o.source }
func (o Observation) Value() json.RawMessage { return slices.Clone(o.value) }
func (o Observation) clone() Observation     { o.value = slices.Clone(o.value); return o }

type Invocation struct {
	ID     ID           `json:"id"`
	Target Capability   `json:"target"`
	Inputs []ID         `json:"inputs"`
	Grant  []Capability `json:"grant"`
}

func (i Invocation) clone() Invocation {
	i.Inputs = slices.Clone(i.Inputs)
	i.Grant = append([]Capability{}, i.Grant...)
	return i
}

type Action struct {
	Kind      string            `json:"kind"` // author | invoke | return
	RequestID string            `json:"requestId,omitempty"`
	Value     json.RawMessage   `json:"value,omitempty"`
	Target    Capability        `json:"target,omitempty"`
	Inputs    []ID              `json:"inputs,omitempty"`
	Grant     []Capability      `json:"grant"` // empty is an empty grant, never ambient authority
	Call      ID                `json:"call,omitempty"`
	Receipt   string            `json:"receipt,omitempty"`
	Outputs   []json.RawMessage `json:"outputs,omitempty"`
}

type Record struct {
	Mission  string `json:"mission"`
	Previous string `json:"previous"`
	Sequence uint64 `json:"sequence"`
	Action   Action `json:"action"`
}

func (r Record) clone() Record {
	r.Action.Value = slices.Clone(r.Action.Value)
	r.Action.Inputs = slices.Clone(r.Action.Inputs)
	r.Action.Grant = slices.Clone(r.Action.Grant)
	r.Action.Outputs = slices.Clone(r.Action.Outputs)
	for i, b := range r.Action.Outputs {
		r.Action.Outputs[i] = slices.Clone(b)
	}
	return r
}
func digest(domain string, v any) string {
	b, err := json.Marshal([]any{domain, v})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func (r Record) Stamp() string { return digest("skyra/human-root/record/v1", r) }

// State is immutable at the API boundary. Observation scope, invocation
// receipts and request deduplication are derived by replay, never supplied by
// a model or restored from a serialized mutable State.
type State struct {
	config   Config
	head     string
	sequence uint64
	scope    []Observation
	calls    map[ID]Invocation
	returned map[ID]bool
	requests map[string]completed
	pending  *Record
}
type completed struct {
	fingerprint string
	ids         []ID
	call        ID
}

func New(c Config) (State, error) {
	if c.Mission == "" || c.Human == "" {
		return State{}, errors.New("a root needs a Mission and human principal")
	}
	for _, caps := range [][]Capability{c.Entrypoints, c.Ceiling} {
		for i, cap := range caps {
			if cap == "" || slices.Contains(caps[:i], cap) {
				return State{}, errors.New("empty or duplicate capability")
			}
		}
	}
	c.Entrypoints = append([]Capability{}, c.Entrypoints...)
	c.Ceiling = append([]Capability{}, c.Ceiling...)
	// Capability-set order must not change the pinned root configuration.
	slices.Sort(c.Entrypoints)
	slices.Sort(c.Ceiling)
	return State{config: c, head: digest("skyra/human-root/config/v1", c), calls: map[ID]Invocation{}, returned: map[ID]bool{}, requests: map[string]completed{}}, nil
}
func (s State) Head() string     { return s.head }
func (s State) Sequence() uint64 { return s.sequence }
func (s State) Awaiting() bool   { return s.pending != nil }
func (s State) Config() Config {
	c := s.config
	c.Entrypoints = slices.Clone(c.Entrypoints)
	c.Ceiling = slices.Clone(c.Ceiling)
	return c
}
func (s State) Scope() []Observation {
	out := make([]Observation, len(s.scope))
	for i, o := range s.scope {
		out[i] = o.clone()
	}
	return out
}
func (s State) Lookup(id ID) (Observation, bool) {
	for _, o := range s.scope {
		if o.id == id {
			return o.clone(), true
		}
	}
	return Observation{}, false
}
func (s State) Invocation(id ID) (Invocation, bool) { i, ok := s.calls[id]; return i.clone(), ok }
func (s State) PendingCalls() []Invocation {
	out := []Invocation{}
	for id, i := range s.calls {
		if !s.returned[id] {
			out = append(out, i.clone())
		}
	}
	slices.SortFunc(out, func(a, b Invocation) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return out
}

// These are the decisions mirrored by Invariants/HumanRoot.lean.
func MayPass(scope, inputs []ID) bool {
	for _, id := range inputs {
		if !slices.Contains(scope, id) {
			return false
		}
	}
	return true
}
func MayGrant(ceiling, grant []Capability) bool {
	for _, c := range grant {
		if !slices.Contains(ceiling, c) {
			return false
		}
	}
	return true
}
func MayInvoke(entries []Capability, target Capability, scope, inputs []ID, ceiling, grant []Capability) bool {
	return slices.Contains(entries, target) && MayPass(scope, inputs) && MayGrant(ceiling, grant)
}

// Event constructors are for trusted adapters. The worker instruction parser
// does not import this package or expose any of these operations to a model.
// Ingress is assigned by the trusted transport, never parsed from a payload.
// A worker can send JSON naming a human, but still arrives on WorkerIngress.
type Ingress uint8

const (
	HumanIngress Ingress = iota + 1
	RuntimeIngress
	StorageIngress
	WorkerIngress
)

func MayAuthor(from Ingress) bool { return from == HumanIngress }

type Event interface{ event() }
type Authored struct {
	RequestID string
	Value     json.RawMessage
}
type Invoke struct {
	RequestID string
	Target    Capability
	Inputs    []ID
	Grant     []Capability
}
type Returned struct {
	Call    ID
	Receipt string
	Outputs []json.RawMessage
}
type Committed struct{ Stamp string }

func (Authored) event()  {}
func (Invoke) event()    {}
func (Returned) event()  {}
func (Committed) event() {}

type EffectKind string

const (
	Commit    EffectKind = "commit"
	Accepted  EffectKind = "accepted"
	Dispatch  EffectKind = "dispatch"
	Duplicate EffectKind = "duplicate"
	Refused   EffectKind = "refused"
)

type Effect struct {
	Call         ID
	Kind         EffectKind
	Record       *Record
	Stamp        string
	Observations []ID
	Invocation   *Invocation
	// Passed contains exactly Invocation.Inputs, in order, retaining identity
	// and origin. It never contains the rest of the root's scope.
	Passed []Observation
	Reason string
}

func refuse(s State, why string) (State, Effect) { return s, Effect{Kind: Refused, Reason: why} }

// canonical accepts one complete JSON value and retains number precision.
// Values are data; nested fields claiming an author or grant have no authority.
func canonical(raw json.RawMessage) (json.RawMessage, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err != nil {
		return nil, err
	}
	if !json.Valid(raw) {
		return nil, errors.New("expected one JSON value")
	}
	b, err := json.Marshal(v)
	return b, err
}
func requestKey(a Action) string {
	if a.Kind == "return" {
		return "return/" + string(a.Call)
	}
	return "human/" + a.RequestID
}

// Step plans one durable record. Only an acknowledgement of that exact record
// applies the decision or releases a Dispatch. No I/O, time, randomness, child
// execution, model inference or changes to the worker tree live here.
func Step(s State, from Ingress, e Event) (State, Effect) {
	if s.config.Mission == "" {
		return refuse(s, "root is not configured")
	}
	if ack, ok := e.(Committed); ok {
		if from != StorageIngress {
			return refuse(s, "commit acknowledgements require storage ingress")
		}
		if s.pending == nil || ack.Stamp != s.pending.Stamp() {
			return refuse(s, "commit acknowledgement does not match the pending record")
		}
		return apply(s, *s.pending)
	}
	if s.pending != nil {
		return refuse(s, "a record is awaiting durable commit")
	}
	var a Action
	switch e := e.(type) {
	case Authored:
		if !MayAuthor(from) {
			return refuse(s, "only human ingress may author at the root")
		}
		v, err := canonical(e.Value)
		if err != nil {
			return refuse(s, "invalid observation JSON")
		}
		a = Action{Kind: "author", RequestID: e.RequestID, Value: v, Grant: []Capability{}}
	case Invoke:
		if from != HumanIngress {
			return refuse(s, "only human ingress may invoke a root child")
		}
		a = Action{Kind: "invoke", RequestID: e.RequestID, Target: e.Target, Inputs: slices.Clone(e.Inputs), Grant: append([]Capability{}, e.Grant...)}
	case Returned:
		if from != RuntimeIngress {
			return refuse(s, "child returns require verified runtime ingress")
		}
		outputs := make([]json.RawMessage, len(e.Outputs))
		for i, b := range e.Outputs {
			v, err := canonical(b)
			if err != nil {
				return refuse(s, "invalid returned observation JSON")
			}
			outputs[i] = v
		}
		a = Action{Kind: "return", Call: e.Call, Receipt: e.Receipt, Outputs: outputs, Grant: []Capability{}}
	default:
		return refuse(s, "unsupported root event")
	}
	key := requestKey(a)
	fingerprint := digest("skyra/human-root/action/v1", a)
	if old, ok := s.requests[key]; ok {
		if old.fingerprint != fingerprint {
			return refuse(s, "request identity already belongs to different work")
		}
		return s, Effect{Kind: Duplicate, Call: old.call, Observations: slices.Clone(old.ids)} // never dispatch a committed call twice
	}
	switch a.Kind {
	case "author":
		if a.RequestID == "" {
			return refuse(s, "a human observation needs a request identity")
		}
	case "invoke":
		if a.RequestID == "" {
			return refuse(s, "an invocation needs a human request identity")
		}
		scope := make([]ID, len(s.scope))
		for i, o := range s.scope {
			scope[i] = o.id
		}
		if !MayInvoke(s.config.Entrypoints, a.Target, scope, a.Inputs, s.config.Ceiling, a.Grant) {
			return refuse(s, "invocation exceeds root scope or authority")
		}
	case "return":
		if _, ok := s.calls[a.Call]; !ok || s.returned[a.Call] {
			return refuse(s, "no pending child invocation")
		}
		if a.Receipt == "" {
			return refuse(s, "a child return needs a verified runtime receipt")
		}
	}
	r := Record{Mission: s.config.Mission, Previous: s.head, Sequence: s.sequence + 1, Action: a}
	s.pending = &r
	copy := r.clone()
	return s, Effect{Kind: Commit, Record: &copy, Stamp: r.Stamp()}
}

func apply(s State, r Record) (State, Effect) {
	// Copy every mutable collection. Older states and effects remain snapshots.
	s.scope = append([]Observation{}, s.scope...)
	calls := make(map[ID]Invocation, len(s.calls))
	for id, c := range s.calls {
		calls[id] = c
	}
	s.calls = calls
	returned := make(map[ID]bool, len(s.returned))
	for id, b := range s.returned {
		returned[id] = b
	}
	s.returned = returned
	requests := make(map[string]completed, len(s.requests))
	for k, v := range s.requests {
		requests[k] = v
	}
	s.requests = requests
	s.head, s.sequence, s.pending = r.Stamp(), r.Sequence, nil
	effect := Effect{Kind: Accepted, Stamp: s.head}
	a := r.Action
	remember := completed{fingerprint: digest("skyra/human-root/action/v1", a)}
	add := func(value json.RawMessage, source Source) {
		id := ID(digest("skyra/human-root/observation/v1", []any{s.head, source.Position}))
		s.scope = append(s.scope, Observation{id, source, slices.Clone(value)})
		effect.Observations = append(effect.Observations, id)
	}
	switch a.Kind {
	case "author":
		add(a.Value, Source{Human: s.config.Human, RequestID: a.RequestID})
	case "invoke":
		call := Invocation{ID: ID(s.head), Target: a.Target, Inputs: slices.Clone(a.Inputs), Grant: append([]Capability{}, a.Grant...)}
		s.calls[call.ID] = call
		remember.call = call.ID
		out := call.clone()
		effect.Kind = Dispatch
		effect.Call = call.ID
		effect.Invocation = &out
		for _, id := range call.Inputs {
			o, _ := s.Lookup(id)
			effect.Passed = append(effect.Passed, o)
		}
	case "return":
		call := s.calls[a.Call]
		for i, value := range a.Outputs {
			add(value, Source{Call: a.Call, Target: call.Target, Receipt: a.Receipt, Position: i})
		}
		s.returned[a.Call] = true
	}
	remember.ids = slices.Clone(effect.Observations)
	s.requests[requestKey(a)] = remember
	return s, effect
}

// Replay checks each persisted decision by asking Step to derive it again.
// It consumes the acknowledgements but intentionally discards Dispatch effects:
// an unknown execution outcome is not permission to repeat the execution.
func Replay(c Config, records []Record) (State, error) {
	s, err := New(c)
	if err != nil {
		return State{}, err
	}
	for _, record := range records {
		a := record.Action
		var event Event
		switch a.Kind {
		case "author":
			event = Authored{a.RequestID, a.Value}
		case "invoke":
			event = Invoke{a.RequestID, a.Target, a.Inputs, a.Grant}
		case "return":
			event = Returned{a.Call, a.Receipt, a.Outputs}
		default:
			return State{}, errors.New("unknown root record kind")
		}
		from := HumanIngress
		if a.Kind == "return" {
			from = RuntimeIngress
		}
		next, effect := Step(s, from, event)
		if effect.Kind != Commit || effect.Record.Stamp() != record.Stamp() {
			return State{}, errors.New("root record does not derive from its inputs")
		}
		s, effect = Step(next, StorageIngress, Committed{record.Stamp()})
		if effect.Kind == Refused {
			return State{}, errors.New("root record acknowledgement refused")
		}
	}
	return s, nil
}
