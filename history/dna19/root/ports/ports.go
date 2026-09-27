// Package ports owns the host's wires: authenticated bindings, the framed Move
// endpoints built on them, and the helper processes whose stdio they carry. It
// knows nothing about contexts, frames or scheduling.
package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	p "skyra/history/dna19/protocol"
	"skyra/history/dna19/root/parent"
	"skyra/history/dna19/root/trace"
)

// Each replacement gets a new, never-reused binding object and generation.
// Its pointer is trusted host provenance, never decoded from the wire.
type Binding struct {
	owner      p.ContextID
	actual     uint64
	target     p.ExecutableID
	generation uint64
	external   bool
	live       atomic.Bool
}

func (b *Binding) Owner() p.ContextID     { return b.owner }
func (b *Binding) Actual() uint64         { return b.actual }
func (b *Binding) Target() p.ExecutableID { return b.target }
func (b *Binding) Generation() uint64     { return b.generation }
func (b *Binding) External() bool         { return b.external }
func (b *Binding) Live() bool             { return b.live.Load() }

type Endpoint struct {
	ingress    *Ingress
	egress     *Egress
	connection io.ReadWriteCloser
	binding    *Binding
	once       sync.Once
}

func (e *Endpoint) Ingress() *Ingress        { return e.ingress }
func (e *Endpoint) Egress() *Egress          { return e.egress }
func (e *Endpoint) Binding() *Binding        { return e.binding }
func (e *Endpoint) Conn() io.ReadWriteCloser { return e.connection }

type Ingress struct {
	endpoint *Endpoint
	mu       sync.Mutex
	decoder  *json.Decoder
}
type Egress struct {
	endpoint *Endpoint
	mu       sync.Mutex
	encoder  *json.Encoder
}

func (in *Ingress) Binding() *Binding { return in.endpoint.binding }
func (in *Ingress) Close() error      { return in.endpoint.Close() }
func (out *Egress) Close() error      { return out.endpoint.Close() }

// Rejection marks a Move refused at a boundary. It is contained to that Move;
// any other error is loss of the transport or of a runtime invariant.
type Rejection struct{ error }

func Reject(err error) error { return &Rejection{err} }

func Bind(connection io.ReadWriteCloser, owner p.ContextID, actual uint64, target p.ExecutableID, generation uint64, external bool) *Endpoint {
	b := &Binding{owner: owner, actual: actual, target: target, generation: generation, external: external}
	b.live.Store(true)
	e := &Endpoint{connection: connection, binding: b}
	e.ingress = &Ingress{endpoint: e, decoder: json.NewDecoder(connection)}
	e.ingress.decoder.DisallowUnknownFields()
	e.egress = &Egress{endpoint: e, encoder: json.NewEncoder(connection)}
	return e
}

func (e *Endpoint) Close() error {
	var err error
	e.once.Do(func() { e.binding.live.Store(false); err = e.connection.Close() })
	return err
}

func (in *Ingress) Receive(ctx context.Context) (p.Move, error) {
	in.mu.Lock()
	defer in.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return p.Move{}, err
	}
	if !in.endpoint.binding.live.Load() {
		return p.Move{}, fmt.Errorf("closed binding")
	}
	stop := context.AfterFunc(ctx, func() { _ = in.endpoint.Close() })
	defer stop()
	var m p.Move
	err := in.decoder.Decode(&m)
	if ctx.Err() != nil {
		return p.Move{}, ctx.Err()
	}
	if err == nil {
		if validation := m.Validate(); validation != nil {
			err = &Rejection{validation}
		}
	}
	return m, err
}

func (out *Egress) Send(ctx context.Context, m p.Move) error {
	out.mu.Lock()
	defer out.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	b := out.endpoint.binding
	if !b.live.Load() {
		return fmt.Errorf("closed binding")
	}
	if b.target != 0 && (m.Target != 0 || m.Invoke != nil) && b.target != m.Target {
		return fmt.Errorf("fixed destination mismatch")
	}
	if err := m.Validate(); err != nil {
		return err
	}
	stop := context.AfterFunc(ctx, func() { _ = out.endpoint.Close() })
	defer stop()
	err := out.encoder.Encode(m)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

// Pair and PrepareParent are trusted connection plumbing. No handles escape into
// RootProgram, runtime state, or executable protocol inputs.
func Pair(owner p.ContextID, actual uint64, target p.ExecutableID, external bool) (*Endpoint, *Endpoint) {
	a, b := net.Pipe()
	return Bind(a, owner, actual, target, 1, external), Bind(b, owner, 0, 0, 1, false)
}
func PrepareParent(owner p.ContextID, actual uint64) (*Endpoint, <-chan error) {
	a, b := net.Pipe()
	done := make(chan error, 1)
	go func() { done <- parent.Serve(b) }()
	return Bind(a, owner, actual, p.Parent, 1, false), done
}

type duplex struct {
	io.ReadCloser
	writer io.WriteCloser
}

func (d *duplex) Write(b []byte) (int, error) { return d.writer.Write(b) }
func (d *duplex) Close() error {
	a := d.writer.Close()
	b := d.ReadCloser.Close()
	if a != nil {
		return a
	}
	return b
}

type Helper struct {
	command  *exec.Cmd
	endpoint *Endpoint
	waited   bool
}

func (h *Helper) Endpoint() *Endpoint { return h.endpoint }

// Waited reports whether the process has been waited on.
func (h *Helper) Waited() bool { return h.waited }

// Unreaped reports a started process that has not been waited to exit.
func (h *Helper) Unreaped() bool {
	return h.command.Process != nil && (!h.waited || h.command.ProcessState == nil)
}

// PrepareHelper creates host plumbing only; launch occurs after Active publication.
func PrepareHelper(ctx context.Context, path string, owner p.ContextID, actual uint64) (*Helper, error) {
	cmd := exec.CommandContext(ctx, path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}
	return &Helper{command: cmd, endpoint: Bind(&duplex{ReadCloser: stdout, writer: stdin}, owner, actual, p.Echo, 1, false)}, nil
}

// PreparePeer prepares an external executable in its own process group, so a
// reap can also stop whatever it supervises. Launch occurs after publication.
func PreparePeer(bin string, args []string, env []string, stderr io.Writer, owner p.ContextID, port uint64, target p.ExecutableID) (*Helper, error) {
	cmd := exec.Command(bin, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(os.Environ(), env...)
	cmd.Stderr = stderr
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}
	return &Helper{command: cmd, endpoint: Bind(&duplex{ReadCloser: stdout, writer: stdin}, owner, port, target, 1, false)}, nil
}

func (h *Helper) Launch() error { return h.command.Start() }
func (h *Helper) Close() error {
	_ = h.endpoint.Close()
	if h.command.Process == nil || h.waited {
		return nil
	}
	// Shutdown does not depend on a cooperative foreign executable reading EOF.
	_ = h.command.Process.Kill()
	err := h.command.Wait()
	h.waited = true
	if _, killed := err.(*exec.ExitError); killed {
		return nil
	}
	return err
}

// Reap closes a peer's endpoint and gives it two seconds to exit on EOF before
// killing its process group. transition is recorded once the peer is reaped.
func (h *Helper) Reap(t *trace.Trace, transition string) {
	if h == nil {
		return
	}
	_ = h.endpoint.Close()
	if h.command.Process == nil || h.waited {
		return
	}
	done := make(chan error, 1)
	go func() { done <- h.command.Wait() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = syscall.Kill(-h.command.Process.Pid, syscall.SIGKILL)
		<-done
	}
	h.waited = true
	t.Record("launcher", "root", 0, 0, transition)
}
