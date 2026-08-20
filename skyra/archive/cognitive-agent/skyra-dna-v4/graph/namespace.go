package graph

import (
	"fmt"

	"skyra-v1/src/skyra/spacetime/kernel"
)

// A Namespace is an agent's base node: what it is called, and the root of everything it
// can reach. IT IS THE GRAPH — a filesystem is its root inode, and you do not hold a
// *Filesystem, you hold `/`.
//
// It also holds this agent's Behavior, which makes a Namespace approximately "the agent."
// That is a convenience, not a claim: name+reach and methods could be split later if
// something needs one without the other. What must NOT change is that the behavior never
// leaves — see Mount.
//
// APPEND-ONLY. Nodes and observations are added and never rewritten or removed, which is
// what lets an execution walk the graph while learning happens: a stack cannot be standing
// on something that ceased to exist. Single-threaded for now (specs/2026-08-03/graph.md).
type Namespace struct {
	// Name is what this agent is called, and what another agent mounts it by.
	Name string

	// root is `/` — see Root.
	root *Node

	// behavior is this agent's methods. It is coupled to this agent's nodes and to
	// nothing else, ever.
	behavior Behavior

	// mounts are the boundaries this agent may cross, keyed by LOCAL name. Granted at
	// construction — an agent does not mount for itself, because granting yourself
	// reach is the self-selection failure.
	mounts map[string]*Mount

	// index is a derived, rebuildable cache — not structure. The base node is
	// authoritative; this exists so the observation worker need not walk from root to
	// find what an observation bears on. Same status dnav3 gives BuildResolutionIndex:
	// deletable, rebuildable, cheap to be wrong about.
	nodes map[NodeID]*Node
	obs   map[ObservationID]Observation

	// runner holds this agent's live executions.
	//
	// ONE STACK (specs/2026-08-03/execution.md, 08-06): one subgraph is active at a time,
	// connected, entered at a root. So this holds at most one, and the runner's capacity,
	// eviction, and round-robin are unreachable — see kernel/runner.go.
	//
	// Waves never block (actions are residual), so an impulse arriving mid-wave is not a
	// case that arises single-threaded. When an impulse queue exists it must be FIFO:
	// admission may not consult attention.
	runner *kernel.Runner

	waveSeq   int
	expectSeq int

	// Gates is INSTRUMENTATION, not model state. EXPECT runs ungated
	// (specs/2026-08-03/behaviors.md, 08-05); this records what the gate WOULD have said
	// so the trigger can be calibrated from data rather than from argument.
	//
	// Unbounded on purpose — these runs are disposable, and the series is the point.
	Gates []GateVerdict
}

// A GateVerdict is one evaluation of the gap condition, recorded and not enforced.
//
// What is measurable TODAY is the aggregate that workers.md already names — *"a pattern of
// low confidence across whatever was in view"* — expressed as the share of the frontier that
// found no home. When a boundary exists, Unplaced becomes confidence-weighted rather than a
// bare count, and nothing else here changes.
type GateVerdict struct {
	WaveID   string
	Node     NodeID
	Frontier int     // candidates in view
	Placed   int     // how many found a home
	Residual int     // how many did not
	Unplaced float64 // Residual / Frontier — the aggregate signal
	Fires    bool    // what the gate WOULD have said
	Reason   string
}

// GapThreshold is the share of an unplaceable frontier that would fire the gate.
//
// 1.0 — placed NOTHING at all — is the most conservative reading of "could not place it,"
// and it is a guess. It is a variable so a run can sweep it.
var GapThreshold = 1.0

// recordGate evaluates the gap condition and files the verdict. It never gates anything.
func (n *Namespace) recordGate(waveID string, node NodeID, frontier, residual int) GateVerdict {
	v := GateVerdict{
		WaveID: waveID, Node: node,
		Frontier: frontier, Residual: residual, Placed: frontier - residual,
	}
	switch {
	case frontier == 0:
		v.Reason = "nothing in view — the gate cannot speak"
	default:
		v.Unplaced = float64(residual) / float64(frontier)
		v.Fires = v.Unplaced >= GapThreshold
		if v.Fires {
			v.Reason = "read against the frontier and placed none of it"
		} else {
			v.Reason = "placed some of what was in view"
		}
	}
	n.Gates = append(n.Gates, v)
	return v
}

