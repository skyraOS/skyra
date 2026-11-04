package tree

import (
	"bytes"
	"encoding/json"
	"slices"
)

// Step is the tree machine (proposal 56): it interprets an event and decides
// what happens next, as effects for the shell to perform. It is pure: the
// tree it is handed is never changed. Every decision it makes is one of
// core.go's, which is Core.lean's.
//
// This is step 4a: the tree. The store and the history are held in memory;
// reading and verifying them, and the write-once commits, are step 4b. With
// remember off, a guard's answer is recorded and the agent always goes back
// to its model.

// MaxFanOut is the most instructions one call may carry, as the kernel's
// fan-out.
const MaxFanOut = 32

// The output shapes of what the apprentice's own contracts make: text, or a
// list (the packaged reach and the delegate computation). baseOf keeps them
// apart by contract, so they need not differ.
var (
	shapeText = Schema(`{"type":"string"}`)
	shapeList = Schema(`{"type":"array"}`)
)

// The descriptions create_delegation and create_act fill in.
const (
	delegationText = "Compose and Delegate the following goal: "
	actText        = "Act on the following goal: "
)

// wrapperShape is wrapper<T> for a kind, the shape every return comes back in:
// { <kind>: [T] }.
func wrapperShape(k Kind) Schema {
	return Schema(`{"properties":{"` + k.String() + `":{"type":"array"}},"required":["` + k.String() + `"],"type":"object"}`)
}

// --- state ---

// A Tree is the whole tree, held in one machine, so an Observation is only
// ever made by Observe.
type Tree struct {
	App     Apprentice
	Actor   Contract    // the contract that runs every child, from Started
	Store   Store       // base types known, one guard each
	History []InputBase // inputs seen
	Agents  map[AgentID]*Agent
	Depth   int // the bound

	reading *reading // a CEL read in flight: the next event must answer it
}

type reading struct {
	agent AgentID
	base  ReturnBase
}

// An AgentID is the agent's path: "0", "0.1", "0.1.0".
type AgentID string

const rootID AgentID = "0"

type Phase uint8

const (
	_               Phase = iota
	PhaseThinking         // its turn is out, or about to be
	PhaseInvoking         // waiting on what its perform_effects bring back
	PhaseDelegating       // waiting on its children
	PhaseDone
	PhaseFaulted
)

type Agent struct {
	ID       AgentID
	Parent   AgentID  // "" for the root
	Self     Contract // the root's from Started, a child's the tree's actor; never a payload
	Kind     Kind
	Offer    Offer
	Depth    int
	Input    InputBase     // what it was invoked under
	Received []Observation // what it was passed
	Scope    []Observation // grows by Receive
	Created  []Observation // what its model made directly: never returned
	Names    []Named       // 0, 1, 2, …: the integers the model refers to
	Waiting  []Pending     // what its last call started
	Phase    Phase
	Returned []Observation // when done: what it returned, or its exit
	children int
}

// A Named is an observation the model may refer to, and the return base
// type it came back under.
type Named struct {
	Obs  Observation
	Base ReturnBase
}

// A Pending is an instruction of a call still in flight: a perform_effect,
// or a child.
type Pending struct {
	Element  int
	Child    AgentID // "" for a perform_effect
	Contract Contract
	Input    InputBase
	Done     bool
}

// NewTree is the tree before Started: the apprentice and the depth bound.
func NewTree(a Apprentice, depth int) Tree {
	return Tree{App: a, Store: EmptyStore(), Agents: map[AgentID]*Agent{}, Depth: depth}
}

// Root is the root agent, once Started.
func (t Tree) Root() (Agent, bool) {
	a, ok := t.Agents[rootID]
	if !ok {
		return Agent{}, false
	}
	return *a, true
}

// Done: the root has returned, exited or faulted.
func (t Tree) Done() bool {
	a, ok := t.Agents[rootID]
	return ok && (a.Phase == PhaseDone || a.Phase == PhaseFaulted)
}

