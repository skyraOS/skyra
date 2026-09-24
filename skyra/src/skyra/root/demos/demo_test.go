package demos_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/boot"
	"skyra-v1/src/skyra/root/demos"
	kt "skyra-v1/src/skyra/root/kernel/kerneltest"
	"skyra-v1/src/skyra/root/trace"
)

var (
	guard    = kt.Guard
	must     = kt.Must
	rejected = kt.Rejected
	await    = kt.Await
)

func TestProcessDemoAndLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "echo")
	build := exec.CommandContext(guard(t), "go", "build", "-o", path, "../echo")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build echo: %v\n%s", err, output)
	}
	config, _ := json.Marshal(boot.Bootstrap{EchoPath: path, Observations: []string{"hello", "again"}})
	t.Run("boot persistent echo and parent", func(t *testing.T) {
		var output bytes.Buffer
		must(t, demos.Entrypoint(guard(t), []string{string(config)}, &output))
		decoder := json.NewDecoder(&output)
		var events []trace.Event
		for decoder.More() {
			var event trace.Event
			must(t, decoder.Decode(&event))
			events = append(events, event)
		}
		launches, results, resumes, reaped := 0, 0, 0, 0
		published := false
		committed := map[p.FrameID]bool{}
		settled := map[p.FrameID]bool{}
		returned := map[p.FrameID]bool{}
		for i, e := range events {
			if e.Number != uint64(i+1) {
				t.Fatal("nonmonotonic trace")
			}
			if strings.HasPrefix(e.Transition, "context published") {
				published = true
			}
			if strings.HasPrefix(e.Transition, "echo started") {
				launches++
				if !published {
					t.Fatal("launch before publication")
				}
			}
			if e.Transition == "invocation admitted" && !published {
				t.Fatal("admission before publication")
			}
			if e.Transition == "transition committed" {
				committed[e.Frame] = true
			}
			if e.Transition == "return admitted" {
				returned[e.Frame] = true
			}
			if e.Transition == "frame settled" {
				settled[e.Frame] = true
				if !returned[e.Frame] {
					t.Fatal("return bypassed admission")
				}
			}
			if e.Transition == "invocation sent" && !committed[e.Frame] {
				t.Fatal("send preceded commit")
			}
			if e.Transition == "root result sent" {
				results++
				if !committed[e.Frame] || !settled[e.Frame] {
					t.Fatal("result preceded committed settlement")
				}
			}
			if e.Transition == "parent resumed" {
				resumes++
			}
			if e.Transition == "echo reaped" {
				reaped++
			}
		}
		if launches != 1 || results != 3 || resumes != 1 || reaped != 1 {
			t.Fatalf("launch/results/resume/reap = %d/%d/%d/%d", launches, results, resumes, reaped)
		}
	})
	t.Run("complete extended demo", func(t *testing.T) {
		var output bytes.Buffer
		must(t, demos.Run(guard(t), []string{string(config)}, &output))
		decoder := json.NewDecoder(&output)
		found := map[string]bool{}
		var last uint64
		reaped := 0
		for decoder.More() {
			var event trace.Event
			must(t, decoder.Decode(&event))
			if event.Number != last+1 {
				t.Fatal("demo trace numbering reset")
			}
			last = event.Number
			found[event.Transition] = true
			if event.Transition == "echo reaped" {
				reaped++
			}
		}
		for _, transition := range []string{"loop survived out-of-horizon creation rejection", "cross-context round trip verified", "authorized creation and new address use verified"} {
			if !found[transition] {
				t.Fatal("missing demo evidence:", transition)
			}
		}
		if reaped != 2 {
			t.Fatal("demo helper processes not reaped")
		}
	})
	t.Run("lifecycle text stays child data and process is reaped", func(t *testing.T) {
		ctx := guard(t)
		r, err := boot.Root(ctx, []string{string(config)}, &trace.Trace{})
		must(t, err)
		defer func() {
			r.Shutdown()
			if r.Helper.Unreaped() {
				t.Error("helper not reaped")
			}
		}()
		loopCtx, cancel := context.WithCancel(ctx)
		done := make(chan error, 1)
		go func() { done <- r.Kernel.Loop(loopCtx, r.Ingress) }()
		defer func() {
			cancel()
			_ = await(t, ctx, done)
		}()
		for i, text := range []string{"save", "close"} {
			m := p.Move{Identity: uint64(i + 1), Target: p.Parent, Invoke: &p.Invoke{Args: p.Arguments("parent", text)}}
			must(t, r.Client.Egress().Send(ctx, m))
			result, err := r.Results.Ingress().Receive(ctx)
			must(t, err)
			if result.Return == nil || string(result.Return.Observation.Bytes) != "parent:"+text {
				t.Fatal("child data changed lifecycle")
			}
			must(t, r.Kernel.Scheduler().WaitIdle(ctx, "root"))
			if _, ok := r.Kernel.State().Active.Lookup("root"); !ok {
				t.Fatal("child data retired context")
			}
		}
	})
	t.Run("boot launch failure unwinds", func(t *testing.T) {
		bad, _ := json.Marshal(boot.Bootstrap{EchoPath: filepath.Join(t.TempDir(), "missing"), Observations: []string{"a", "b"}})
		tr := &trace.Trace{}
		_, err := boot.Root(guard(t), []string{string(bad)}, tr)
		rejected(t, err)
		retired := false
		for _, e := range tr.Snapshot() {
			if e.Transition == "context retired and reservations released" {
				retired = true
			}
		}
		if !retired {
			t.Fatal("boot failed without rollback")
		}
	})
}
