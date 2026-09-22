package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func launchProvider(path string, args []string) (*provider, error) {
	cmd := exec.Command(path, args...)
	// Parent signing material never reaches the provider.
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "SKYRA_PARENT_") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		in.Close()
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		in.Close()
		out.Close()
		return nil, err
	}
	p := &provider{command: cmd, send: make(chan any, 64), events: make(chan rpc, 64), errors: make(chan error, 2)}
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		d := json.NewDecoder(out)
		for {
			var m rpc
			if err := d.Decode(&m); err != nil {
				select {
				case p.errors <- err:
				case <-done:
				}
				return
			}
			select {
			case p.events <- m:
			case <-done:
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		e := json.NewEncoder(in)
		for {
			select {
			case <-done:
				return
			case m := <-p.send:
				if err := e.Encode(m); err != nil {
					select {
					case p.errors <- err:
					case <-done:
					}
					return
				}
			}
		}
	}()
	var once sync.Once
	p.close = func() {
		once.Do(func() {
			close(done)
			in.Close()
			out.Close()
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			wg.Wait()
			fmtReaped()
		})
	}
	return p, nil
}
func fmtReaped() { _, _ = io.WriteString(os.Stderr, "provider reaped\n") }
func (a *actor) send(value any) error {
	select {
	case a.provider.send <- value:
		return nil
	default:
		return io.ErrShortWrite
	}
}
func (a *actor) request(kind, method string, params any, r request) error {
	a.sequence++
	id := a.sequence
	r.kind = kind
	if r.deadline.IsZero() {
		r.deadline = time.Now().Add(5 * time.Second)
	}
	key := jsonID(id)
	a.requests[key] = r
	if err := a.send(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		delete(a.requests, key)
		return err
	}
	return nil
}
func jsonID(n uint64) string { raw, _ := json.Marshal(n); return string(raw) }
