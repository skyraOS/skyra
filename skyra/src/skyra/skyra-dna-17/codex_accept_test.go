package dna17

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// In-process fake App Server transport. Deterministic channels only — no timing
// sleeps are used to order concurrency. The 2s guards on blocking channel reads are
// failure fail-safes, not concurrency waits.
// ---------------------------------------------------------------------------

type fakeReceive struct {
	frame []byte
	err   error
}

type fakeTransport struct {
	sent chan []byte
	recv chan fakeReceive
	done chan struct{}
	once sync.Once
}

func newFakeTransport() *fakeTransport {
	return &fakeTransport{sent: make(chan []byte, 32), recv: make(chan fakeReceive, 32), done: make(chan struct{})}
}

func (t *fakeTransport) Send(ctx context.Context, frame []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.done:
		return errors.New("closed")
	case t.sent <- append([]byte(nil), frame...):
		return nil
	}
}

func (t *fakeTransport) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-t.done:
		return nil, errors.New("closed")
	case r := <-t.recv:
		return append([]byte(nil), r.frame...), r.err
	}
}

func (t *fakeTransport) Close() error {
	t.once.Do(func() { close(t.done) })
	return nil
}

func (t *fakeTransport) deliver(value string) { t.recv <- fakeReceive{frame: []byte(value)} }

func receiveMessage(t *testing.T, transport *fakeTransport) map[string]any {
	t.Helper()
	select {
	case frame := <-transport.sent:
		var decoded map[string]any
		if err := json.Unmarshal(frame, &decoded); err != nil {
			t.Fatal(err)
		}
		return decoded
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for a Codex output frame")
		return nil
	}
}

// ---------------------------------------------------------------------------
// Fixture: handshake, boot CU, submit parent Invoke A, reach the running turn.
// ---------------------------------------------------------------------------

const (
	codexSurface   SurfaceID = "codex"
	weatherSurface SurfaceID = "weather"
)

func startFixture(t *testing.T) (*CodexDriver, *fakeTransport, *CU, chan RootResult) {
	t.Helper()
	transport := newFakeTransport()
	driver, err := NewCodexDriver(context.Background(), transport)
	if err != nil {
		t.Fatal(err)
	}
	// Handshake: initialize -> {result} -> initialized -> ready.
	if m := receiveMessage(t, transport); m["method"] != "initialize" {
		t.Fatalf("first frame = %#v", m)
	}
	transport.deliver(`{"id":1,"result":{}}`)
	if m := receiveMessage(t, transport); m["method"] != "initialized" {
		t.Fatalf("second frame = %#v", m)
	}

	surfaces := map[SurfaceID]Surface{codexSurface: driver, weatherSurface: EchoSurface{}}
	results := make(chan RootResult, 4)
	cu, err := NewCU(surfaces, func(r RootResult) error { results <- r; return nil })
	if err != nil {
		driver.Close()
		t.Fatal(err)
	}

	horizon := Horizon{Grants: map[SurfaceID]Grant{
		codexSurface:   {SurfaceID: codexSurface},
		weatherSurface: {SurfaceID: weatherSurface},
	}}
	parent := Move{
		PossibilityID: "possibility",
		Target:        codexSurface,
		Invoke: &Invoke{
			InvocationID: "parent",
			Action:       "answer",
			Params:       []Observation{{Bytes: []byte("weather?")}},
			Focus:        []Capability{{Target: weatherSurface, Action: "forecast"}},
		},
	}
	if err := cu.Submit(context.Background(), parent, horizon); err != nil {
		cu.Close()
		driver.Close()
		t.Fatal(err)
	}
	if m := receiveMessage(t, transport); m["method"] != "thread/start" {
		t.Fatalf("thread start = %#v", m)
	}
	return driver, transport, cu, results
}

func advanceToRunning(t *testing.T, transport *fakeTransport) {
	t.Helper()
	transport.deliver(`{"id":101,"result":{"thread":{"id":"thread_parent"}}}`)
	if m := receiveMessage(t, transport); m["method"] != "turn/start" {
		t.Fatalf("turn start = %#v", m)
	}
	transport.deliver(`{"id":102,"result":{"turn":{"id":"turn_parent","status":"inProgress"}}}`)
}

// ---------------------------------------------------------------------------
// Happy path: the full dynamic-tool round trip (workorder "First acceptance path").
// ---------------------------------------------------------------------------

