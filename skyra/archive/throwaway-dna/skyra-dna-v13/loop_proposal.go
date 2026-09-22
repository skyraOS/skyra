// Package dna probes the proposed direct execution loop. Registry, validation,
// dispatch reporting, and CTU wiring are implemented. Scheduler execution and
// output binding remain explicit blockers; see boundary-crossings.md.
//
// Decisions carried from specs/2026-09-07 /main-loop.md:
//   - Scheduler owns registers and continuation state and is their single writer.
//   - Continuation Controller authorizes and dispatches without retaining run state.
//   - Drivers execute; registered interpretation normalizes their output into Moves.
//     Every Move, including Ret, is an invocation.
//   - Dispatch acknowledgement is separate from the driver's observation payload.
//
// Method signatures and event carriers below are a proposal of the agreed
// boundaries. Compilation proves only that these declarations fit together.
package dna

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v12/causal"
)

// Reuse only these leaf value types. v12's Move, Body, Contract, address topology,
// and runtime ownership are superseded and are deliberately not imported here.
type Observation = c.Observation
type Action = c.Name
type EntityID = c.EntityID
type HostID = c.HostID
type PossibilityID = c.PossibilityID

// MoveID and EngramID are the SAME identity: hash(Action + ordered Params).
// Neither routing nor invocation history enters that hash. Encoding is still open.
type MoveID struct{ Value string }
type EngramID = MoveID
type InvocationID struct{ Value string }

// Target is a wire reference, not the entity/host object graph. Address names the
// stable EntityAddress representation; Host selects its hosting environment.
// Network locators and representation implementation remain behind the driver.
type Target struct {
	Entity  EntityID
	Address string
	Host    HostID
}

// These carry the existing capability shape onto the revised address reference.
// Focus selects Horizon grants; it neither grants authority nor declares children.
// Signature encoding and trusted root configuration remain implementation seams.
type Capability struct {
	Target    Target
	Action    Action
	ExpiresAt int64
	Sig       c.Signature
}
type Horizon struct{ Capabilities []Capability }

// Move is the universal driver wire envelope. No register travels with it.
// Action is the sole semantic operation name; Params are ordered observations.
//
// PROPOSED correlation: ParentFd names a parent INVOCATION's resident register.
// This changes the older ParentFd *MoveID shape. Using MoveID alone would merge
// simultaneous uses of the same remembered operation. Whether to rename ParentFd
// to ParentInvocation is open; no separate numeric descriptor is introduced.
type Move struct {
	ID         MoveID
	Invocation InvocationID
	ParentFd   *InvocationID // nil for a root invocation
	Root       PossibilityID
	Target     Target
	Action     Action
	Params     []Observation
	Op         Opcode
	Horizon    Horizon
	Focus      []Capability
}

// Only Call and Ret are needed to review the first direct loop. This is not a
// claim that the full opcode set is closed. Wire encoding remains unspecified.
// The opcode interface is Scheduler's protocol for invocation-driven changes.
// Opcodes describe those semantics; they do not execute drivers themselves.
type Opcode interface{ opcode() }
type Call struct{}

// Ret has its OWN Move.Invocation. InReplyTo identifies the earlier invocation
// whose result this is; Move.ParentFd identifies the receiving parent register.
// These answer different questions and must not reuse Ret's own identity.
// Operation errors/timeouts are observations here, not dispatch errors below.
type Ret struct {
	InReplyTo InvocationID
	Value     Observation
}

func (Call) opcode() {}
func (Ret) opcode()  {}

// Contract is schema-only. RawMessage does not imply a selected schema dialect
// or validation algorithm. Parent dependency/contract checking stays stubbed.
type Contract struct{ Schema json.RawMessage }

// Register associates child invocations with their asynchronously returned
// observations, keyed by the child's InvocationID. Arrival alone does not establish
// that the return contract is satisfied. The map does not preserve declaration
// order; how to retain order if required remains OPEN.
type Register struct {
	Contract Contract
	Children map[InvocationID]RegisterEntry
}

// ChildInvocation records the invocation on its parent. The same invocation also
// has an independent Continuation entry in Scheduler's table, keyed by
// Move.Invocation. Its own register and lifecycle belong to that independent entry.
// InvocationID must equal Move.Invocation and the parent register's map key.
type ChildInvocation struct {
	InvocationID InvocationID
	Move         Move
}

