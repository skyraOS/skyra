// Package tree is the functional core of one tool trees: 55's decisions,
// ported one to one from controller/apprentice/invariants/Invariants/Core.lean
// and checked against every vector the Lean writes (56, "Decisions").
//
// Fields are unexported, so Identify, TypeOf, ApprenticeOf, Observe and
// RootObservation are the only constructors: Go's form of the private
// constructor. A Store changes only through See and Rewrite. The package is
// pure: it touches no OS, tree or network (TestCoreIsPure).
package tree

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"
)

// --- identity ---

// A Binary is a binary by its hash, with its scheme: `cdhash:` and 40 hex
// digits on macOS, `sha256:` and 64 on Linux.
type Binary string

// A Schema is a shape by its canonical JSON, as contract.Normal writes it.
type Schema string

func hexDigits(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for i := 0; i < len(s); i++ {
		if c := s[i]; !('0' <= c && c <= '9' || 'a' <= c && c <= 'f') {
			return false
		}
	}
	return true
}

// WellFormed: a known scheme and its digest.
func (b Binary) WellFormed() bool {
	s := string(b)
	if rest, ok := strings.CutPrefix(s, "cdhash:"); ok {
		return hexDigits(rest, 40)
	}
	if rest, ok := strings.CutPrefix(s, "sha256:"); ok {
		return hexDigits(rest, 64)
	}
	return false
}

// WellFormed is the one check on a shape, wherever one is taken. Non-empty
// for now; checking that it is canonical goes here.
func (s Schema) WellFormed() bool { return s != "" }

// A Contract is a binary and the shape of its input. Every one is made by
// Identify.
type Contract struct {
	binary Binary
	shape  Schema
}

func (c Contract) Binary() Binary { return c.binary }
func (c Contract) Shape() Schema  { return c.shape }

// Identify is the one way to a contract. It refuses a binary without a known
// scheme and digest, or an empty shape.
func Identify(b Binary, s Schema) (Contract, bool) {
	if !b.WellFormed() || !s.WellFormed() {
		return Contract{}, false
	}
	return Contract{b, s}, true
}

// An ObservationType is the contract an observation came back through, and
// its own output shape. Every one is made by TypeOf.
type ObservationType struct {
	contract Contract
	output   Schema
}

func (t ObservationType) Contract() Contract { return t.contract }
func (t ObservationType) Output() Schema     { return t.output }

// TypeOf is the one way to an observation type. It refuses an output shape
// that is not well formed.
func TypeOf(c Contract, output Schema) (ObservationType, bool) {
	if !output.WellFormed() {
		return ObservationType{}, false
	}
	return ObservationType{c, output}, true
}

// --- observe ---

// The Apprentice is one binary and its contracts, each of that binary and
// none twice. Observe and Root are privileged: never loaded into a reach.
// Description and Reach are run by delegate's own CEL, before the model's
// point: never loaded either. Every one is made by ApprenticeOf.
type Apprentice struct {
	binary            Binary
	observe           Contract
	root              Contract
	exit              Contract
	ret               Contract
	effectObservation Contract
	invoke            Contract
	delegate          Contract
	goal              Contract
	description       Contract // create_description
	reach             Contract // package_reach
}

// ApprenticeOf is the one way to an apprentice. It refuses a contract of
// another binary, or one contract in two places.
func ApprenticeOf(b Binary, observe, root, exit, ret, effectObservation, invoke, delegate, goal, description, reach Contract) (Apprentice, bool) {
	a := Apprentice{b, observe, root, exit, ret, effectObservation, invoke, delegate, goal, description, reach}
	cs := a.Contracts()
	for i, c := range cs {
		if c.binary != b || slices.Contains(cs[:i], c) {
			return Apprentice{}, false
		}
	}
	return a, true
}

func (a Apprentice) Binary() Binary              { return a.binary }
func (a Apprentice) Observe() Contract           { return a.observe }
func (a Apprentice) Root() Contract              { return a.root }
func (a Apprentice) Exit() Contract              { return a.exit }
func (a Apprentice) Return() Contract            { return a.ret }
func (a Apprentice) EffectObservation() Contract { return a.effectObservation }
func (a Apprentice) Invoke() Contract            { return a.invoke }
func (a Apprentice) Delegate() Contract          { return a.delegate }
func (a Apprentice) Goal() Contract              { return a.goal }
func (a Apprentice) Description() Contract       { return a.description }
func (a Apprentice) Reach() Contract             { return a.reach }

func (a Apprentice) Contracts() []Contract {
	return []Contract{a.observe, a.root, a.exit, a.ret, a.effectObservation, a.invoke, a.delegate, a.goal,
		a.description, a.reach}
}

