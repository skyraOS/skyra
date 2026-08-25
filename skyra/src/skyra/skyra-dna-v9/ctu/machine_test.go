package ctu_test

import (
	"bytes"
	"strconv"
	"testing"

	c "skyra-v1/src/skyra/skyra-dna-v9/causal"
	"skyra-v1/src/skyra/skyra-dna-v9/ctu"
	"skyra-v1/src/skyra/skyra-dna-v9/drivers"
	"skyra-v1/src/skyra/skyra-dna-v9/entity"
	"skyra-v1/src/skyra/skyra-dna-v9/fabric"
	"skyra-v1/src/skyra/skyra-dna-v9/kernel"
)

func key(name string, iter uint32) c.RegKey {
	return c.RegKey{Name: c.Name{Value: name}, Tag: c.Tag{Iter: iter}}
}

type trap struct{ calls int }

func (t *trap) Trap(_ *c.Possibility, _ *c.Move, _ []c.Observation) (c.Content, c.Outcome) {
	t.calls++
	return c.Content{Bytes: []byte("trapped")}, c.Success
}

func securityFor(t *testing.T, names ...string) (*fabric.Security, fabric.Authority, fabric.SegmentVerifier, c.Horizon) {
	t.Helper()
	ctus := make([]c.Name, len(names))
	for i, name := range names {
		ctus[i] = c.Name{Value: name}
	}
	security, out := fabric.NewSecurity(ctus...)
	if out != c.Success {
		t.Fatal("security")
	}
	horizon, out := security.Minter().Mint(c.Name{Value: "account"})
	if out != c.Success {
		t.Fatal("horizon")
	}
	return security, security.Authority(), security.Verifier(), horizon
}

func TestDAGParksAcrossTwoCTUsThenRecognizeRedrivesAndRememberGathers(t *testing.T) {
	security, authority, verifier, horizon := securityFor(t, "client", "worker")
	clientCustody, _ := security.Custody(c.Name{Value: "client"})
	workerCustody, _ := security.Custody(c.Name{Value: "worker"})
	transport := fabric.NewLocal(authority, verifier)
	driver := drivers.NewFile(t.TempDir() + "/graph.json")
	body := []c.MoveTemplate{
		{Position: 0, Target: c.Name{Value: "seed"}, Writes: c.Name{Value: "seed-out"}, Contract: c.Contract{Name: c.Name{Value: "seed"}}},
		{Position: 1, Target: c.Name{Value: "left"}, Reads: []c.Name{{Value: "seed-out"}}, Writes: c.Name{Value: "left-out"}, Contract: c.Contract{Name: c.Name{Value: "left"}}},
		{Position: 2, Target: c.Name{Value: "right"}, Reads: []c.Name{{Value: "seed-out"}}, Writes: c.Name{Value: "right-out"}, Contract: c.Contract{Name: c.Name{Value: "right"}}},
		{Position: 3, Target: c.Name{Value: "join"}, Reads: []c.Name{{Value: "left-out"}, {Value: "right-out"}}, Writes: c.Name{Value: "result"}, Contract: c.Contract{Name: c.Name{Value: "join"}}},
	}
	engram := c.Engram{Name: c.Name{Value: "dag"}, Body: body, Return: key("result", 0)}
	if driver.Seed(engram) != c.Success {
		t.Fatal("seed engram")
	}
	system := kernel.New(driver, transport, verifier, authority, security.Minter(), nil)
	root, out := system.Activate(engram.Name, horizon)
	if out != c.Success || len(root.Moves) != 4 {
		t.Fatalf("activate body: %v", out)
	}
	client := ctu.New("client", transport, clientCustody, authority, nil)
	worker := ctu.New("worker", transport, workerCustody, authority, nil)
	if transport.Register("client", client) != c.Success || transport.Register("worker", worker) != c.Success {
		t.Fatal("register endpoints")
	}
	client.RegisterHost(c.Name{Value: "seed"}, c.Name{Value: "seed-host"}, entity.Function(func(_ []c.Observation, _ c.Payload) (c.Content, c.Outcome) {
		return c.Content{Bytes: []byte("S")}, c.Success
	}))
	client.RegisterEntity(c.Name{Value: "join"}, c.Name{Value: "join-unit"}, entity.Function(func(operands []c.Observation, _ c.Payload) (c.Content, c.Outcome) {
		return c.Content{Bytes: append(append([]byte{}, operands[0].Content.Bytes...), operands[1].Content.Bytes...)}, c.Success
	}))
	worker.RegisterEntity(c.Name{Value: "left"}, c.Name{Value: "left-unit"}, entity.Function(func(operands []c.Observation, _ c.Payload) (c.Content, c.Outcome) {
		return c.Content{Bytes: append([]byte("L"), operands[0].Content.Bytes...)}, c.Success
	}))
	worker.RegisterEntity(c.Name{Value: "right"}, c.Name{Value: "right-unit"}, entity.Function(func(operands []c.Observation, _ c.Payload) (c.Content, c.Outcome) {
		return c.Content{Bytes: append([]byte("R"), operands[0].Content.Bytes...)}, c.Success
	}))
	client.RegisterRemote(c.Name{Value: "left"}, c.Target{Endpoint: "worker"})
	client.RegisterRemote(c.Name{Value: "right"}, c.Target{Endpoint: "worker"})

	if out = client.Drive(&root); out != c.NotYet {
		t.Fatalf("park outcome: %v", out)
	}
	if root.Result != nil || len(root.InFlight) != 2 || transport.Pending() != 2 {
		t.Fatalf("cut did not park: result=%v in-flight=%d queued=%d", root.Result, len(root.InFlight), transport.Pending())
	}
	if out = transport.Drain(); out != c.Success {
		t.Fatalf("drain events: %v", out)
	}
	if root.Result == nil || !bytes.Equal(root.Result.Content.Bytes, []byte("LSRS")) {
		t.Fatalf("join result: %+v", root.Result)
	}
	if len(root.InFlight) != 0 || root.Retired != len(root.Moves) {
		t.Fatalf("not quiescent: frontier=%d watermark=%d", len(root.InFlight), root.Retired)
	}
	cone, out := system.Remember(&root, *root.Result)
	if out != c.Success || len(cone.Segments) != 4 {
		t.Fatalf("gather: outcome=%v segments=%d", out, len(cone.Segments))
	}
	child := root
	parent := c.MoveID{Value: "cut"}
	child.Parent = &parent
	if _, out = system.Remember(&child, *root.Result); out != c.Failure {
		t.Fatalf("non-root remembered: %v", out)
	}
}

