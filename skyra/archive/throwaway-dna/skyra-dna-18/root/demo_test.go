package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	p "skyra-v1/src/skyra/skyra-dna-18/protocol"
)

func TestProcessDemoAndLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "echo")
	build := exec.CommandContext(guard(t), "go", "build", "-o", path, "../echo")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build echo: %v\n%s", err, output)
	}
	config, _ := json.Marshal(Bootstrap{EchoPath: path, Observations: []string{"hello", "again"}})
	t.Run("boot persistent echo and parent", func(t *testing.T) {
		var output bytes.Buffer
		must(t, entrypoint(guard(t), []string{string(config)}, &output))
		decoder := json.NewDecoder(&output)
		var events []Event
		for decoder.More() {
			var event Event
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
	t.Run("lifecycle text stays child data and process is reaped", func(t *testing.T) {
		ctx := guard(t)
		trace := &Trace{}
		r, err := boot(ctx, []string{string(config)}, trace)
		must(t, err)
		defer func() {
			r.shutdown()
			if r.helper.command.ProcessState == nil || !r.helper.waited {
				t.Error("helper not reaped")
			}
		}()
		loopCtx, cancel := context.WithCancel(ctx)
		done := make(chan error, 1)
		go func() { done <- r.kernel.loop(loopCtx, r.ingress) }()
		defer func() {
			cancel()
			_ = await(t, ctx, done)
		}()
		for i, text := range []string{"save", "close"} {
			m := p.Move{Identity: uint64(i + 1), Target: p.Parent, Invoke: &p.Invoke{Args: p.Arguments("parent", text)}}
			must(t, r.client.egress.Send(ctx, m))
			result, err := r.results.ingress.Receive(ctx)
			must(t, err)
			if result.Return == nil || string(result.Return.Observation.Bytes) != "parent:"+text {
				t.Fatal("child data changed lifecycle")
			}
			must(t, r.kernel.scheduler.waitIdle(ctx, "root"))
			if _, ok := r.kernel.runtimeState.active.lookup("root"); !ok {
				t.Fatal("child data retired context")
			}
		}
	})
	t.Run("boot launch failure unwinds", func(t *testing.T) {
		bad, _ := json.Marshal(Bootstrap{EchoPath: filepath.Join(t.TempDir(), "missing"), Observations: []string{"a", "b"}})
		trace := &Trace{}
		_, err := boot(guard(t), []string{string(bad)}, trace)
		rejected(t, err)
		retired := false
		for _, e := range trace.snapshot() {
			if e.Transition == "context retired and reservations released" {
				retired = true
			}
		}
		if !retired {
			t.Fatal("boot failed without rollback")
		}
	})
}

func TestOnePendingChildAndCorrelation(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	parent := f.start(t, invoke(1, "parent"))
	childRequest := invoke(2, "child")
	childRequest.Source = parent.Invoke.Frame
	childRequest.State = []byte("continuation")
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, childRequest))
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, childRequest))
	done := runAsync(f.ctx, f.kernel)
	child, err := f.peer.ingress.Receive(f.ctx)
	must(t, err)
	must(t, await(t, f.ctx, done))
	rejected(t, f.kernel.scheduler.Run(f.ctx, f.kernel.dispatch))
	if child.Invoke.ReturnTo != parent.Invoke.Frame || child.Invoke.Frame == parent.Invoke.Frame {
		t.Fatal("missing cactus fork")
	}
	forged := returned(child, "wrong")
	forged.Return.Answering = parent.Invoke.Frame
	rejected(t, f.kernel.admit(f.ctx, f.echo.ingress, forged))
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, returned(child, "child")))
	done = runAsync(f.ctx, f.kernel)
	resume, err := f.peer.ingress.Receive(f.ctx)
	must(t, err)
	must(t, await(t, f.ctx, done))
	if resume.Source != parent.Invoke.Frame || resume.Return.Answering != child.Invoke.Frame || string(resume.State) != "continuation" {
		t.Fatal("incorrect correlated resume")
	}
	rejected(t, f.kernel.admit(f.ctx, f.echo.ingress, returned(child, "duplicate")))
	result := f.finish(t, returned(parent, "finished"))
	if result.Identity != 1 || string(result.Return.Observation.Bytes) != "finished" {
		t.Fatal("wrong root result")
	}
}
