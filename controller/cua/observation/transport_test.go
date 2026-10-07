package observation

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func socketPath(t *testing.T, name string) string {
	t.Helper()
	// Darwin's Unix socket path limit is shorter than named t.TempDir paths.
	dir, err := os.MkdirTemp("/private/tmp", "skyra-cua-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, name)
}

func fakeDaemon(t *testing.T, version, mode, hash string) (string, *[]map[string]any, *sync.Mutex) {
	t.Helper()
	path := socketPath(t, "driver.sock")
	l, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	var requests []map[string]any
	var mu sync.Mutex
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				s := bufio.NewScanner(c)
				for s.Scan() {
					var q map[string]any
					if json.Unmarshal(s.Bytes(), &q) != nil {
						return
					}
					mu.Lock()
					requests = append(requests, q)
					mu.Unlock()
					var result any
					switch q["method"] {
					case "metadata":
						result = map[string]any{"driver_version": version}
					case "authorization_status":
						result = map[string]any{"permission_mode": mode, "capability_manifest_configured": true, "capability_manifest_valid": true, "capability_manifest_approved_at_startup": true, "capability_manifest": map[string]string{"sha256": hash}}
					case "call":
						result = map[string]any{"content": []any{map[string]string{"type": "text", "text": "Count: 42"}}, "structuredContent": map[string]string{"snapshot_id": "fixture-1"}}
					default:
						return
					}
					if json.NewEncoder(c).Encode(map[string]any{"ok": true, "result": result}) != nil {
						return
					}
				}
			}()
		}
	}()
	return path, &requests, &mu
}

func TestReadVerifiesScopeBeforeSendingOnlyObservation(t *testing.T) {
	s := Scope{DriverVersion: "test-version", ManifestHash: strings.Repeat("a", 64), BundleID: "com.example.Fixture", PID: 7, WindowID: 9}
	path, requests, mu := fakeDaemon(t, s.DriverVersion, "bounded", s.ManifestHash)
	c, err := Dial(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	r := c.Observe(s, Request{PID: 7, WindowID: 9, Query: "Count"})
	if r.Error != nil || r.Text != "Count: 42" {
		t.Fatalf("%+v", r)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(*requests) != 3 {
		t.Fatalf("requests: %+v", requests)
	}
	q := (*requests)[2]
	args := q["args"].(map[string]any)
	if q["name"] != "get_window_state" || args["include_screenshot"] != true || args["max_dimension"] != float64(MaxDimension) || args["max_elements"] != float64(250) || args["max_depth"] != float64(25) || args["query"] != "Count" {
		t.Fatalf("%+v", q)
	}
	if _, ok := args["screenshot_out_file"]; ok {
		t.Fatal("output path forwarded")
	}
}

func TestMismatchedDaemonCannotObserve(t *testing.T) {
	s := Scope{DriverVersion: "v1", ManifestHash: strings.Repeat("a", 64), BundleID: "com.example.Fixture", PID: 7, WindowID: 9}
	for _, tc := range []struct{ version, mode, hash string }{
		{"v2", "bounded", s.ManifestHash}, {"v1", "standard", s.ManifestHash}, {"v1", "bounded", strings.Repeat("b", 64)},
	} {
		path, requests, mu := fakeDaemon(t, tc.version, tc.mode, tc.hash)
		c, err := Dial(context.Background(), path)
		if err != nil {
			t.Fatal(err)
		}
		r := c.Observe(s, Request{PID: 7, WindowID: 9})
		c.Close()
		if r.Error == nil || r.Error.Code != "scope_mismatch" {
			t.Fatalf("%+v", r)
		}
		mu.Lock()
		for _, q := range *requests {
			if q["method"] == "call" {
				t.Errorf("unverified call: %+v", q)
			}
		}
		mu.Unlock()
	}
}

func TestDifferentWindowDoesNotContactDriver(t *testing.T) {
	s := Scope{PID: 7, WindowID: 9}
	r := (&Client{}).Observe(s, Request{PID: 7, WindowID: 10})
	if r.Error == nil || r.Error.Code != "outside_scope" {
		t.Fatalf("%+v", r)
	}
}

func TestUnresponsiveDaemonRespectsDeadline(t *testing.T) {
	path := socketPath(t, "idle.sock")
	l, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		c, err := l.Accept()
		if err == nil {
			defer c.Close()
			bufio.NewReader(c).ReadBytes('\n')
			time.Sleep(100 * time.Millisecond)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	c, err := Dial(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.Verify(Scope{DriverVersion: "v1", ManifestHash: strings.Repeat("a", 64), BundleID: "com.example.Fixture", PID: 7, WindowID: 9}); err == nil {
		t.Fatal("accepted an unresponsive driver")
	}
}
