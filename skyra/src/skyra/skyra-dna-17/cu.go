package dna17

import (
	"context"
	"errors"
	"sync"
)

// inbound queue capacity. The spec leaves queue capacity open; a generous buffered
// channel stands in for the "one inbound queue" here (a single-process
// simplification — see issues.md). The durable unfinished-work records live in
// state, not in this channel.
const inboundCapacity = 4096

// CU is the collapsed compute unit: one interpreter goroutine draining one inbound
// queue, one StateExecutor (the writer), and a pool of concurrent workers.
type CU struct {
	state   *StateExecutor
	inbound chan Move

	ctx    context.Context
	cancel context.CancelFunc

	onRoot   func(RootResult) error
	rootOnce sync.Map // InvocationID -> struct{}: exactly one root result per root

	workers  sync.WaitGroup
	loopDone chan struct{}
}

// NewCU boots the runtime over a set of registered Surfaces and a root-result sink.
// It starts the single interpreter loop.
func NewCU(surfaces map[SurfaceID]Surface, onRoot func(RootResult) error) (*CU, error) {
	if onRoot == nil {
		return nil, errors.New("a root result sink is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cu := &CU{
		state:    newStateExecutor(surfaces),
		inbound:  make(chan Move, inboundCapacity),
		ctx:      ctx,
		cancel:   cancel,
		onRoot:   onRoot,
		loopDone: make(chan struct{}),
	}
	cu.state.enqueue = cu.Interpret // even the writer re-enters through the one door
	go cu.run()
	return cu, nil
}

// Interpret is the single public door: it enqueues a Move for interpretation.
// External submission (via Submit) and worker emissions both use it. There is no
// second door for returns — a Return is a Move.
func (cu *CU) Interpret(move Move) {
	select {
	case cu.inbound <- move:
	case <-cu.ctx.Done():
	}
}

// Submit opens a Possibility with the supplied Horizon and posts its first Move.
func (cu *CU) Submit(ctx context.Context, move Move, horizon Horizon) error {
	if move.Invoke == nil {
		return errors.New("submit requires a first Invoke move")
	}
	if move.PossibilityID == "" {
		return errors.New("submit requires a possibility id")
	}
	return cu.state.Open(ctx, OpenPossibility{Identity: move.PossibilityID, First: move, Horizon: horizon})
}

// run is the interpreter: sequential. It sleeps while the queue is empty, wakes on
// a Move, decodes it, runs read+commit in one turn, then dispatches the authorized
// executions and delivers any root results — both OUTSIDE the writer lock.
func (cu *CU) run() {
	defer close(cu.loopDone)
	for {
		select {
		case <-cu.ctx.Done():
			return
		case move := <-cu.inbound:
			cu.step(move)
		}
	}
}

func (cu *CU) step(move Move) {
	decide, err := decode(move)
	if err != nil {
		return // undecodable; nothing to commit
	}
	result, err := cu.state.InterpretTurn(cu.ctx, move, decide)
	if err != nil {
		return // fatal interpret error for this move; records preserved
	}
	// All effects happen AFTER the lock released inside InterpretTurn: dispatch the
	// authorized executions, re-enter any corrective/propagated Moves through the one
	// door, and deliver root results. The commit itself did none of this.
	cu.Dispatch(result.authorized)
	for _, m := range result.enqueue {
		cu.Interpret(m)
	}
	for _, rr := range result.rootResults {
		cu.deliverRoot(rr)
	}
}

// deliverRoot notifies the sink exactly once per root invocation.
func (cu *CU) deliverRoot(rr RootResult) {
	if _, loaded := cu.rootOnce.LoadOrStore(rr.Invocation, struct{}{}); loaded {
		return
	}
	_ = cu.onRoot(rr)
}

// Snapshot exposes a detached invocation view for tests.
func (cu *CU) Snapshot(inv InvocationID) (InvocationView, bool) { return cu.state.snapshot(inv) }

// Close stops the interpreter and waits for in-flight workers to drain. It is
// idempotent and does not itself close user-supplied Surfaces/transports.
func (cu *CU) Close() {
	cu.cancel()
	<-cu.loopDone
	cu.workers.Wait()
}