// Mine: a contract of the apprentice's binary.
func (a Apprentice) Mine(c Contract) bool { return c.binary == a.binary }

// A Kind is what a node is, decided from its reach.
type Kind uint8

const (
	_            Kind = iota
	KindEffect        // no contract in reach: it writes
	KindAct           // one: it acts, invoking it
	KindDelegate      // two or more: it delegates
)

func (k Kind) String() string {
	switch k {
	case KindEffect:
		return "effect"
	case KindAct:
		return "act"
	case KindDelegate:
		return "delegate"
	}
	return "kind?"
}

// An Observation is its type and the observations it was made from. Observe
// and RootObservation are the only ways to make one.
type Observation struct {
	typ     ObservationType
	members []Observation
}

func (o Observation) Type() ObservationType { return o.typ }

func (o Observation) Members() []Observation { return slices.Clone(o.members) }

// An effect has no members; a computation has members.
func (o Observation) IsEffect() bool      { return len(o.members) == 0 }
func (o Observation) IsComputation() bool { return len(o.members) != 0 }

// Equal is structural, as Lean's decidable equality is.
func (o Observation) Equal(p Observation) bool {
	return o.typ == p.typ && slices.EqualFunc(o.members, p.members, Observation.Equal)
}

func contains(xs []Observation, o Observation) bool {
	return slices.ContainsFunc(xs, o.Equal)
}

func nodup(xs []Observation) bool {
	for i, x := range xs {
		if contains(xs[:i], x) {
			return false
		}
	}
	return true
}

// effectThrough is what an effect may come through, by the kind of node that
// sends it: an exit, from every kind; a goal, a description and a packaged
// reach, from a delegate; what an effect node wrote; what came back through
// an act node's contract, outside the apprentice.
func effectThrough(a Apprentice, k Kind, c Contract) bool {
	return c == a.exit || k == KindDelegate && (c == a.goal || c == a.description || c == a.reach) ||
		k == KindEffect && c == a.effectObservation || k == KindAct && !a.Mine(c)
}

// Admits is what Observe admits. No members: an effect, through what its
// kind may make one through. Members: a computation, only from a delegate,
// through delegate, none twice.
func Admits(a Apprentice, k Kind, t ObservationType, ms []Observation) bool {
	if len(ms) == 0 {
		return effectThrough(a, k, t.contract)
	}
	return k == KindDelegate && t.contract == a.delegate && nodup(ms)
}

// Observe is the one function that creates observations from an
// instruction. The caller never says what it sends: an effect or a
// computation is decided by what it has, and admitted by the kind of node
// that sent it.
func Observe(a Apprentice, k Kind, t ObservationType, ms []Observation) (Observation, bool) {
	if !Admits(a, k, t, ms) {
		return Observation{}, false
	}
	return Observation{t, slices.Clone(ms)}, true
}

// RootObservation is the root's endowment: an effect through root, with no
// instruction before it. Privileged: no node sends it.
func RootObservation(a Apprentice, t ObservationType) (Observation, bool) {
	if t.contract != a.root {
		return Observation{}, false
	}
	return Observation{t, nil}, true
}

// --- scope ---

// A Dir is which way an observation is going.
type Dir uint8

const (
	_ Dir = iota
	DirInvoke
	DirRet
)

// MayPass: everything passed is in the sender's scope.
func MayPass(sender, passed []Observation) bool {
	for _, p := range passed {
		if !contains(sender, p) {
			return false
		}
	}
	return true
}

// Receive is the receiver's scope after the edge. A child starts with only
// what it is passed; a parent keeps its scope and adds what is returned.
func Receive(d Dir, receiver, passed []Observation) []Observation {
	if d == DirInvoke {
		return slices.Clone(passed)
	}
	return append(slices.Clone(receiver), passed...)
}

// Lineage is an observation and every observation it was made from.
func Lineage(o Observation) []Observation {
	out := []Observation{o}
	for _, m := range o.members {
		out = append(out, Lineage(m)...)
	}
	return out
}

// NodeMayReturn: a node's return holds nothing in the lineage of what it was
// passed.
func NodeMayReturn(received, returned []Observation) bool {
	var seen []Observation
	for _, r := range received {
		seen = append(seen, Lineage(r)...)
	}
	for _, o := range returned {
		for _, x := range Lineage(o) {
			if contains(seen, x) {
				return false
			}
		}
	}
	return true
}

// MayReturn is what return may pass: something; and for an act or
// delegate node, nothing its model created directly. An effect node may
// return what it wrote. created is the frame's own: what its model made,
// never a field on an observation.
func MayReturn(k Kind, created, returned []Observation) bool {
	if len(returned) == 0 {
		return false
	}
	if k == KindEffect {
		return true
	}
	for _, r := range returned {
		if contains(created, r) {
			return false
		}
	}
	return true
}

