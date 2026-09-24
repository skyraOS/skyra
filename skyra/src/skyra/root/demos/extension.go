package demos

import (
	"context"
	"fmt"
	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/boot"
	"skyra-v1/src/skyra/root/creation"
	ext "skyra-v1/src/skyra/root/extension"
	"skyra-v1/src/skyra/root/trace"
	"strconv"
)

// extension runs after the baseline demo, with fresh bootstrap authority.
func extension(ctx context.Context, args []string, t *trace.Trace) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	r, err := boot.Root(ctx, args, t)
	if err != nil {
		return err
	}
	defer r.Shutdown()
	x, err := ext.Setup(ctx, r.Kernel, r.Execution, ext.GrantPolicy{Echo: true})
	if err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- r.Kernel.Loop(ctx, append(r.Ingress, x.Inputs...)) }()
	defer func() { cancel(); <-done; x.Close() }()
	// B cannot use A's creation capability. Admission succeeds, realization rejects.
	bad := p.Move{Identity: 900, Target: creation.Capability, Invoke: &p.Invoke{Args: p.Arguments("create", "1")}}
	if err := r.Kernel.Admit(ctx, x.Inputs[1], bad); err != nil {
		return err
	}
	if err := r.Kernel.Scheduler().WaitIdle(ctx, "B"); err != nil {
		return err
	}
	t.Record("demo", "B", 900, 0, "loop survived out-of-horizon creation rejection")
	call := func(id uint64, target p.ExecutableID, operation, text string) (p.Move, error) {
		m := p.Move{Identity: id, Target: target, Invoke: &p.Invoke{Args: p.Arguments(operation, text)}}
		if err := r.Client.Egress().Send(ctx, m); err != nil {
			return p.Move{}, err
		}
		result, err := r.Results.Ingress().Receive(ctx)
		if err != nil {
			return result, err
		}
		if result.Identity != id || result.Context != "root" || result.Return == nil {
			return result, fmt.Errorf("incorrect extension result: %+v", result)
		}
		if err := r.Kernel.Scheduler().WaitIdle(ctx, "root"); err != nil {
			return result, err
		}
		return result, nil
	}
	roundtrip, err := call(10, x.Child.Activation().ID, "echo", "across contexts")
	if err != nil {
		return err
	}
	if string(roundtrip.Return.Observation.Bytes) != "across contexts" {
		return fmt.Errorf("cross-context observation changed")
	}
	t.Record("demo", "root", 10, roundtrip.Return.Answering, "cross-context round trip verified")
	created, err := call(11, creation.Capability, "create", "1")
	if err != nil {
		return err
	}
	id, err := strconv.ParseUint(string(created.Return.Observation.Bytes), 10, 64)
	if err != nil {
		return err
	}
	used, err := call(12, p.ExecutableID(id), "echo", "new address works")
	if err != nil {
		return err
	}
	if string(used.Return.Observation.Bytes) != "new address works" {
		return fmt.Errorf("created result changed")
	}
	t.Record("demo", "root", 12, used.Return.Answering, "authorized creation and new address use verified")
	return nil
}
