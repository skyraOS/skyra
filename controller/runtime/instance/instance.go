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
	"strings"
	"syscall"
	"time"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/handshake"
	"skyra/controller/runtime/park"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/trace"
	"skyra/controller/runtime/tree"

	"skyra/controller/runtime/hold"
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
	under, err := object(root, "horizons", named.Horizon)
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
	_, err = Drive(root, up, nil, machine.Instance{Floor: root, Sockets: hold.Sockets(), Phase: machine.Intake}, machine.Received{Under: under, In: in})
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
				// The core verifies the Horizon and counts the descriptors
				// against it (41, V1 and V3): nothing is checked here.
				self, err := object(root, "horizons", release.Horizon)
				if err != nil {
					for _, f := range files {
						f.Close()
					}
					_ = handshake.Send(up, handshake.Accepted{Error: "released Horizon: " + err.Error()})
					return s, fmt.Errorf("released Horizon: %w", err)
				}
				stack = files
				accepting = release.Horizon
				acceptingRecord = s.R
				trace.Logf("release  Horizon %s read (%d descriptors)", record.Short(release.Horizon), len(files))
				e = machine.Released{Self: self, Frames: len(files)}
			case machine.Run:
				trace.Logf("execute  %s under rendered sandbox", filepath.Base(x.Surface.Path))
				e = run(x)
			case machine.StartResident:
				e = startResident(x, root)
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
				if err := handshake.Send(down, handshake.Release{Horizon: id}, frames...); err != nil {
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
				if err := park.Send(x.Host, park.Frame{Root: root, Stamp: s.Stamp, Frames: s.Self.Depth(), Payload: x.Payload, Files: stack}); err != nil {
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
// to the frame itself. Discharges are read from each child's frame. Every
// record goes to the core as read, and the core verifies it; the walk decodes
// a resolve only to find the round before it.
func gather(root string, frame *os.File, waiter horizon.Horizon) (machine.Event, error) {
	f, err := object(root, "records", waiter.Frame)
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
			inv, err := object(root, "records", c)
			if err != nil {
				return nil, err
			}
			child := machine.Child{Invoke: inv}
			if result, err := tree.Result(root, c); err == nil {
				d, err := object(root, "records", result)
				if err != nil {
					return nil, err
				}
				child.Discharge = &d
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
// the rendered profile, and read its one line of output; anything after it is discarded.
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
	var stderr boundedBuffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	var line firstLine
	cmd.Stdout = &line
	// A descendant left holding stdout or stderr cannot hold the step open.
	cmd.WaitDelay = time.Second
	if err := cmd.Start(); err != nil {
		return machine.Emitted{Err: err.Error()}
	}
	waitErr := cmd.Wait()
	if line.overflow {
		return machine.Emitted{Err: "target output exceeded limit"}
	}
	if ctx.Err() != nil {
		return machine.Emitted{Err: "target timed out"}
	}
	if waitErr != nil && len(line.b) == 0 && stderr.Len() == 0 {
		return machine.Emitted{Err: waitErr.Error()}
	}
	emission := line.b
	if len(emission) == 0 {
		// What the target said on the way out is the observation.
		if lines := bytes.Split(bytes.TrimSpace(stderr.Bytes()), []byte("\n")); len(lines[len(lines)-1]) > 0 {
			return machine.Emitted{Err: "target exited without emitting: " + string(lines[len(lines)-1])}
		}
	}
	trace.Logf("execute  target emitted %d bytes (sha256 %s)", len(emission), record.Short(record.Sum(emission)))
	return machine.Emitted{Out: emission}
}

// startResident launches a registered stream target under its Horizon profile,
// behind a holder that serves its edge's socket (runtime/hold, proposal 38).
// If the edge already has a live resident, that one is the answer: recorded as
// reused, so every turn of a conversation shares one. The resident sees the
// holder as stdin and stdout and exits on EOF when the holder lets it go.
// Nothing of the stream stays with this instance.
func startResident(x machine.StartResident, root string) machine.Event {
	if incarnation := hold.Hello(x.Socket); incarnation != "" {
		trace.Logf("execute  the edge already holds resident %s: reused", incarnation)
		out, _ := json.Marshal(map[string]any{"to": x.Horizon.Frame, "payload": map[string]any{"incarnation": incarnation, "reused": true}})
		return machine.Started{Out: out}
	}
	bin, err := os.ReadFile(x.Run.Surface.Path)
	if err != nil {
		return machine.Started{Err: err.Error()}
	}
	if record.Sum(bin) != x.Run.Surface.SHA256 {
		return machine.Started{Err: "surface binary does not match its registered hash"}
	}
	profile, env, work := x.Run.Profile, x.Run.Env, ""
	var prepared *preparedResident
	switch x.Run.Surface.ResidentProfile {
	case "darwin-appserver-v1":
		prepared, err = prepareAppserver(x, root)
	case "darwin-claude-v1":
		prepared, err = prepareClaude(x, root)
	}
	if err != nil {
		return machine.Started{Err: err.Error()}
	}
	if prepared != nil {
		profile, env, work = prepared.profile, prepared.env, prepared.work
	}
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		if prepared != nil {
			prepared.abort()
		}
		return machine.Started{Err: err.Error()}
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
	// The resident outlives this step, so it keeps none of this process's
	// output: what it says on stderr goes to its edge's log.
	logs, err := edgeLog(x.Socket)
	if err != nil {
		local.Close()
		if prepared != nil {
			prepared.abort()
		}
		return machine.Started{Err: err.Error()}
	}
	defer logs.Close()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = peer, peer, logs
	cmd.Env, cmd.Dir = env, work
	if prepared != nil {
		cmd.ExtraFiles = []*os.File{prepared.watch}
	}
	if err := cmd.Start(); err != nil {
		local.Close()
		if prepared != nil {
			prepared.abort()
		}
		return machine.Started{Err: err.Error()}
	}
	if prepared != nil {
		prepared.actorStarted()
	}
	incarnation := fmt.Sprintf("%d-%d", cmd.Process.Pid, time.Now().UnixNano())
	pid := cmd.Process.Pid
	cmd.Process.Release()
	owner := []string{}
	if x.Horizon.Resume != nil {
		owner = []string{root, x.Horizon.Resume.Frame}
	}
	if err := startHolder(local, logs, x.Socket, incarnation, owner...); err != nil {
		return machine.Started{Err: "resident holder: " + err.Error()}
	}
	out, _ := json.Marshal(map[string]any{"to": x.Horizon.Frame, "payload": map[string]any{
		"pid": pid, "incarnation": incarnation,
		"profile": record.Sum([]byte(profile))}})
	return machine.Started{Out: out}
}

// edgeLog is where a resident and its holder write diagnostics: beside the
// edge's socket, in the user's private directory, appended across restarts.
func edgeLog(socket string) (*os.File, error) {
	return os.OpenFile(strings.TrimSuffix(socket, ".sock")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
}

// skyraBinary is the skyra executable a holder runs as: this one, except
// under a test, which is not skyra.
var skyraBinary = os.Executable

// startHolder hands the resident's end of its stream to a holder serving the
// edge's socket, and returns once that holder answers for the incarnation.
// Closing local here leaves the holder the stream's only keeper: when it lets
// go, the resident reads EOF and exits.
func startHolder(local, logs *os.File, socket, incarnation string, owner ...string) error {
	defer local.Close()
	self, err := skyraBinary()
	if err != nil {
		return err
	}
	args := []string{"hold", socket, incarnation}
	args = append(args, owner...)
	cmd := exec.Command(self, args...)
	cmd.ExtraFiles = []*os.File{local}
	cmd.Stderr = logs // like the resident, it outlives this step and keeps none of its output
	if err := cmd.Start(); err != nil {
		return err
	}
	cmd.Process.Release()
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if hold.Hello(socket) == incarnation {
			return nil
		}
	}
	return errors.New("did not start serving the edge")
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
	r, err := object(root, "records", st.Stamp) // the core verifies it
	if err != nil {
		return down, cmd, nil, err
	}
	return down, cmd, machine.Stamped{Record: r}, nil
}

// object reads what the tree holds under a key, unverified: the core
// verifies it.
func object(root, kind, key string) (machine.Object, error) {
	b, err := tree.Read(root, kind, key)
	return machine.Object{Key: key, Raw: b}, err
}

const maxEmission = 16 << 20

// firstLine keeps a target's one line of output, without its newline, and
// discards the rest so a target writing more is never blocked on a full pipe.
// A line longer than maxEmission stops the copy: the target's next write fails.
type firstLine struct {
	b              []byte
	done, overflow bool
}

func (l *firstLine) Write(p []byte) (int, error) {
	if l.done {
		return len(p), nil
	}
	n := len(p)
	if i := bytes.IndexByte(p, '\n'); i >= 0 {
		p, l.done = p[:i], true
	}
	if len(l.b)+len(p) > maxEmission {
		l.overflow = true
		return 0, errors.New("target output exceeded limit")
	}
	l.b = append(l.b, p...)
	return n, nil
}

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
