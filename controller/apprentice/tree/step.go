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
// Each Step returns at most one effect, the kernel's cycle, Read* Commit?
// Execute*: a Read, answered by a ReadResult; a Commit, answered by
// Committed; or an Execute, whose results are the next events. What an event
// decides before a read, it holds in the tree until the cycle completes, so
// nothing leaves the process before its records are durable. While a read or
// a commit is in flight (Awaiting), the shell holds every other event back.
//
// The store and the history are folds of write-once records, read raw and
// verified here. With remember off, a guard's answer is recorded and the
// agent always goes back to its model.

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

	held *held // a cycle in flight: what it decided, and what it waits on
}

// held is a cycle in flight: the writes and actions decided so far, and a
// read (cont) or a commit (committing) whose answer is the next event.
type held struct {
	writes     []Write
	owners     []AgentID // whose decision each write is
	acts       []Action
	read       *cont
	committing bool
}

// cont is what a read's answer goes on to do.
type cont struct {
	kind    contKind
	agent   AgentID
	base    ReturnBase
	names   []int   // what came back, by name: a guard's inputs
	started Started // the root, while the history is read
}

type contKind uint8

const (
	_           contKind = iota
	contHistory          // the input records, before the root is provisioned
	contChain            // a return base type's guard records
	contCEL              // a guard's answer
)

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
	Gave     []int         // the names of what it returned
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
		b.Gave = slices.Clone(a.Gave)
		c.Agents[id] = &b
	}
	if t.held != nil {
		h := *t.held
		h.writes, h.owners, h.acts = slices.Clone(h.writes), slices.Clone(h.owners), slices.Clone(h.acts)
		if h.read != nil {
			k := *h.read
			k.names = slices.Clone(k.names)
			h.read = &k
		}
		c.held = &h
	}
	return c
}

// Awaiting: a read or a commit is in flight, and only its answer may be the
// next event. The shell holds every other event back until it is answered.
func (t Tree) Awaiting() bool { return t.held != nil }

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

// Step returns at most one Effect: a Read, a Commit or an Execute.
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

// HistoryKey asks for the input records: a JSON array of input base ids, or
// nothing.
type HistoryKey struct{}

// ChainKey asks for a return base type's guard records: a JSON array of
// guard objects, oldest first, or nothing.
type ChainKey struct{ Base ReturnBase }

func (HistoryKey) key() {}
func (ChainKey) key()   {}

type Write interface{ write() }

// RecordWrite is record/1: decisions, values, faults. Values never enter the
// state; the shell keeps them under their names.
type RecordWrite struct {
	Kind  string
	Agent AgentID
	Body  map[string]any
}

func (RecordWrite) write() {}

// GuardWrite is a guard, write-once under its refinement id: its base type,
// its code and the refinement id it replaces (none for the default). A retry,
// or a sibling's same default, finds it Identical.
type GuardWrite struct {
	ID    RefinementID
	Base  ReturnBase
	Bytes []byte
}

// HistoryWrite appends an input record: the input base id, as JSON.
type HistoryWrite struct{ Bytes []byte }

func (GuardWrite) write()   {}
func (HistoryWrite) write() {}

// guardBytes is a guard object's canonical encoding.
func guardBytes(b ReturnBase, g Guard, replaces RefinementID) []byte {
	var guard any = "default"
	if cel, ok := g.Written(); ok {
		guard = map[string]string{"written": cel}
	}
	return canonical(map[string]any{"base": b.id, "guard": guard, "replaces": string(replaces)})
}

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
	owners []AgentID
	acts   []Action
	read   *cont
	key    Key // the read's key, once decided
}

func Step(t Tree, e Event) (Tree, []Effect) {
	r := &run{t: t.clone()}
	h := r.t.held
	r.t.held = nil
	switch {
	case h != nil && h.read != nil:
		rr, ok := e.(ReadResult)
		if !ok {
			return t, nil // the shell holds it back: see Awaiting
		}
		r.writes, r.owners, r.acts = h.writes, h.owners, h.acts
		r.answer(*h.read, rr)
	case h != nil && h.committing:
		c, ok := e.(Committed)
		if !ok {
			return t, nil
		}
		r.acts = h.acts
		r.committed(h, c)
	default:
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
		case ReadResult, Committed:
			r.record("ignored", "", map[string]any{"why": "nothing was asked"})
		}
	}
	return r.next()
}

