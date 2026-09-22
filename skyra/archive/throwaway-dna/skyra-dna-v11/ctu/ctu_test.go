package ctu

import (
	"sync"
	"testing"
	"time"

	c "skyra-v1/src/skyra/skyra-dna-v11/causal"
)

var (
	selfTarget   = target("self")
	entryTarget  = target("entry")
	localTarget  = target("local")
	remoteTarget = target("remote")
	faultTarget  = target("missing")
	runAction    = c.Name{Value: "run"}
	entryAction  = c.Name{Value: "invoke"}
	regA         = reg("a")
	regResult    = reg("result")
)

type fakeDriver struct {
	decl    DriverDecl
	perform func(c.Name, []c.Observation) (c.Observation, c.Outcome)
}

func (d fakeDriver) Decl() DriverDecl { return d.decl }
func (d fakeDriver) Perform(action c.Name, in []c.Observation) (c.Observation, c.Outcome) {
	return d.perform(action, in)
}

type fakeEgress struct {
	mu    sync.Mutex
	moves []c.Move
	wake  chan c.Move
}

func (e *fakeEgress) Egress(_ c.Target, move c.Move) c.Outcome {
	e.mu.Lock()
	e.moves = append(e.moves, move)
	e.mu.Unlock()
	if e.wake != nil {
		e.wake <- move
	}
	return c.Success
}

func (e *fakeEgress) sent() []c.Move {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]c.Move(nil), e.moves...)
}

func TestLocalSingleMove(t *testing.T) {
	body := c.Body{
		ID:     c.BodyID{Value: "local-single"},
		Moves:  []c.MoveTemplate{move(localTarget, nil, regResult)},
		Return: regResult,
	}
	driver := returningDriver(localTarget, func(_ []c.Observation) c.Observation { return observation("done") })
	machine := testMachine(body, map[c.Target]Driver{localTarget: driver}, nil)
	possibility := c.PossibilityID{Value: "p-local"}

	if out := machine.Transfer(request(body.ID, possibility, localTarget)); out != c.Success {
		t.Fatalf("Transfer() = %v, want Success", out)
	}
	if got := len(machine.Core.Work[c.Ready]); got != 1 {
		t.Fatalf("Ready after Transfer = %d, want 1", got)
	}
	if out := machine.Core.Transition(machine.Core.Work); out != c.Success {
		t.Fatalf("Transition() = %v, want Success", out)
	}

	ctx := machine.Core.Cache[possibility]
	if ctx.Result == nil || string(ctx.Result.Content.Bytes) != "done" {
		t.Fatalf("Result = %#v, want done", ctx.Result)
	}
	assertNoWork(t, machine.Core.Work)
}

func TestDependentFlipCascades(t *testing.T) {
	body := c.Body{
		ID: c.BodyID{Value: "dependent"},
		Moves: []c.MoveTemplate{
			move(localTarget, nil, regA),
			move(localTarget, []c.RegKey{regA}, regResult),
		},
		Return: regResult,
	}
	driver := returningDriver(localTarget, func(in []c.Observation) c.Observation {
		if len(in) == 0 {
			return observation("first")
		}
		return observation(string(in[0].Content.Bytes) + "+second")
	})
	machine := testMachine(body, map[c.Target]Driver{localTarget: driver}, nil)
	possibility := c.PossibilityID{Value: "p-dependent"}

	if out := machine.Transfer(request(body.ID, possibility, localTarget)); out != c.Success {
		t.Fatalf("Transfer() = %v, want Success", out)
	}
	if got := len(machine.Core.Work[c.Waiting]); got != 1 {
		t.Fatalf("Waiting before Transition = %d, want 1", got)
	}
	if out := machine.Core.Transition(machine.Core.Work); out != c.Success {
		t.Fatalf("Transition() = %v, want Success", out)
	}

	ctx := machine.Core.Cache[possibility]
	if ctx.Result == nil || string(ctx.Result.Content.Bytes) != "first+second" {
		t.Fatalf("Result = %#v, want dependent output", ctx.Result)
	}
	assertNoWork(t, machine.Core.Work)
}