func TestDispatchIsHostBelievedEntityCheckedAndKernelTrapped(t *testing.T) {
	security, authority, _, horizon := securityFor(t, "one")
	custody, _ := security.Custody(c.Name{Value: "one"})
	trapper := &trap{}
	machine := ctu.New("one", nil, custody, authority, trapper)
	wrong := c.Contract{Name: c.Name{Value: "different"}}
	machine.RegisterHost(c.Name{Value: "host"}, c.Name{Value: "host-unit"}, entity.Function(func(_ []c.Observation, _ c.Payload) (c.Content, c.Outcome) {
		return c.Content{Bytes: []byte("believed")}, c.Success
	}))
	host := &c.Possibility{ID: c.PossibilityID{Value: "host"}, Horizon: horizon, Return: key("out", 0), Registers: map[c.RegKey]c.Register{key("out", 0): {Contract: wrong}}, Moves: []c.Move{{Target: c.Name{Value: "host"}, Writes: key("out", 0), Contract: wrong, Site: c.Local{}}}, InFlight: make(map[c.MoveID]c.RegKey)}
	if out := machine.Drive(host); out != c.Success || host.Result == nil {
		t.Fatalf("host was contract-gated: %v", out)
	}

	machine.RegisterEntity(c.Name{Value: "entity"}, c.Name{Value: "entity-unit"}, entity.Function(func(_ []c.Observation, _ c.Payload) (c.Content, c.Outcome) {
		return c.Content{Bytes: []byte("unchecked")}, c.Success
	}))
	foreign := &c.Possibility{ID: c.PossibilityID{Value: "entity"}, Horizon: horizon, Return: key("out", 0), Registers: map[c.RegKey]c.Register{key("out", 0): {Contract: wrong}}, Moves: []c.Move{{Target: c.Name{Value: "entity"}, Writes: key("out", 0), Contract: wrong, Site: c.Local{}}}, InFlight: make(map[c.MoveID]c.RegKey)}
	if out := machine.Drive(foreign); out != c.Failure {
		t.Fatalf("entity escaped contract check: %v", out)
	}

	peerCalls := 0
	kernelName := c.Name{Value: "kernel.remember"}
	machine.RegisterEntity(kernelName, c.Name{Value: "fake-peer"}, entity.Function(func(_ []c.Observation, _ c.Payload) (c.Content, c.Outcome) {
		peerCalls++
		return c.Content{}, c.Success
	}))
	if machine.RegisterKernel(kernelName) != c.Success {
		t.Fatal("register kernel trap")
	}
	kernelMove := &c.Possibility{ID: c.PossibilityID{Value: "trap"}, Horizon: horizon, Return: key("out", 0), Registers: map[c.RegKey]c.Register{key("out", 0): {Contract: c.Contract{Name: kernelName}}}, Moves: []c.Move{{Target: kernelName, Writes: key("out", 0), Contract: c.Contract{Name: kernelName}, Site: c.Local{}}}, InFlight: make(map[c.MoveID]c.RegKey)}
	if out := machine.Drive(kernelMove); out != c.Success || trapper.calls != 1 || peerCalls != 0 {
		t.Fatalf("kernel dispatched as peer: out=%v traps=%d peers=%d", out, trapper.calls, peerCalls)
	}
}

