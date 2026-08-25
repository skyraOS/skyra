package ctu_test

import (
	"bytes"
	"path/filepath"
	"testing"

	c "skyra-v1/src/skyra/skyra-dna-v7/causal"
	"skyra-v1/src/skyra/skyra-dna-v7/ctu"
	macos "skyra-v1/src/skyra/skyra-dna-v7/drivers/mac-os"
	ports "skyra-v1/src/skyra/skyra-dna-v7/drivers/mac-os/port-subsystem"
	"skyra-v1/src/skyra/skyra-dna-v7/entity"
	"skyra-v1/src/skyra/skyra-dna-v7/kernel"
)

type fabricFunc func(c.Target, c.Move, *c.Observation) c.Outcome

func (f fabricFunc) Approach(from c.Target, move c.Move, returned *c.Observation) c.Outcome {
	return f(from, move, returned)
}

func TestMoveFlowsRequestToEntityAndResponseToParent(t *testing.T) {
	echo := c.Name{Value: "echo"}
	storage := macos.New(filepath.Join(t.TempDir(), "graph.json"))
	if got := storage.Remember(c.Cone{Engrams: []c.Engram{{Name: echo}}}); got != c.Success {
		t.Fatalf("seed durable graph: got %v", got)
	}

	fabric := ports.New(nil)
	client := ctu.New("client", fabric, nil)
	server := ctu.New("server", fabric, kernel.New(storage, nil))
	server.RegisterEntity(echo, c.Name{Value: "echo-worker"}, entity.Function(
		func(_ []c.Observation, payload c.Payload) (c.Content, c.Outcome) {
			return c.Content{Bytes: append([]byte("echo:"), payload.Body...)}, c.Success
		},
	))
	if got := fabric.Register("client", client); got != c.Success {
		t.Fatalf("register client: got %v", got)
	}
	if got := fabric.Register("server", server); got != c.Success {
		t.Fatalf("register server: got %v", got)
	}

	correlation := c.Correlation{ID: "request-1"}
	if got := client.Expect(correlation, c.Contract{Name: echo}); got != c.Success {
		t.Fatalf("install return contract: got %v", got)
	}
	move := c.Move{
		Header:      c.Request,
		Target:      c.Target{Endpoint: "server", Name: echo},
		Horizon:     c.Horizon{Root: c.Name{Value: "account"}},
		Payload:     c.Payload{Body: []byte("hello")},
		Correlation: correlation,
	}
	if got := client.Move(move); got != c.Success {
		t.Fatalf("MOVE custody: got %v", got)
	}

	observation, outcome, complete := client.Result(correlation)
	if !complete || outcome != c.Success {
		t.Fatalf("response result: complete=%v outcome=%v", complete, outcome)
	}
	if !bytes.Equal(observation.Content.Bytes, []byte("echo:hello")) {
		t.Fatalf("response content: got %q", observation.Content.Bytes)
	}
	if observation.Provenance.Resource != echo || observation.Provenance.Invocation != correlation.ID {
		t.Fatalf("response provenance: got %+v", observation.Provenance)
	}
	if got := client.Orient().Count; got != 0 {
		t.Fatalf("outstanding after response: got %d", got)
	}
}

func TestBoundNarrowsWithoutMutatingParentAndThinkCapsAtFive(t *testing.T) {
	machine := ctu.New("endpoint", nil, nil)
	root := c.Horizon{Root: c.Name{Value: "account"}, Caveats: []c.Caveat{{Constraint: "a"}}}
	narrowed := machine.Bound(root, c.Caveat{Constraint: "b"})
	if len(root.Caveats) != 1 || len(narrowed.Caveats) != 2 {
		t.Fatalf("attenuation mutated parent or failed to append: parent=%v child=%v", root, narrowed)
	}
	for i := 0; i < 6; i++ {
		name := c.Name{Value: string(rune('a' + i))}
		machine.RegisterEntity(name, name, entity.Function(func([]c.Observation, c.Payload) (c.Content, c.Outcome) {
			return c.Content{}, c.Success
		}))
	}
	if got := len(machine.Think(c.Goal{}, c.Frame{})); got != 5 {
		t.Fatalf("THINK cap: got %d", got)
	}
}

func TestRecognizeMakesContractMismatchATotalFailure(t *testing.T) {
	fabric := fabricFunc(func(c.Target, c.Move, *c.Observation) c.Outcome { return c.Success })
	machine := ctu.New("parent", fabric, nil)
	correlation := c.Correlation{ID: "mismatch"}
	if got := machine.Expect(correlation, c.Contract{Name: c.Name{Value: "expected"}}); got != c.Success {
		t.Fatal(got)
	}
	request := c.Move{
		Header:      c.Request,
		Target:      c.Target{Endpoint: "child", Name: c.Name{Value: "expected"}},
		Horizon:     c.Horizon{Root: c.Name{Value: "account"}},
		Payload:     c.Payload{Body: []byte("request")},
		Correlation: correlation,
	}
	if got := machine.Move(request); got != c.Success {
		t.Fatal(got)
	}
	response := request
	response.Header = c.Response
	returned := c.Observation{Provenance: c.Provenance{
		Resource:   c.Name{Value: "wrong"},
		Invocation: correlation.ID,
	}}
	if got := machine.Recognize(response, returned); got != c.Failure {
		t.Fatalf("RECOGNIZE mismatch: got %v", got)
	}
	_, outcome, complete := machine.Result(correlation)
	if !complete || outcome != c.Failure {
		t.Fatalf("total mismatch result: complete=%v outcome=%v", complete, outcome)
	}
	if got := machine.Orient().Count; got != 0 {
		t.Fatalf("returned custody still outstanding: %d", got)
	}
}

func TestFabricRefusesUnverifiedCaveatWithoutReadingPayload(t *testing.T) {
	fabric := ports.New(nil)
	client := ctu.New("client", fabric, nil)
	if got := fabric.Register("client", client); got != c.Success {
		t.Fatal(got)
	}
	move := c.Move{
		Header:  c.Response,
		Target:  c.Target{Endpoint: "client"},
		Horizon: c.Horizon{Root: c.Name{Value: "account"}, Caveats: []c.Caveat{{Constraint: "narrow"}}},
		Payload: c.Payload{Body: []byte("opaque")},
	}
	if got := fabric.Approach(c.Target{Endpoint: "server"}, move, nil); got != c.Failure {
		t.Fatalf("unverified caveat: got %v", got)
	}
}