func (t Tree) clone() Tree {
	c := t
	c.History = slices.Clone(t.History)
	c.Agents = make(map[AgentID]*Agent, len(t.Agents))
	for id, a := range t.Agents {
		b := *a
		b.Offer = Offer{slices.Clone(a.Offer.Invoke), slices.Clone(a.Offer.PassOnly), slices.Clone(a.Offer.Apprentice)}
		b.Received = slices.Clone(a.Received)
		b.Scope = slices.Clone(a.Scope)
		b.Created = slices.Clone(a.Created)
		b.Names = slices.Clone(a.Names)
		b.Waiting = slices.Clone(a.Waiting)
		b.Returned = slices.Clone(a.Returned)
		c.Agents[id] = &b
	}
	if t.reading != nil {
		r := *t.reading
		c.reading = &r
	}
	return c
}

// --- events ---

type Event interface{ event() }

// Started: the root, its contract, its reach and its kind, given by whoever
// started it; the actor that runs every child; the endowment's output shape.
type Started struct {
	Self      Contract
	Reach     []Contract
	Kind      Kind
	Actor     Contract
	Endowment Schema
}

// ReadResult answers a Read: raw and unverified, per key, in order.
type ReadResult struct{ Found [][]byte }

// Committed answers a Commit, an outcome per write.
type Committed struct{ Outcomes []Outcome }

// Replied is a turn's reply, raw: interpreted into a Call.
type Replied struct {
	Agent AgentID
	Out   []byte
}

// Returned: what a perform_effect brought back, with its shape.
type Returned struct {
	Agent   AgentID
	Element int // which instruction of the call it answers, by position
	Shape   Schema
	Err     string
}

// Stopped: a node's crash, timeout or kill. Always a fault.
type Stopped struct {
	Agent  AgentID
	Reason string
}

func (Started) event()    {}
func (ReadResult) event() {}
func (Committed) event()  {}
func (Replied) event()    {}
func (Returned) event()   {}
func (Stopped) event()    {}

type Outcome uint8

const (
	_ Outcome = iota
	OutcomeCreated
	OutcomeIdentical
	OutcomeConflict
	OutcomeLost
)

// --- what a reply holds ---

// A Call is what a reply holds: one call, many instructions.
type Call []Instruction

type Instruction interface{ instruction() }

type Goal struct{ Text string }        // a delegate's goal, an effect
type CreateInput struct{ Text string } // an act node's input, an effect
type Delegate struct {                 // creates nothing: passes what is referred to
	Observations []Ref
	Contracts    []Contract
}
type PerformEffect struct { // the generic dispatch
	Contract Contract
	Input    Ref
}
type Return struct{ Members []Ref }
type Exit struct{ Reason string }

func (Goal) instruction()          {}
func (CreateInput) instruction()   {}
func (Delegate) instruction()      {}
func (PerformEffect) instruction() {}
func (Return) instruction()        {}
func (Exit) instruction()          {}

// A Ref is what the model refers to: an integer, never content.
type Ref interface{ ref() }
type InScope int // an observation in the agent's scope, by its name
type Earlier int // the result of an earlier instruction of this call, by position

func (InScope) ref() {}
func (Earlier) ref() {}

// --- effects ---

// Effects come in order: a Commit, then an Execute, then at most one Read,
// last, as what follows it waits on its answer.
type Effect interface{ effect() }

type Read struct{ Keys []Key }
type Commit struct{ Writes []Write }
type Execute struct{ Acts []Action } // one instruction/1

func (Read) effect()    {}
func (Commit) effect()  {}
func (Execute) effect() {}

type Key interface{ key() }

// CELKey asks for a guard's answer over named values, from apprentice/refine.
// CEL is the written guard's code; empty for the default, which the shell
// registered.
type CELKey struct {
	Base   ReturnBase
	CEL    string
	Inputs []int // the observations it reads, by name
}

func (CELKey) key() {}

type Write interface{ write() }

// RecordWrite is record/1: decisions, values, faults. Values never enter the
// state; the shell keeps them under their names.
type RecordWrite struct {
	Kind  string
	Agent AgentID
	Body  map[string]any
}

func (RecordWrite) write() {}

type Action interface{ action() }