func NewNamespace(name string, b Behavior) *Namespace {
	return &Namespace{
		Name:     name,
		behavior: b,
		mounts:   map[string]*Mount{},
		nodes:    map[NodeID]*Node{},
		obs:      map[ObservationID]Observation{},
		runner:   &kernel.Runner{},
	}
}

// SetCapacity bounds how many executions may be live. Finite attention as a RESOURCE
// LIMIT: at capacity an arrival evicts rather than being refused, so ingress never closes.
func (n *Namespace) SetCapacity(c int) { n.runner.Capacity = c }

// Spawn starts an execution at one of this agent's own nodes and admits it to the runner.
// Admission never fails; the returned stack is live but has not stepped.
//
// The evicted stack, if any, is returned so the caller can see what was displaced — its
// transcript is valid and says how far it got.
func (n *Namespace) Spawn(node *Node, impulse any) (*kernel.Stack, *kernel.Stack) {
	h := n.Open(node)
	if h == nil {
		return nil, nil
	}
	n.waveSeq++
	h.WaveID = fmt.Sprintf("%s/wave:%d", n.Name, n.waveSeq)
	s := kernel.NewStack(h.WaveID, h, h, impulse)
	return s, n.runner.Admit(s)
}

// Step advances one of this agent's executions by one action. Round-robin across every
// live stack — deliberately the dumbest policy, because scheduling by trust or activation
// would hand the controller authority over which executions make progress.
func (n *Namespace) Step() bool { return n.runner.Step() }

// RunAll drains every live execution, interleaved.
func (n *Namespace) RunAll() { n.runner.RunAll() }

// Live is how many executions this agent currently holds.
func (n *Namespace) Live() int { return n.runner.Live() }

// Root is `/` — the expectation everything else hangs beneath, and the ordinary place an
// execution begins.
//
// # The root is a STANDING expectation
//
// It never closes, and that is what makes it a root rather than a very large task. The
// world's instruments decide it rather than any label — see
// cognitive-agent/skyra-dna-v3/affordance.go:
//
//	a target is STANDING iff none of its triples have an exact channel.
//	the purpose is what makes the being MOVE; it is not what makes it LEARN.
//
// It still carries a perceive like everything else. GRADED IS NOT CLOSED — support is
// assigned to it normally; it simply never reaches P → +1, because a standing claim has no
// exact channel to close it.
//
// A standing root is also why an execution can outlive any single turn: a root that never
// closes never stops promoting, so the wave ends on its deadline rather than on completion.
func (n *Namespace) Root() (*Node, bool) {
	if n == nil || n.root == nil {
		return nil, false
	}
	return n.root, true
}

// Open couples one of THIS agent's nodes with THIS agent's behavior. It is the entry point
// into the graph, and the only place a Space is made.
//
//	THE COUPLING IS CHOSEN BY THE AGENT. NEVER BY THE DATA.
//
// Nothing forces a walk to start at Root — opening any node is `chroot`, not a violation,
// and it is how a resumed execution re-enters partway down.
func (n *Namespace) Open(node *Node) *Handle {
	return n.OpenWith(node, n.behavior)
}

// OpenWith couples a node to a DIFFERENT behavior of this agent's. It is how one phase of
// a turn hands off to the next: same node, different methods, which is exactly the case
// BehaviorID exists to tell apart in the transcript (questions.md Q7).
func (n *Namespace) OpenWith(node *Node, b Behavior) *Handle {
	if n == nil || node == nil || b == nil {
		return nil
	}
	return Open(node, b)
}

// Ground creates `/` — the only node with no incoming edge, because it is the only one
// nothing grounds. Everything else arrives through AddExpectation or AddAction, carrying the
// observation that is its edge.
//
// The root is STANDING: no exact channel, so it never closes. It still carries a perceive,
// because graded is not closed.
func (n *Namespace) Ground(label string, perceive Act) *Node {
	if n == nil || n.root != nil {
		return n.root
	}
	perceive.Perceive = true
	root := &Node{ID: NodeID(n.Name + ":/"), Label: label}
	n.nodes[root.ID] = root
	n.root = root
	n.attachPerceive(root, perceive, Observation{
		ID:      ObservationID(string(root.ID) + "#constitutive"),
		Support: 0,
	})
	return root
}

