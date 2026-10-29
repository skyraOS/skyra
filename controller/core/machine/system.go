package machine

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	"skyra/controller/core/record"
)

// The Boundary as an actor (next/32, decisions 16, 21). Every Boundary holds the
// name system, and its contracts are the syscalls: a syscall is an ordinary
// emission to it, as a Unix process calls the kernel without knowing which
// container it runs in. The runtime answers them itself, because what they
// do (create a Boundary, read its history, add a contract to it) changes or
// reads Boundary state, which only the kernel holds. No registry may name
// system, or anything under system.
//
//	{"to": "system", "payload": {"narrow": {"actor": "<name in reach>", "reach": [...], "authority": {...}, "without": [...]}}}
//	   → returns {"reference": "<ref>"}: a narrowed Boundary for that actor, now in the caller's reach.
//	     "without" drops whole kinds of authority (read, write, mach, connect,
//	     scratch) from what is held, so a procedure can narrow without naming paths
//	{"to": "<ref>", "payload": <the actor's input>}
//	   → an ordinary invoke of the actor, under the narrowed Boundary
//	{"to": "system", "payload": {"extend-reach": {"schema": <a contract>}}}
//	   → returns {"ok": true}: the emitter's own reach entry holds one more contract,
//	     throughout its Boundary
//	{"to": "system", "payload": {"trace": {"observation": <stamp>, "contract": {"target": <name>, "schema": <a contract>}}}}
//	   → returns the trace: the stamps from the observation up to the invoke made
//	     through that contract, with every record of the work beneath it (H4)
//
// A narrowing may only drop what is held. "reach" and "authority" are what
// the actor keeps; left out, it keeps all of the caller's. An extension adds a
// contract, never authority: the actor is the emitter, derived, never named,
// so an actor can only extend itself (decision 17).
const (
	System       = "system"
	SystemPrefix = "system."

	Narrowing = "narrow"
	Extending = "extend-reach"
	Tracing   = "trace"
)

// maxDepth bounds the frame stack a narrowing can grow from, whatever the
// actor: a scope that keeps creating scopes is refused, not left to recurse.
const maxDepth = 64

var (
	str      = &contract.Schema{Type: "string"}
	strs     = &contract.Schema{Type: "array", Items: str}
	grants   = &contract.Schema{Type: "array", Items: &contract.Schema{Type: "object", Required: []string{"path", "scope"}, Fields: map[string]*contract.Schema{"path": str, "scope": str}}}
	narrowIn = &contract.Schema{Type: "object", Required: []string{"actor"}, Fields: map[string]*contract.Schema{
		"actor": str,
		"reach": strs,
		"tools": strs,
		"authority": {Type: "object", Fields: map[string]*contract.Schema{
			"read": grants, "write": grants, "mach": strs, "connect": strs, "scratch": str,
		}},
		"without": strs,
	}}
	extendIn = &contract.Schema{Type: "object", Required: []string{"schema"}, Fields: map[string]*contract.Schema{
		"schema": {Type: "any"},
	}}
	traceIn = &contract.Schema{Type: "object", Required: []string{"contract"}, Fields: map[string]*contract.Schema{
		"observation":  str,
		"invocation":   str,
		"observations": strs,
		"contract": {Type: "object", Required: []string{"target", "schema"}, Fields: map[string]*contract.Schema{
			"target": str, "schema": {Type: "any"},
		}},
	}}
)

// call is a syscall's contract: an object naming exactly that syscall.
func call(name string, in *contract.Schema) *contract.Schema {
	return &contract.Schema{Title: name, Type: "object", Required: []string{name}, Fields: map[string]*contract.Schema{name: in}}
}

// Syscalls is the surface every Boundary holds under the name system: no
// binary, and one contract per syscall.
func Syscalls() boundary.Surface {
	return boundary.Surface{System: "boundary", Contracts: []*contract.Schema{
		call(Narrowing, narrowIn),
		call(Extending, extendIn),
		call(Tracing, traceIn),
	}}
}

// IsSystem reports whether a name is the runtime's: system, a syscall's
// reference, or anything else under system.
func IsSystem(name string) bool { return name == System || strings.HasPrefix(name, SystemPrefix) }

// syscall is the syscall a payload to system names, and its arguments.
func syscall(payload json.RawMessage) (string, json.RawMessage) {
	var o map[string]json.RawMessage
	if json.Unmarshal(payload, &o) != nil || len(o) != 1 {
		return "", nil
	}
	for name, args := range o {
		return name, args
	}
	return "", nil
}

// changesWaiter reports whether an invoke changes the Boundary of the caller
// waiting on it: narrow records a reference there, extend-reach a contract.
func changesWaiter(r record.Record, s boundary.Surface) bool {
	if s.System != "boundary" {
		return false
	}
	name, _ := syscall(r.Payload)
	return name == Narrowing || name == Extending
}