// Turn is one run of an agent's actor; its reply is Replied.
type Turn struct {
	Agent     AgentID
	Binary    Binary
	Authority string // the agent's grant: the shell's
	Offer     Offer
	Names     []int // in scope; the shell renders their values
}

// InvokeContract is an act node's perform_effect; what comes back is
// Returned.
type InvokeContract struct {
	Agent     AgentID
	Contract  Contract
	Authority string // the act node's grant: the shell's
	Element   int    // its position in the call
	Input     int    // the observation dispatched, by name; the shell sends its value
}

func (Turn) action()           {}
func (InvokeContract) action() {}

// --- Step ---

type run struct {
	t      Tree
	writes []Write
	acts   []Action
	read   *Read
}

func Step(t Tree, e Event) (Tree, []Effect) {
	r := &run{t: t.clone()}
	if r.t.reading != nil {
		if rr, ok := e.(ReadResult); ok {
			r.answered(rr)
		} else {
			r.record("ignored", "", map[string]any{"why": "a read is in flight"})
		}
		return r.t, r.effects()
	}
	switch e := e.(type) {
	case Started:
		r.started(e)
	case Replied:
		r.replied(e)
	case Returned:
		r.returned(e)
	case Stopped:
		if a := r.live(e.Agent); a != nil {
			r.fault(a, "stopped: "+e.Reason)
		}
	case Committed:
		// Step 4a: records are not write-once yet; nothing waits on them.
	case ReadResult:
		r.record("ignored", "", map[string]any{"why": "no read is in flight"})
	}
	return r.t, r.effects()
}

func (r *run) effects() []Effect {
	var out []Effect
	if len(r.writes) > 0 {
		out = append(out, Commit{r.writes})
	}
	if len(r.acts) > 0 {
		out = append(out, Execute{r.acts})
	}
	if r.read != nil {
		out = append(out, *r.read)
	}
	return out
}

func (r *run) record(kind string, a AgentID, body map[string]any) {
	r.writes = append(r.writes, RecordWrite{kind, a, body})
}

// live is an agent that may still take an event; anything else is recorded
// and ignored.
func (r *run) live(id AgentID) *Agent {
	a, ok := r.t.Agents[id]
	if !ok || a.Phase == PhaseDone || a.Phase == PhaseFaulted {
		r.record("ignored", id, map[string]any{"why": "no live agent"})
		return nil
	}
	return a
}

// --- provisioning ---

func (r *run) started(e Started) {
	if len(r.t.Agents) > 0 {
		r.record("ignored", rootID, map[string]any{"why": "already started"})
		return
	}
	app := r.t.App
	r.t.Actor = e.Actor
	a := &Agent{ID: rootID, Self: e.Self}
	r.t.Agents[rootID] = a
	k, ok := RootKind(app, e.Self, e.Reach, e.Kind)
	if !ok || (k != KindAct && k != KindDelegate) {
		r.fault(a, "the root reaches its own binary, or has no kind")
		return
	}
	a.Kind = k
	a.Offer = Offered(app, k, counted(app, e.Reach))
	rt, ok := TypeOf(app.root, e.Endowment)
	if !ok {
		r.fault(a, "the endowment's shape is not well formed")
		return
	}
	o, _ := RootObservation(app, rt)
	base, _ := BaseOf(rt, nil)
	a.Input = r.invoke(a.ID, e.Self, []ReturnBase{base})
	a.Received = []Observation{o}
	a.Scope = Receive(DirInvoke, nil, a.Received)
	a.Names = []Named{{o, base}}
	r.turn(a)
}

// invoke is down an edge: the input base type, and do's answer, recorded.
// With remember off the invoke guard is not consulted: no store holds one.
func (r *run) invoke(id AgentID, c Contract, passed []ReturnBase) InputBase {
	i := CreateInvoke(c, passed)
	seen := Do(r.t.History, i)
	if seen == SeenUnseen {
		r.t.History = append(r.t.History, i)
	}
	r.record("invoke", id, map[string]any{"input": i.id, "do": seenWord(seen)})
	return i
}

func seenWord(s Seen) string {
	if s == SeenSeen {
		return "seen"
	}
	return "unseen"
}

