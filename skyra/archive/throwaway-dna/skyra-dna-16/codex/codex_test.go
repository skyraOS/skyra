package codex

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type codexReceive struct {
	frame []byte
	err   error
}

type fakeCodexTransport struct {
	sent chan []byte
	recv chan codexReceive
	done chan struct{}
	once sync.Once
}

func newFakeCodexTransport() *fakeCodexTransport {
	return &fakeCodexTransport{sent: make(chan []byte, 32), recv: make(chan codexReceive, 32), done: make(chan struct{})}
}

func (transport *fakeCodexTransport) Send(ctx context.Context, frame []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-transport.done:
		return errors.New("closed")
	case transport.sent <- append([]byte(nil), frame...):
		return nil
	}
}

func (transport *fakeCodexTransport) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-transport.done:
		return nil, errors.New("closed")
	case result := <-transport.recv:
		return append([]byte(nil), result.frame...), result.err
	}
}

func (transport *fakeCodexTransport) Close() error {
	transport.once.Do(func() { close(transport.done) })
	return nil
}

func (transport *fakeCodexTransport) deliver(value string) {
	transport.recv <- codexReceive{frame: []byte(value)}
}

func receiveCodexMessage(t *testing.T, transport *fakeCodexTransport) map[string]any {
	t.Helper()
	select {
	case frame := <-transport.sent:
		var decoded map[string]any
		if err := json.Unmarshal(frame, &decoded); err != nil {
			t.Fatal(err)
		}
		return decoded
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Codex output")
		return nil
	}
}

func startCodexFixture(t *testing.T) (*CodexDriver, *fakeCodexTransport, CTU, chan RootResult) {
	t.Helper()
	transport := newFakeCodexTransport()
	driver, err := NewCodexDriver(context.Background(), transport)
	if err != nil {
		t.Fatal(err)
	}
	if message := receiveCodexMessage(t, transport); message["method"] != "initialize" {
		t.Fatalf("first message = %#v", message)
	}
	transport.deliver(`{"id":1,"result":{}}`)
	if message := receiveCodexMessage(t, transport); message["method"] != "initialized" {
		t.Fatalf("second message = %#v", message)
	}
	weather := Capability{Target: Target{Entity: EntityID{Value: "weather"}, Address: "weather"}, Action: Action{Value: "forecast"}}
	rootEndpoint := Endpoint{Address: "codex", Registration: DriverRegistration{Driver: driver, Input: CodexInput{}}}
	kernel := &FixtureKernel{Root: rootEndpoint, Local: []Endpoint{{Address: "weather", Registration: DriverRegistration{Driver: &EchoDriver{}, Input: ProbeInput{}}}}}
	results := make(chan RootResult, 2)
	runtime, err := NewCTU(BootConfig{KernelAddress: "kernel", Kernel: kernel, KernelInput: CodexInput{}, Root: rootEndpoint}, func(result RootResult) error {
		results <- result
		return nil
	})
	if err != nil {
		driver.Close()
		t.Fatal(err)
	}
	move := Move{Root: PossibilityID{Value: "possibility"}, Target: Target{Address: "codex"}, Op: Invoke{
		InvocationID: InvocationID{Value: "parent"}, Action: Action{Value: "answer"}, Params: []Observation{testObservation("weather?")},
		Horizon: Horizon{Capabilities: []Capability{{Target: Target{Address: "codex"}, Action: Action{Value: "answer"}}, weather}}, Focus: []Capability{weather},
	}}
	if err := runtime.Submit(context.Background(), move); err != nil {
		driver.Close()
		t.Fatal(err)
	}
	threadStart := receiveCodexMessage(t, transport)
	if threadStart["method"] != "thread/start" {
		t.Fatalf("thread start = %#v", threadStart)
	}
	return driver, transport, runtime, results
}

func advanceCodexToRunning(t *testing.T, transport *fakeCodexTransport) {
	t.Helper()
	transport.deliver(`{"id":101,"result":{"thread":{"id":"thread_parent"}}}`)
	turnStart := receiveCodexMessage(t, transport)
	if turnStart["method"] != "turn/start" {
		t.Fatalf("turn start = %#v", turnStart)
	}
	transport.deliver(`{"id":102,"result":{"turn":{"id":"turn_parent","status":"inProgress"}}}`)
}

