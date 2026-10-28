package machine

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// world is an in-memory shell and tree for whole lineages (41, V8): every
// instance is a proc, every target a function, and every Step call is
// checked against the loop's grammar. Nothing touches the OS.
type world struct {
	t       *testing.T
	objects map[string][]byte // kind/hash
	frames  map[string]*memFrame
	accepts map[string]string
	target  func(h horizon.Horizon, stdin []byte) []byte
	queue   []queued
	steps   int
}

type memFrame struct {
	result  string
	slots   map[string]map[string]int // round → child → position
	emitted map[string]string
	claims  map[string]string
}

type proc struct {
	s       Instance
	stack   []string // the handle table of frames: stamps
	parent  *proc
	kids    []*proc
	pending []Action // the rest of a batch of executes, after one that waits on another proc
	exited  bool
	err     string
}

type queued struct {
	p *proc
	e Event
}

func newWorld(t *testing.T, target func(horizon.Horizon, []byte) []byte) *world {
	return &world{t: t, objects: map[string][]byte{}, frames: map[string]*memFrame{}, accepts: map[string]string{}, target: target}
}

func (w *world) frame(stamp string) *memFrame {
	f := w.frames[stamp]
	if f == nil {
		f = &memFrame{slots: map[string]map[string]int{}, emitted: map[string]string{}, claims: map[string]string{}}
		w.frames[stamp] = f
	}
	return f
}

// boot commits genesis, as boot.go does, and runs the operator's seed to the
// end of the lineage. It returns what discharged the operator's round.
func (w *world) boot(reach map[string]horizon.Surface, seed string) record.Record {
	w.t.Helper()
	g := record.Record{Version: 2, State: record.Invoke, Reach: reach, Author: "operator", Nonce: "n"}
	o := HeldRecord(g)
	w.objects["records/"+o.Key] = o.Raw
	w.frame(o.Key)
	h0 := horizon.Horizon{Stamp: o.Key, Reach: reach, Frame: o.Key}
	root := &proc{s: Instance{Floor: "/floor/.skyra", R: g, Phase: Running, Self: h0}, stack: []string{o.Key}}
	w.queue = append(w.queue, queued{root, Executed{Obs: []byte(seed)}})
	for len(w.queue) > 0 {
		q := w.queue[0]
		w.queue = w.queue[1:]
		w.run(q.p, q.e)
		if w.steps > 10000 {
			w.t.Fatal("the lineage did not end")
		}
	}
	// The operator's round resolves like any other: boot reads its claim.
	claim := w.frame(o.Key).claims[o.Key]
	resolve, err := RecordOf(Object{Key: claim, Raw: w.objects["records/"+claim]})
	if err != nil || len(resolve.Deps) != 1 {
		w.t.Fatalf("the operator's round never resolved: %v", err)
	}
	r, err := RecordOf(Object{Key: resolve.Deps[0], Raw: w.objects["records/"+resolve.Deps[0]]})
	if err != nil {
		w.t.Fatal(err)
	}
	return r
}

// run steps p from e until it waits on another proc or exits.
func (w *world) run(p *proc, e Event) {
	w.t.Helper()
	if rest := p.pending; rest != nil {
		p.pending = nil
		e = w.acts(p, rest)
	}
	for e != nil && !p.exited {
		var effs []Effect
		p.s, effs = Step(p.s, e)
		w.steps++
		grammar(w.t, effs)
		e = nil
		switch x := effs[0].(type) {
		case Read:
			found := make([][]byte, len(x.Keys))
			for i, k := range x.Keys {
				found[i] = w.read(p, k)
			}
			e = ReadResult{Found: found}
		case Commit:
			outs := make([]Outcome, len(x.Writes))
			for i, wr := range x.Writes {
				outs[i] = w.write(p, wr)
			}
			e = Committed{Outcomes: outs}
		default:
			e = w.acts(p, acts(effs))
		}
	}
}