func (r *run) turn(a *Agent) {
	a.Phase = PhaseThinking
	a.Waiting = nil
	names := make([]int, len(a.Names))
	for i := range names {
		names[i] = i
	}
	r.acts = append(r.acts, Turn{Agent: a.ID, Binary: a.Self.binary, Offer: a.Offer, Names: names})
}

// --- faults ---

// fault ends an agent; its parent takes nothing from it.
func (r *run) fault(a *Agent, why string) {
	a.Phase = PhaseFaulted
	a.Waiting = nil
	r.record("fault", a.ID, map[string]any{"why": why})
	r.up(a)
}

// --- interpret: a turn's reply ---

func (r *run) replied(e Replied) {
	a := r.live(e.Agent)
	if a == nil {
		return
	}
	if a.Phase != PhaseThinking {
		r.fault(a, "a reply while not thinking")
		return
	}
	call, why := parseCall(e.Out, a.Offer)
	if why == "" {
		before, writes, acts := r.t.clone(), len(r.writes), len(r.acts)
		why = r.interpret(a, call)
		if why != "" { // all or nothing: undo what the call did, then fault
			r.t, r.writes, r.acts = before, r.writes[:writes], r.acts[:acts]
			a = r.t.Agents[e.Agent]
		}
	}
	if why != "" {
		r.fault(a, why)
	}
}

// parseCall reads a reply's wire format: a JSON array, one object per
// instruction, observations by integer or {"earlier": n}, contracts by their
// index in the Offer (pass-only for delegate, invoke for perform_effect).
func parseCall(out []byte, o Offer) (Call, string) {
	var raw []map[string]json.RawMessage
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, "the reply is not an array of instructions"
	}
	if len(raw) == 0 {
		return nil, "an empty call"
	}
	if len(raw) > MaxFanOut {
		return nil, "a call of more than 32 instructions"
	}
	var call Call
	for _, m := range raw {
		if len(m) != 1 {
			return nil, "an instruction is one object of one key"
		}
		for k, v := range m {
			ins, why := parseInstruction(k, v, o)
			if why != "" {
				return nil, why
			}
			call = append(call, ins)
		}
	}
	return call, ""
}

func strict(v json.RawMessage, into any) bool {
	d := json.NewDecoder(bytes.NewReader(v))
	d.DisallowUnknownFields()
	return d.Decode(into) == nil && !d.More()
}

func parseInstruction(k string, v json.RawMessage, o Offer) (Instruction, string) {
	text := func() (string, string) {
		var s string
		if !strict(v, &s) {
			return "", k + " takes a string"
		}
		return s, ""
	}
	switch k {
	case "goal":
		s, why := text()
		return Goal{s}, why
	case "create_input":
		s, why := text()
		return CreateInput{s}, why
	case "exit":
		s, why := text()
		return Exit{s}, why
	case "return":
		var b struct {
			Members []json.RawMessage `json:"members"`
		}
		if !strict(v, &b) {
			return nil, "return takes members"
		}
		refs, why := parseRefs(b.Members)
		return Return{refs}, why
	case "delegate":
		var b struct {
			Observations []json.RawMessage `json:"observations"`
			Contracts    []int             `json:"contracts"`
		}
		if !strict(v, &b) {
			return nil, "delegate takes observations and contracts"
		}
		refs, why := parseRefs(b.Observations)
		if why != "" {
			return nil, why
		}
		var cs []Contract
		for _, i := range b.Contracts {
			if i < 0 || i >= len(o.PassOnly) {
				return nil, "delegate passes a contract it does not hold"
			}
			cs = append(cs, o.PassOnly[i])
		}
		return Delegate{refs, cs}, ""
	case "perform_effect":
		var b struct {
			Contract *int            `json:"contract"`
			Input    json.RawMessage `json:"input"`
		}
		if !strict(v, &b) || b.Contract == nil || b.Input == nil {
			return nil, "perform_effect takes a contract and an input"
		}
		if *b.Contract < 0 || *b.Contract >= len(o.Invoke) {
			return nil, "perform_effect of a contract it may not invoke"
		}
		ref, why := parseRef(b.Input)
		return PerformEffect{o.Invoke[*b.Contract], ref}, why
	}
	return nil, "no instruction " + k
}