// ChildObservation identifies the child being answered, using Ret.InReplyTo.
// InvocationID is not the returning Move's own, new invocation identity.
type ChildObservation struct {
	InvocationID InvocationID
	Observation  Observation
}

// The invocation creates an entry; its correlated return fills Observation later.
// Both child IDs must match the map key. Nil Observation means no result arrived.
type RegisterEntry struct {
	Invocation  ChildInvocation
	Observation *ChildObservation
}

// These are the current provisional local states, not wire opcodes or driver
// events. The distinction between Dispatched and InFlight, transition timing,
// initial state, and completion/retirement semantics remain OPEN. Zero is unset.
// Eligibility is not represented by a separate Ready state.
type State uint8

const (
	Dispatched State = iota + 1
	InFlight
	Faulted
)

// Continuation directly holds an invocation's Move, lifecycle, and register.
// This replaces the intermediate Fd wrapper. A parent's child entry records the
// child invocation and observation; the child's own register lives here under
// its InvocationID, rather than being nested again inside the parent's register.
type Continuation struct {
	Move     Move
	State    State
	Register Register
}

// Scheduler owns one resident table including registers and lifecycle. Keying by
// InvocationID separates live occurrences while Move.ID still selects the engram.
// No controller/driver receives a pointer into this map.
//
// DECIDED: one Write(OpcodeEvent) handles both register and lifecycle mutations.
// Move processing, validation failures, and dispatch outcomes have distinct event
// forms within that protocol. There is no separate OnDispatch write method.
// Driver-emitted Moves pass through Validate; Horizon is checked at dispatch.
//
// DECIDED: Write triggers private run(move), which traces affected dependencies
// and uses private selectMoves(move) before dispatch. Go reserves "select",
// hence the name selectMoves.
// Dispatch outcomes schedule further work only when their outcome warrants it.
//
// DECIDED: use one scheduler mutex for register/lifecycle access and selection.
// Concurrent Write calls are allowed; their mutations are serialized.
// A child Ret can arrive before or concurrently with its dispatch acknowledgement.
// A late acknowledgement must not regress state established by that return.
// Selection must account for handoff under the lock to prevent duplicate dispatch.
// Release the lock BEFORE calling CC or a driver: an immediate callback must be
// able to re-enter Scheduler without deadlocking. Pass detached Move snapshots
// across that boundary, not mutable references into scheduler-owned state.
// These notes establish the synchronization rule, not an implemented guarantee.
type Scheduler struct {
	mu            sync.Mutex
	controller    Controller
	continuations map[InvocationID]Continuation
}

// NewScheduler receives the ordinary CC instance created first at startup. CC is
// a stable dependency, not a global/static controller or a per-invocation object.
// Scheduler retains it; moves are supplied to Write and private run afterward.
func NewScheduler(cc Controller) *Scheduler {
	return &Scheduler{controller: cc, continuations: make(map[InvocationID]Continuation)}
}

// OpcodeEvent is the internal Scheduler protocol, not a replacement for the
// driver Move wire format. These proposed variants keep payloads distinct while
// giving CC a single write callback. Scheduler composes state changes according
// to the event and, for a Move, its Op. Exact opcode handler methods remain OPEN.
// The private marker specifies the variants; it is not an implementation of them.
type OpcodeEvent interface{ opcodeEvent() }

// A successfully validated Move is passed directly to Write. It is not yet
// authorized to execute. A Ret is still a new Move with its own invocation ID.
func (Move) opcodeEvent() {}

// ValidationFailure reports an unroutable or malformed Move through the SAME
// scheduler ingress, without submitting it as executable work. Scheduler records
// the failure and decides recovery. An unchanged invalid call must not be blindly
// redispatched. Missing/unusable invocation identity and producer correlation are
// still OPEN: invalid input cannot choose arbitrary existing scheduler entries.
type ValidationFailure struct {
	Move Move
	Err  error
}

func (ValidationFailure) opcodeEvent() {}

// DispatchResult answers two questions: did Horizon authorization pass, and did
// delivery to the driver succeed? Authorized distinguishes authorization denial
// from delivery failure. Nil Err requires Authorized and successful acceptance;
// it never establishes completion. A later driver result is a new Move.
// Uncertain remote delivery cannot safely imply "not run"; retry policy is OPEN.
type DispatchResult struct {
	Invocation InvocationID
	Authorized bool
	Err        error
}

