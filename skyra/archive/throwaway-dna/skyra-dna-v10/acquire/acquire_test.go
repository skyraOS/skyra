package acquire_test

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"skyra-v1/src/skyra/skyra-dna-v10/acquire"
	c "skyra-v1/src/skyra/skyra-dna-v10/causal"
	"skyra-v1/src/skyra/skyra-dna-v10/engine"
	"skyra-v1/src/skyra/skyra-dna-v10/kernel"
	"skyra-v1/src/skyra/skyra-dna-v10/privilege"
)

func onlyPending(t *testing.T, p *c.Possibility) c.MoveID {
	t.Helper()
	if len(p.InFlight) != 1 {
		t.Fatalf("want one pending grant, got %d", len(p.InFlight))
	}
	for id := range p.InFlight {
		return id
	}
	panic("unreachable")
}

func TestAcquireEndToEnd(t *testing.T) {
	var receipt bytes.Buffer
	trace := log.New(&receipt, "", 0)
	e := c.Endpoint{Namespace: "dns", Address: "203.0.113.7:443", Triple: c.Triple{Domain: "inet", Type: "stream", Protocol: "tcp"}}
	fake := &acquire.FakeNetwork{Names: []c.Name{{Value: "weather.example"}}, Endpoint: e, Reply: map[string]any{"forecast": "sunny", "temperature_c": 19}}
	store := kernel.NewStore(trace)
	privilegePeerCalls := 0
	store.Seed(engine.FunctionResource{ResourceName: c.Name{Value: engine.PrivilegeName}, Trust: c.Entity, Runner: func(_ []c.Observation, _ c.Move) (c.Observation, c.Outcome) {
		privilegePeerCalls++
		return c.Observation{}, c.Success
	}})
	boundary := privilege.New(privilege.NovelPolicy{}, store, store, trace)
	kit := acquire.Kit{Registry: store, Fake: fake, Trace: trace}
	kit.Install(store)
	machine := engine.New(store, boundary, trace)
	intent := c.Observation{Content: c.Content{Bytes: []byte(acquire.Purpose)}, Segment: c.Segment{By: c.Name{Value: "user"}}}

	p := acquire.NewPossibility("acquire.first", intent)
	trace.Printf("ASSERT start acquire horizon.grants=%d kernel.mints=%d (read-only)\n", len(p.Horizon.Grants), store.MintCount())
	if out := machine.Drive(p); out != c.NotYet {
		t.Fatalf("poke should park: %v", out)
	}
	if store.MintCount() != 0 || len(p.Horizon.Grants) != 0 || fake.ReachCalls != 0 || privilegePeerCalls != 0 {
		t.Fatalf("read/fault path minted, contacted endpoint, or peer-dispatched privilege")
	}
	pokeID := onlyPending(t, p)
	if _, ok := p.Moves[2].Site.(c.Cut); !ok {
		t.Fatalf("parked permission was not represented as a Cut: %T", p.Moves[2].Site)
	}
	trace.Printf("ASSERT after issuing Drive returned outcome=NOTYET result=nil in_flight=%d kernel.mints=%d endpoint_probe_calls=%d privilege_peer_calls=%d\n", len(p.InFlight), store.MintCount(), fake.ReachCalls, privilegePeerCalls)
	poke, out := boundary.Approve(pokeID)
	if out != c.Success || poke.Kind != c.PokeGrant {
		t.Fatal("poke approval")
	}
	if out = machine.Recognize(pokeID, poke); out != c.NotYet {
		t.Fatalf("register should park: %v", out)
	}
	registerID := onlyPending(t, p)
	if fake.UnsafeCalls != 0 || store.MintCount() != 0 {
		t.Fatalf("test was unsafe or minted")
	}
	if fake.ReachCalls != 1 || fake.HandshakeCalls != 1 || fake.ChallengeCalls != 1 || fake.DescribeCalls != 1 {
		t.Fatalf("safe probe leaves did not each run once: reach=%d handshake=%d challenge=%d describe=%d", fake.ReachCalls, fake.HandshakeCalls, fake.ChallengeCalls, fake.DescribeCalls)
	}
	register, out := boundary.Approve(registerID)
	if out != c.Success || register.Kind != c.RegisterGrant {
		t.Fatal("register approval")
	}
	if store.MintCount() != 1 {
		t.Fatalf("kernel did not mint exactly once")
	}
	if out = machine.Recognize(registerID, register); out != c.Success {
		t.Fatalf("finish acquisition: %v", out)
	}
	cap, ok := p.Result.Value.(c.Capability)
	if !ok || cap.Block != c.Entity {
		t.Fatalf("bad capability: %#v", p.Result.Value)
	}
	stored, exists := store.Existing(e)
	if !exists || !reflect.DeepEqual(stored.Contract, cap.Contract) {
		t.Fatalf("registry did not retain Capability.Contract: stored=%#v returned=%#v", stored.Contract, cap.Contract)
	}
	trace.Printf("ADMIT -> %s\nREGISTRY READBACK endpoint=%s block=%s contract.response={name:%s required:%v}\n", acquire.FormatCapability(cap), cap.Endpoint.Address, cap.Block, cap.Contract.Fitness.Response.Name.Value, cap.Contract.Fitness.Response.RequiredFields)
	if p.Retired != len(p.Moves) || len(p.InFlight) != 0 {
		t.Fatalf("not quiescent: retired=%d moves=%d flight=%d", p.Retired, len(p.Moves), len(p.InFlight))
	}
	trace.Printf("QUIESCENCE retired=%d/%d in_flight=%d result=Capability\n", p.Retired, len(p.Moves), len(p.InFlight))

	good := acquire.UsePossibility("use.good", cap)
	if out = machine.Drive(good); out != c.Success {
		t.Fatalf("honoring entity reply: %v", out)
	}
	if good.Result.Segment.By != cap.Endpoint.Key() {
		t.Fatalf("entity reply was checked but not stamped: %+v", good.Result.Segment)
	}
	trace.Printf("USE ENTITY honoring reply=%v -> SUCCESS\n", fake.Reply)
	fake.Reply = map[string]any{"forecast": "sunny"}
	bad := acquire.UsePossibility("use.bad", cap)
	if out = machine.Drive(bad); out != c.Failure {
		t.Fatalf("violating entity reply passed: %v", out)
	}
	trace.Printf("USE ENTITY violating reply=%v -> FAILURE\n", fake.Reply)
	store.Seed(engine.FunctionResource{ResourceName: c.Name{Value: "host.weather"}, Trust: c.Host, Runner: func(_ []c.Observation, _ c.Move) (c.Observation, c.Outcome) {
		return c.Observation{Value: fake.Reply}, c.Success
	}})
	hostOut := c.RegKey{Name: c.Name{Value: "host.out"}}
	host := &c.Possibility{ID: c.PossibilityID{Value: "use.host"}, Registers: map[c.RegKey]c.Register{hostOut: {}}, InFlight: map[c.MoveID]c.Pending{}, Return: hostOut, Moves: []c.Move{{ID: c.MoveID{Value: "host-use"}, Target: c.Name{Value: "host.weather"}, Writes: hostOut, Contract: cap.Contract.Fitness.Response, Site: c.Local{}}}}
	if out = machine.Drive(host); out != c.Success {
		t.Fatalf("host should be believed: %v", out)
	}
	trace.Printf("USE HOST same violating reply=%v -> SUCCESS (BELIEVED, no contract gate)\n", fake.Reply)

	beforeDecisions, beforeWakes, beforeMints := boundary.DecisionCount(), boundary.WakeCount(), store.MintCount()
	second := acquire.NewPossibility("acquire.second", intent)
	if out = machine.Drive(second); out != c.Success {
		t.Fatalf("reacquire: %v", out)
	}
	if boundary.DecisionCount() != beforeDecisions || boundary.WakeCount() != beforeWakes || store.MintCount() != beforeMints || len(second.InFlight) != 0 || len(second.Horizon.Grants) != 0 {
		t.Fatal("known fast path crossed boundary")
	}
	if privilegePeerCalls != 0 {
		t.Fatal("privilege trap was dispatched as a peer")
	}
	trace.Printf("REACQUIRE known=true -> SUCCESS faults=0 grants=0 wakes=0 new_mints=0 retired=%d/%d\n", second.Retired, len(second.Moves))

	for _, want := range []string{"FAULT target=" + engine.PrivilegeName, "PARK move=reach", "RECOGNIZE", "KERNEL MINT", "result=FAIL", "REACQUIRE known=true"} {
		if !strings.Contains(receipt.String(), want) {
			t.Fatalf("trace missing %q", want)
		}
	}
	fmt.Print(receipt.String())
}
