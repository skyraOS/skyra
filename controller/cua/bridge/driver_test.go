package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"skyra/controller/cua/observation"
)

var testScope = observation.Scope{DriverVersion: "0.34.0", ManifestHash: strings.Repeat("a", 64), BundleID: "com.example.Fixture", PID: 7, WindowID: 9}

// unreachable is an upstream every launch of which fails: for requests the
// bridge must refuse before reaching any daemon.
func unreachable(t *testing.T) *upstream {
	t.Helper()
	return newUpstream(context.Background(), testScope, func(context.Context, int) (*daemon, error) {
		t.Error("launched a daemon for a request refused before connecting")
		return nil, errors.New("no daemon")
	})
}

// fake is a Cua daemon: it verifies as testScope, and its calls either
// observe or report the latched idle timeout.
type fake struct {
	socket  string
	expired bool
	stopped bool
	calls   int
	mu      sync.Mutex
}

func startFake(t *testing.T, expired bool) *fake {
	t.Helper()
	// Darwin's Unix socket path limit is shorter than named t.TempDir paths.
	dir, err := os.MkdirTemp("/private/tmp", "skyra-cua-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	f := &fake{socket: filepath.Join(dir, "driver.sock"), expired: expired}
	l, err := net.Listen("unix", f.socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go f.serve(c)
		}
	}()
	return f
}

func (f *fake) serve(c net.Conn) {
	defer c.Close()
	s := bufio.NewScanner(c)
	for s.Scan() {
		var q map[string]any
		if json.Unmarshal(s.Bytes(), &q) != nil {
			return
		}
		reply := map[string]any{"ok": true}
		switch q["method"] {
		case "metadata":
			reply["result"] = map[string]any{"driver_version": testScope.DriverVersion}
		case "authorization_status":
			reply["result"] = map[string]any{"permission_mode": "bounded", "capability_manifest_configured": true, "capability_manifest_valid": true,
				"capability_manifest_approved_at_startup": true, "capability_manifest": map[string]string{"sha256": testScope.ManifestHash}}
		case "call":
			f.mu.Lock()
			f.calls++
			f.mu.Unlock()
			if f.expired {
				reply = map[string]any{"ok": false, "error": "Policy loading error: capability manifest idle timeout exceeded"}
			} else {
				reply["result"] = map[string]any{"content": []any{map[string]string{"type": "text", "text": "Count: 42"}}, "structuredContent": map[string]string{"snapshot_id": "fixture-1"}}
			}
		default:
			return
		}
		if json.NewEncoder(c).Encode(reply) != nil {
			return
		}
	}
}

func (f *fake) daemon() *daemon {
	return &daemon{Socket: f.socket, PID: 1, stop: func() { f.mu.Lock(); f.stopped = true; f.mu.Unlock() }}
}

// fleet launches the given fakes in order, one per launch.
func fleet(t *testing.T, fakes ...*fake) (func(context.Context, int) (*daemon, error), *int) {
	launches := 0
	return func(context.Context, int) (*daemon, error) {
		launches++
		if launches > len(fakes) {
			return nil, errors.New("no more daemons")
		}
		return fakes[launches-1].daemon(), nil
	}, &launches
}

func request(t *testing.T, up *upstream) observation.Result {
	t.Helper()
	return handle(context.Background(), strings.NewReader(`{"pid":7,"window_id":9}`+"\n"), up, testScope)
}

func TestIdleExpiredDaemonIsReplacedAndTheObservationRetried(t *testing.T) {
	old, next := startFake(t, true), startFake(t, false)
	launch, launches := fleet(t, old, next)
	up := newUpstream(context.Background(), testScope, launch)
	out := request(t, up)
	if out.Error != nil || !strings.Contains(out.Text, "Count: 42") {
		t.Fatalf("%+v", out)
	}
	if *launches != 2 || !old.stopped || next.stopped || old.calls != 1 || next.calls != 1 {
		t.Fatalf("launches=%d old stopped=%v calls=%d, next stopped=%v calls=%d", *launches, old.stopped, old.calls, next.stopped, next.calls)
	}
	if out = request(t, up); out.Error != nil || *launches != 2 {
		t.Fatalf("the replacement was not kept: %+v, launches=%d", out, *launches)
	}
}