// AddExpectation creates an expectation beneath a parent, WITH its perceive, WITH the observation
// that is the edge grounding it. All three or none.
//
//	AN UNFALSIFIABLE EXPECTATION IS UNCONSTRUCTIBLE.
//
// specs/2026-08-03/graph.md promotes the falsifiability invariant from a check to a
// constitutive rule. This is where that becomes true: there is no call that produces an
// expectation without a perceive, so there is no such state to detect.
//
// The perceive's own edge carries the SAME observation that licensed the expectation. The
// thing that made the agent form this expectation is what made it commit to looking, so
// they are not separately evidenced.
func (n *Namespace) AddExpectation(parent NodeID, name, label string, perceive Act, via Observation) (*Node, error) {
	p, err := n.directory(parent)
	if err != nil {
		return nil, err
	}
	perceive.Perceive = true
	node := &Node{ID: childID(parent, name), Label: label}
	if _, taken := n.nodes[node.ID]; taken {
		return nil, fmt.Errorf("%q already exists", node.ID)
	}
	n.nodes[node.ID] = node
	n.attachPerceive(node, perceive, via)
	n.link(p, name, node, via)
	return node, nil
}

// AddAction attaches an action beneath an expectation. Actions are leaves and own nothing.
func (n *Namespace) AddAction(parent NodeID, name string, a Act, via Observation) (*Node, error) {
	p, err := n.directory(parent)
	if err != nil {
		return nil, err
	}
	node := &Node{ID: childID(parent, name), Label: name, Act: &a}
	if _, taken := n.nodes[node.ID]; taken {
		return nil, fmt.Errorf("%q already exists", node.ID)
	}
	n.nodes[node.ID] = node
	n.link(p, name, node, via)
	return node, nil
}

// directory resolves a parent and refuses to attach beneath a FILE.
//
//	A FILE NEVER BECOMES A DIRECTORY.
//
// This answers specs/2026-08-03/questions.md Q2. Kind is derived from composition, so
// nothing structural would stop an entry appearing under an action and silently promoting
// it to an expectation. Unix cannot do that, and neither can this: an action that needs
// decomposing was never an action, and the right move is to build an expectation instead.
// Rejecting the growth is what keeps the derivation honest.
func (n *Namespace) directory(id NodeID) (*Node, error) {
	p, ok := n.nodes[id]
	if !ok {
		return nil, fmt.Errorf("no node %q in namespace %q", id, n.Name)
	}
	if p.IsFile() {
		return nil, fmt.Errorf("%q is an action; actions are leaves and cannot be grown into expectations", id)
	}
	return p, nil
}

// link appends the entry and the observation that IS that entry. They are written together
// because an edge without an observation is not a thing this graph can hold — the structure
// IS the evidence.
func (n *Namespace) link(parent *Node, name string, child *Node, via Observation) {
	if via.ID == "" {
		via.ID = ObservationID(string(parent.ID) + "→" + string(child.ID))
	}
	if _, written := n.obs[via.ID]; !written {
		n.obs[via.ID] = via // immutable once written
	}
	parent.Entries = append(parent.Entries, Entry{Name: name, Target: child.ID, Via: via.ID})
}

// childID joins a parent path and an entry name. The root already ends in "/", so a naive
// concatenation would produce "skyra://respond" — the same reason path.Join exists.
func childID(parent NodeID, name string) NodeID {
	sep := "/"
	if len(parent) > 0 && parent[len(parent)-1] == '/' {
		sep = ""
	}
	return NodeID(string(parent) + sep + name)
}

func (n *Namespace) attachPerceive(owner *Node, a Act, via Observation) {
	p := &Node{ID: childID(owner.ID, "@perceive"), Label: "did it happen", Act: &a}
	n.nodes[p.ID] = p
	n.link(owner, "@perceive", p, via)
}

func (n *Namespace) Node(id NodeID) (*Node, bool) {
	node, ok := n.nodes[id]
	return node, ok
}