func parseRefs(vs []json.RawMessage) ([]Ref, string) {
	var refs []Ref
	for _, v := range vs {
		ref, why := parseRef(v)
		if why != "" {
			return nil, why
		}
		refs = append(refs, ref)
	}
	return refs, ""
}

func parseRef(v json.RawMessage) (Ref, string) {
	var n int
	if strict(v, &n) {
		if n < 0 {
			return nil, "a negative name"
		}
		return InScope(n), ""
	}
	var e struct {
		Earlier *int `json:"earlier"`
	}
	if strict(v, &e) && e.Earlier != nil && *e.Earlier >= 0 {
		return Earlier(*e.Earlier), ""
	}
	return nil, "a reference is a name or {\"earlier\": n}"
}

// interpret runs a call, instruction by instruction, each checked on its
// own. It returns why the call is a fault, or "".
func (r *run) interpret(a *Agent, call Call) string {
	if len(call) > 1 {
		for _, ins := range call {
			switch ins.(type) {
			case Return, Exit:
				return "return and exit go alone"
			}
		}
	}
	made := make([]int, len(call)) // the name each instruction made, or -1
	for i, ins := range call {
		made[i] = -1
		resolve := func(ref Ref) (int, string) { return resolveRef(a, made, i, ref) }
		switch in := ins.(type) {
		case Goal:
			if a.Kind != KindDelegate {
				return "goal is a delegate's"
			}
			n, why := r.make(a, r.t.App.goal, in.Text, true)
			if why != "" {
				return why
			}
			made[i] = n
		case CreateInput:
			if a.Kind != KindAct {
				return "create_input is an act node's"
			}
			n, why := r.make(a, r.t.App.createInput, in.Text, true)
			if why != "" {
				return why
			}
			made[i] = n
		case Delegate:
			if a.Kind != KindDelegate {
				return "delegate is a delegate's"
			}
			var names []int
			for _, ref := range in.Observations {
				n, why := resolve(ref)
				if why != "" {
					return why
				}
				names = append(names, n)
			}
			if why := r.delegate(a, i, names, in.Contracts); why != "" {
				return why
			}
		case PerformEffect:
			if a.Kind != KindAct {
				return "perform_effect is an act node's"
			}
			n, why := resolve(in.Input)
			if why != "" {
				return why
			}
			r.performEffect(a, i, in.Contract, n)
		case Return:
			var returned []Observation
			for _, ref := range in.Members {
				n, why := resolve(ref)
				if why != "" {
					return why
				}
				returned = append(returned, a.Names[n].Obs)
			}
			if !MayReturn(a.Created, returned) || !NodeMayReturn(a.Received, returned) {
				return "a return of nothing, of what its model made, or of what it was passed"
			}
			r.finish(a, returned)
			return ""
		case Exit:
			t, _ := TypeOf(r.t.App.exit, shapeText)
			o, ok := Observe(r.t.App, a.Kind, t, nil)
			if !ok {
				return "observe refused the exit"
			}
			r.record("value", a.ID, map[string]any{"exit": in.Reason})
			r.finish(a, []Observation{o})
			return ""
		}
	}
	if len(a.Waiting) == 0 { // only goals and inputs: the model goes on
		r.turn(a)
	} else if a.Kind == KindAct {
		a.Phase = PhaseInvoking
	} else {
		a.Phase = PhaseDelegating
		r.ready(a) // every child it started may have faulted already
	}
	return ""
}

// resolveRef is check 9: a name in scope, or an earlier instruction's
// result, never a later one's, and only one that made something now.
func resolveRef(a *Agent, made []int, i int, ref Ref) (int, string) {
	switch ref := ref.(type) {
	case InScope:
		n := int(ref)
		if n >= len(a.Names) || !MayPass(a.Scope, []Observation{a.Names[n].Obs}) {
			return 0, "a name not in scope"
		}
		return n, ""
	case Earlier:
		j := int(ref)
		if j >= i || made[j] < 0 {
			return 0, "an earlier reference to nothing made yet"
		}
		return made[j], ""
	}
	return 0, "no reference"
}

