package kernel_test

import (
	"strings"
	"testing"

	p "skyra-v1/src/skyra/protocol"
)

func expectFault(t *testing.T, m p.Move, want string) {
	t.Helper()
	if m.Fault == nil || !strings.Contains(m.Fault.Message, want) {
		t.Fatalf("want fault %q: %+v", want, m)
	}
}
func TestOnePendingChildAndCorrelation(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	parent := f.Start(t, invoke(1, "parent"))
	childRequest := invoke(2, "child")
	childRequest.Source = parent.Invoke.Frame
	childRequest.State = []byte("continuation")
	must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), childRequest))
	must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), childRequest))
	done := runAsync(f.Ctx, f.Kernel)
	child, err := f.Peer.Ingress().Receive(f.Ctx)
	must(t, err)
	must(t, await(t, f.Ctx, done))
	rejected(t, f.Kernel.Scheduler().Run(f.Ctx, f.Kernel.Dispatch))
	if child.Invoke.ReturnTo != parent.Invoke.Frame || child.Invoke.Frame == parent.Invoke.Frame {
		t.Fatal("missing cactus fork")
	}
	forged := returned(child, "wrong")
	forged.Return.Answering = parent.Invoke.Frame
	rejected(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), forged))
	must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), returned(child, "child")))
	done = runAsync(f.Ctx, f.Kernel)
	resume, err := f.Peer.Ingress().Receive(f.Ctx)
	must(t, err)
	must(t, await(t, f.Ctx, done))
	if resume.Source != parent.Invoke.Frame || resume.Return.Answering != child.Invoke.Frame || string(resume.State) != "continuation" {
		t.Fatal("incorrect correlated resume")
	}
	rejected(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), returned(child, "duplicate")))
	result := f.Finish(t, returned(parent, "finished"))
	if result.Identity != 1 || string(result.Return.Observation.Bytes) != "finished" {
		t.Fatal("wrong root result")
	}
}
func TestWaitingFaultInvalidatesDescendants(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	parent := f.Start(t, invoke(1, "parent"))
	request := invoke(2, "child")
	request.Source = parent.Invoke.Frame
	request.State = []byte("saved")
	must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), request))
	done := runAsync(f.Ctx, f.Kernel)
	child, err := f.Peer.Ingress().Receive(f.Ctx)
	must(t, err)
	must(t, await(t, f.Ctx, done))
	fault := p.Move{Identity: 3, Context: "A", Source: parent.Invoke.Frame, Fault: &p.Fault{Frame: parent.Invoke.Frame, Message: "contamination"}}
	must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), fault))
	done = runAsync(f.Ctx, f.Kernel)
	result, err := f.Result.Ingress().Receive(f.Ctx)
	must(t, err)
	must(t, await(t, f.Ctx, done))
	expectFault(t, result, "contamination")
	rejected(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), returned(child, "late")))
	rejected(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), returned(parent, "clean")))
}