// acts performs a batch of executes. It returns the next event for p, or
// nil if p now waits on another proc (the rest of the batch is kept).
func (w *world) acts(p *proc, as []Action) Event {
	for i, a := range as {
		rest := as[i+1:]
		switch x := a.(type) {
		case SendStamp:
			w.queue = append(w.queue, queued{p.parent, Executed{Obs: []byte(x.Stamp)}})
			return nil // waits for its release
		case Acknowledge:
			parent := p.parent
			w.queue = append(w.queue, queued{parent, Executed{}})
			if x.Err != "" {
				continue
			}
			if len(rest) > 0 {
				w.t.Fatalf("acknowledge is the last act of its batch: %#v", as)
			}
			return Executed{} // custody acknowledged: it goes on
		case Run:
			return Executed{Obs: w.target(p.s.Self, x.Stdin)}
		case Spawn:
			child := &proc{s: Instance{Floor: p.s.Floor, Phase: Intake}, parent: p}
			p.kids = append(p.kids, child)
			under := Object{Key: x.Under, Raw: w.objects["horizons/"+x.Under]}
			w.queue = append(w.queue, queued{child, Received{Under: under, In: x.Out}})
			return nil // waits for the child's stamp
		case ReleaseChild:
			child := p.kids[x.Index]
			child.stack = p.stack
			if x.NewFrame != "" {
				child.stack = append([]string{x.NewFrame}, p.stack...)
			}
			// The release waits for the child's acknowledgment, which comes
			// back as p's next event; the rest of the batch then runs.
			p.pending = rest
			w.queue = append(w.queue, queued{child, Executed{Obs: []byte(x.Horizon), Handles: len(child.stack)}})
			return nil
		case KillSuccessors:
			for _, k := range p.kids {
				k.exited = true
			}
		case Pop:
			p.stack = p.stack[1:]
		case Park:
			w.t.Fatal("no hosts in this world")
		case Exit:
			p.exited, p.err = true, x.Err
			if x.Err != "" && !strings.Contains(x.Err, "already") {
				w.t.Logf("exit: %s", x.Err)
			}
			return nil
		}
	}
	return Executed{}
}

func (w *world) read(p *proc, k Key) []byte {
	switch k := k.(type) {
	case ObjectKey:
		return w.objects[k.Kind+"/"+k.Hash]
	case SlotsKey:
		f := w.frame(p.stack[1])
		slots, ok := f.slots[k.Round]
		if !ok {
			return nil
		}
		names := make([]string, 0, len(slots))
		for n := range slots {
			names = append(names, n)
		}
		sort.SliceStable(names, func(i, j int) bool { return slots[names[i]] < slots[names[j]] })
		b, _ := json.Marshal(names)
		return b
	case EmissionKey:
		if e, ok := w.frame(k.Frame).emitted[k.Round]; ok {
			return []byte(e)
		}
	case ResultKey:
		if r := w.frame(k.Frame).result; r != "" {
			return []byte(r)
		}
	}
	return nil
}

func (w *world) write(p *proc, wr Write) Outcome {
	exclusive := func(m map[string]string, k, v string) Outcome {
		if _, ok := m[k]; ok {
			return Lost
		}
		m[k] = v
		return Created
	}
	switch x := wr.(type) {
	case ObjectWrite:
		k := x.Kind + "/" + record.Sum(x.Bytes)
		if old, ok := w.objects[k]; ok {
			if string(old) != string(x.Bytes) {
				return Conflict
			}
			return Identical
		}
		w.objects[k] = x.Bytes
		return Created
	case AcceptanceWrite:
		return exclusive(w.accepts, x.Stamp, x.Horizon)
	case FrameWrite:
		if _, ok := w.frames[x.Stamp]; ok {
			return Lost
		}
		w.frame(x.Stamp)
		return Created
	case EmissionWrite:
		return exclusive(w.frame(p.stack[0]).emitted, x.Round, x.Hash)
	case SlotWrite:
		f := w.frame(p.stack[0])
		if f.slots[x.Round] == nil {
			f.slots[x.Round] = map[string]int{}
		}
		if _, ok := f.slots[x.Round][x.Child]; ok {
			return Lost
		}
		f.slots[x.Round][x.Child] = x.Pos
		return Created
	case ResultWrite:
		f := w.frame(p.stack[0])
		if f.result != "" {
			return Lost
		}
		f.result = x.Stamp
		return Created
	case ClaimWrite:
		return exclusive(w.frame(p.stack[1]).claims, x.Round, x.Stamp)
	}
	w.t.Fatalf("unknown write %T", wr)
	return Conflict
}

