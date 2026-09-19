package dna

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
)

func testObservation(value string) Observation {
	observation := Observation{}
	observation.Content.Bytes = []byte(value)
	return observation
}

func testRootMove(rootAddress, echoAddress string) Move {
	action := Action{Value: "echo"}
	rootTarget := Target{Address: rootAddress}
	echoTarget := Target{Address: echoAddress}
	return Move{
		Root:   PossibilityID{Value: "possibility"},
		Target: rootTarget,
		Op: Invoke{
			InvocationID: InvocationID{Value: "root"},
			Action:       action,
			Params:       []Observation{testObservation("hello")},
			Horizon: Horizon{Capabilities: []Capability{
				{Target: rootTarget, Action: action},
				{Target: echoTarget, Action: action},
			}},
		},
	}
}

func TestIntegratedLoop(t *testing.T) {
	root := &RootDriver{EchoTarget: Target{Address: "echo"}}
	echo := &EchoDriver{}
	validator := ProbeInput{}
	rootEndpoint := Endpoint{Address: "root", Registration: DriverRegistration{Driver: root, Input: validator}}
	kernel := &FixtureKernel{Root: rootEndpoint, Local: []Endpoint{{Address: "echo", Registration: DriverRegistration{Driver: echo, Input: validator}}}}
	result := make(chan RootResult, 1)
	runtime, err := NewCTU(BootConfig{KernelAddress: "kernel", Kernel: kernel, KernelInput: validator, Root: rootEndpoint}, func(value RootResult) error {
		result <- value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Submit(context.Background(), testRootMove("root", "echo")); err != nil {
		t.Fatal(err)
	}
	completed := <-result
	if completed.Err != nil || completed.Observation == nil || string(completed.Observation.Content.Bytes) != "hellohello" {
		t.Fatalf("unexpected result: %#v", completed)
	}
}

type stateCommandDriver struct {
	scheduler *Scheduler
	sawCommit bool
}

func (driver *stateCommandDriver) Ingress(_ context.Context, input DriverInput, output func(DriverEmission, error) error) error {
	state := DriverState{Version: 1, Payload: []byte("committed")}
	encoded, _ := json.Marshal([]ProbeOutput{{Kind: "return", Value: ptrObservation(testObservation("done"))}})
	return output(DriverEmission{State: state, Raw: string(encoded), Commands: []HostCommand{{ID: "one", Payload: []byte("send")}}}, nil)
}

func (*stateCommandDriver) Egress(raw string) ([]Move, error) { return parseProbe(raw) }

func (driver *stateCommandDriver) Execute(_ context.Context, _ HostCommand) error {
	snapshot, exists := driver.scheduler.Snapshot(InvocationID{Value: "root"})
	driver.sawCommit = exists && string(snapshot.DriverState.Payload) == "committed" && snapshot.Result != nil
	return nil
}

func TestStateCommitsBeforeCommandAndSnapshotsAreDetached(t *testing.T) {
	driver := &stateCommandDriver{}
	registry := NewDriverRegistry()
	if err := registry.Register("root", DriverRegistration{Driver: driver, Input: ProbeInput{}}); err != nil {
		t.Fatal(err)
	}
	controller, _ := NewContinuationController(registry.Resolver())
	scheduler, _ := NewScheduler(controller, func(RootResult) error { return nil })
	driver.scheduler = scheduler
	move := testRootMove("root", "root")
	invocation := move.Op.(Invoke)
	invocation.Horizon.Capabilities = []Capability{{Target: move.Target, Action: invocation.Action}}
	move.Op = invocation
	if err := scheduler.Transition(context.Background(), MoveBatch{Moves: []Move{move}}); err != nil {
		t.Fatal(err)
	}
	if !driver.sawCommit {
		t.Fatal("host command executed before state and return committed")
	}
	snapshot, _ := scheduler.Snapshot(InvocationID{Value: "root"})
	snapshot.DriverState.Payload[0] = 'X'
	again, _ := scheduler.Snapshot(InvocationID{Value: "root"})
	if string(again.DriverState.Payload) != "committed" {
		t.Fatal("snapshot mutated resident driver state")
	}
	if err := scheduler.Transition(context.Background(), DriverTransition{
		Source: InvocationID{Value: "root"}, Sequence: 99,
		State: DriverState{Version: 1, Payload: []byte("stale")},
	}); err != nil {
		t.Fatal(err)
	}
	again, _ = scheduler.Snapshot(InvocationID{Value: "root"})
	if string(again.DriverState.Payload) != "committed" {
		t.Fatal("stale transition changed terminal state")
	}
}

type invalidTransitionDriver struct{}

func (*invalidTransitionDriver) Ingress(_ context.Context, _ DriverInput, output func(DriverEmission, error) error) error {
	bad := Move{Target: Target{Address: "forbidden"}, Op: Invoke{Action: Action{Value: "nope"}, Params: []Observation{testObservation("x")}}}
	encoded, _ := json.Marshal([]ProbeOutput{{Kind: "invoke", Target: bad.Target, Action: bad.Op.(Invoke).Action, Params: bad.Op.(Invoke).Params}})
	return output(DriverEmission{State: DriverState{Version: 1, Payload: []byte("must rollback")}, Raw: string(encoded)}, nil)
}

func (*invalidTransitionDriver) Egress(raw string) ([]Move, error) { return parseProbe(raw) }

func TestRejectedOutputRollsBackDriverStateAndFaults(t *testing.T) {
	driver := &invalidTransitionDriver{}
	registry := NewDriverRegistry()
	if err := registry.Register("root", DriverRegistration{Driver: driver, Input: ProbeInput{}}); err != nil {
		t.Fatal(err)
	}
	controller, _ := NewContinuationController(registry.Resolver())
	faults := make(chan RootResult, 1)
	scheduler, _ := NewScheduler(controller, func(result RootResult) error { faults <- result; return nil })
	move := testRootMove("root", "root")
	invocation := move.Op.(Invoke)
	invocation.Horizon.Capabilities = []Capability{{Target: move.Target, Action: invocation.Action}}
	move.Op = invocation
	err := scheduler.Transition(context.Background(), MoveBatch{Moves: []Move{move}})
	if err == nil {
		t.Fatal("expected rejected child transition")
	}
	snapshot, _ := scheduler.Snapshot(InvocationID{Value: "root"})
	if len(snapshot.DriverState.Payload) != 0 || snapshot.State != Faulted {
		t.Fatalf("transition was not rolled back and faulted: %#v", snapshot)
	}
	select {
	case result := <-faults:
		if result.Err == nil {
			t.Fatal("fault notification missing error")
		}
	default:
		t.Fatal("missing root fault notification")
	}
}

func TestConcurrentRegistryInstallation(t *testing.T) {
	registry := NewDriverRegistry()
	registration := DriverRegistration{Driver: &EchoDriver{}, Input: ProbeInput{}}
	var wait sync.WaitGroup
	errorsSeen := make(chan error, 8)
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() { defer wait.Done(); errorsSeen <- registry.Register("same", registration) }()
	}
	wait.Wait()
	close(errorsSeen)
	var successes int
	for err := range errorsSeen {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful registrations = %d", successes)
	}
}

var _ CommandExecutor = (*stateCommandDriver)(nil)
