// Package instance is the shell around the machine: it feeds events to
// machine.Step, performs the effects Step returns, and turns whatever the
// world does next into the next event. It owns processes (the target and the
// successor) and makes no decisions.
package instance

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"skyra/controller/handshake"
	"skyra/controller/horizon"
	"skyra/controller/machine"
	"skyra/controller/record"
	"skyra/controller/trace"
	"skyra/controller/tree"
)

// Run is one run of skyra, started by a parent that holds fd 3.
func Run(root string) error {
	up, err := handshake.Conn(os.NewFile(3, "handshake"))
	if err != nil {
		return fmt.Errorf("an instance must be started by a parent: %w", err)
	}
	defer up.Close()
	var under horizon.Horizon // the Horizon the bytes were emitted under
	if _, err := handshake.Recv(up, &under); err != nil {
		return err
	}
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	_, err = Drive(root, up, nil, machine.Instance{Floor: root, Phase: machine.Intake}, machine.Received{Under: under, In: in})
	return err
}

// Drive is the loop: feed an event to machine.Step, perform the effects it
// returns, and turn the result into the next event. Every decision is made in
// Step. Drive only does I/O.
func Drive(root string, up *net.UnixConn, node *os.File, s machine.Instance, e machine.Event) (machine.Instance, error) {
	var (
		down  *net.UnixConn // handshake with the successor
		child *exec.Cmd
	)
	defer func() {
		if down != nil {
			down.Close()
		}
	}()
	for e != nil {
		var effs []machine.Effect
		s, effs = machine.Step(s, e)
		e = nil
		for _, eff := range effs {
			switch x := eff.(type) {
			case machine.StoreInput:
				if _, err := tree.Put(root, "inputs", x.In); err != nil {
					return s, err
				}
			case machine.Commit:
				stamp, err := tree.Commit(root, x.R)
				if err != nil {
					return s, err
				}
				if x.R.Supervisor {
					trace.Logf("fault    %s (committed %s)", x.R.Reason, record.Short(stamp))
				} else {
					trace.Who = "inst " + record.Short(stamp)
					trace.Logf("intake   committed %s  %s", record.Short(stamp), record.Describe(x.R))
				}
			case machine.SendStamp:
				if err := handshake.Send(up, handshake.StampMsg{Stamp: x.Stamp}, nil); err != nil {
					return s, err
				}
				var h horizon.Horizon
				n, err := handshake.Recv(up, &h)
				if err != nil {
					return s, fmt.Errorf("no Horizon, cannot execute: %w", err)
				}
				if node = n; node != nil {
					defer node.Close()
				}
				trace.Logf("release  Horizon received (reach %v)", h.Names())
				e = machine.Released{Self: h, Node: node != nil}
			case machine.Run:
				trace.Logf("execute  %s under rendered sandbox", filepath.Base(x.Surface.Path))
				e = run(x)
			case machine.Spawn:
				var err error
				if down, child, e, err = spawn(root, x); err != nil {
					return s, err
				}
			case machine.KillSuccessor:
				trace.Logf("handoff  killing successor")
				child.Process.Kill()
			case machine.SendHorizon:
				childNode := node
				if x.NewFrame {
					f, err := tree.OpenFrame(root, x.Child.Stamp)
					if err != nil {
						return s, err
					}
					defer f.Close()
					childNode = f
				}
				if err := handshake.Send(down, x.Child, childNode); err != nil {
					return s, err
				}
				trace.Logf("handoff  verified and released %s, custody passed, exiting", record.Short(x.Child.Stamp))
				child.Process.Release() // D2: the parent does not wait on the child
			case machine.Settle:
				trace.Logf("release  settles the parent frame: result → %s", record.Short(x.Result))
				if err := tree.Settle(node, x.Result); err != nil {
					return s, err
				}
			case machine.Exit:
				if x.Err != "" {
					return s, errors.New(x.Err)
				}
				return s, nil
			}
		}
	}
	return s, errors.New("step stopped without exiting")
}

// run performs a Run effect: check the binary's content hash, run it under
// the rendered profile, and read its one line of output.
func run(x machine.Run) machine.Event {
	bin, err := os.ReadFile(x.Surface.Path)
	if err != nil {
		return machine.Emitted{Err: err.Error()}
	}
	if record.Sum(bin) != x.Surface.SHA256 {
		return machine.Emitted{Err: "surface binary does not match its registered hash"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", "-p", x.Profile, x.Surface.Path)
	cmd.Stdin = bytes.NewReader(x.Stdin)
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		return machine.Emitted{Err: err.Error()}
	}
	if err := cmd.Start(); err != nil {
		return machine.Emitted{Err: err.Error()}
	}
	line, _ := bufio.NewReader(out).ReadBytes('\n')
	_ = cmd.Wait()
	emission := bytes.TrimSuffix(line, []byte("\n"))
	trace.Logf("execute  target emitted %d bytes (sha256 %s)", len(emission), record.Short(record.Sum(emission)))
	return machine.Emitted{Out: emission}
}

// spawn performs a Spawn effect: start skyra again with the bytes on stdin,
// send it the Horizon they were emitted under, and wait for its stamp.
func spawn(root string, x machine.Spawn) (*net.UnixConn, *exec.Cmd, machine.Event, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, nil, nil, err
	}
	down, childEnd, err := handshake.Pair()
	if err != nil {
		return nil, nil, nil, err
	}
	cmd := exec.Command(self, "instance", root)
	cmd.Stdin = bytes.NewReader(x.Out)
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{childEnd} // becomes the child's fd 3
	if err := cmd.Start(); err != nil {
		return nil, nil, nil, err
	}
	childEnd.Close()
	if err := handshake.Send(down, x.Under, nil); err != nil {
		return nil, nil, nil, err
	}
	var st handshake.StampMsg
	if _, err := handshake.Recv(down, &st); err != nil {
		_ = cmd.Wait()
		return down, cmd, machine.Stamped{Err: err.Error()}, nil
	}
	if r, err := tree.Load(root, st.Stamp); err == nil { // Load refuses a record that does not hash to its stamp
		return down, cmd, machine.Stamped{Stamp: st.Stamp, R: &r}, nil
	}
	return down, cmd, machine.Stamped{Stamp: st.Stamp}, nil
}
