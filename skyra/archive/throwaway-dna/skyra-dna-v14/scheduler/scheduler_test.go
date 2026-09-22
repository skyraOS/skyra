package scheduler

import (
	"context"
	"testing"

	"skyra-v1/src/skyra/skyra-dna-v14/controller"
	"skyra-v1/src/skyra/skyra-dna-v14/driver"
	"skyra-v1/src/skyra/skyra-dna-v14/probe"
	"skyra-v1/src/skyra/skyra-dna-v14/protocol"
)

func TestRunSuppressesSnapshotsAfterRootCompletion(t *testing.T) {
	dispatches := 0
	authority := &probe.ProbeAuthority{CheckMove: func(context.Context, protocol.Move) error { dispatches++; return nil }}
	registry := driver.NewDriverRegistry()
	for address, endpoint := range map[string]driver.Driver{"root": &probe.RootDriver{EchoTarget: protocol.Target{Address: "echo"}}, "echo": &probe.EchoDriver{}} {
		if err := registry.Register(address, driver.DriverRegistration{Driver: endpoint, Input: probe.ProbeInput{}}); err != nil {
			t.Fatal(err)
		}
	}
	cc, err := controller.NewContinuationController(registry, authority)
	if err != nil {
		t.Fatal(err)
	}
	results := 0
	s, err := NewScheduler(cc, func(protocol.RootResult) error { results++; return nil })
	if err != nil {
		t.Fatal(err)
	}
	id := protocol.InvocationID{Value: "root"}
	move := protocol.Move{Target: protocol.Target{Address: "root"}, Op: protocol.Invoke{InvocationID: id, Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{{}}}}
	if err := s.Write(context.Background(), protocol.MoveBatch{Moves: []protocol.Move{move}}); err != nil {
		t.Fatal(err)
	}
	if results != 1 || dispatches != 4 {
		t.Fatalf("results=%d dispatches=%d", results, dispatches)
	}
	root, _ := s.Snapshot(id)
	var selected []protocol.Continuation
	for _, childID := range root.Move.Op.(protocol.Invoke).Register.Order {
		child, _ := s.Snapshot(childID)
		selected = append(selected, child)
	}
	if err := s.run(selected); err != nil {
		t.Fatal(err)
	}
	if dispatches != 4 {
		t.Fatal("completed workflow dispatched again")
	}
}