func TestCutRoundTripLandsAndContinues(t *testing.T) {
	body := c.Body{
		ID: c.BodyID{Value: "cut"},
		Moves: []c.MoveTemplate{
			move(remoteTarget, nil, regA),
			move(localTarget, []c.RegKey{regA}, regResult),
		},
		Return: regResult,
	}
	egress := &fakeEgress{}
	driver := returningDriver(localTarget, func(in []c.Observation) c.Observation {
		return observation(string(in[0].Content.Bytes) + "+local")
	})
	machine := testMachine(body, map[c.Target]Driver{localTarget: driver}, egress)
	possibility := c.PossibilityID{Value: "p-cut"}

	if out := machine.Transfer(request(body.ID, possibility, localTarget, remoteTarget)); out != c.Success {
		t.Fatalf("Transfer() = %v, want Success", out)
	}
	if out := machine.Core.Transition(machine.Core.Work); out != c.NotYet {
		t.Fatalf("first Transition() = %v, want NotYet", out)
	}
	sent := egress.sent()
	if len(sent) != 1 || sent[0].Header != c.Request {
		t.Fatalf("egress = %#v, want one Request", sent)
	}
	cut := sent[0]
	if cut.Seed == nil || cut.Seed.ReturnTo == nil || cut.Seed.ReturnTo.Fd != cut.ID || cut.Seed.ReturnTo.To != selfTarget {
		t.Fatalf("cut route = %#v, want self + cut fd", cut.Seed)
	}
	if _, ok := machine.Core.Work[c.Issued][cut.ID]; !ok {
		t.Fatal("cut did not park in Issued")
	}

	returned := observation("remote")
	if out := machine.Transfer(c.Move{ID: cut.ID, Header: c.Response, Returned: &returned}); out != c.Success {
		t.Fatalf("response Transfer() = %v, want Success", out)
	}
	if _, ok := machine.Core.Work[c.Issued][cut.ID]; ok {
		t.Fatal("landed cut remains in Issued")
	}
	if out := machine.Core.Transition(machine.Core.Work); out != c.Success {
		t.Fatalf("second Transition() = %v, want Success", out)
	}
	ctx := machine.Core.Cache[possibility]
	if ctx.Result == nil || string(ctx.Result.Content.Bytes) != "remote+local" {
		t.Fatalf("Result = %#v, want remote+local", ctx.Result)
	}
}

func TestUnprovisionedTargetFaults(t *testing.T) {
	body := c.Body{
		ID:     c.BodyID{Value: "fault"},
		Moves:  []c.MoveTemplate{move(faultTarget, nil, regResult)},
		Return: regResult,
	}
	egress := &fakeEgress{}
	machine := testMachine(body, nil, egress)

	if out := machine.Transfer(request(body.ID, c.PossibilityID{Value: "p-fault"}, faultTarget)); out != c.Success {
		t.Fatalf("Transfer() = %v, want Success", out)
	}
	if out := machine.Core.Transition(machine.Core.Work); out != c.Fault {
		t.Fatalf("Transition() = %v, want Fault", out)
	}
	if got := len(egress.sent()); got != 0 {
		t.Fatalf("dispatched %d moves on Fault, want 0", got)
	}
	if got := len(machine.Core.Work[c.Ready]); got != 1 {
		t.Fatalf("Ready after Fault = %d, want faulting move retained", got)
	}
}

