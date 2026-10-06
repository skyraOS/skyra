// Package instance is the shell around the machine: it feeds events to
// machine.Step, performs the effects Step returns, and turns whatever the
// world does next into the next event. It owns processes (the target and the
// successor) and makes no decisions.
package instance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/custody"
	"skyra/controller/runtime/handshake"
	"skyra/controller/runtime/park"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/trace"
	"skyra/controller/runtime/tree"
)

// Run is one run of skyra, started by a parent that holds fd 3.
func Run(root string) error {
	up, err := handshake.Conn(os.NewFile(3, "handshake"))
	if err != nil {
		return fmt.Errorf("an instance must be started by a parent: %w", err)
	}
	defer up.Close()
	var named handshake.Under // the Horizon the bytes were emitted under, by id
	if _, err := handshake.Recv(up, &named); err != nil {
		return err
	}
	under, err := tree.LoadHorizon(root, named.Horizon)
	if err != nil {
		return fmt.Errorf("the Horizon this input was emitted under: %w", err)
	}
	in, err := io.ReadAll(io.LimitReader(os.Stdin, maxEmission+1))
	if err != nil {
		return err
	}
	if len(in) > maxEmission {
		return errors.New("instance input too large")
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
func Drive(root string, up *net.UnixConn, stack []*os.File, s machine.Instance, e machine.Event) (result machine.Instance, failure error) {
	var (
		downs           []*net.UnixConn // handshakes with the successors, in instruction order
		children        []*exec.Cmd
		accepting       string // release awaiting the machine's validation
		acceptingRecord record.Record
	)
	defer func() {
		if failure != nil && len(stack) > 0 {
			if err := tree.PublishFailure(root, stack[0], tree.Failure{Stamp: s.Stamp, Message: failure.Error()}); err != nil {
				trace.Logf("recovery required: %v; recording failure: %v", failure, err)
			}
		}
		for _, down := range downs {
			down.Close()
		}
		for _, f := range stack {
			f.Close()
		}
	}()
	for e != nil {
		var effs []machine.Effect
		s, effs = machine.Step(s, e)
		if accepting != "" {
			for _, eff := range effs {
				if stop, ok := eff.(machine.Exit); ok && stop.Err != "" {
					_ = handshake.Send(up, handshake.Accepted{Error: stop.Err})
					return s, errors.New(stop.Err)
				}
			}
			if err := tree.Accept(root, acceptingRecord, accepting); err != nil {
				return s, err
			}
			if err := handshake.Send(up, handshake.Accepted{}); err != nil {
				return s, err
			}
			accepting = ""
		}
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
				var release handshake.Release
				files, err := handshake.Recv(up, &release)
				if err != nil {
					return s, fmt.Errorf("no Horizon, cannot execute: %w", err)
				}
				h, err := tree.LoadHorizon(root, release.Horizon)
				if err != nil {
					for _, f := range files {
						f.Close()
					}
					_ = handshake.Send(up, handshake.Accepted{Error: "released Horizon: " + err.Error()})
					return s, fmt.Errorf("released Horizon: %w", err)
				}
				if len(files) != h.Depth()+len(release.Resources) {
					for _, f := range files {
						f.Close()
					}
					_ = handshake.Send(up, handshake.Accepted{Error: "descriptor count does not match release"})
					return s, fmt.Errorf("release has %d descriptors; expected %d frames and %d resources", len(files), h.Depth(), len(release.Resources))
				}
				if err := custody.Validate(release.Resources, files[h.Depth():]); err != nil {
					for _, f := range files {
						f.Close()
					}
					_ = handshake.Send(up, handshake.Accepted{Error: err.Error()})
					return s, err
				}
				stack = files
				accepting = release.Horizon
				acceptingRecord = s.R
				trace.Logf("release  Horizon %s looked up (reach %v, %d frames, %d resources)", record.Short(release.Horizon), h.Names(), h.Depth(), len(release.Resources))
				e = machine.Released{Self: h, Frames: h.Depth(), Resources: release.Resources}
			case machine.Run:
				trace.Logf("execute  %s under rendered sandbox", filepath.Base(x.Surface.Path))
				var conn *os.File
				resourceIndex, incarnation := -1, ""
				if x.ConnectionOwner != "" {
					var request struct {
						Incarnation string `json:"incarnation"`
					}
					if err := json.Unmarshal(x.Stdin, &request); err != nil || request.Incarnation == "" {
						e = machine.Emitted{Err: "connection operation needs an incarnation"}
						break
					}
					var err error
					conn, err = custody.Find(s.Resources, stack[s.Self.Depth():], x.ConnectionOwner, request.Incarnation)
					if err != nil {
						e = machine.Emitted{Err: err.Error()}
						break
					}
					incarnation = request.Incarnation
					for i, r := range s.Resources {
						if r.Owner == x.ConnectionOwner && r.Incarnation == incarnation {
							resourceIndex = s.Self.Depth() + i
							break
						}
					}
				}
				e = run(x, conn)
				if x.CloseConnection && resourceIndex >= 0 {
					stack[resourceIndex].Close()
					stack = append(stack[:resourceIndex:resourceIndex], stack[resourceIndex+1:]...)
					out := e.(machine.Emitted)
					out.ReleasedResource = incarnation
					e = out
				}
			case machine.StartResident:
				var file *os.File
				e, file = startResident(x, root)
				if file != nil {
					stack = append(stack, file)
				}
			case machine.Spawn:
				down, child, ev, err := spawn(root, x)
				if err != nil {
					return s, err
				}
				downs, children, e = append(downs, down), append(children, child), ev
			case machine.KillSuccessor:
				trace.Logf("handoff  killing %d successor(s)", len(children))
				for _, child := range children {
					child.Process.Kill()
				}
			case machine.PublishEmission:
				if err := tree.PublishEmission(stack[0], x.Round, x.Hash); err != nil {
					return s, err
				}
			case machine.OpenDep:
				if err := tree.OpenDep(stack[0], x.Round, x.Child, x.Pos); err != nil {
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
				id, err := tree.PutHorizon(root, x.Child)
				if err != nil {
					return s, err
				}
				down, child := downs[x.Index], children[x.Index]
				if err := handshake.Send(down, handshake.Release{Horizon: id, Resources: s.Resources}, frames...); err != nil {
					return s, err
				}
				var accepted handshake.Accepted
				if _, err := handshake.Recv(down, &accepted); err != nil {
					return s, fmt.Errorf("successor did not accept custody: %w", err)
				}
				if accepted.Error != "" {
					return s, fmt.Errorf("successor refused custody: %s", accepted.Error)
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
			case machine.ReadStrand:
				var st any
				var err error
				if x.Invocation != "" {
					st, err = replay.ReadTreeInRoom(root, x.Invocation, x.Observations, x.Target, x.Schema, x.Origin, x.Room)
				} else {
					st, err = replay.ReadStrand(root, x.Observation, x.Target, x.Schema, x.Origin)
				}
				switch {
				case replay.Refused(err):
					e = machine.StrandRead{Refused: err.Error()}
				case err != nil:
					return s, err
				default:
					b, _ := json.Marshal(st)
					trace.Logf("strand   returned %d bytes of recorded work", len(b))
					e = machine.StrandRead{Strand: b}
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
			case machine.Park:
				trace.Logf("park     custody passes to the host at %s; waiting on an outside event", x.Host)
				e = machine.Parked{}
				if err := park.Send(x.Host, park.Frame{Root: root, Stamp: s.Stamp, Frames: s.Self.Depth(), Resources: s.Resources, Payload: x.Payload, Files: stack}); err != nil {
					e = machine.Parked{Err: err.Error()}
				}
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
		emission, err := tree.Emission(root, waiter.Frame, round)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		r := machine.Round{Stamp: round, Emission: emission}
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
func run(x machine.Run, connection *os.File) machine.Event {
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
	if connection != nil {
		cmd.ExtraFiles = []*os.File{connection}
	}
	var stderr boundedBuffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return machine.Emitted{Err: err.Error()}
	}
	if err := cmd.Start(); err != nil {
		return machine.Emitted{Err: err.Error()}
	}
	line, readErr := io.ReadAll(io.LimitReader(out, maxEmission+1))
	if len(line) > maxEmission || readErr != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return machine.Emitted{Err: "target output exceeded limit or could not be read"}
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return machine.Emitted{Err: "target timed out"}
	}
	if waitErr != nil && len(line) == 0 && stderr.Len() == 0 {
		return machine.Emitted{Err: waitErr.Error()}
	}
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

// startResident launches a registered stream target under its Horizon profile.
// Only the local socket endpoint survives the handoff. The process sees the
// peer as stdin and stdout, and exits on EOF when custody is dropped.
func startResident(x machine.StartResident, root string) (machine.Event, *os.File) {
	bin, err := os.ReadFile(x.Run.Surface.Path)
	if err != nil {
		return machine.Started{Err: err.Error()}, nil
	}
	if record.Sum(bin) != x.Run.Surface.SHA256 {
		return machine.Started{Err: "surface binary does not match its registered hash"}, nil
	}
	profile, env, work := x.Run.Profile, x.Run.Env, ""
	var prepared *preparedResident
	if x.Run.Surface.ResidentProfile != "" {
		prepared, err = prepareAppserver(x, root)
		if err != nil {
			return machine.Started{Err: err.Error()}, nil
		}
		profile, env, work = prepared.profile, prepared.env, prepared.work
	}
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		if prepared != nil {
			prepared.abort()
		}
		return machine.Started{Err: err.Error()}, nil
	}
	for _, fd := range fds {
		syscall.CloseOnExec(fd)
	}
	local := os.NewFile(uintptr(fds[0]), "resident-connection")
	peer := os.NewFile(uintptr(fds[1]), "resident-peer")
	defer peer.Close()
	argv := []string{"-p", profile, x.Run.Surface.Path}
	if prepared != nil {
		argv = []string{"-f", filepath.Join(filepath.Dir(prepared.work), "actor.sb"), x.Run.Surface.Path}
	}
	argv = append(argv, x.Run.Surface.Args...)
	cmd := exec.Command("/usr/bin/sandbox-exec", argv...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = peer, peer, os.Stderr
	cmd.Env, cmd.Dir = env, work
	if prepared != nil {
		cmd.ExtraFiles = []*os.File{prepared.watch}
	}
	if err := cmd.Start(); err != nil {
		local.Close()
		if prepared != nil {
			prepared.abort()
		}
		return machine.Started{Err: err.Error()}, nil
	}
	if prepared != nil {
		prepared.actorStarted()
	}
	incarnation := fmt.Sprintf("%d-%d", cmd.Process.Pid, time.Now().UnixNano())
	resource := custody.Resource{Kind: "connection", Owner: x.Owner, Incarnation: incarnation}
	out, _ := json.Marshal(map[string]any{"to": x.Horizon.Frame, "payload": map[string]any{
		"pid": cmd.Process.Pid, "incarnation": incarnation,
		"profile": record.Sum([]byte(profile))}})
	cmd.Process.Release()
	return machine.Started{Out: out, Resource: resource}, local
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
	id, err := tree.PutHorizon(root, x.Under)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := handshake.Send(down, handshake.Under{Horizon: id}); err != nil {
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

const maxEmission = 16 << 20

type boundedBuffer struct{ bytes.Buffer }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if remaining := (64 << 10) - b.Len(); remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		b.Buffer.Write(p)
	}
	return n, nil
}