func (n *Namespace) Observation(id ObservationID) (Observation, bool) {
	o, ok := n.obs[id]
	return o, ok
}

// Len is how many nodes this namespace holds.
func (n *Namespace) Len() int { return len(n.nodes) }

// A Mount is ONE DIRECTION of a boundary, held only by the side that crosses it.
//
// Note what is absent: there is no field for the target's behavior, and there cannot be.
// A crossing spawns a wave AT the target, and the target couples its own nodes with its own
// methods on its own side. B's behavior never travels.
//
// There is no shared edge. A's record of B and B's record of A are separate objects with
// separate histories, and they are allowed to disagree — A may find B reliable while B
// finds A abusive. Both true, neither derivable from the other. A shared object would force
// agreement and destroy the asymmetry.
type Mount struct {
	// As is the LOCAL name — what this agent calls the target.
	As string

	// Target is the target namespace's own name.
	Target string

	// send is the transport. In a situated world the world resolves it, which is what
	// makes delivery a real process that can fail rather than a lookup that cannot.
	send Send
}

// Send delivers an impulse toward a named namespace and reports what happened to the
// SEND — never what the target thought about it.
type Send func(target string, impulse any) Receipt

// A Receipt is a fact about this agent's own action, which is the only reason it may come
// back synchronously. dnav3's rule, from observation.go:
//
//	LEGITIMATE   "a wave of mine passed here at tick N" — a fact about my own history.
//	FORBIDDEN    "there is a trace toward the kitchen, therefore food is in the kitchen"
//
// So a Receipt may say delivered, refused, or lost. It may NEVER carry what the target
// concluded — that has to arrive later as an observation, and be graded like anything else.
type Receipt struct {
	Target    string
	WaveID    string
	Delivered bool

	// Refused is the target declining. That is the target's decision alone, never a
	// negotiation — A spawns, B refuses, A observes. Repeated refusal becomes a low
	// edge weight in A's graph: a LEARNED fact about B rather than a permission bit
	// somebody maintains.
	Refused bool
	Reason  string
}

// Grant wires a boundary this agent may cross. It is called by whatever CONSTRUCTS the
// agent, never by the agent itself — granting yourself reach is the same failure as a node
// choosing its own behavior.
func (n *Namespace) Grant(as, target string, send Send) *Mount {
	m := &Mount{As: as, Target: target, send: send}
	n.mounts[as] = m
	return m
}

// Cross spawns a wave at a mounted namespace and returns the receipt for the send.
//
// It does NOT return what the target realized, and it does not wait for it. A's execution
// continues and finishes without B's answer; B's answer arrives later as an observation.
// Waiting here would make A the collector of its own send, which is the shape the
// Perception Invariant forbids.
func (n *Namespace) Cross(as string, impulse any) (Receipt, error) {
	m, ok := n.mounts[as]
	if !ok {
		// Not a permission failure — an unmounted name is simply not a name here.
		// Reach was decided at construction; there is nothing to check at runtime.
		return Receipt{}, fmt.Errorf("namespace %q has no mount named %q", n.Name, as)
	}
	if m.send == nil {
		return Receipt{Target: m.Target}, fmt.Errorf("mount %q has no transport", as)
	}
	return m.send(m.Target, impulse), nil
}

// Accept is the receiving side. B decides — alone — whether an arriving wave gets a Space,
// couples it with B's OWN behavior, and admits it to B's own runner as a NEW execution.
//
// A wave never spans agents. The arriving impulse starts a fresh stack here, with its own
// WaveID, and the sender records that WaveID as the edge of the wave tree. Correlation is
// explicit rather than inferred.
//
// Refusal is an outcome, not a negotiation. The error returned here is what the sender
// observes as a refused Receipt.
func (n *Namespace) Accept(from string, impulse any) (*kernel.Stack, error) {
	if n == nil {
		return nil, fmt.Errorf("no namespace")
	}
	root, ok := n.Root()
	if !ok {
		return nil, fmt.Errorf("namespace %q has no root to enter", n.Name)
	}
	s, _ := n.Spawn(root, impulse)
	return s, nil
}