func TestRunParksAndTransferWakesIt(t *testing.T) {
	body := c.Body{
		ID:     c.BodyID{Value: "park-wake"},
		Moves:  []c.MoveTemplate{move(localTarget, nil, regResult)},
		Return: regResult,
	}
	responses := make(chan c.Move, 1)
	egress := &fakeEgress{wake: responses}
	driver := returningDriver(localTarget, func(_ []c.Observation) c.Observation { return observation("awake") })
	machine := testMachine(body, map[c.Target]Driver{localTarget: driver}, egress)
	go machine.Run()

	select {
	case move := <-responses:
		t.Fatalf("Run dispatched before Transfer: %#v", move)
	default:
	}

	requestMove := request(body.ID, c.PossibilityID{Value: "p-wake"}, localTarget)
	requestMove.Seed.ReturnTo = &c.Route{
		To: target("caller"),
		Fd: c.MoveID{Possibility: c.PossibilityID{Value: "parent"}, Writes: regResult},
	}
	if out := machine.Transfer(requestMove); out != c.Success {
		t.Fatalf("Transfer() = %v, want Success", out)
	}

	select {
	case response := <-responses:
		if response.Header != c.Response || response.Returned == nil || string(response.Returned.Content.Bytes) != "awake" {
			t.Fatalf("woken response = %#v", response)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not wake and drain after Transfer")
	}
}

func TestRunEmitsCompletionAfterFinalCutLanding(t *testing.T) {
	body := c.Body{
		ID:     c.BodyID{Value: "cut-is-return"},
		Moves:  []c.MoveTemplate{move(remoteTarget, nil, regResult)},
		Return: regResult,
	}
	crossings := make(chan c.Move, 2)
	egress := &fakeEgress{wake: crossings}
	machine := testMachine(body, nil, egress)
	go machine.Run()

	requestMove := request(body.ID, c.PossibilityID{Value: "p-cut-return"}, remoteTarget)
	requestMove.Seed.ReturnTo = &c.Route{
		To: target("caller"),
		Fd: c.MoveID{Possibility: c.PossibilityID{Value: "parent-cut"}, Writes: regResult},
	}
	if out := machine.Transfer(requestMove); out != c.Success {
		t.Fatalf("Transfer() = %v, want Success", out)
	}

	cut := receiveMove(t, crossings)
	if cut.Header != c.Request {
		t.Fatalf("first crossing = %v, want Request", cut.Header)
	}
	returned := observation("remote-final")
	if out := machine.Transfer(c.Move{ID: cut.ID, Header: c.Response, Returned: &returned}); out != c.Success {
		t.Fatalf("response Transfer() = %v, want Success", out)
	}

	response := receiveMove(t, crossings)
	if response.Header != c.Response || response.ID != requestMove.Seed.ReturnTo.Fd || response.Returned == nil {
		t.Fatalf("completion response = %#v", response)
	}
}

func testMachine(body c.Body, local map[c.Target]Driver, out Egress) *Machine {
	bodies := c.BodyCache{body.ID: body}
	core := &Core{
		Self:   selfTarget,
		Bodies: bodies,
		Interp: InterpretFunc(func(_ *c.Context, move *c.Move) (EligibleSet, c.Outcome) {
			if move.Template.Target == faultTarget {
				return EligibleSet{}, c.Fault
			}
			return EligibleSet{Targets: []c.Target{move.Template.Target}}, c.Success
		}),
		Local: local,
		Out:   out,
	}
	return New(core, CachedMaterializer{Bodies: bodies}, VerifyFunc(func(c.Capability, c.Target, c.Name) bool {
		return true
	}))
}

func request(body c.BodyID, possibility c.PossibilityID, capabilities ...c.Target) c.Move {
	horizon := c.Horizon{Capabilities: []c.Capability{{Target: entryTarget, Action: entryAction}}}
	for _, target := range capabilities {
		horizon.Capabilities = append(horizon.Capabilities, c.Capability{Target: target, Action: runAction})
	}
	return c.Move{
		Template: c.MoveTemplate{Target: entryTarget, Action: entryAction},
		Header:   c.Request,
		Seed: &c.Context{
			ID:        possibility,
			Body:      body,
			Registers: make(map[c.RegKey]c.Register),
			Horizon:   horizon,
		},
	}
}

func returningDriver(target c.Target, perform func([]c.Observation) c.Observation) Driver {
	return fakeDriver{
		decl: DriverDecl{Target: target, Actions: []c.Name{runAction}},
		perform: func(_ c.Name, in []c.Observation) (c.Observation, c.Outcome) {
			return perform(in), c.Success
		},
	}
}

func move(target c.Target, reads []c.RegKey, writes c.RegKey) c.MoveTemplate {
	return c.MoveTemplate{
		Target: target,
		Action: runAction,
		Reads:  reads,
		Writes: writes,
		Return: true,
	}
}

func target(value string) c.Target { return c.Target{Name: c.Name{Value: value}} }
func reg(value string) c.RegKey    { return c.RegKey{Name: c.Name{Value: value}} }

func observation(value string) c.Observation {
	return c.Observation{Content: c.Content{Bytes: []byte(value)}}
}

func receiveMove(t *testing.T, moves <-chan c.Move) c.Move {
	t.Helper()
	select {
	case move := <-moves:
		return move
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for egress")
		return c.Move{}
	}
}

func assertNoWork(t *testing.T, work c.WorkList) {
	t.Helper()
	for _, state := range []c.MoveState{c.Waiting, c.Ready, c.Issued} {
		if got := len(work[state]); got != 0 {
			t.Fatalf("work[%v] = %d, want empty", state, got)
		}
	}
}