func (DispatchResult) opcodeEvent() {}

// Write is the sole event ingress and owns both register and lifecycle changes.
// A Move's opcode determines invocation semantics: child admission records the
// invocation on its parent and creates its independent continuation; Ret records
// a correlated observation. Validate failures and dispatch outcomes record their
// state consequences through this same method, not a second write callback.
// Protect changes with mu, then run affected work when warranted. Dispatch success
// alone must not redispatch the acknowledged invocation. Late acknowledgements
// cannot regress state already established by a child return.
func (s *Scheduler) Write(event OpcodeEvent) error {
	// Admission and lifecycle policy are deliberately not guessed. No mutation
	// occurs before the missing execution context surface is resolved.
	return errors.New("v13 scheduler Write unfinished: context propagation and opcode lifecycle require decisions")
}

// selectMoves traces affected dependencies from move and selects eligible work.
// Caller holds mu; selection accounts for handoff before releasing that lock.
// Returned Moves must be detached snapshots, including slices and observation
// bytes. A batch is selection convenience, not a driver completion barrier.
func (s *Scheduler) selectMoves(move Move) []Move {
	panic("v13 selectMoves unfinished: eligibility and lifecycle require decisions")
}

// run receives the changed Move; controller is already supplied at startup.
// Lock around selectMoves, then UNLOCK before dispatching each selected Move via
// s.controller.Dispatch with s.Write as the outcome callback. CC.execute binds
// driver output to its source invocation, runs that driver's Egress, and sends
// normalized Moves through Validate, also with s.Write as the callback.
// For events identifying an existing invocation, Write obtains its resident Move
// as the dependency-tracing starting point; it never treats an acknowledgement
// as a fresh invocation.
// An immediate driver return can therefore safely re-enter Scheduler.
// No eligible work does not imply workflow completion. Context propagation and
// shutdown remain OPEN; cancellation cannot prove that remote work stopped.
func (s *Scheduler) run(move Move) error {
	return errors.New("v13 run unfinished: no submitted context is available through Write/run")
}

// Controller mediates between Scheduler's continuation state and DriverRegistry's
// endpoint state. It knows the OpcodeEvent protocol but owns neither mutable store.
// Both validation and dispatch report through the same supplied write callback.
// DECIDED: these callback/return signatures and private execute's signature below
// are accepted. Dispatch writes authorization/handoff outcomes; its returned error
// reports failure to deliver that report. Validate uses the same scheduler callback.
// execute returns immediate ingress acceptance/failure to Dispatch and retains
// invocation correlation in its output closure, not an executor struct.
type Controller interface {
	Validate(ctx context.Context, move Move, write func(OpcodeEvent) error) error
	Dispatch(ctx context.Context, move Move, write func(OpcodeEvent) error) error
}

// ContinuationController retains dependencies, not invocation or registry state.
// The separately owned registry may contain local drivers or remote proxies.
type ContinuationController struct {
	registry  *DriverRegistry
	authority Authority
}

// NewContinuationController is called before NewScheduler at startup. CTU owns
// the registry and supplies it here; Scheduler receives this same CC instance.
func NewContinuationController(registry *DriverRegistry, authority Authority) *ContinuationController {
	return &ContinuationController{registry: registry, authority: authority}
}

// Validate checks Move structure, target registration, and endpoint input schema.
// It does NOT check Horizon and cannot guarantee execution will complete. Success
// writes the Move; failure writes ValidationFailure. The returned error reports
// inability to process/deliver the report, not a second recovery notification.
// Parent dependency/return-contract evaluation remains on the continuation side.
func (cc *ContinuationController) Validate(ctx context.Context, move Move, write func(OpcodeEvent) error) error {
	if write == nil {
		return errors.New("nil scheduler write callback")
	}
	move = cloneMove(move)
	report := func(err error) error { return write(ValidationFailure{Move: move, Err: err}) }
	if ctx == nil {
		return report(errors.New("nil context"))
	}
	if err := ctx.Err(); err != nil {
		return report(err)
	}
	if err := validateStructure(move); err != nil {
		return report(err)
	}
	if cc == nil || cc.registry == nil {
		return report(errors.New("missing driver registry"))
	}
	registration, err := cc.registry.Lookup(move.Target.Address)
	if err != nil {
		return report(err)
	}
	// An absent schema imposes no input schema check for this probe. A declared
	// schema cannot be evaluated without a dialect and Params mapping decision.
	if len(registration.Schema) != 0 {
		return report(errors.New("v13 input schema validation unfinished: dialect and Params mapping unspecified"))
	}
	return write(move)
}