// grammar is the loop's shape (41, V2) on one Step: one read, one commit, or
// a batch of executes; an exit is the last act; and only an act that leaves
// the failure domain is forced.
func grammar(t *testing.T, effs []Effect) {
	t.Helper()
	if len(effs) == 0 {
		t.Fatal("a step with no effect: the instance would hang")
	}
	switch effs[0].(type) {
	case Read, Commit:
		if len(effs) != 1 {
			t.Fatalf("a read or a commit is answered before anything else: %#v", effs)
		}
		return
	}
	for i, e := range effs {
		x, ok := e.(Execute)
		if !ok {
			t.Fatalf("executes end a cycle: nothing follows them in it: %#v", effs)
		}
		if _, exit := x.Act.(Exit); exit && i != len(effs)-1 {
			t.Fatalf("an exit is the last act: %#v", effs)
		}
		switch x.Act.(type) {
		case Run, Park:
			if !x.Force {
				t.Fatalf("an act that leaves the failure domain is forced (40): %#v", x)
			}
		default:
			if x.Force {
				t.Fatalf("a handshake is not forced (40): %#v", x)
			}
		}
	}
}

// The whole round trip, through the in-memory world: the operator invokes the
// agent, which calls echo, waits, resumes with what echo said, and returns.
// Every Step on the way keeps the loop's grammar.
func TestALineageKeepsTheLoopsGrammar(t *testing.T) {
	agent := func(h horizon.Horizon, stdin []byte) []byte {
		var c Continuation
		if err := json.Unmarshal(stdin, &c); err != nil {
			t.Fatalf("agent stdin: %s", stdin)
		}
		if len(c.Steps) == 0 {
			return []byte(`{"to":"echo","payload":{"text":"hi"}}`)
		}
		said, _ := json.Marshal("echo said " + string(c.Steps[0].Value))
		return []byte(fmt.Sprintf(`{"to":%q,"payload":%s}`, c.Frame, said))
	}
	echo := func(h horizon.Horizon, stdin []byte) []byte {
		var in struct{ Text string }
		_ = json.Unmarshal(stdin, &in)
		return []byte(fmt.Sprintf(`{"to":%q,"payload":%q}`, h.Frame, in.Text))
	}
	w := newWorld(t, func(h horizon.Horizon, stdin []byte) []byte {
		if h.Surface.Continuation {
			return agent(h, stdin)
		}
		return echo(h, stdin)
	})
	r := w.boot(both, `{"to":"agent","payload":{"text":"hi"}}`)
	if r.State != record.Return || string(r.Payload) != `"echo said \"hi\""` {
		t.Fatalf("the operator's frame is discharged by the agent's return: %+v", r)
	}
}

// Fan-out through the world: one emission of two calls, one round of two
// slots, resolved once, the waiter resumed with both steps in order.
func TestAFanOutLineageKeepsTheLoopsGrammar(t *testing.T) {
	w := newWorld(t, func(h horizon.Horizon, stdin []byte) []byte {
		if !h.Surface.Continuation {
			var in struct{ Text string }
			_ = json.Unmarshal(stdin, &in)
			return []byte(fmt.Sprintf(`{"to":%q,"payload":%q}`, h.Frame, in.Text))
		}
		var c Continuation
		_ = json.Unmarshal(stdin, &c)
		if len(c.Steps) == 0 {
			return []byte(`[{"to":"echo","payload":{"text":"b"}},{"to":"echo","payload":{"text":"a"}}]`)
		}
		var said []string
		for _, s := range c.Steps {
			said = append(said, string(s.Value))
		}
		b, _ := json.Marshal(strings.Join(said, " "))
		return []byte(fmt.Sprintf(`{"to":%q,"payload":%s}`, c.Frame, b))
	})
	r := w.boot(both, `{"to":"agent","payload":{"text":"hi"}}`)
	if r.State != record.Return || string(r.Payload) != `"\"b\" \"a\""` {
		t.Fatalf("the agent resumes once with both steps, in emission order: %+v", r)
	}
}
