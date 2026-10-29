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

// The syscalls. An emission to system is answered by the kernel itself: it
// is in no Boundary's reach and held by none, so no registry, narrowing or
// c-list can grant or drop it. Nothing runs; the result is derived.
//
//	{"to": "system", "payload": {"narrow": {"actor": "<name in reach>", "reach": [...], "contracts": [{"target": <name>, "schema": <a contract>}, ...], "authority": {...}, "without": [...]}}}
//	   → returns {"handle": "<sha256>"}: the kernel provisions a narrowed Boundary for that
//	     actor, stores it under its hash, and hands the hash back. That is all it keeps:
//	     no Boundary records the handle, so the caller keeps its own table of them.
//	     "contracts" is its c-list: the contracts it holds.
//	     "without" drops whole kinds of authority (read, write, mach, connect,
//	     scratch) from what is held, so a procedure can narrow without naming paths
//	{"to": "<handle>", "payload": <the actor's input>}
//	   → an ordinary invoke of the actor, under the provisioned Boundary. Any
//	     Boundary that could provision the same grant itself may invoke it: a
//	     handle never widens
//	{"to": "system", "payload": {"export": {"schema": <a contract>}}}
//	   → returns {"ok": true}: the emitter's own reach entry declares one more contract,
//	     throughout its Boundary, which is another Boundary (a new Hash). A
//	     Boundary narrowed by contract does not hold it: an export never widens
//	{"to": "system", "payload": {"trace": {"observation": <stamp>, "contract": {"target": <name>, "schema": <a contract>}}}}
//	   → returns the trace: the stamps from the observation up to the invoke made
//	     through that contract, with every record of the work beneath it (H4)
//
// A narrowing may only drop what is held. "reach", "contracts" and "authority"
// are what the actor keeps; left out, it keeps all of the caller's. An export adds a
// contract, never authority: the actor is the emitter, derived, never named,
// so an actor can only extend itself (decision 17).
const (
	System       = "system"
	SystemPrefix = "system."

	Narrowing = "narrow"
	Exporting = "export"
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
		"contracts": {Type: "array", Items: &contract.Schema{Type: "object", Required: []string{"target", "schema"}, Fields: map[string]*contract.Schema{
			"target": str, "schema": {Type: "any"},
		}}},
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

// Syscalls is every syscall's contract: what a payload to system must fit.
func Syscalls() []*contract.Schema {
	return []*contract.Schema{call(Narrowing, narrowIn), call(Exporting, extendIn), call(Tracing, traceIn)}
}

// IsSystem reports whether a name is the runtime's: system, or anything else
// under system. No registry may name one.
func IsSystem(name string) bool { return name == System || strings.HasPrefix(name, SystemPrefix) }

// isSyscall reports whether h is a syscall's frame: it runs nothing.
func isSyscall(h boundary.Boundary) bool { return h.Surface == nil && h.Name == System }

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

// invokeSystem is intake's derivation for an emission to system: the payload
// fits exactly one syscall, and its arguments ask for nothing h does not hold.
func invokeSystem(h boundary.Boundary, payload any, attempt json.RawMessage) record.Record {
	if _, err := selected(Syscalls(), payload); err != nil {
		return faulted(record.Contract, System+": "+err.Error(), attempt)
	}
	if kind, err := checkSyscall(h, contract.Canonical(payload)); err != nil {
		return faulted(kind, System+": "+err.Error(), attempt)
	}
	return record.Record{State: record.Invoke, To: System, Payload: contract.Canonical(payload)}
}

// changesWaiter reports whether an invoke changes the Boundary of the caller
// waiting on it: export adds a contract there. A narrowing changes nothing:
// what it provisions is stored under its handle, not in the caller.
func changesWaiter(r record.Record) bool {
	if r.State != record.Invoke || r.To != System {
		return false
	}
	name, _ := syscall(r.Payload)
	return name == Exporting
}

// provision is the grant an invoke of narrow stores, if r is one.
func provision(h boundary.Boundary, r record.Record) (Object, bool) {
	if r.State != record.Invoke || r.To != System {
		return Object{}, false
	}
	name, args := syscall(r.Payload)
	if name != Narrowing {
		return Object{}, false
	}
	g, err := narrowing(h, args)
	if err != nil {
		return Object{}, false
	}
	return HeldGrant(g), true
}

// HandleOf is the handle an emission under h addresses, or "": an address
// shaped as a hash that is neither a name in reach nor h's own slot. Intake
// reads the grant stored under it before it derives the record.
func HandleOf(h boundary.Boundary, out []byte) string {
	v, err := contract.Decode(out)
	o, _ := v.(map[string]any)
	to, _ := o["to"].(string)
	if _, named := h.Reach[to]; err != nil || named || to == h.Frame || !isStamp(to) {
		return ""
	}
	return to
}

// reprovisions checks that h could provision g itself: narrowing h to
// everything g names derives g again, pins included. So a handle is
// authority only under a Boundary that holds at least what it grants, and the
// kernel needs no record of who was handed it.
func reprovisions(h boundary.Boundary, g boundary.Grant) error {
	reach := g.Reach
	if reach == nil {
		reach = []string{}
	}
	ask := map[string]any{"actor": g.Actor, "reach": reach, "authority": g.Authority}
	if g.Held != nil {
		kept := []map[string]any{}
		for name, held := range *g.Held {
			for _, c := range held {
				kept = append(kept, map[string]any{"target": name, "schema": c})
			}
		}
		ask["contracts"] = kept
	}
	args, _ := json.Marshal(ask)
	again, err := narrowing(h, args)
	if err != nil {
		return err
	}
	if HeldGrant(again).Key != HeldGrant(g).Key {
		return errors.New("it grants what this Boundary does not hold")
	}
	return nil
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
	case Exporting:
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
		Contracts *[]struct {
			Target string          `json:"target"`
			Schema json.RawMessage `json:"schema"`
		} `json:"contracts"`
	}
	if err := json.Unmarshal(payload, &ask); err != nil {
		return boundary.Grant{}, err
	}
	if _, ok := h.Reach[ask.Actor]; !ok || IsSystem(ask.Actor) {
		return boundary.Grant{}, fmt.Errorf("%q is not an actor in reach", ask.Actor)
	}
	g := boundary.Grant{Actor: ask.Actor, Authority: h.Authority, Held: h.Held}
	if h.Depth() > maxDepth {
		return g, fmt.Errorf("narrowing exceeds the runtime frame limit of %d", maxDepth)
	}
	if ask.Contracts != nil {
		// What the actor keeps, by contract: each one held here. An empty
		// list keeps no other contract; left out, the caller's are kept.
		held := boundary.CList{}
		for _, kept := range *ask.Contracts {
			s, err := schemaOf(kept.Schema)
			if err != nil {
				return g, err
			}
			if IsSystem(kept.Target) || !h.Holds(kept.Target, s) {
				return g, fmt.Errorf("narrowing keeps a contract of %q that is not held", kept.Target)
			}
			if !slices.ContainsFunc(held[kept.Target], func(c *contract.Schema) bool { return contract.Same(c, s) }) {
				held[kept.Target] = append(held[kept.Target], s)
			}
		}
		g.Held = &held
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
	// The actor keeps what it runs through, so a child can run; held by
	// contract, it also keeps its own contracts, so it can be narrowed to in
	// turn if it is in reach. Only what the caller holds is kept.
	uses := h.Reach[ask.Actor].Uses
	if ask.Reach != nil {
		for _, name := range uses {
			if _, ok := h.Reach[name]; ok && !slices.Contains(g.Reach, name) {
				g.Reach = append(g.Reach, name)
			}
		}
	}
	if ask.Contracts != nil {
		for _, name := range append([]string{ask.Actor}, uses...) {
			if held := h.HeldOf(name); len(held) > 0 && slices.Contains(g.Reach, name) {
				(*g.Held)[name] = held
			}
		}
	}
	slices.Sort(g.Reach)
	if ask.Contracts != nil {
		for name := range *g.Held {
			if !slices.Contains(g.Reach, name) {
				return boundary.Grant{}, fmt.Errorf("narrowing holds a contract of %q but does not keep it in reach", name)
			}
		}
	}
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
	g.Pins = map[string]string{ask.Actor: boundary.Identity(h.Reach[ask.Actor])}
	for _, name := range g.Reach {
		g.Pins[name] = boundary.Identity(h.Reach[name])
	}
	return g, nil
}

// extending checks an export payload under h, the emitter's Boundary
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

// extension derives the contract an export adds, under h.
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

// syscallResult is what a syscall emits when it runs: nothing executes, the
// result is derived. narrow returns the handle of the grant intake stored,
// derived again from the caller's Boundary.
func syscallResult(s Instance) []byte {
	var result any = map[string]bool{"ok": true}
	if name, args := syscall(s.R.Payload); name == Narrowing {
		g, _ := narrowing(*s.Self.Resume, args)
		result = map[string]string{"handle": HeldGrant(g).Key}
	}
	out, _ := json.Marshal(map[string]any{"to": s.Self.Frame, "payload": result})
	return out
}

// kept is a reach restricted to the names a grant keeps.
func kept(reach map[string]boundary.Surface, names []string) map[string]boundary.Surface {
	out := map[string]boundary.Surface{}
	for _, name := range names {
		if s, ok := reach[name]; ok {
			out[name] = s
		}
	}
	return out
}