// Dispatch resolves the driver and checks Horizon at access, then calls
// execute(move, driver). It reports lookup/authorization failure or ingress
// acceptance/failure through write(DispatchResult). Its returned error describes
// failure to deliver that report, not a second dispatch outcome.
// Never hold Scheduler or registry locks across execute or write callbacks.
func (cc *ContinuationController) Dispatch(ctx context.Context, move Move, write func(OpcodeEvent) error) error {
	if write == nil {
		return errors.New("nil scheduler write callback")
	}
	move = cloneMove(move)
	outcome := DispatchResult{Invocation: move.Invocation}
	report := func(err error) error { outcome.Err = err; return write(outcome) }
	if ctx == nil {
		return report(errors.New("nil context"))
	}
	if err := ctx.Err(); err != nil {
		return report(err)
	}
	if cc == nil || cc.registry == nil {
		return report(errors.New("missing driver registry"))
	}
	registration, err := cc.registry.Lookup(move.Target.Address)
	if err != nil {
		return report(err)
	}
	if nilDependency(cc.authority) {
		return report(errors.New("missing authority"))
	}
	if err := cc.authority.Check(ctx, cloneMove(move)); err != nil {
		return report(err)
	}
	outcome.Authorized = true
	return report(cc.execute(ctx, move, registration.Driver, write))
}

// execute belongs to CC and is private to dispatch. No executor struct is needed.
// It creates a callback bound to a detached copy of the source Move and the exact
// driver instance selected for ingress. That callback can outlive this method:
//
//	driver.Ingress(move, output callback)
//	  output(raw) -> same driver.Egress(raw)
//	              -> bind each normalized Move to the source invocation
//	              -> cc.Validate(normalized, write)
//
// Invocation correlation lives in the closure, never a shared driver.currentID
// field. Scheduler remains the sole owner of registers and lifecycle state.
// Parser-produced identity, parent links, and Horizon are not trusted assertions:
// execute binds/checks them against the source context before validation. Forward
// versus back comes from the registered driver's interpretation; how that mapping
// fills routing/correlation fields is still OPEN. Every output Move gets its own
// InvocationID; a Ret answers prior work rather than reusing its identity.
//
// Return value is the driver's immediate ingress acceptance error. Dispatch uses
// it to report handoff. Later callback errors are returned to the driver and must
// be surfaced, not swallowed or recast as a failed initial handoff. Egress parse
// failures must reach Scheduler's failure protocol with bound source context;
// the exact error opcode is still OPEN. This is not an implemented recovery path.
func (cc *ContinuationController) execute(ctx context.Context, move Move, driver Driver, write func(OpcodeEvent) error) error {
	// Do not start a driver whose output cannot yet be safely bound and reported.
	return errors.New("v13 execute unfinished: output routing, identity encoding, and bound parse-failure protocol require decisions")
}

// DriverRegistry owns endpoint registration separately from CC and Scheduler.
// Address is the stable EntityAddress key; network route selection stays behind
// the driver. Registration and lookup synchronize access to this map.
//
// DECIDED: registration binds BOTH ingress and egress interpretation:
//
//	dispatch: Move -> registered ingress interpretation -> driver
//	output: driver output -> registered egress interpretation -> Validate -> Write
//
// Both interpretations are methods of the registered driver instance. Ingress
// interprets a Move for the endpoint; Egress interprets its raw output into Moves.
// CC.execute facilitates those methods and binds invocation context. Validate
// checks the normalized Moves before reporting through Scheduler's protocol.
// Scheduler still understands the opcode's state-changing semantics.
//
// A return-only endpoint's egress interpretation always produces a return from
// its observation. The output need not restate "back" each time. An LLM endpoint's
// interpretation can distinguish a tool request (forward) from an answer (back).
// Direction must be established by this explicit endpoint contract or a signal
// understood by its interpretation, never guessed from arbitrary payload content.
// Return-only is an endpoint property, NOT a human/LLM versus other-driver rule:
// compiled workflows and remote runtimes may also initiate child work.
//
// Runtime dispatch binds the output to the registered endpoint and originating
// invocation. Output does not choose its own parser, identity, or correlation.
// A back emission identifies the invocation answered through that binding while
// becoming a fresh Move with its own InvocationID. Interpretation grants no
// authority; the resulting work still faces Horizon authorization at dispatch.
//
// Registering one driver registers both methods together. There are no separate
// interpreter objects. The callback/return signatures below are agreed;
// output framing and asynchronous transport-failure recovery remain OPEN.
type DriverRegistry struct {
	mu      sync.RWMutex
	drivers map[string]DriverRegistration
}

