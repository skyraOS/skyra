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
//
// stack is the frame stack this instance holds: stack[0] is the frame its
// reply discharges, stack[1] the frame of the continuation waiting on it, and
// so on down to the operator's frame.
func Drive(root string, up *net.UnixConn, stack []*os.File, s machine.Instance, e machine.Event) (machine.Instance, error) {
	var (
		down  *net.UnixConn // handshake with the successor
		child *exec.Cmd
	)
	defer func() {
		if down != nil {
			down.Close()
		}
		for _, f := range stack {
			f.Close()
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
				switch {
				case x.R.Supervisor:
					trace.Logf("fault    %s: %s (committed %s)", x.R.Fault.Kind, x.R.Fault.Cause, record.Short(stamp))
				case x.R.State == record.Resolve:
					trace.Logf("gather   every slot of %s discharged: committed resolve %s", record.Short(x.R.Frame), record.Short(stamp))
				default:
					trace.Who = "inst " + record.Short(stamp)
					trace.Logf("intake   committed %s  %s", record.Short(stamp), record.Describe(x.R))
				}
			case machine.SendStamp:
				if err := handshake.Send(up, handshake.StampMsg{Stamp: x.Stamp}); err != nil {
					return s, err
				}
				var h horizon.Horizon
				files, err := handshake.Recv(up, &h)
				if err != nil {
					return s, fmt.Errorf("no Horizon, cannot execute: %w", err)
				}
				stack = files
				trace.Logf("release  Horizon received (reach %v, %d frames)", h.Names(), len(files))
				e = machine.Released{Self: h, Frames: len(files)}
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
			case machine.OpenDep:
				if err := tree.OpenDep(stack[0], x.Round, x.Child); err != nil {
					return s, err
				}
			case machine.SendHorizon:
				frames := stack
				if x.NewFrame {
					f, err := tree.OpenFrame(root, x.Child.Stamp)
					if err != nil {
						return s, err
					}
					defer f.Close()
					frames = append([]*os.File{f}, stack...)
				}
				if err := handshake.Send(down, x.Child, frames...); err != nil {
					return s, err
				}
				trace.Logf("handoff  verified and released %s, custody passed, exiting", record.Short(x.Child.Stamp))
				child.Process.Release() // D2: the parent does not wait on the child
			case machine.Settle:
				trace.Logf("settle   discharges the slot of %s: result → %s", record.Short(s.Self.Frame), record.Short(x.Result))
				if err := tree.Settle(stack[0], x.Result); err != nil {
					return s, err
				}
			case machine.GatherWaiter:
				var err error
				if e, err = gather(root, stack[1], *s.Self.Resume); err != nil {
					return s, err
				}
			case machine.TryClaim:
				won, err := tree.Claim(stack[1], x.Round, x.Stamp)
				if err != nil {
					return s, err
				}
				if won {
					trace.Logf("claim    won: resumes %s", record.Short(s.Self.Resume.Frame))
				}
				e = machine.Claimed{Won: won}
			case machine.Pop:
				stack[0].Close()
				stack = stack[1:]
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

// gather reads what a waiting continuation waits on: its frame's record and
// every round it opened, following resolve records back from the latest round
// to the frame itself. Every record is loaded by stamp and refused unless it
// hashes to it. Discharges are read from each child's frame.
func gather(root string, frame *os.File, waiter horizon.Horizon) (machine.Event, error) {
	f, err := tree.Load(root, waiter.Frame)
	if err != nil {
		return nil, err
	}
	var rounds []machine.Round
	for round := waiter.Stamp; ; {
		children, err := tree.Children(frame, round)
		if err != nil {
			return nil, err
		}
		r := machine.Round{Stamp: round}
		for _, c := range children {
			inv, err := tree.Load(root, c)
			if err != nil {
				return nil, err
			}
			child := machine.Child{Stamp: c, Invoke: inv}
			if result, err := tree.Result(root, c); err == nil {
				d, err := tree.Load(root, result)
				if err != nil {
					return nil, err
				}
				child.Result, child.Discharge = result, &d
			}
			r.Children = append(r.Children, child)
		}
		rounds = append([]machine.Round{r}, rounds...)
		if round == waiter.Frame {
			return machine.Settled{Frame: f, Rounds: rounds}, nil
		}
		prev, err := tree.Load(root, round)
		if err != nil {
			return nil, err
		}
		if prev.State != record.Resolve || prev.Frame != waiter.Frame {
			return nil, fmt.Errorf("round %s is not a resolve of frame %s", record.Short(round), record.Short(waiter.Frame))
		}
		round = prev.Round
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", "-p", x.Profile, x.Surface.Path)
	cmd.Stdin = bytes.NewReader(x.Stdin)
	cmd.Env = x.Env
	var stderr bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
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
	if len(emission) == 0 {
		// What the target said on the way out is the observation.
		if lines := bytes.Split(bytes.TrimSpace(stderr.Bytes()), []byte("\n")); len(lines[len(lines)-1]) > 0 {
			return machine.Emitted{Err: "target exited without emitting: " + string(lines[len(lines)-1])}
		}
	}
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
	if err := handshake.Send(down, x.Under); err != nil {
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
