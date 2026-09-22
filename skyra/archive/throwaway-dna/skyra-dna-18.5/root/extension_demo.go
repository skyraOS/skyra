package main

import (
	"context"
	"fmt"
	"skyra-v1/src/skyra/skyra-dna-18.5/creation"
	p "skyra-v1/src/skyra/skyra-dna-18.5/protocol"
	"strconv"
)

// extensionDemo runs after the baseline demo, with fresh bootstrap authority.
func extensionDemo(ctx context.Context, args []string, trace *Trace) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	r, err := boot(ctx, args, trace)
	if err != nil {
		return err
	}
	defer r.shutdown()
	x, err := setupExtension(ctx, r.kernel, r.execution, grantPolicy{echo: true})
	if err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- r.kernel.loop(ctx, append(r.ingress, x.inputs...)) }()
	defer func() { cancel(); <-done; x.close() }()
	// B cannot use A's creation capability. Admission succeeds, realization rejects.
	bad := p.Move{Identity: 900, Target: creation.Capability, Invoke: &p.Invoke{Args: p.Arguments("create", "1")}}
	if err := r.kernel.admit(ctx, x.inputs[1], bad); err != nil {
		return err
	}
	if err := r.kernel.scheduler.waitIdle(ctx, "B"); err != nil {
		return err
	}
	trace.record("demo", "B", 900, 0, "loop survived out-of-horizon creation rejection")
	call := func(id uint64, target p.ExecutableID, operation, text string) (p.Move, error) {
		m := p.Move{Identity: id, Target: target, Invoke: &p.Invoke{Args: p.Arguments(operation, text)}}
		if err := r.client.egress.Send(ctx, m); err != nil {
			return p.Move{}, err
		}
		result, err := r.results.ingress.Receive(ctx)
		if err != nil {
			return result, err
		}
		if result.Identity != id || result.Context != "root" || result.Return == nil {
			return result, fmt.Errorf("incorrect extension result: %+v", result)
		}
		if err := r.kernel.scheduler.waitIdle(ctx, "root"); err != nil {
			return result, err
		}
		return result, nil
	}
	roundtrip, err := call(10, x.child.Activation().ID, "echo", "across contexts")
	if err != nil {
		return err
	}
	if string(roundtrip.Return.Observation.Bytes) != "across contexts" {
		return fmt.Errorf("cross-context observation changed")
	}
	trace.record("demo", "root", 10, roundtrip.Return.Answering, "cross-context round trip verified")
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
	trace.record("demo", "root", 12, used.Return.Answering, "authorized creation and new address use verified")
	return nil
}