func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{drivers: make(map[string]DriverRegistration)}
}

// Register establishes the endpoint implementation and its exposed schema.
// Registration permission/bootstrap policy and replacement semantics remain OPEN;
// the method must not silently replace an in-use binding.
func (r *DriverRegistry) Register(address string, registration DriverRegistration) error {
	if r == nil {
		return errors.New("nil driver registry")
	}
	if strings.TrimSpace(address) == "" {
		return errors.New("empty driver address")
	}
	if nilDependency(registration.Driver) {
		return errors.New("nil driver")
	}
	if len(registration.Schema) != 0 && !json.Valid(registration.Schema) {
		return errors.New("invalid schema JSON")
	}
	registration.Schema = slices.Clone(registration.Schema)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.drivers[address]; exists {
		return fmt.Errorf("driver address %q already registered", address)
	}
	if r.drivers == nil {
		r.drivers = make(map[string]DriverRegistration)
	}
	r.drivers[address] = registration
	return nil
}

// Lookup returns a registration snapshot, not mutable registry access. Schema
// bytes must be detached as well. Dispatch
// uses the selected binding through the handoff; validation is not a reservation.
func (r *DriverRegistry) Lookup(address string) (DriverRegistration, error) {
	if r == nil {
		return DriverRegistration{}, errors.New("nil driver registry")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	registration, exists := r.drivers[address]
	if !exists {
		return DriverRegistration{}, fmt.Errorf("driver address %q not registered", address)
	}
	registration.Schema = slices.Clone(registration.Schema)
	return registration, nil
}

// Authority checks established grants, never mints them. Checking at dispatch is
// necessary because a capability can expire while a Move waits in Scheduler.
// Root keys/policy are configuration of the checker, not controller run state.
type Authority interface {
	Check(ctx context.Context, move Move) error
}

// Driver owns exactly the two endpoint interpretation methods. The same instance
// is used for both directions of an execution. Many executions may share it;
// per-invocation correlation belongs to execute's bound callback.
// DECIDED: Ingress takes the output callback and returns an acceptance error;
// Egress returns candidate Moves or a parsing error. The output callback invokes
// Egress on that same driver, binds source context, then validates each Move.
type Driver interface {
	// Ingress receives a Move, interprets it for the endpoint, and starts work.
	// Nil error acknowledges acceptance only. output may be called later or
	// before Ingress returns; do not wait for child results to acknowledge.
	// The callback receives an opaque response string, not proof of safe data.
	// It must receive a parseable response unit, not arbitrary streaming chunks;
	// endpoint-specific buffering belongs in the driver. Callback errors must
	// be observed. Async transport failure/shutdown reporting remains OPEN.
	Ingress(ctx context.Context, move Move, output func(string) error) error

	// Egress interprets one response into zero or more candidate Moves. A list
	// supports parallel tool requests. These Moves still need execute's source
	// binding and Validate; parsing never grants authority. A return-only driver
	// produces Ret; other endpoints can express forward/back through their own
	// registered interpretation. Parsing errors are distinct from operation
	// errors, which can be successfully interpreted Ret observations.
	Egress(raw string) ([]Move, error)
}

// DriverRegistration keeps the endpoint schema in the registry so Driver itself
// has just Ingress and Egress. Schema is the EntityAddress input contract exposed
// to Validate, not a second interpretation object. This carrier is PROPOSED.
// Schema dialect and translation validation remain OPEN.
type DriverRegistration struct {
	Driver Driver
	Schema json.RawMessage
}

// CTU is the public mediation boundary for the kernel and other host wiring.
// Its implementation and component fields are private so callers cannot construct
// an empty concrete CTU or bypass it to mutate Scheduler's registers. A nil interface
// is still possible in Go; construction must return either a valid CTU or an error.
type CTU interface {
	Submit(ctx context.Context, move Move) error
	RegisterDriver(address string, registration DriverRegistration) error
}

type ctu struct {
	scheduler  *Scheduler
	controller Controller
	registry   *DriverRegistry
}

// NewCTU creates registry -> CC -> Scheduler in that order, registering the supplied
// root driver first. It must validate required dependencies before exposing CTU.
// Scheduler holds the same controller used by Submit; CC holds the CTU's registry.
// Kernel remains the root driver, not Scheduler's owner. No workflow starts here.
func NewCTU(rootAddress string, root DriverRegistration, authority Authority) (CTU, error) {
	if nilDependency(authority) {
		return nil, errors.New("nil authority")
	}
	registry := NewDriverRegistry()
	if err := registry.Register(rootAddress, root); err != nil {
		return nil, err
	}
	controller := NewContinuationController(registry, authority)
	return &ctu{scheduler: NewScheduler(controller), controller: controller, registry: registry}, nil
}

// Submit passes the Move to CC.Validate with Scheduler.Write as its callback.
func (ct *ctu) Submit(ctx context.Context, move Move) error {
	if ct == nil || ct.scheduler == nil || nilDependency(ct.controller) {
		return errors.New("uninitialized CTU")
	}
	return ct.controller.Validate(ctx, move, ct.scheduler.Write)
}

// RegisterDriver exposes the owned registry's registration seam through CTU.
// Access to this host API must follow the root's registration policy (still OPEN).
func (ct *ctu) RegisterDriver(address string, registration DriverRegistration) error {
	if ct == nil {
		return errors.New("nil CTU")
	}
	return ct.registry.Register(address, registration)
}

// Deferred deliberately: durable Engram's executable representation, filesystem
// Resolve -> continuation installation, registry replacement policy, and root-result delivery.
// The first trace can use already-registered drivers. It must not claim that this
// provides paging, compilation, Remember, or a completed end-to-end workflow.
// Review those surfaces separately before freezing an overnight implementation.

// nilDependency rejects typed nil interface dependencies as well as nil interfaces.
func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

// validateStructure checks only envelope invariants already stated by the proposal.
// ID hash verification, schema semantics, and parent contracts remain unfinished.
func validateStructure(move Move) error {
	if move.ID.Value == "" || move.Invocation.Value == "" {
		return errors.New("missing move or invocation identity")
	}
	if move.Target.Address == "" || move.Action.Value == "" {
		return errors.New("missing target address or action")
	}
	if move.ParentFd != nil && (move.ParentFd.Value == "" || *move.ParentFd == move.Invocation) {
		return errors.New("invalid parent invocation")
	}
	switch op := move.Op.(type) {
	case Call:
	case *Call:
		if op == nil {
			return errors.New("nil Call opcode")
		}
	case Ret:
		return validateReply(move, op)
	case *Ret:
		if op == nil {
			return errors.New("nil Ret opcode")
		}
		return validateReply(move, *op)
	default:
		return errors.New("missing or unsupported opcode")
	}
	return nil
}

func validateReply(move Move, ret Ret) error {
	if ret.InReplyTo.Value == "" || ret.InReplyTo == move.Invocation {
		return errors.New("Ret must answer a distinct invocation")
	}
	// A nil parent could be a root result; its delivery remains a separate blocker.
	return nil
}

func cloneMove(move Move) Move {
	if move.ParentFd != nil {
		parent := *move.ParentFd
		move.ParentFd = &parent
	}
	move.Params = slices.Clone(move.Params)
	for i := range move.Params {
		move.Params[i].Content.Bytes = slices.Clone(move.Params[i].Content.Bytes)
	}
	move.Horizon.Capabilities = cloneCapabilities(move.Horizon.Capabilities)
	move.Focus = cloneCapabilities(move.Focus)
	switch op := move.Op.(type) {
	case Ret:
		op.Value.Content.Bytes = slices.Clone(op.Value.Content.Bytes)
		move.Op = op
	case *Ret:
		if op != nil {
			ret := *op
			ret.Value.Content.Bytes = slices.Clone(ret.Value.Content.Bytes)
			move.Op = &ret
		}
	case *Call:
		if op != nil {
			move.Op = &Call{}
		}
	}
	return move
}

func cloneCapabilities(capabilities []Capability) []Capability {
	capabilities = slices.Clone(capabilities)
	for i := range capabilities {
		capabilities[i].Sig.Bytes = slices.Clone(capabilities[i].Sig.Bytes)
	}
	return capabilities
}
