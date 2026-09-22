package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	p "skyra-v1/src/skyra/skyra-dna-19/protocol"
	"syscall"
	"time"
)

type codexConfig struct {
	Binary, Descriptor, Provider, Echo string
	ProviderArgs                       []string
	Grants                             map[string]bool
	Stderr                             io.Writer
}
type codexBoot struct {
	kernel          *Kernel
	execution       *ExecutionContext
	actor           *helper
	echo            *helper
	client, results *endpoint
	inputs          []*Ingress
	cancel          context.CancelFunc
	done            chan error
	connection      string
	ready           bool
	closed          bool
}

func bootCodex(ctx context.Context, c codexConfig, trace *Trace) (r *codexBoot, err error) {
	r = &codexBoot{kernel: newKernel(trace)}
	defer func() {
		if err != nil {
			r.close()
		}
	}()
	raw, err := os.ReadFile(c.Descriptor)
	if err != nil {
		return r, err
	}
	d, err := p.LoadDescriptor(raw)
	if err != nil {
		return r, err
	}
	r.kernel.runtimeState.memory.register(definition(d, nil))
	trace.record("memory", "root", 0, 0, "Codex descriptor registered")
	keyText, namespace := randomToken(), randomToken()
	key, _ := hex.DecodeString(keyText)
	cmd := exec.Command(c.Binary, append([]string{c.Provider}, c.ProviderArgs...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(os.Environ(), "SKYRA_PARENT_KEY="+keyText, "SKYRA_PARENT_NAMESPACE="+namespace)
	cmd.Stderr = c.Stderr
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return r, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return r, err
	}
	r.actor = &helper{command: cmd, endpoint: bind(&duplex{ReadCloser: stdout, writer: stdin}, "root", 20, p.Codex, 1, false)}
	r.echo, err = prepareHelper(ctx, c.Echo, "root", 21)
	if err != nil {
		return r, err
	}
	external, client := pair("root", 22, 0, true)
	sink, results := pair("root", 23, 0, false)
	r.client, r.results = client, results
	e := newExecution("root", trace, sink.egress)
	r.execution = e
	e.bookkeeper.key = key
	e.bookkeeper.namespace = namespace
	for i, item := range []struct {
		ep  *endpoint
		def Executable
	}{{r.actor.endpoint, definition(d, c.Grants)}, {r.echo.endpoint, Executable{ID: p.Echo, Contracts: []ArgContract{{Operation: "echo"}}}}, {external, Executable{}}, {sink, Executable{}}} {
		if err = e.install(uint64(i), item.ep, item.def); err != nil {
			return r, err
		}
	}
	r.kernel.runtimeState.memory.retain(e)
	if err = r.kernel.runtimeState.active.provision(ctx, e, []uint64{20, 21, 22, 23}); err != nil {
		return r, err
	}
	if err = r.echo.launch(); err != nil {
		return r, err
	}
	if err = r.actor.launch(); err != nil {
		return r, err
	}
	trace.record("launcher", "root", 0, 0, "Codex and echo started")
	r.inputs = []*Ingress{external.ingress, r.actor.endpoint.ingress, r.echo.endpoint.ingress}
	loopCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.done = make(chan error, 1)
	go func() { r.done <- r.kernel.loop(loopCtx, r.inputs) }()
	digest := sha256.Sum256(raw)
	connectCtx, stop := context.WithTimeout(ctx, 6*time.Second)
	defer stop()
	result, err := r.call(connectCtx, 1, p.Args("connect", "configuration", "default", "descriptor", hex.EncodeToString(digest[:])))
	if err != nil {
		return r, err
	}
	if result.Return == nil || result.Return.Observation.Key != "connection" || len(result.Return.Observation.Bytes) != 64 {
		return r, fmt.Errorf("connect failed: %+v", result.Fault)
	}
	r.connection = string(result.Return.Observation.Bytes)
	r.ready = true
	trace.record("root", "root", 1, result.Return.Answering, "connect validated; ready")
	return r, nil
}
func (r *codexBoot) call(ctx context.Context, id uint64, args []p.Observation) (p.Move, error) {
	if err := r.client.egress.Send(ctx, p.Move{Identity: id, Target: p.Codex, Invoke: &p.Invoke{Args: args}}); err != nil {
		return p.Move{}, err
	}
	return r.results.ingress.Receive(ctx)
}
func (r *codexBoot) close() {
	if r.closed {
		return
	}
	r.closed = true
	r.ready = false
	if r.cancel != nil {
		r.cancel()
		<-r.done
	}
	if r.execution != nil {
		r.kernel.runtimeState.active.retire(r.execution.identity)
		r.execution.invalidate()
	}
	if r.client != nil {
		r.client.Close()
	}
	if r.results != nil {
		r.results.Close()
	}
	if r.actor != nil {
		r.actor.endpoint.Close()
		if r.actor.command.Process != nil && !r.actor.waited {
			done := make(chan error, 1)
			go func() { done <- r.actor.command.Wait() }()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				_ = syscall.Kill(-r.actor.command.Process.Pid, syscall.SIGKILL)
				<-done
			}
			r.actor.waited = true
			r.kernel.trace.record("launcher", "root", 0, 0, "Codex reaped (provider supervised by Codex)")
		}
	}
	if r.echo != nil {
		_ = r.echo.close()
	}
	r.kernel.runtimeState.memory.mu.Lock()
	delete(r.kernel.runtimeState.memory.executables, p.Codex)
	delete(r.kernel.runtimeState.memory.contexts, "root")
	r.kernel.runtimeState.memory.mu.Unlock()
}
func codexDemo(ctx context.Context, c codexConfig, out io.Writer) error {
	trace := &Trace{output: out}
	r, err := bootCodex(ctx, c, trace)
	if err != nil {
		return err
	}
	defer r.close()
	profiles, err := r.execution.Capabilities(p.Codex)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		trace.record("planner", "root", 0, 0, "admitted profile: "+profile.Operation)
	}
	p1, err := r.execution.mintParent(p.Codex)
	if err != nil {
		return err
	}
	p2, err := r.execution.mintParent(p.Codex)
	if err != nil {
		return err
	}
	calls := [][]p.Observation{
		p.Args("start-thread", "connection", r.connection, "parent", p1, "text", "hello", "tools", `["echo"]`),
		p.Args("submit-input", "parent", p1, "text", "again"),
		p.Args("start-thread", "connection", r.connection, "parent", p2, "text", "other", "tools", `["echo"]`),
		p.Args("submit-input", "parent", p1, "text", "tool:child"),
		p.Args("submit-input", "parent", p2, "text", "taint"),
	}
	for i, args := range calls {
		result, err := r.call(ctx, uint64(i+2), args)
		if err != nil {
			return err
		}
		if i == 4 {
			if result.Fault == nil {
				return fmt.Errorf("contamination returned success")
			}
			trace.record("demo", "root", result.Identity, result.Source, "contamination rejected: "+result.Fault.Message)
		} else {
			if result.Return == nil {
				return fmt.Errorf("unexpected command fault: %+v", result.Fault)
			}
			trace.record("demo", "root", result.Identity, result.Source, "command result: "+string(result.Return.Observation.Bytes))
		}
	}
	return nil
}