// --- the two edges ---

// An InputBase is an input base type, by its hash: the binary invoked, and
// the return base types of what it was passed, in order. Every one is made
// by CreateInvoke.
type InputBase struct{ id string }

// A ReturnBase is a return base type, by its hash: the input base type it
// answers, and the shape of what came back. No guard is in it. Every one is
// made by CreateReturn, which takes an input base type, so every return
// answers an invoke.
type ReturnBase struct{ id string }

func (i InputBase) String() string  { return i.id }
func (r ReturnBase) String() string { return r.id }

// hashes is how the core hashes; the vector tests pass Lean's toy ones.
type hashes struct {
	input func(Binary, []ReturnBase) string
	ret   func(InputBase, Schema) string
}

// The domain tags. Each kind of id hashes its own tag first, so no two kinds
// of id share one, whatever their parts; the version keeps a changed format
// from matching old records.
const (
	inputTag      = "skyra/input/v1"
	returnTag     = "skyra/return/v1"
	refinementTag = "skyra/refinement/v1"
)

// canonical is the JSON of v as JSON.stringify writes it: no HTML escaping,
// no trailing newline.
func canonical(v any) []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		panic(err) // strings, lists and maps of strings always encode
	}
	return bytes.TrimSuffix(b.Bytes(), []byte("\n"))
}

func sum(v any) string {
	h := sha256.Sum256(canonical(v))
	return "sha256:" + hex.EncodeToString(h[:])
}

var sha = hashes{
	input: func(b Binary, passed []ReturnBase) string {
		ps := make([]string, len(passed))
		for i, p := range passed {
			ps[i] = p.id
		}
		return sum([]any{inputTag, string(b), ps})
	},
	ret: func(i InputBase, shape Schema) string {
		return sum([]any{returnTag, i.id, string(shape)})
	},
}

// CreateInvoke is down an edge (create_invoke): the one way to an input base
// type. It refuses a binary without a known scheme and digest.
func CreateInvoke(b Binary, passed []ReturnBase) (InputBase, bool) {
	return createInvoke(sha, b, passed)
}

func createInvoke(h hashes, b Binary, passed []ReturnBase) (InputBase, bool) {
	if !b.WellFormed() {
		return InputBase{}, false
	}
	return InputBase{h.input(b, passed)}, true
}

// CreateReturn is up an edge (create_return): the one way to a return base
// type. It takes the input base type it answers, and refuses a shape that is
// not well formed.
func CreateReturn(i InputBase, shape Schema) (ReturnBase, bool) {
	return createReturn(sha, i, shape)
}

func createReturn(h hashes, i InputBase, shape Schema) (ReturnBase, bool) {
	if !shape.WellFormed() {
		return ReturnBase{}, false
	}
	return ReturnBase{h.ret(i, shape)}, true
}

type Seen uint8

const (
	_ Seen = iota
	SeenSeen
	SeenUnseen
)

// Do: an input base type came down an edge. Seen when the system has been
// handed it before.
func Do(history []InputBase, i InputBase) Seen {
	if slices.Contains(history, i) {
		return SeenSeen
	}
	return SeenUnseen
}

// --- refinement types ---

// A RefinementID is a refinement type, by its hash: a return base type and
// its guard's code.
type RefinementID string

// A Guard is the default until the model writes one.
type Guard struct {
	written bool
	cel     string
}

func DefaultGuard() Guard           { return Guard{} }
func WrittenGuard(cel string) Guard { return Guard{true, cel} }

func (g Guard) Written() (string, bool) { return g.cel, g.written }

// RefinementOf is a refinement type's id: a return base type and its guard,
// the default as "default", a written one as {"written": cel}.
func RefinementOf(b ReturnBase, g Guard) RefinementID {
	var guard any = "default"
	if cel, ok := g.Written(); ok {
		guard = map[string]string{"written": cel}
	}
	return RefinementID(sum([]any{refinementTag, b.id, guard}))
}

// A Store is two levels, kept apart. known is the base level: the return
// base types the system recognizes, newest first. A hash not in it is
// unseen; it has no slot. guards is the refinement level: one slot per known
// return base type.
type Store struct {
	known  []ReturnBase
	guards map[ReturnBase]Guard
}

func EmptyStore() Store { return Store{} }

func (s Store) Known() []ReturnBase { return slices.Clone(s.known) }

// Guard is a return base type's slot: none for one the system does not
// recognize.
func (s Store) Guard(b ReturnBase) (Guard, bool) {
	g, ok := s.guards[b]
	return g, ok
}