// next is the cycle's next effect: the read it waits on, else the commit
// of what it decided, else the actions, once their records are durable.
func (r *run) next() (Tree, []Effect) {
	if r.read != nil {
		r.t.held = &held{writes: r.writes, owners: r.owners, acts: r.acts, read: r.read}
		return r.t, []Effect{Read{[]Key{r.key}}}
	}
	if len(r.writes) > 0 {
		r.t.held = &held{writes: r.writes, owners: r.owners, acts: r.acts, committing: true}
		return r.t, []Effect{Commit{r.writes}}
	}
	var acts []Action
	for _, a := range r.acts { // an agent faulted since is not run
		if ag := r.t.Agents[actAgent(a)]; ag != nil && ag.Phase != PhaseFaulted && ag.Phase != PhaseDone {
			acts = append(acts, a)
		}
	}
	if len(acts) > 0 {
		return r.t, []Effect{Execute{acts}}
	}
	return r.t, nil
}

func actAgent(a Action) AgentID {
	switch a := a.(type) {
	case Turn:
		return a.Agent
	case InvokeContract:
		return a.Agent
	}
	return ""
}

// committed takes a commit's outcomes: Created and Identical are what
// write-once allows; Conflict is corruption, and faults whose write it was.
// Nothing is Lost: no write races yet.
func (r *run) committed(h *held, c Committed) {
	if len(c.Outcomes) != len(h.writes) {
		for _, id := range h.owners {
			if a := r.t.Agents[id]; a != nil && a.Phase != PhaseFaulted {
				r.fault(a, "a commit answered for other writes")
			}
		}
		return
	}
	for i, o := range c.Outcomes {
		if o == OutcomeCreated || o == OutcomeIdentical {
			continue
		}
		if a := r.t.Agents[h.owners[i]]; a != nil && a.Phase != PhaseFaulted {
			r.fault(a, "a write conflicted: corruption")
		}
	}
}

func (r *run) write(w Write, owner AgentID) {
	r.writes = append(r.writes, w)
	r.owners = append(r.owners, owner)
}

func (r *run) record(kind string, a AgentID, body map[string]any) {
	r.write(RecordWrite{kind, a, body}, a)
}

