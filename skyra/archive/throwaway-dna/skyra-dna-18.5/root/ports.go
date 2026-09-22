package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os/exec"

	"skyra-v1/src/skyra/skyra-dna-18.5/parent"
	p "skyra-v1/src/skyra/skyra-dna-18.5/protocol"
)

func bind(connection io.ReadWriteCloser, owner p.ContextID, actual uint64, target p.ExecutableID, generation uint64, external bool) *endpoint {
	b := &binding{owner: owner, actual: actual, target: target, generation: generation, external: external}
	b.live.Store(true)
	e := &endpoint{connection: connection, binding: b}
	e.ingress = &Ingress{endpoint: e, decoder: json.NewDecoder(connection)}
	e.ingress.decoder.DisallowUnknownFields()
	e.egress = &Egress{endpoint: e, encoder: json.NewEncoder(connection)}
	return e
}

func (e *endpoint) Close() error {
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
			err = &moveRejection{validation}
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

// pair and prepareParent are trusted connection plumbing. No handles escape into
// RootProgram, RuntimeState, or executable protocol inputs.
func pair(owner p.ContextID, actual uint64, target p.ExecutableID, external bool) (*endpoint, *endpoint) {
	a, b := net.Pipe()
	return bind(a, owner, actual, target, 1, external), bind(b, owner, 0, 0, 1, false)
}
func prepareParent(owner p.ContextID, actual uint64) (*endpoint, <-chan error) {
	a, b := net.Pipe()
	done := make(chan error, 1)
	go func() { done <- parent.Serve(b) }()
	return bind(a, owner, actual, p.Parent, 1, false), done
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

// prepareHelper creates host plumbing only; launch occurs after Active publication.
func prepareHelper(ctx context.Context, path string, owner p.ContextID, actual uint64) (*helper, error) {
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
	return &helper{command: cmd, endpoint: bind(&duplex{ReadCloser: stdout, writer: stdin}, owner, actual, p.Echo, 1, false)}, nil
}

func (h *helper) launch() error { return h.command.Start() }
func (h *helper) close() error {
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