func (s Store) set(b ReturnBase, g Guard) Store {
	guards := make(map[ReturnBase]Guard, len(s.guards)+1)
	for k, v := range s.guards {
		guards[k] = v
	}
	guards[b] = g
	return Store{s.known, guards}
}

// See: a return of return base type b came back. On unseen, b becomes
// known, and its slot is made with its one guard, the default.
func See(s Store, b ReturnBase) (Seen, Store) {
	if slices.Contains(s.known, b) {
		return SeenSeen, s
	}
	t := s.set(b, DefaultGuard())
	t.known = append([]ReturnBase{b}, s.known...)
	return SeenUnseen, t
}

// Rewrite: a guard broke, the model ran, and wrote another. It overwrites
// the one slot. An unknown return base type has no slot to overwrite.
func Rewrite(s Store, b ReturnBase, cel string) (Store, bool) {
	if !slices.Contains(s.known, b) {
		return Store{}, false
	}
	return s.set(b, WrittenGuard(cel)), true
}

// Refinement is a return base type's refinement type: its base and its one
// guard, hashed by refine.
func Refinement(refine func(ReturnBase, Guard) RefinementID, s Store, b ReturnBase) (RefinementID, bool) {
	g, ok := s.guards[b]
	if !ok {
		return "", false
	}
	return refine(b, g), true
}

// --- the guards ---

// A ReturnResult is the return guard's: what the apprentice does after a
// return comes back.
type ReturnResult uint8

const (
	_              ReturnResult = iota
	ReturnContinue              // take the next instruction; no model
	ReturnFallback              // an assumption broke; the model runs
)

// ReturnOf: an unseen return broke an assumption; a seen one continues only
// while its guard holds.
func ReturnOf(n Seen, guardHolds bool) ReturnResult {
	if n == SeenSeen && guardHolds {
		return ReturnContinue
	}
	return ReturnFallback
}

// An InvokeResult is the invoke guard's: what the core does when a child is
// invoked.
type InvokeResult uint8

const (
	_               InvokeResult = iota
	InvokeProcedure              // the invoke guard holds: the remembered procedure runs; no model
	InvokeModel                  // an assumption broke; the model runs
)

func InvokeOf(n Seen, guardHolds bool) InvokeResult {
	if n == SeenSeen && guardHolds {
		return InvokeProcedure
	}
	return InvokeModel
}

// --- what a node is ---

// counted is the reach, the apprentice's taken out.
func counted(a Apprentice, reach []Contract) []Contract {
	var out []Contract
	for _, c := range reach {
		if !a.Mine(c) {
			out = append(out, c)
		}
	}
	return out
}

func kind(cs []Contract) Kind {
	var distinct []Contract
	for _, c := range cs {
		if !slices.Contains(distinct, c) {
			distinct = append(distinct, c)
		}
	}
	switch len(distinct) {
	case 0:
		return KindEffect
	case 1:
		return KindAct
	}
	return KindDelegate
}

// Classify is what the core runs, in the observe phase. The apprentice's
// contracts count for neither the self check nor the kind. It refuses an
// agent whose own binary is in the rest of its reach, at every count.
func Classify(a Apprentice, self Contract, reach []Contract) (Kind, bool) {
	rest := counted(a, reach)
	for _, c := range rest {
		if c.binary == self.binary {
			return 0, false
		}
	}
	return kind(rest), true
}

// Loaded is the apprentice's contracts for a kind, loaded into its reach
// after the kind is known.
func Loaded(a Apprentice, k Kind) []Contract {
	switch k {
	case KindEffect:
		return []Contract{a.exit, a.ret, a.effectObservation}
	case KindAct:
		return []Contract{a.exit, a.ret, a.invoke}
	case KindDelegate:
		return []Contract{a.exit, a.ret, a.delegate, a.goal}
	}
	return nil
}

// --- pass time and provision time ---

// Passable: every contract passed is held by the parent.
func Passable(held, passed []Contract) bool {
	for _, p := range passed {
		if !slices.Contains(held, p) {
			return false
		}
	}
	return true
}

// An Offer is what an agent is provisioned with, from its kind: the
// contracts it may invoke, those it may only pass on, and the apprentice's
// contracts loaded for its kind.
type Offer struct {
	Invoke, PassOnly, Apprentice []Contract
}

func Offered(a Apprentice, k Kind, passed []Contract) Offer {
	switch k {
	case KindAct:
		return Offer{slices.Clone(passed), nil, Loaded(a, k)}
	case KindDelegate:
		return Offer{nil, slices.Clone(passed), Loaded(a, k)}
	}
	return Offer{nil, nil, Loaded(a, k)}
}