func TestSwitchMoveUnfoldsTwoTaggedIterationsAndTerminatesByQuiescence(t *testing.T) {
	security, authority, _, horizon := securityFor(t, "loop")
	custody, _ := security.Custody(c.Name{Value: "loop"})
	machine := ctu.New("loop", nil, custody, authority, nil)
	incName, switchName := c.Name{Value: "increment"}, c.Name{Value: "switch"}
	carried0, computed0, exit := key("carried", 0), key("computed", 0), key("exit", 0)
	initial := c.Observation{Content: c.Content{Bytes: []byte("0")}}
	p := &c.Possibility{ID: c.PossibilityID{Value: "loop"}, Horizon: horizon, Return: exit, InFlight: make(map[c.MoveID]c.RegKey), Registers: map[c.RegKey]c.Register{
		carried0: {Value: &initial}, computed0: {Contract: c.Contract{Name: incName}}, key("carried", 1): {Contract: c.Contract{Name: switchName}}, exit: {Contract: c.Contract{Name: switchName}},
	}}
	p.Moves = []c.Move{
		{Position: 0, Target: incName, Reads: []c.RegKey{carried0}, Writes: computed0, Contract: c.Contract{Name: incName}, Site: c.Local{}},
		{Position: 1, Target: switchName, Reads: []c.RegKey{computed0}, Writes: key("carried", 1), Contract: c.Contract{Name: switchName}, Site: c.Local{}},
	}
	iterations := 0
	machine.RegisterEntity(incName, c.Name{Value: "alu"}, entity.Function(func(operands []c.Observation, _ c.Payload) (c.Content, c.Outcome) {
		iterations++
		value, _ := strconv.Atoi(string(operands[0].Content.Bytes))
		return c.Content{Bytes: []byte(strconv.Itoa(value + 1))}, c.Success
	}))
	machine.RegisterEntity(switchName, c.Name{Value: "switch-unit"}, entity.Function(func(operands []c.Observation, _ c.Payload) (c.Content, c.Outcome) {
		value, _ := strconv.Atoi(string(operands[0].Content.Bytes))
		for i := range p.Moves {
			if p.Moves[i].Target == switchName && p.Moves[i].State == c.Issued {
				if value < 2 {
					iter := p.Moves[i].Writes.Tag.Iter
					computed, next := key("computed", iter), key("carried", iter+1)
					p.Registers[computed] = c.Register{Contract: c.Contract{Name: incName}}
					p.Registers[next] = c.Register{Contract: c.Contract{Name: switchName}}
					p.Moves = append(p.Moves,
						c.Move{Position: 0, Target: incName, Reads: []c.RegKey{p.Moves[i].Writes}, Writes: computed, Contract: c.Contract{Name: incName}, Site: c.Local{}},
						c.Move{Position: 1, Target: switchName, Reads: []c.RegKey{computed}, Writes: next, Contract: c.Contract{Name: switchName}, Site: c.Local{}},
					)
				} else {
					p.Moves[i].Writes = exit
				}
				break
			}
		}
		return c.Content{Bytes: append([]byte(nil), operands[0].Content.Bytes...)}, c.Success
	}))
	if out := machine.Drive(p); out != c.Success {
		t.Fatalf("loop drive: %v", out)
	}
	if iterations != 2 || len(p.Moves) != 4 || p.Result == nil || string(p.Result.Content.Bytes) != "2" {
		t.Fatalf("loop result: iterations=%d moves=%d result=%+v", iterations, len(p.Moves), p.Result)
	}
	if p.Registers[key("carried", 1)].Value == nil || p.Moves[2].Reads[0].Tag.Iter != 1 || p.Retired != len(p.Moves) || len(p.InFlight) != 0 {
		t.Fatalf("tags/frontier/watermark did not drain: %+v", p)
	}
}
