package extension_test

import (
	"context"
	"strconv"
	"testing"

	p "skyra/history/dna19/protocol"
	"skyra/history/dna19/root/creation"
	"skyra/history/dna19/root/execution"
	"skyra/history/dna19/root/extension"
	kt "skyra/history/dna19/root/kernel/kerneltest"
	"skyra/history/dna19/root/ports"
)

var (
	must       = kt.Must
	rejected   = kt.Rejected
	newFixture = kt.New
	traceCount = kt.TraceCount
	liveLoop   = kt.LiveLoop
)

func TestCrossContextAndCreation(t *testing.T) {
	a := newFixture(t, nil, "A", 10, false)
	ctx, cancel := context.WithCancel(a.Ctx)
	defer cancel()
	x, err := extension.Setup(ctx, a.Kernel, a.Execution, extension.GrantPolicy{Echo: true})
	must(t, err)
	defer x.Close()
	stop := liveLoop(t, ctx, a.Kernel, append([]*ports.Ingress{a.External.Ingress()}, x.Inputs...))
	defer stop()
	call := func(id uint64, target p.ExecutableID, op, text string) p.Move {
		t.Helper()
		must(t, a.Client.Egress().Send(ctx, p.Move{Identity: id, Target: target, Invoke: &p.Invoke{Args: p.Arguments(op, text)}}))
		result, err := a.Result.Ingress().Receive(ctx)
		must(t, err)
		if result.Context != "A" || result.Identity != id {
			t.Fatal("lost caller correlation")
		}
		must(t, a.Kernel.Scheduler().WaitIdle(ctx, "A"))
		return result
	}
	// B cannot invoke A's capability, even when it knows the numeric address.
	bad := p.Move{Identity: 900, Target: creation.Capability, Invoke: &p.Invoke{Args: p.Arguments("create", "1")}}
	must(t, a.Kernel.Admit(ctx, x.Inputs[1], bad))
	must(t, a.Kernel.Scheduler().WaitIdle(ctx, "B"))
	if traceCount(a.Kernel, "B", "dispatch rejected:") != 1 {
		t.Fatal("B authority violation accepted")
	}
	before := a.Kernel.State().Memory.Len()
	if x.Minted() {
		t.Fatal("unauthorized creation minted")
	}
	for _, id := range []uint64{21, 22} {
		result := call(id, x.Child.Activation().ID, "echo", "round trip")
		if result.Return == nil || string(result.Return.Observation.Bytes) != "round trip" {
			t.Fatal("wrong cross-context result")
		}
	}
	if traceCount(a.Kernel, "B", "root sink resolved") != 2 || traceCount(a.Kernel, "A", "root result sent") != 2 {
		t.Fatal("result bypassed own sink")
	}
	if a.Execution.Mappings()[0].Actual == x.Child.Mappings()[0].Actual {
		t.Fatal("virtual port spaces alias")
	}
	for _, m := range x.Child.Mappings() {
		if m.Executable.ID != 0 && m.Executable.ID != p.Echo {
			t.Fatal("B received excess delegation")
		}
	}
	// The invocation has capability access but requests a non-grantable template.
	denied := call(23, creation.Capability, "create", "2")
	if denied.Fault == nil || x.Minted() {
		t.Fatal("creation exceeded grant policy")
	}
	result := call(24, creation.Capability, "create", "1")
	if result.Return == nil || result.Return.Observation.Key != "executable" {
		t.Fatal("missing new address")
	}
	n, err := strconv.ParseUint(string(result.Return.Observation.Bytes), 10, 64)
	must(t, err)
	if n < 1000 || a.Kernel.State().Memory.Len() != before+1 {
		t.Fatal("address was not freshly minted")
	}
	used := call(25, p.ExecutableID(n), "echo", "minted")
	if used.Return == nil || string(used.Return.Observation.Bytes) != "minted" {
		t.Fatal("new executable unusable")
	}
	exhausted := call(26, creation.Capability, "create", "1")
	if exhausted.Fault == nil {
		t.Fatal("one-address creation budget exceeded")
	}
	must(t, a.Execution.Inspect(func(h execution.Horizon) error {
		minted, err := h.Resolve(p.ExecutableID(n))
		must(t, err)
		if len(minted.Contracts) != 1 || minted.Contracts[0].Operation != "echo" {
			t.Fatal("created authority widened")
		}
		return nil
	}))
}

func TestDelegationDeniedWithoutGrant(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	_, err := extension.Setup(f.Ctx, f.Kernel, f.Execution, extension.GrantPolicy{})
	rejected(t, err)
	if _, ok := f.Kernel.State().Active.Lookup("B"); ok {
		t.Fatal("ungrantable child published")
	}
	if len(f.Execution.Mappings()) != 3 {
		t.Fatal("denied delegation changed creator horizon")
	}
}