// checkSyscall checks a syscall's arguments against the Boundary it was
// emitted under, at intake: a narrowing that asks for anything not held, or
// an extension that is not the emitter's to make, is a fault, of this kind.
func checkSyscall(h boundary.Boundary, payload json.RawMessage) (string, error) {
	switch name, args := syscall(payload); name {
	case Narrowing:
		if _, err := narrowing(h, args); err != nil {
			return record.Boundary, err
		}
	case Extending:
		return extending(h, args)
	case Tracing:
		if _, err := tracing(h, args); err != nil {
			return record.Boundary, err
		}
	}
	return "", nil
}

// tracing checks a trace payload under h at intake: the observation is a
// stamp, and the contract it names is one the requester holds. Whether the
// observation was requested through it is the walk's to find (ReadTrace).
func tracing(h boundary.Boundary, payload json.RawMessage) (TraceKey, error) {
	var ask struct {
		Observation  string   `json:"observation"`
		Invocation   string   `json:"invocation"`
		Observations []string `json:"observations"`
		Contract     struct {
			Target string          `json:"target"`
			Schema json.RawMessage `json:"schema"`
		} `json:"contract"`
	}
	if err := json.Unmarshal(payload, &ask); err != nil {
		return TraceKey{}, err
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(payload, &fields)
	_, single := fields["observation"]
	_, root := fields["invocation"]
	_, leaves := fields["observations"]
	if single == (root || leaves) || !single && (!root || !leaves || len(ask.Observations) == 0) {
		return TraceKey{}, fmt.Errorf("trace requires exactly observation, or invocation and nonempty observations")
	}
	if single && !isStamp(ask.Observation) {
		return TraceKey{}, fmt.Errorf("observation %q is not a stamp", ask.Observation)
	}
	if !single {
		if !isStamp(ask.Invocation) {
			return TraceKey{}, fmt.Errorf("invocation %q is not a stamp", ask.Invocation)
		}
		for _, leaf := range ask.Observations {
			if !isStamp(leaf) {
				return TraceKey{}, fmt.Errorf("observation %q is not a stamp", leaf)
			}
		}
	}
	target, ok := h.Reach[ask.Contract.Target]
	if !ok || IsSystem(ask.Contract.Target) {
		return TraceKey{}, fmt.Errorf("%q is not an actor in reach", ask.Contract.Target)
	}
	named, err := schemaOf(ask.Contract.Schema)
	if err != nil {
		return TraceKey{}, err
	}
	for _, c := range target.All() {
		if contract.Same(c, named) {
			read := TraceKey{Observation: ask.Observation, Invocation: ask.Invocation, Observations: slices.Clone(ask.Observations), Target: ask.Contract.Target, Schema: c, Origin: h.Origin}
			if !single {
				outer := h
				for outer.Resume != nil {
					outer = *outer.Resume
				}
				read.Genesis = outer.Frame
			}
			return read, nil
		}
	}
	return TraceKey{}, fmt.Errorf("%s holds no such contract", ask.Contract.Target)
}

func isStamp(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

// narrowing derives the Grant a narrow payload asks for under h, or says why
// it reaches past the Boundary. It is checked at intake, so a narrowing that
// asks for anything not held is a boundary fault.
func narrowing(h boundary.Boundary, payload json.RawMessage) (boundary.Grant, error) {
	var ask struct {
		Actor     string              `json:"actor"`
		Reach     *[]string           `json:"reach"`
		Authority *boundary.Authority `json:"authority"`
		Without   []string            `json:"without"`
		Tools     *[]string           `json:"tools"`
	}
	if err := json.Unmarshal(payload, &ask); err != nil {
		return boundary.Grant{}, err
	}
	if _, ok := h.Reach[ask.Actor]; !ok || IsSystem(ask.Actor) {
		return boundary.Grant{}, fmt.Errorf("%q is not an actor in reach", ask.Actor)
	}
	g := boundary.Grant{Actor: ask.Actor, Authority: h.Authority, WorkTools: h.WorkTools}
	if h.Depth() > maxDepth {
		return g, fmt.Errorf("narrowing exceeds the runtime frame limit of %d", maxDepth)
	}
	if ask.Tools != nil {
		allowed := map[string]boundary.ToolBinding{}
		available := h.AvailableTools()
		for _, name := range *ask.Tools {
			binding, ok := boundary.ResolveTool(available, name)
			if !ok {
				return g, fmt.Errorf("tool %q is not available", name)
			}
			allowed[name] = binding
		}
		g.WorkTools = &allowed
	}
	if ask.Reach == nil {
		for name := range h.Reach {
			g.Reach = append(g.Reach, name)
		}
	} else {
		for _, name := range *ask.Reach {
			if _, ok := h.Reach[name]; !ok {
				return boundary.Grant{}, fmt.Errorf("narrowing keeps %q, which is not in reach", name)
			}
			if !slices.Contains(g.Reach, name) {
				g.Reach = append(g.Reach, name)
			}
		}
	}
	slices.Sort(g.Reach)
	if ask.Authority != nil {
		if !h.Authority.Allows(*ask.Authority) {
			return boundary.Grant{}, fmt.Errorf("narrowing asks for authority outside the Boundary")
		}
		g.Authority = *ask.Authority
	}
	for _, kind := range ask.Without {
		switch kind {
		case "read":
			g.Authority.Read = nil
		case "write":
			g.Authority.Write = nil
		case "mach":
			g.Authority.Mach = nil
		case "connect":
			g.Authority.Connect = nil
		case "scratch":
			g.Authority.Scratch = ""
		default:
			return boundary.Grant{}, fmt.Errorf("narrowing drops %q, which is not a kind of authority", kind)
		}
	}
	if !g.Authority.Covers() {
		g.Authority.Workspace = "" // its read grant was narrowed away
	}
	return g, nil
}

// extending checks an extend-reach payload under h, the emitter's Boundary
// (decision 17): the actor is the emitter, the name its invoke resolved to;
// the schema is one the kernel can check; and (actor, schema) is not
// already there. It returns the fault kind for what fails.
func extending(h boundary.Boundary, payload json.RawMessage) (string, error) {
	_, err := extension(h, payload)
	switch {
	case err == nil:
		return "", nil
	case errors.Is(err, errNotTheEmitters):
		return record.Boundary, err
	}
	return record.Contract, err
}

var errNotTheEmitters = errors.New("not the emitter's to extend")

// extension derives the contract an extend-reach adds, under h.
func extension(h boundary.Boundary, payload json.RawMessage) (*contract.Schema, error) {
	var ask struct {
		Schema json.RawMessage `json:"schema"`
	}
	if err := json.Unmarshal(payload, &ask); err != nil {
		return nil, err
	}
	s, err := schemaOf(ask.Schema)
	if err != nil {
		return nil, err
	}
	switch {
	case h.Surface == nil || h.Name == "":
		return nil, fmt.Errorf("%w: only an actor extends itself", errNotTheEmitters)
	case h.Forked:
		// Its siblings hold this Boundary too, without the extension: whoever
		// resolves the round last would decide whether the waiter has it.
		return nil, fmt.Errorf("%w: one of several siblings cannot extend the Boundary they share", errNotTheEmitters)
	}
	actor, ok := h.Reach[h.Name]
	if !ok {
		return nil, fmt.Errorf("%w: %q is not in its own Boundary's reach", errNotTheEmitters, h.Name)
	}
	for _, c := range actor.All() {
		if contract.Same(c, s) {
			return nil, fmt.Errorf("%s already holds this contract", h.Name)
		}
	}
	return s, nil
}

// schemaOf reads a contract a syscall names: strictly, and well formed. Two
// contracts are the same when they read to the same schema.
func schemaOf(b json.RawMessage) (*contract.Schema, error) {
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.DisallowUnknownFields()
	var s contract.Schema
	if err := d.Decode(&s); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	if err := s.Check(); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	return &s, nil
}

// extended is a Boundary and every frame below it that shares its Boundary
// (the same origin), with one more contract on actor's reach entry. Maps are
// copied, never written in place: other frames share them.
func extended(h boundary.Boundary, origin, actor string, s *contract.Schema) boundary.Boundary {
	if h.Origin == origin {
		if entry, ok := h.Reach[actor]; ok {
			reach := make(map[string]boundary.Surface, len(h.Reach))
			for name, surface := range h.Reach {
				reach[name] = surface
			}
			entry.Contracts = append(slices.Clone(entry.Contracts), s)
			reach[actor] = entry
			h.Reach = reach
		}
	}
	if h.Resume != nil {
		below := extended(*h.Resume, origin, actor, s)
		h.Resume = &below
	}
	return h
}

// runFacet is what a syscall emits when it runs: nothing executes, the result
// is derived. narrow returns the reference: the name the waiting caller now
// holds it under, meaningful only in the reach it was recorded into.
func runFacet(s Instance) []byte {
	var result any = map[string]bool{"ok": true}
	if name, _ := syscall(s.R.Payload); name == Narrowing {
		result = map[string]string{"reference": refName(s.Self.Refs)}
	}
	out, _ := json.Marshal(map[string]any{"to": s.Self.Frame, "payload": result})
	return out
}

// refName names the next reference a Boundary holding refs receives. It is a
// counter, not the invoke's stamp, because a continuation is shown its
// references and hashes never reach a model (next/30, F1). It is in the
// runtime's namespace, which no registry may name, so it never shadows a
// surface.
func refName(refs map[string]boundary.Grant) string {
	for n := len(refs) + 1; ; n++ {
		name := fmt.Sprintf("%sref.%d", SystemPrefix, n)
		if _, held := refs[name]; !held {
			return name
		}
	}
}

// kept is a reach restricted to the names a grant keeps. Every Boundary holds
// system: its syscalls are the Boundary's own (decision 16), and none of them
// can add authority.
func kept(reach map[string]boundary.Surface, names []string) map[string]boundary.Surface {
	out := map[string]boundary.Surface{}
	if s, ok := reach[System]; ok {
		out[System] = s
	}
	for _, name := range names {
		if s, ok := reach[name]; ok {
			out[name] = s
		}
	}
	return out
}