// make is an effect through a contract of the apprentice, from text: down
// the edge and back up (baseOf), named. byModel marks what the model made.
func (r *run) make(a *Agent, c Contract, text string, byModel bool) (int, string) {
	t, _ := TypeOf(c, shapeText)
	o, ok := Observe(r.t.App, a.Kind, t, nil)
	if !ok {
		return 0, "observe refused"
	}
	base, _ := BaseOf(t, nil)
	if byModel {
		a.Created = append(a.Created, o)
	}
	a.Scope = append(a.Scope, o)
	a.Names = append(a.Names, Named{o, base})
	n := len(a.Names) - 1
	r.record("value", a.ID, map[string]any{"name": n, "text": text})
	return n, ""
}

// --- delegate: the act transformation, and a child provisioned ---

func (r *run) delegate(a *Agent, element int, names []int, cs []Contract) string {
	app := r.t.App
	if !Passable(a.Offer.PassOnly, cs) {
		return "delegate passes a contract it does not hold"
	}
	var passed []Named
	hasGoal := false
	for _, n := range names {
		passed = append(passed, a.Names[n])
		hasGoal = hasGoal || a.Names[n].Obs.typ.contract == app.goal
	}
	if !hasGoal {
		return "delegate passes no goal"
	}
	id := a.ID + "." + AgentID(itoa(a.children))
	a.children++
	child := &Agent{ID: id, Parent: a.ID, Self: r.t.Actor, Depth: a.Depth + 1}
	r.t.Agents[id] = child
	a.Waiting = append(a.Waiting, Pending{Element: element, Child: id})
	if child.Depth > r.t.Depth {
		r.fault(child, "the depth bound")
		return ""
	}
	k, ok := Classify(app, r.t.Actor, cs)
	if !ok {
		r.fault(child, "refused: the actor in its own reach, or nothing passed")
		return ""
	}
	child.Kind = k
	// delegate's CEL, before the model's point: the parts for the child's kind.
	var members []Named
	for _, c := range Parts(app, k) {
		shape, value := shapeText, any(actText)
		switch c {
		case app.createDelegation:
			value = delegationText
		case app.reach:
			shape, value = shapeList, len(cs)
		}
		t, _ := TypeOf(c, shape)
		o, ok := Observe(app, KindDelegate, t, nil)
		if !ok {
			return "observe refused a part"
		}
		base, _ := BaseOf(t, nil)
		members = append(members, Named{o, base})
		r.record("value", id, map[string]any{"part": c.shape, "value": value})
	}
	members = append(members, passed...)
	var obs []Observation
	var bases []ReturnBase
	for _, m := range members {
		obs = append(obs, m.Obs)
		bases = append(bases, m.Base)
	}
	ct, _ := TypeOf(app.delegate, shapeList)
	comp, ok := Observe(app, KindDelegate, ct, obs)
	if !ok {
		return "delegate passes one observation twice"
	}
	compBase, _ := BaseOf(ct, bases)
	child.Input = r.invoke(id, r.t.Actor, []ReturnBase{compBase})
	child.Offer = Offered(app, k, cs)
	child.Received = []Observation{comp}
	child.Scope = Receive(DirInvoke, nil, append([]Observation{comp}, obs...))
	child.Names = append([]Named{{comp, compBase}}, members...)
	r.turn(child)
	return ""
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}

// --- perform_effect: down an edge, and what comes back ---

func (r *run) performEffect(a *Agent, element int, c Contract, n int) {
	i := r.invoke(a.ID, c, []ReturnBase{a.Names[n].Base})
	a.Waiting = append(a.Waiting, Pending{Element: element, Contract: c, Input: i})
	r.acts = append(r.acts, InvokeContract{Agent: a.ID, Contract: c, Element: element, Input: n})
}

