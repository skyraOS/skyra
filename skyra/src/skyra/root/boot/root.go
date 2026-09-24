// Package boot owns bringing a root system up and tearing it down: decoding
// startup, preparing peer processes and wires, registering executables,
// installing and publishing the root context, provisioning parent authority,
// and reaping everything on close. Once a Kernel's loop runs, boot is done.
package boot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/execution"
	"skyra-v1/src/skyra/root/kernel"
	"skyra-v1/src/skyra/root/ports"
	"skyra-v1/src/skyra/root/trace"
)

type Bootstrap struct {
	EchoPath     string   `json:"echo"`
	Observations []string `json:"observations"`
}
type RootProgram struct {
	arguments  []string
	Config     Bootstrap
	Kernel     *kernel.Kernel
	Execution  *execution.ExecutionContext
	Ingress    []*ports.Ingress
	Egress     []*ports.Egress
	Client     *ports.Endpoint
	Results    *ports.Endpoint
	Helper     *ports.Helper
	parentDone <-chan error
	trace      *trace.Trace
}

func DecodeStartup(args []string) (Bootstrap, error) {
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

// Root boots the baseline echo/parent system. The caller runs its Kernel loop.
func Root(ctx context.Context, args []string, t *trace.Trace) (_ *RootProgram, err error) {
	config, err := DecodeStartup(args)
	if err != nil {
		return nil, err
	}
	t.Record("root", "root", 0, 0, "startup decoded")
	r := &RootProgram{arguments: append([]string(nil), args...), Config: config, Kernel: kernel.New(t), trace: t}
	defer func() {
		if err != nil {
			r.Shutdown()
		}
	}()
	h, err := ports.PrepareHelper(ctx, config.EchoPath, "root", 10)
	if err != nil {
		return nil, err
	}
	r.Helper = h
	external, client := ports.Pair("root", 11, 0, true)
	rootSink, result := ports.Pair("root", 12, 0, false)
	parentEndpoint, parentDone := ports.PrepareParent("root", 13)
	r.Client, r.Results = client, result
	r.Ingress = []*ports.Ingress{external.Ingress(), h.Endpoint().Ingress(), parentEndpoint.Ingress()}
	r.Egress = []*ports.Egress{rootSink.Egress(), h.Endpoint().Egress(), parentEndpoint.Egress()}
	r.parentDone = parentDone
	r.Execution = execution.New("root", t, rootSink.Egress())
	echo := execution.Executable{ID: p.Echo, Contracts: []execution.ArgContract{{Operation: "echo"}}}
	parentDefinition := execution.Executable{ID: p.Parent, Contracts: []execution.ArgContract{{Operation: "parent"}}}
	rt := r.Kernel.State()
	rt.Memory.Register(echo)
	rt.Memory.Register(parentDefinition)
	rt.Memory.Register(r.Execution.Activation())
	for i, install := range []struct {
		endpoint   *ports.Endpoint
		definition execution.Executable
	}{{h.Endpoint(), echo}, {parentEndpoint, parentDefinition}, {external, execution.Executable{}}, {rootSink, execution.Executable{}}} {
		if err := r.Execution.Install(uint64(i), install.endpoint, install.definition); err != nil {
			return nil, err
		}
	}
	rt.Memory.Retain(r.Execution)
	if err := rt.Active.Provision(ctx, r.Execution, []uint64{10, 11, 12, 13}); err != nil {
		return nil, err
	}
	if err := h.Launch(); err != nil {
		return nil, err
	}
	t.Record("launcher", "root", 0, 0, "echo started (bootstrap/OS plumbing)")
	return r, nil
}

func (r *RootProgram) Shutdown() {
	if r.Execution != nil {
		r.Kernel.State().Active.Retire(r.Execution.Identity())
		r.Execution.Invalidate()
	}
	for _, in := range r.Ingress {
		_ = in.Close()
	}
	for _, out := range r.Egress {
		_ = out.Close()
	}
	if r.Client != nil {
		_ = r.Client.Close()
	}
	if r.Results != nil {
		_ = r.Results.Close()
	}
	if r.Helper != nil {
		err := r.Helper.Close()
		if err != nil {
			r.trace.Record("launcher", "root", 0, 0, "reap failed: "+err.Error())
		} else if r.Helper.Waited() {
			r.trace.Record("launcher", "root", 0, 0, "echo reaped")
		}
	}
	if r.parentDone != nil {
		<-r.parentDone
		r.parentDone = nil
	}
}