func TestAnObservationIsRetriedOnlyOnce(t *testing.T) {
	launch, launches := fleet(t, startFake(t, true), startFake(t, true), startFake(t, false))
	out := request(t, newUpstream(context.Background(), testScope, launch))
	if out.Error == nil || !strings.Contains(out.Error.Message, "idle timeout exceeded") || *launches != 2 {
		t.Fatalf("%+v, launches=%d", out, *launches)
	}
}

func TestAStaleDaemonIsReplacedOnceForEveryRequestThatSawIt(t *testing.T) {
	launch, launches := fleet(t, startFake(t, false), startFake(t, false))
	up := newUpstream(context.Background(), testScope, launch)
	first, err := up.get()
	if err != nil {
		t.Fatal(err)
	}
	a, errA := up.replace(first)
	b, errB := up.replace(first)
	if errA != nil || errB != nil || a != b || *launches != 2 {
		t.Fatalf("a=%p b=%p launches=%d: %v %v", a, b, *launches, errA, errB)
	}
}

func TestFailedLaunchesBackOffExponentially(t *testing.T) {
	launches := 0
	up := newUpstream(context.Background(), testScope, func(context.Context, int) (*daemon, error) {
		launches++
		return nil, errors.New("Cua did not start")
	})
	now := time.Unix(0, 0)
	up.now = func() time.Time { return now }
	for i, want := range []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second} {
		out := request(t, up)
		if out.Error == nil || out.Error.Code != "unavailable" || !strings.Contains(out.Error.Message, "desktop driver unavailable") || launches != i+1 {
			t.Fatalf("attempt %d: %+v, launches=%d", i, out, launches)
		}
		// Inside the backoff a request is refused without launching.
		now = now.Add(want - time.Millisecond)
		if out = request(t, up); out.Error == nil || !strings.Contains(out.Error.Message, "backing off") || launches != i+1 {
			t.Fatalf("attempt %d inside the backoff: %+v, launches=%d", i, out, launches)
		}
		now = now.Add(time.Millisecond)
	}
	up.failures = 20
	up.retryAt = time.Time{}
	request(t, up)
	if got := up.retryAt.Sub(now); got != up.max {
		t.Fatalf("backoff %s, want the cap %s", got, up.max)
	}
}

func TestALaunchThatFailsVerificationIsStopped(t *testing.T) {
	f := startFake(t, false)
	wrong := testScope
	wrong.ManifestHash = strings.Repeat("b", 64)
	up := newUpstream(context.Background(), wrong, func(context.Context, int) (*daemon, error) { return f.daemon(), nil })
	if _, err := up.get(); err == nil || !f.stopped {
		t.Fatalf("kept a daemon holding another manifest: %v, stopped=%v", err, f.stopped)
	}
}

func TestOnlyALapsedOrMissingDaemonIsStale(t *testing.T) {
	r := observation.NewResult(testScope)
	for _, c := range []struct {
		code, message string
		want          bool
	}{
		{"transport_error", "driver: Policy loading error: capability manifest idle timeout exceeded", true},
		{"transport_error", "driver: Policy loading error: capability manifest expired", true},
		{"unavailable", "dial unix /x: connect: no such file or directory", true},
		{"transport_error", "incomplete frame: EOF", true},
		{"transport_error", "read unix ->/x: i/o timeout", false},
		{"invalid_response", "multiple screenshots", false},
		{"outside_scope", "request does not name the selected window", false},
	} {
		if got := stale(r.Failed(c.code, errors.New(c.message))); got != c.want {
			t.Errorf("%s %q: stale=%v", c.code, c.message, got)
		}
	}
	if stale(r) {
		t.Error("a success is stale")
	}
}
