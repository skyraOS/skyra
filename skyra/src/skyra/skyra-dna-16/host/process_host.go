package host

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"sync"
)

// ProcessSpec is supplied by trusted kernel composition. Path is executed
// directly; no shell expansion is performed. Environment is explicit rather than
// inherited accidentally.
type ProcessSpec struct {
	Path string
	Args []string
	Dir  string
	Env  []string
}

// ProcessHost owns one external process and its stdio message transport.
type ProcessHost struct {
	command   *exec.Cmd
	transport *JSONLTransport
	stdin     io.WriteCloser
	done      chan struct{}
	waitErr   error
	stderr    bytes.Buffer
	closeOnce sync.Once
	mu        sync.Mutex
}

func StartProcessHost(ctx context.Context, spec ProcessSpec) (*ProcessHost, error) {
	if ctx == nil || spec.Path == "" {
		return nil, errors.New("process host requires context and executable path")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, spec.Path, spec.Args...)
	command.Dir = spec.Dir
	command.Env = append([]string(nil), spec.Env...)
	host := &ProcessHost{command: command, done: make(chan struct{})}
	command.Stderr = &host.stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := command.Start(); err != nil {
		_ = stdin.Close()
		return nil, err
	}
	transport, err := NewJSONLTransport(stdout, stdin, stdout)
	if err != nil {
		_ = stdin.Close()
		_ = command.Process.Kill()
		_ = command.Wait()
		return nil, err
	}
	host.transport = transport
	host.stdin = stdin
	go func() {
		err := command.Wait()
		host.mu.Lock()
		host.waitErr = err
		host.mu.Unlock()
		close(host.done)
	}()
	return host, nil
}

func (host *ProcessHost) Transport() driver.Transport { return host.transport }
func (host *ProcessHost) Done() <-chan struct{}       { return host.done }

func (host *ProcessHost) WaitError() error {
	select {
	case <-host.done:
		host.mu.Lock()
		defer host.mu.Unlock()
		return host.waitErr
	default:
		return nil
	}
}

func (host *ProcessHost) Stderr() string {
	host.mu.Lock()
	defer host.mu.Unlock()
	return host.stderr.String()
}

func (host *ProcessHost) Close() error {
	var closeErr error
	host.closeOnce.Do(func() {
		closeErr = host.transport.Close()
		if err := host.stdin.Close(); closeErr == nil {
			closeErr = err
		}
		select {
		case <-host.done:
		default:
			if host.command.Process != nil {
				if err := host.command.Process.Kill(); closeErr == nil {
					closeErr = err
				}
			}
			<-host.done
		}
	})
	return closeErr
}
