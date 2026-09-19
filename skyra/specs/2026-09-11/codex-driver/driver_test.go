package codexdriver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"skyra-v1/src/skyra/skyra-dna-15/protocol"
)

type receiveResult struct {
	message []byte
	err     error
}

type fakeTransport struct {
	sent   chan []byte
	recv   chan receiveResult
	closed chan struct{}
	once   sync.Once
}

func newFakeTransport() *fakeTransport {
	return &fakeTransport{
		sent:   make(chan []byte, 16),
		recv:   make(chan receiveResult, 16),
		closed: make(chan struct{}),
	}
}

func (f *fakeTransport) Send(ctx context.Context, message []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-f.closed:
		return errors.New("closed")
	case f.sent <- append([]byte(nil), message...):
		return nil
	}
}

func (f *fakeTransport) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.closed:
		return nil, errors.New("closed")
	case result := <-f.recv:
		return append([]byte(nil), result.message...), result.err
	}
}

func (f *fakeTransport) Close() error {
	f.once.Do(func() { close(f.closed) })
	return nil
}

func (f *fakeTransport) deliver(message string) {
	f.recv <- receiveResult{message: []byte(message)}
}

func receiveJSON(t *testing.T, ch <-chan []byte) map[string]any {
	t.Helper()
	message := <-ch
	var decoded map[string]any
	if err := json.Unmarshal(message, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func receiveOutput(t *testing.T, ch <-chan outputResult) outputResult {
	t.Helper()
	return <-ch
}

type outputResult struct {
	raw string
	err error
}

func outputChannel() (chan outputResult, func(string, error) error) {
	results := make(chan outputResult, 1)
	return results, func(raw string, err error) error {
		results <- outputResult{raw: raw, err: err}
		return nil
	}
}

func testMove() protocol.Move {
	parameter := protocol.Observation{}
	parameter.Segment.By.Value = "location"
	parameter.Content.Bytes = []byte(`"Boston"`)
	capability := protocol.Capability{
		Target: protocol.Target{Entity: protocol.EntityID{Value: "weather"}, Address: "api"},
		Action: protocol.Action{Value: "forecast"},
	}
	return protocol.Move{
		Root: protocol.PossibilityID{Value: "possibility"},
		Op: protocol.Invoke{
			InvocationID: protocol.InvocationID{Value: "parent"},
			Action:       protocol.Action{Value: "answer"},
			Params:       []protocol.Observation{parameter},
			Focus:        []protocol.Capability{capability},
		},
	}
}

func startDriver(t *testing.T) (*Driver, *fakeTransport, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	transport := newFakeTransport()
	d, err := New(ctx, transport)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	initialize := receiveJSON(t, transport.sent)
	if initialize["method"] != "initialize" {
		t.Fatalf("first message = %#v", initialize)
	}
	initialized := receiveJSON(t, transport.sent)
	if initialized["method"] != "initialized" {
		t.Fatalf("second message = %#v", initialized)
	}
	return d, transport, cancel
}

func startInvocation(t *testing.T, d *Driver, transport *fakeTransport, output func(string, error) error) {
	t.Helper()
	if err := d.Ingress(context.Background(), testMove(), output); err != nil {
		t.Fatal(err)
	}
	threadStart := receiveJSON(t, transport.sent)
	if threadStart["method"] != "thread/start" {
		t.Fatalf("message = %#v", threadStart)
	}
	params := threadStart["params"].(map[string]any)
	tools := params["dynamicTools"].([]any)
	if len(tools) != 1 || tools[0].(map[string]any)["name"] != "cap_000" {
		t.Fatalf("tools = %#v", tools)
	}
	transport.deliver(`{"id":2,"result":{"thread":{"id":"thr_parent"}}}`)
	turnStart := receiveJSON(t, transport.sent)
	if turnStart["method"] != "turn/start" {
		t.Fatalf("message = %#v", turnStart)
	}
	turnParams := turnStart["params"].(map[string]any)
	input := turnParams["input"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(input, `"method":"answer"`) || !strings.Contains(input, `"location"`) {
		t.Fatalf("turn input = %s", input)
	}
	transport.deliver(`{"id":3,"result":{"turn":{"id":"turn_parent","status":"inProgress"}}}`)
}

func TestIngressAndEgressReturn(t *testing.T) {
	d, transport, cancel := startDriver(t)
	defer cancel()
	defer d.Close()

	outputs, output := outputChannel()
	startInvocation(t, d, transport, output)
	transport.deliver(`{"method":"item/completed","params":{"threadId":"thr_parent","turnId":"turn_parent","completedAtMs":1,"item":{"type":"agentMessage","id":"item_1","text":"sunny"}}}`)
	transport.deliver(`{"method":"turn/completed","params":{"threadId":"thr_parent","turn":{"id":"turn_parent","items":[],"status":"completed"}}}`)

	result := receiveOutput(t, outputs)
	if result.err != nil {
		t.Fatal(result.err)
	}
	moves, err := d.Egress(result.raw)
	if err != nil {
		t.Fatal(err)
	}
	returned, ok := moves[0].Op.(protocol.Return)
	if !ok || string(returned.Observation.Content.Bytes) != "sunny" {
		t.Fatalf("moves = %#v", moves)
	}
}

func TestToolCallBecomesChildAndReturnAnswersRequest(t *testing.T) {
	d, transport, cancel := startDriver(t)
	defer cancel()
	defer d.Close()

	firstOutputs, firstOutput := outputChannel()
	startInvocation(t, d, transport, firstOutput)
	transport.deliver(`{"method":"item/tool/call","id":88,"params":{"arguments":{"units":"c","city":"Boston"},"callId":"call_1","namespace":null,"threadId":"thr_parent","tool":"cap_000","turnId":"turn_parent"}}`)

	result := receiveOutput(t, firstOutputs)
	if result.err != nil {
		t.Fatal(result.err)
	}
	moves, err := d.Egress(result.raw)
	if err != nil {
		t.Fatal(err)
	}
	child, ok := moves[0].Op.(protocol.Invoke)
	if !ok || moves[0].Target.Address != "api" || child.Action.Value != "forecast" {
		t.Fatalf("child = %#v", moves[0])
	}
	if len(child.Params) != 2 || child.Params[0].Segment.By.Value != "city" || child.Params[1].Segment.By.Value != "units" {
		t.Fatalf("params = %#v", child.Params)
	}

	childID := protocol.InvocationID{Value: "child"}
	observation := protocol.Observation{}
	observation.Content.Bytes = []byte(`{"temperature":20}`)
	resumed := testMove()
	invocation := resumed.Op.(protocol.Invoke)
	invocation.Register = protocol.Register{
		Order: []protocol.InvocationID{childID},
		Children: map[protocol.InvocationID]protocol.RegisterEntry{
			childID: {Observation: &protocol.ChildObservation{InvocationID: childID, Observation: observation}},
		},
	}
	resumed.Op = invocation
	secondOutputs, secondOutput := outputChannel()
	if err := d.Ingress(context.Background(), resumed, secondOutput); err != nil {
		t.Fatal(err)
	}
	response := receiveJSON(t, transport.sent)
	if response["id"] != float64(88) {
		t.Fatalf("response = %#v", response)
	}
	resultBody := response["result"].(map[string]any)
	if resultBody["success"] != true {
		t.Fatalf("result = %#v", resultBody)
	}

	transport.deliver(`{"method":"item/completed","params":{"threadId":"thr_parent","turnId":"turn_parent","completedAtMs":1,"item":{"type":"agentMessage","id":"item_2","text":"20 C"}}}`)
	transport.deliver(`{"method":"turn/completed","params":{"threadId":"thr_parent","turn":{"id":"turn_parent","items":[],"status":"completed"}}}`)
	final := receiveOutput(t, secondOutputs)
	if final.err != nil {
		t.Fatal(final.err)
	}
	moves, err = d.Egress(final.raw)
	if err != nil {
		t.Fatal(err)
	}
	returned := moves[0].Op.(protocol.Return)
	if string(returned.Observation.Content.Bytes) != "20 C" {
		t.Fatalf("return = %#v", returned)
	}
}