func TestCodexDynamicToolRoundTrip(t *testing.T) {
	driver, transport, cu, results := startFixture(t)
	defer cu.Close()
	defer driver.Close()

	advanceToRunning(t, transport)

	// (4) receive a dynamic ToolCall C -> (5) commit waiting + child Invoke B ->
	// (6) child B runs through the ordinary loop (echo) -> (7) child Return B applied
	// to parent -> (8) commit resumed state and answer ToolCall C.
	transport.deliver(`{"method":"item/tool/call","id":88,"params":{"arguments":{"city":"Boston"},"callId":"call_1","namespace":null,"threadId":"thread_parent","tool":"cap_000","turnId":"turn_parent"}}`)
	response := receiveMessage(t, transport)
	if response["id"] != float64(88) {
		t.Fatalf("tool response = %#v", response)
	}

	// (9) clean final agent message + (10) turn completion -> terminal Return A.
	transport.deliver(`{"method":"item/completed","params":{"threadId":"thread_parent","turnId":"turn_parent","item":{"type":"agentMessage","text":"Boston is clear"}}}`)
	transport.deliver(`{"method":"turn/completed","params":{"threadId":"thread_parent","turn":{"id":"turn_parent","status":"completed"}}}`)

	// (11) deliver exactly one root result.
	select {
	case r := <-results:
		if r.Err != nil || r.Observation == nil || string(r.Observation.Bytes) != "Boston is clear" {
			t.Fatalf("root result = %#v (obs=%q, err=%v)", r, obsText(r.Observation), r.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the root result")
	}

	if snap, ok := cu.Snapshot("parent"); !ok || snap.Phase != PhaseReturned {
		t.Fatalf("parent phase = %#v", snap)
	}

	// Exactly one root result.
	select {
	case extra := <-results:
		t.Fatalf("second root result produced: %#v", extra)
	default:
	}
}

// ---------------------------------------------------------------------------
// Rejection path: a native effect-bearing item contaminates and faults, and a
// later agent message + turn/completed CANNOT produce a Skyra Return.
// ---------------------------------------------------------------------------

func TestCodexNativeItemContaminatesAndBlocksReturn(t *testing.T) {
	driver, transport, cu, results := startFixture(t)
	defer cu.Close()
	defer driver.Close()

	advanceToRunning(t, transport)

	// Inject a native effect-bearing item/started -> contamination -> fault Move ->
	// exactly one root fault result.
	transport.deliver(`{"method":"item/started","params":{"threadId":"thread_parent","turnId":"turn_parent","item":{"type":"commandExecution","command":"uname"}}}`)
	select {
	case r := <-results:
		if r.Err == nil {
			t.Fatalf("native item did not fault the root: %#v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the contamination fault")
	}

	snap, ok := cu.Snapshot("parent")
	if !ok || snap.Phase != PhaseFaulted {
		t.Fatalf("parent phase = %#v", snap)
	}
	state, err := decodeCodexState(snap.DriverState)
	if err != nil || !state.Contaminated || state.Phase != codexFaulted {
		t.Fatalf("driver state = %#v (err=%v)", state, err)
	}

	// A later clean agent message + turn/completed must NOT resurrect a Return. The
	// fd is dead, so these emissions are rejected by the settlement fence.
	transport.deliver(`{"method":"item/completed","params":{"threadId":"thread_parent","turnId":"turn_parent","item":{"type":"agentMessage","text":"must not escape"}}}`)
	transport.deliver(`{"method":"turn/completed","params":{"threadId":"thread_parent","turn":{"id":"turn_parent","status":"completed"}}}`)

	// Deterministic barrier (no sleep): a malformed frame makes the driver's single
	// receive goroutine fail and close Done AFTER it has processed the two frames
	// above in FIFO order. Once Done is closed we know they were consumed.
	transport.deliver(`not-json`)
	select {
	case <-driver.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the receive goroutine barrier")
	}

	// The two post-fault emissions committed nothing (fd dead), so no Move was
	// enqueued and no second root result can exist.
	select {
	case extra := <-results:
		t.Fatalf("post-fault emission produced a result: %#v", extra)
	default:
	}
}

// ---------------------------------------------------------------------------
// Sanity: the collapsed loop on a plain echo Surface (no provider), covering
// invoke -> execute -> return -> root result through the ordinary path.
// ---------------------------------------------------------------------------

func TestEchoLoopReturnsRoot(t *testing.T) {
	results := make(chan RootResult, 1)
	cu, err := NewCU(map[SurfaceID]Surface{"echo": EchoSurface{}}, func(r RootResult) error { results <- r; return nil })
	if err != nil {
		t.Fatal(err)
	}
	defer cu.Close()
	move := Move{
		PossibilityID: "p",
		Target:        "echo",
		Invoke:        &Invoke{InvocationID: "root", Action: "echo", Params: []Observation{{Bytes: []byte("ping")}}},
	}
	horizon := Horizon{Grants: map[SurfaceID]Grant{"echo": {SurfaceID: "echo"}}}
	if err := cu.Submit(context.Background(), move, horizon); err != nil {
		t.Fatal(err)
	}
	select {
	case r := <-results:
		if r.Err != nil || r.Observation == nil || string(r.Observation.Bytes) != "ping" {
			t.Fatalf("root result = %#v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for echo root result")
	}
}

// TestUnauthorizedTargetTraps proves the pre-execution trap: an invoke whose target
// is outside the Horizon faults instead of executing.
func TestUnauthorizedTargetTraps(t *testing.T) {
	results := make(chan RootResult, 1)
	cu, err := NewCU(map[SurfaceID]Surface{"echo": EchoSurface{}}, func(r RootResult) error { results <- r; return nil })
	if err != nil {
		t.Fatal(err)
	}
	defer cu.Close()
	move := Move{
		PossibilityID: "p",
		Target:        "echo",
		Invoke:        &Invoke{InvocationID: "root", Action: "echo", Params: []Observation{{Bytes: []byte("ping")}}},
	}
	// Empty horizon: "echo" is not granted.
	if err := cu.Submit(context.Background(), move, Horizon{Grants: map[SurfaceID]Grant{}}); err != nil {
		t.Fatal(err)
	}
	select {
	case r := <-results:
		if r.Err == nil {
			t.Fatalf("unauthorized target did not trap: %#v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the trap fault")
	}
}

func obsText(o *Observation) string {
	if o == nil {
		return "<nil>"
	}
	return string(o.Bytes)
}