func TestCodexDynamicToolRoundTrip(t *testing.T) {
	driver, transport, _, results := startCodexFixture(t)
	defer driver.Close()
	advanceCodexToRunning(t, transport)
	transport.deliver(`{"method":"item/tool/call","id":88,"params":{"arguments":{"city":"Boston"},"callId":"call_1","namespace":null,"threadId":"thread_parent","tool":"cap_000","turnId":"turn_parent"}}`)
	response := receiveCodexMessage(t, transport)
	if response["id"] != float64(88) {
		t.Fatalf("tool response = %#v", response)
	}
	transport.deliver(`{"method":"item/completed","params":{"threadId":"thread_parent","turnId":"turn_parent","item":{"type":"agentMessage","text":"Boston is clear"}}}`)
	transport.deliver(`{"method":"turn/completed","params":{"threadId":"thread_parent","turn":{"id":"turn_parent","status":"completed"}}}`)
	select {
	case result := <-results:
		if result.Err != nil || result.Observation == nil || string(result.Observation.Content.Bytes) != "Boston is clear" {
			t.Fatalf("result = %#v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Codex result")
	}
}

func TestCodexNativeItemContaminatesAndFaults(t *testing.T) {
	driver, transport, runtime, results := startCodexFixture(t)
	defer driver.Close()
	advanceCodexToRunning(t, transport)
	transport.deliver(`{"method":"item/started","params":{"threadId":"thread_parent","turnId":"turn_parent","item":{"type":"commandExecution","command":"uname"}}}`)
	select {
	case result := <-results:
		if result.Err == nil {
			t.Fatal("native item did not fault root")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for contamination fault")
	}
	snapshot, exists := runtime.Snapshot(InvocationID{Value: "parent"})
	if !exists || snapshot.State != Faulted {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	state, err := decodeCodexState(snapshot.DriverState)
	if err != nil || !state.Contaminated || state.Phase != CodexFaulted {
		t.Fatalf("driver state = %#v, %v", state, err)
	}
	transport.deliver(`{"method":"item/completed","params":{"threadId":"thread_parent","turnId":"turn_parent","item":{"type":"agentMessage","text":"must not escape"}}}`)
	transport.deliver(`{"method":"turn/completed","params":{"threadId":"thread_parent","turn":{"id":"turn_parent","status":"completed"}}}`)
	select {
	case duplicate := <-results:
		t.Fatalf("contaminated invocation produced another result: %#v", duplicate)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestCodexRejectsToolOutsideFocus(t *testing.T) {
	driver, transport, runtime, results := startCodexFixture(t)
	defer driver.Close()
	advanceCodexToRunning(t, transport)
	transport.deliver(`{"method":"item/tool/call","id":89,"params":{"arguments":{},"callId":"call_bad","namespace":null,"threadId":"thread_parent","tool":"cap_999","turnId":"turn_parent"}}`)
	select {
	case result := <-results:
		if result.Err == nil {
			t.Fatal("unknown tool did not fault root")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for unknown-tool fault")
	}
	snapshot, _ := runtime.Snapshot(InvocationID{Value: "parent"})
	state, err := decodeCodexState(snapshot.DriverState)
	if err != nil || !state.Contaminated {
		t.Fatalf("unknown tool did not persist contamination: %#v, %v", state, err)
	}
}

func TestCodexRestoreFaultsWhenLiveRequestCannotBeReconciled(t *testing.T) {
	transport := newFakeCodexTransport()
	driver, err := NewCodexDriver(context.Background(), transport)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.Close()
	_ = receiveCodexMessage(t, transport)
	transport.deliver(`{"id":1,"result":{}}`)
	_ = receiveCodexMessage(t, transport)
	state, _ := encodeCodexState(CodexInvocationState{Phase: CodexRunning, ThreadID: "lost", TurnID: "lost"})
	emissions := make(chan DriverEmission, 1)
	move := Move{Target: Target{Address: "codex"}, Op: Invoke{InvocationID: InvocationID{Value: "restored"}, Action: Action{Value: "answer"}}}
	if err := driver.Ingress(context.Background(), DriverInput{Move: move, State: state}, func(emission DriverEmission, _ error) error {
		emissions <- emission
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	emission := <-emissions
	if emission.Failure == nil {
		t.Fatal("unreconcilable restore did not fault")
	}
}

func TestLiveCodexAppServerHandshake(t *testing.T) {
	if os.Getenv("SKYRA_LIVE_CODEX") != "1" {
		t.Skip("set SKYRA_LIVE_CODEX=1 to run the installed App Server handshake")
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	version, err := codexBinaryVersion(context.Background(), binary)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	host, err := StartCodexHost(context.Background(), CodexLaunchPolicy{
		Binary: binary, ExpectedVersion: version, CodexHome: directory, WorkingDir: directory,
		ExecutableSearchPath: []string{filepath.Dir(binary), "/usr/bin", "/bin"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close()
	driver, err := NewCodexDriver(context.Background(), host.Transport())
	if err != nil {
		t.Fatal(err)
	}
	defer driver.Close()
	select {
	case <-driver.ready:
	case <-driver.Done():
		t.Fatalf("App Server exited during handshake: %v: %s", host.WaitError(), host.Stderr())
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for live App Server handshake")
	}
}
