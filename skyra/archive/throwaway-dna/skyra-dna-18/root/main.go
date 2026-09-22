package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	p "skyra-v1/src/skyra/skyra-dna-18/protocol"
)

func decodeStartup(args []string) (Bootstrap, error) {
	var config Bootstrap
	if len(args) != 1 {
		return config, fmt.Errorf("expected one JSON startup argument")
	}
	decoder := json.NewDecoder(strings.NewReader(args[0]))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return config, err
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return config, fmt.Errorf("trailing startup data")
	}
	if config.EchoPath == "" || len(config.Observations) != 2 {
		return config, fmt.Errorf("startup requires echo path and exactly two observations")
	}
	return config, nil
}

func boot(ctx context.Context, args []string, trace *Trace) (_ *RootProgram, err error) {
	config, err := decodeStartup(args)
	if err != nil {
		return nil, err
	}
	trace.record("root", "root", 0, 0, "startup decoded")
	r := &RootProgram{arguments: append([]string(nil), args...), boot: config, kernel: newKernel(trace), trace: trace}
	defer func() {
		if err != nil {
			r.shutdown()
		}
	}()
	h, err := prepareHelper(ctx, config.EchoPath, "root", 10)
	if err != nil {
		return nil, err
	}
	r.helper = h
	external, client := pair("root", 11, 0, true)
	rootSink, result := pair("root", 12, 0, false)
	parentEndpoint, parentDone := prepareParent("root", 13)
	r.client, r.results = client, result
	r.ingress = []*Ingress{external.ingress, h.endpoint.ingress, parentEndpoint.ingress}
	r.egress = []*Egress{rootSink.egress, h.endpoint.egress, parentEndpoint.egress}
	r.parentDone = parentDone
	r.execution = newExecution("root", trace, rootSink.egress)
	echo := Executable{ID: p.Echo, Contracts: []ArgContract{{Operation: "echo"}}}
	parentDefinition := Executable{ID: p.Parent, Contracts: []ArgContract{{Operation: "parent"}}}
	r.kernel.runtimeState.memory.register(echo)
	r.kernel.runtimeState.memory.register(parentDefinition)
	r.kernel.runtimeState.memory.register(r.execution.Activation())
	for i, install := range []struct {
		endpoint   *endpoint
		definition Executable
	}{{h.endpoint, echo}, {parentEndpoint, parentDefinition}, {external, Executable{}}, {rootSink, Executable{}}} {
		if err := r.execution.install(uint64(i), install.endpoint, install.definition); err != nil {
			return nil, err
		}
	}
	r.kernel.runtimeState.memory.retain(r.execution)
	if err := r.kernel.runtimeState.active.provision(ctx, r.execution, []uint64{10, 11, 12, 13}); err != nil {
		return nil, err
	}
	if err := h.launch(); err != nil {
		return nil, err
	}
	trace.record("launcher", "root", 0, 0, "echo started (bootstrap/OS plumbing)")
	return r, nil
}

func (r *RootProgram) shutdown() {
	if r.execution != nil {
		r.kernel.runtimeState.active.retire(r.execution.identity)
		r.execution.invalidate()
	}
	for _, in := range r.ingress {
		_ = in.endpoint.Close()
	}
	for _, out := range r.egress {
		_ = out.endpoint.Close()
	}
	if r.client != nil {
		_ = r.client.Close()
	}
	if r.results != nil {
		_ = r.results.Close()
	}
	if r.helper != nil {
		err := r.helper.close()
		if err != nil {
			r.trace.record("launcher", "root", 0, 0, "reap failed: "+err.Error())
		} else if r.helper.waited {
			r.trace.record("launcher", "root", 0, 0, "echo reaped")
		}
	}
	if r.parentDone != nil {
		<-r.parentDone
		r.parentDone = nil
	}
}

func entrypoint(ctx context.Context, args []string, output io.Writer) error {
	trace := &Trace{output: output}
	r, err := boot(ctx, args, trace)
	if err != nil {
		return err
	}
	loopCtx, cancel := context.WithCancel(ctx)
	loopDone := make(chan error, 1)
	go func() { loopDone <- r.kernel.loop(loopCtx, r.ingress) }()
	defer func() { cancel(); <-loopDone; r.shutdown() }()
	for i, text := range append(append([]string(nil), r.boot.Observations...), "child") {
		target, operation, expected := p.Echo, "echo", text
		if i == 2 {
			target, operation, expected = p.Parent, "parent", "parent:child"
		}
		request := p.Move{Identity: uint64(i + 1), Target: target, Invoke: &p.Invoke{Args: p.Arguments(operation, text)}}
		if err := r.client.egress.Send(ctx, request); err != nil {
			return err
		}
		result, err := r.results.ingress.Receive(ctx)
		if err != nil {
			return err
		}
		if result.Identity != request.Identity || result.Context != "root" || result.Return == nil || string(result.Return.Observation.Bytes) != expected {
			return fmt.Errorf("incorrect correlated result: %+v", result)
		}
		trace.record("root", "root", result.Identity, result.Return.Answering, "result verified: "+expected)
		if err := r.kernel.scheduler.waitIdle(ctx, "root"); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := entrypoint(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