// named records where a name's value comes from, for the shell, which keeps
// every value: never the state (55, invariant 3). from is one of
//
//	{"from":"input"}                               the root's endowment: the tree's input
//	{"from":"goal"|"input"|"exit"|"part", "text"}  text the model or delegate's CEL wrote
//	{"from":"part", "contracts":[binary…]}         a packaged reach
//	{"from":"compose", "names":[n…]}               a computation of the agent's own names
//	{"from":"copy", "agent", "of"}                 another agent's name: passed down or returned up
//	{"from":"result", "element"}                   what the call's element brought back
func (r *run) named(a AgentID, n int, from map[string]any) {
	body := map[string]any{"name": n}
	for k, v := range from {
		body[k] = v
	}
	r.record("name", a, body)
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

// started reads the history first: do needs it from the first invoke.
func (r *run) started(e Started) {
	if len(r.t.Agents) > 0 {
		r.record("ignored", rootID, map[string]any{"why": "already started"})
		return
	}
	r.read, r.key = &cont{kind: contHistory, started: e}, HistoryKey{}
}

// history is check 5: every input record is an input base id, well formed.
func history(found []byte) ([]InputBase, bool) {
	if found == nil {
		return nil, true
	}
	var ids []string
	if json.Unmarshal(found, &ids) != nil {
		return nil, false
	}
	var out []InputBase
	for _, id := range ids {
		if len(id) != len("sha256:")+64 || id[:7] != "sha256:" || !hexDigits(id[7:], 64) {
			return nil, false
		}
		out = append(out, InputBase{id})
	}
	return out, true
}

func (r *run) provision(e Started) {
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
	r.named(a.ID, 0, map[string]any{"from": "input"})
	r.turn(a)
}

// invoke is down an edge: the input base type, and do's answer, recorded.
// With remember off the invoke guard is not consulted: no store holds one.
func (r *run) invoke(id AgentID, c Contract, passed []ReturnBase) InputBase {
	i := CreateInvoke(c, passed)
	seen := Do(r.t.History, i)
	if seen == SeenUnseen {
		r.t.History = append(r.t.History, i)
		r.write(HistoryWrite{canonical(i.id)}, id)
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
			r.t, r.writes, r.owners, r.acts = before, r.writes[:writes], r.owners[:writes], r.acts[:acts]
			r.read, r.key = nil, nil
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
			var gave []int
			for _, ref := range in.Members {
				n, why := resolve(ref)
				if why != "" {
					return why
				}
				returned, gave = append(returned, a.Names[n].Obs), append(gave, n)
			}
			if !MayReturn(a.Created, returned) || !NodeMayReturn(a.Received, returned) {
				return "a return of nothing, of what its model made, or of what it was passed"
			}
			r.finish(a, returned, gave)
			return ""
		case Exit:
			t, _ := TypeOf(r.t.App.exit, shapeText)
			o, ok := Observe(r.t.App, a.Kind, t, nil)
			if !ok {
				return "observe refused the exit"
			}
			base, _ := BaseOf(t, nil)
			a.Names = append(a.Names, Named{o, base})
			r.named(a.ID, len(a.Names)-1, map[string]any{"from": "exit", "text": in.Reason})
			r.finish(a, []Observation{o}, []int{len(a.Names) - 1})
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
	from := "goal"
	if c == r.t.App.createInput {
		from = "input"
	}
	r.named(a.ID, n, map[string]any{"from": from, "text": text})
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
		shape, from := shapeText, map[string]any{"from": "part", "text": actText}
		switch c {
		case app.createDelegation:
			from["text"] = delegationText
		case app.reach:
			var bs []string
			for _, p := range cs {
				bs = append(bs, string(p.binary))
			}
			shape, from = shapeList, map[string]any{"from": "part", "contracts": bs}
		}
		t, _ := TypeOf(c, shape)
		o, ok := Observe(app, KindDelegate, t, nil)
		if !ok {
			return "observe refused a part"
		}
		base, _ := BaseOf(t, nil)
		members = append(members, Named{o, base})
		r.named(id, len(members), from)
	}
	for i, n := range names {
		r.named(id, len(members)+1+i, map[string]any{"from": "copy", "agent": string(a.ID), "of": n})
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
	all := make([]int, len(members))
	for i := range all {
		all[i] = i + 1
	}
	r.named(id, 0, map[string]any{"from": "compose", "names": all})
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
	r.arrive(a, base, []Observation{o}, []map[string]any{{"from": "result", "element": e.Element}})
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
func (r *run) finish(a *Agent, returned []Observation, gave []int) {
	a.Phase = PhaseDone
	a.Returned, a.Gave = returned, gave
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
	var from []map[string]any
	for _, n := range a.Gave {
		from = append(from, map[string]any{"from": "copy", "agent": string(a.ID), "of": n})
	}
	r.arrive(p, base, a.Returned, from)
}

// arrive names what came back under its return base type, then sees it. A
// base type this run has not read is read first, from its guard records.
func (r *run) arrive(a *Agent, base ReturnBase, obs []Observation, from []map[string]any) {
	var names []int
	for i, o := range obs {
		if a.Phase == PhaseInvoking {
			a.Scope = Receive(DirRet, a.Scope, []Observation{o})
		}
		a.Names = append(a.Names, Named{o, base})
		names = append(names, len(a.Names)-1)
		r.named(a.ID, len(a.Names)-1, from[i])
	}
	if slices.Contains(r.t.Store.Known(), base) {
		r.ask(a, base, names)
		return
	}
	r.read, r.key = &cont{kind: contChain, agent: a.ID, base: base, names: names}, ChainKey{base}
}

// ask is a seen base type's guard, asked over what came back.
func (r *run) ask(a *Agent, base ReturnBase, names []int) {
	g, _ := r.t.Store.Guard(base)
	cel, _ := g.Written()
	r.read = &cont{kind: contCEL, agent: a.ID, base: base, names: names}
	r.key = CELKey{Base: base, CEL: cel, Inputs: names}
}

// answer takes a read's answer and goes on.
func (r *run) answer(c cont, rr ReadResult) {
	if len(rr.Found) != 1 {
		if a := r.t.Agents[c.agent]; a != nil {
			r.fault(a, "a read answered for other keys")
		} else {
			r.t.Agents[rootID] = &Agent{ID: rootID, Self: c.started.Self}
			r.fault(r.t.Agents[rootID], "a read answered for other keys")
		}
		return
	}
	found := rr.Found[0]
	switch c.kind {
	case contHistory:
		h, ok := history(found)
		if !ok {
			r.t.Agents[rootID] = &Agent{ID: rootID, Self: c.started.Self}
			r.fault(r.t.Agents[rootID], "a corrupt input record")
			return
		}
		r.t.History = h
		r.provision(c.started)
	case contChain:
		a := r.t.Agents[c.agent]
		store, ok := fold(r.t.Store, c.base, found)
		if !ok {
			r.fault(a, "a corrupt guard chain")
			return
		}
		r.t.Store = store
		if slices.Contains(store.Known(), c.base) {
			r.ask(a, c.base, c.names)
			return
		}
		// Unseen: the default is the base type's first guard (check 14).
		_, r.t.Store = See(r.t.Store, c.base)
		g := DefaultGuard()
		r.write(GuardWrite{RefinementOf(c.base, g), c.base, guardBytes(c.base, g, "")}, a.ID)
		r.record("guard", a.ID, map[string]any{"base": c.base.id, "see": "unseen", "ret": "fallback"})
		r.ready(a)
	case contCEL:
		a := r.t.Agents[c.agent]
		var holds bool
		if !strict(found, &holds) {
			r.fault(a, "a guard's answer is not true or false")
			return
		}
		ret := "fallback"
		if ReturnOf(SeenSeen, holds) == ReturnContinue {
			ret = "continue"
		}
		r.record("guard", a.ID, map[string]any{"base": c.base.id, "see": "seen", "holds": holds, "ret": ret})
		r.ready(a)
	}
}

// fold is checks 1, 3, 4 and 14 over a return base type's guard records:
// each its canonical encoding and for the base type it is read under; the
// first the default, replacing nothing; each after it written, replacing the
// one before. The store's slot is the last.
func fold(s Store, b ReturnBase, found []byte) (Store, bool) {
	if found == nil {
		return s, true
	}
	var chain []json.RawMessage
	if json.Unmarshal(found, &chain) != nil || len(chain) == 0 {
		return s, false
	}
	var prev RefinementID
	for i, raw := range chain {
		var rec struct {
			Base     string          `json:"base"`
			Guard    json.RawMessage `json:"guard"`
			Replaces string          `json:"replaces"`
		}
		if !strict(raw, &rec) || rec.Base != b.id || RefinementID(rec.Replaces) != prev {
			return s, false
		}
		var g Guard
		var def string
		var written struct {
			Written *string `json:"written"`
		}
		switch {
		case strict(rec.Guard, &def) && def == "default" && i == 0:
			g = DefaultGuard()
		case strict(rec.Guard, &written) && written.Written != nil && i > 0:
			g = WrittenGuard(*written.Written)
		default:
			return s, false
		}
		if !bytes.Equal(guardBytes(b, g, prev), raw) {
			return s, false
		}
		if i == 0 {
			_, s = See(s, b)
		} else {
			s, _ = Rewrite(s, b, *written.Written)
		}
		prev = RefinementOf(b, g)
	}
	return s, true
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