func (r *run) returned(e Returned) {
	a := r.live(e.Agent)
	if a == nil {
		return
	}
	p := pendingAt(a, e.Element)
	if a.Phase != PhaseInvoking || p == nil || p.Child != "" || p.Done {
		r.fault(a, "a return for nothing in flight")
		return
	}
	if e.Err != "" {
		r.fault(a, "perform_effect failed: "+e.Err)
		return
	}
	t, ok := TypeOf(p.Contract, e.Shape)
	if !ok {
		r.fault(a, "the shape that came back is not well formed")
		return
	}
	o, ok := Observe(r.t.App, KindAct, t, nil)
	if !ok {
		r.fault(a, "observe refused what came back")
		return
	}
	base, _ := CreateReturn(p.Input, e.Shape)
	p.Done = true
	r.arrive(a, base, []Observation{o})
}

func pendingAt(a *Agent, element int) *Pending {
	for i := range a.Waiting {
		if a.Waiting[i].Element == element && a.Waiting[i].Child == "" {
			return &a.Waiting[i]
		}
	}
	return nil
}

// --- up an edge ---

// finish is an agent's return or exit: done, and passed up.
func (r *run) finish(a *Agent, returned []Observation) {
	a.Phase = PhaseDone
	a.Returned = returned
	r.record("return", a.ID, map[string]any{"members": len(returned)})
	r.up(a)
}

// up hands a done or faulted agent to its parent's Delegating.
func (r *run) up(a *Agent) {
	if a.Parent == "" {
		r.record("done", a.ID, map[string]any{"faulted": a.Phase == PhaseFaulted})
		return
	}
	p := r.t.Agents[a.Parent]
	var w *Pending
	for i := range p.Waiting {
		if p.Waiting[i].Child == a.ID {
			w = &p.Waiting[i]
		}
	}
	if p.Phase != PhaseDelegating && p.Phase != PhaseThinking || w == nil {
		r.record("ignored", a.ID, map[string]any{"why": "its parent no longer waits"})
		return
	}
	w.Done = true
	if a.Phase == PhaseFaulted {
		r.ready(p)
		return
	}
	p.Scope = Receive(DirRet, p.Scope, a.Returned)
	base, _ := CreateReturn(a.Input, wrapperShape(a.Kind))
	r.arrive(p, base, a.Returned)
}

// arrive names what came back under its return base type, then sees it:
// unseen registers the default guard; seen asks for the guard's answer.
func (r *run) arrive(a *Agent, base ReturnBase, obs []Observation) {
	var names []int
	for _, o := range obs {
		if a.Phase == PhaseInvoking {
			a.Scope = Receive(DirRet, a.Scope, []Observation{o})
		}
		a.Names = append(a.Names, Named{o, base})
		names = append(names, len(a.Names)-1)
	}
	seen, store := See(r.t.Store, base)
	r.t.Store = store
	if seen == SeenUnseen {
		r.record("guard", a.ID, map[string]any{"base": base.id, "see": "unseen", "ret": "fallback"})
		r.ready(a)
		return
	}
	g, _ := store.Guard(base)
	cel, _ := g.Written()
	r.read = &Read{[]Key{CELKey{Base: base, CEL: cel, Inputs: names}}}
	r.t.reading = &reading{a.ID, base}
}

// answered takes a guard's answer: trusted, recorded, never consulted with
// remember off.
func (r *run) answered(rr ReadResult) {
	rd := *r.t.reading
	r.t.reading = nil
	a := r.t.Agents[rd.agent]
	var holds bool
	if len(rr.Found) != 1 || json.Unmarshal(rr.Found[0], &holds) != nil {
		r.fault(a, "a guard's answer is not true or false")
		return
	}
	ret := "fallback"
	if ReturnOf(SeenSeen, holds) == ReturnContinue {
		ret = "continue"
	}
	r.record("guard", a.ID, map[string]any{"base": rd.base.id, "see": "seen", "holds": holds, "ret": ret})
	r.ready(a)
}

// ready: once everything its call started is back, the model goes on.
func (r *run) ready(a *Agent) {
	if a.Phase != PhaseInvoking && a.Phase != PhaseDelegating {
		return
	}
	for _, w := range a.Waiting {
		if !w.Done {
			return
		}
	}
	r.turn(a)
}
