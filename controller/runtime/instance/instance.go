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
	"time"

	"skyra/controller/core/boundary"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/handshake"
	"skyra/controller/runtime/log"
	"skyra/controller/runtime/park"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/tree"
)

// Run is one run of skyra, started by a parent that holds fd 3.
func Run(root string) error {
	up, err := handshake.Conn(os.NewFile(3, "handshake"))
	if err != nil {
		return fmt.Errorf("an instance must be started by a parent: %w", err)
	}
	defer up.Close()
	var named handshake.Under // the Boundary the bytes were emitted under, by id
	if _, err := handshake.Recv(up, &named); err != nil {
		return err
	}
	under, err := object(root, "boundaries", named.Boundary)
	if err != nil {
		return fmt.Errorf("the Boundary this input was emitted under: %w", err)
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
// Step. Drive only does I/O: it answers reads with what the tree holds,
// unverified, commits writes and reports their outcomes, and performs acts.
//
// stack is the frame stack this instance holds, its handle table of frames:
// stack[0] is the frame its reply discharges, stack[1] the frame of the
// continuation waiting on it, and so on down to the operator's frame.
func Drive(root string, up *net.UnixConn, stack []*os.File, s machine.Instance, e machine.Event) (result machine.Instance, failure error) {
	var (
		downs    []*net.UnixConn // handshakes with the successors, in instruction order
		children []*exec.Cmd
	)
	defer func() {
		if failure != nil && len(stack) > 0 {
			if err := tree.PublishFailure(root, stack[0], tree.Failure{Stamp: s.Stamp(), Message: failure.Error()}); err != nil {
				log.Logf("recovery required: %v; recording failure: %v", failure, err)
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
		e = nil
		for _, eff := range effs {
			switch x := eff.(type) {
			case machine.Read:
				found := make([][]byte, len(x.Keys))
				for i, k := range x.Keys {
					b, err := readKey(root, stack, k)
					if err != nil {
						return s, err
					}
					found[i] = b
				}
				e = machine.ReadResult{Found: found}
			case machine.Commit:
				outcomes := make([]machine.Outcome, len(x.Writes))
				for i, w := range x.Writes {
					o, err := write(root, stack, w)
					if err != nil {
						return s, err
					}
					outcomes[i] = o
				}
				e = machine.Committed{Outcomes: outcomes}
			case machine.Execute:
				if x.Force {
					// The act leaves the failure domain: everything committed
					// before it is made durable first (40, B).
					if err := force(root); err != nil {
						return s, err
					}
				}
				var err error
				var done bool
				e, done, err = act(root, up, &stack, &downs, &children, x.Act)
				if err != nil || done {
					return s, err
				}
			}
		}
	}
	return s, errors.New("step stopped without exiting")
}

// force is tree.Force, replaced by a test that records the order of acts.
var force = tree.Force

// readKey answers one key of a read with what the tree holds under it, as
// stored: nil if nothing is there.
func readKey(root string, stack []*os.File, k machine.Key) ([]byte, error) {
	switch k := k.(type) {
	case machine.ObjectKey:
		return tree.Read(root, k.Kind, k.Hash)
	case machine.SlotsKey:
		if len(stack) < 2 {
			return nil, errors.New("no waiting frame to read")
		}
		children, err := tree.Children(stack[1], k.Round)
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if children == nil {
			children = []string{}
		}
		return json.Marshal(children)
	case machine.EmissionKey:
		return mutable(tree.Emission(root, k.Frame, k.Round))
	case machine.ResultKey:
		return mutable(tree.Result(root, k.Frame))
	case machine.TraceKey:
		var st any
		var err error
		if k.Invocation != "" {
			st, err = replay.ReadTreeInGenesis(root, k.Invocation, k.Observations, k.Target, k.Schema, k.Origin, k.Genesis)
		} else {
			st, err = replay.ReadTrace(root, k.Observation, k.Target, k.Schema, k.Origin)
		}
		switch {
		case replay.Refused(err):
			return json.Marshal(map[string]string{"refused": err.Error()})
		case err != nil:
			return nil, err
		}
		b, _ := json.Marshal(st)
		log.Logf("trace    returned %d bytes of recorded work", len(b))
		return json.Marshal(map[string]json.RawMessage{"trace": b})
	}
	return nil, fmt.Errorf("unknown read %T", k)
}

// mutable is a mutable name's content, or nil if it does not exist yet.
func mutable(v string, err error) ([]byte, error) {
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

// write performs one write of a commit and reports what it found.
func write(root string, stack []*os.File, w machine.Write) (machine.Outcome, error) {
	won := func(ok bool, err error) (machine.Outcome, error) {
		if err != nil {
			return 0, err
		}
		if !ok {
			return machine.Lost, nil
		}
		return machine.Created, nil
	}
	stored := func(existed bool, err error) (machine.Outcome, error) {
		switch {
		case errors.Is(err, tree.ErrConflict):
			return machine.Conflict, nil
		case err != nil:
			return 0, err
		case existed:
			return machine.Identical, nil
		}
		return machine.Created, nil
	}
	frame := func(i int) (*os.File, error) {
		if len(stack) <= i {
			return nil, fmt.Errorf("no frame at handle %d", i)
		}
		return stack[i], nil
	}
	switch w := w.(type) {
	case machine.ObjectWrite:
		if w.Kind == "boundaries" {
			if err := os.MkdirAll(filepath.Join(root, "boundaries"), 0o700); err != nil {
				return 0, err
			}
		}
		o, err := stored(tree.Store(root, w.Kind, w.Bytes))
		if err == nil && w.Kind == "records" && o != machine.Conflict {
			traceRecord(w.Bytes)
		}
		return o, err
	case machine.AcceptanceWrite:
		return stored(tree.Receipt(root, w.Stamp, w.Boundary))
	case machine.FrameWrite:
		return won(tree.MakeFrame(root, w.Stamp))
	case machine.EmissionWrite:
		f, err := frame(0)
		if err != nil {
			return 0, err
		}
		return won(tree.Emit(f, w.Round, w.Hash))
	case machine.SlotWrite:
		f, err := frame(0)
		if err != nil {
			return 0, err
		}
		return won(tree.Slot(f, w.Round, w.Child, w.Pos))
	case machine.ResultWrite:
		f, err := frame(0)
		if err != nil {
			return 0, err
		}
		log.Logf("settle   discharges the slot of %s: result → %s", record.Short(w.Frame), record.Short(w.Stamp))
		return won(tree.Discharge(f, w.Stamp))
	case machine.ClaimWrite:
		f, err := frame(1)
		if err != nil {
			return 0, err
		}
		o, err := won(tree.Claim(f, w.Round, w.Stamp))
		if o == machine.Created {
			log.Logf("claim    won: resumes %s", record.Short(w.Frame))
		}
		return o, err
	}
	return 0, fmt.Errorf("unknown write %T", w)
}

func traceRecord(b []byte) {
	var r record.Record
	if json.Unmarshal(b, &r) != nil {
		return
	}
	stamp := record.Sum(b)
	switch {
	case r.Supervisor:
		log.Logf("fault    %s: %s (committed %s)", r.Fault.Kind, r.Fault.Cause, record.Short(stamp))
	case r.State == record.Resolve:
		log.Logf("gather   every slot of %s discharged: committed resolve %s", record.Short(r.Frame), record.Short(stamp))
	default:
		log.Who = "inst " + record.Short(stamp)
		log.Logf("intake   committed %s  %s", record.Short(stamp), record.Describe(r))
	}
}

// act performs one execute. It returns the next event, or done once the
// instance has exited.
func act(root string, up *net.UnixConn, stack *[]*os.File, downs *[]*net.UnixConn, children *[]*exec.Cmd, a machine.Action) (machine.Event, bool, error) {
	switch x := a.(type) {
	case machine.SendStamp:
		if err := handshake.Send(up, handshake.StampMsg{Stamp: x.Stamp}); err != nil {
			return nil, false, err
		}
		var release handshake.Release
		files, err := handshake.Recv(up, &release)
		if err != nil {
			return nil, false, fmt.Errorf("no Boundary, cannot execute: %w", err)
		}
		// The core reads the Boundary, verifies it, and counts the
		// descriptors against it (41, V1 and V3): nothing is checked here.
		*stack = files
		log.Logf("release  Boundary %s named (%d descriptors)", record.Short(release.Boundary), len(files))
		return machine.Executed{Obs: []byte(release.Boundary), Handles: len(files)}, false, nil
	case machine.Acknowledge:
		return machine.Executed{}, false, handshake.Send(up, handshake.Accepted{Error: x.Err})
	case machine.Run:
		log.Logf("execute  %s under rendered sandbox", filepath.Base(x.Surface.Path))
		return run(x), false, nil
	case machine.Spawn:
		down, child, ev, err := spawn(root, x)
		if err != nil {
			return nil, false, err
		}
		*downs, *children = append(*downs, down), append(*children, child)
		return ev, false, nil
	case machine.KillSuccessors:
		log.Logf("handoff  killing %d successor(s)", len(*children))
		for _, child := range *children {
			child.Process.Kill()
		}
		return machine.Executed{}, false, nil
	case machine.ReleaseChild:
		frames := *stack
		if x.NewFrame != "" {
			f, err := tree.FrameNode(root, x.NewFrame)
			if err != nil {
				return nil, false, err
			}
			defer f.Close()
			frames = append([]*os.File{f}, *stack...)
		}
		down, child := (*downs)[x.Index], (*children)[x.Index]
		if err := handshake.Send(down, handshake.Release{Boundary: x.Boundary}, frames...); err != nil {
			return nil, false, err
		}
		var accepted handshake.Accepted
		if _, err := handshake.Recv(down, &accepted); err != nil {
			return nil, false, fmt.Errorf("successor did not accept custody: %w", err)
		}
		if accepted.Error != "" {
			return nil, false, fmt.Errorf("successor refused custody: %s", accepted.Error)
		}
		log.Logf("handoff  verified and released %s, custody passed, exiting", record.Short(x.Stamp))
		child.Process.Release() // D2: the parent does not wait on the child
		return machine.Executed{}, false, nil
	case machine.Park:
		log.Logf("park     custody passes to the host at %s; waiting on an outside event", x.Host)
		if err := park.Send(x.Host, park.Frame{Root: root, Stamp: x.Stamp, Frames: x.Frames, Payload: x.Payload, Files: *stack}); err != nil {
			return machine.Executed{Err: err.Error()}, false, nil
		}
		return machine.Executed{}, false, nil
	case machine.Pop:
		(*stack)[0].Close()
		*stack = (*stack)[1:]
		return machine.Executed{}, false, nil
	case machine.Exit:
		if x.Err != "" {
			return nil, true, errors.New(x.Err)
		}
		return nil, true, nil
	}
	return nil, false, fmt.Errorf("unknown act %T", a)
}

// run performs a Run effect: check the binary's content hash, run the copy
// it checked under the rendered profile, and read its one line of output;
// anything after it is discarded.
func run(x machine.Run) machine.Event {
	if x.Exec == "" {
		return machine.Executed{Err: "surface has no verified copy to run"}
	}
	if err := verifiedCopy(x.Surface, x.Exec); err != nil {
		return machine.Executed{Err: err.Error()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", "-p", x.Profile, x.Exec)
	cmd.Stdin = bytes.NewReader(x.Stdin)
	cmd.Env = x.Env
	var stderr boundedBuffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	var line firstLine
	cmd.Stdout = &line
	// A descendant left holding stdout or stderr cannot hold the step open.
	cmd.WaitDelay = time.Second
	if err := cmd.Start(); err != nil {
		return machine.Executed{Err: err.Error()}
	}
	waitErr := cmd.Wait()
	if line.overflow {
		return machine.Executed{Err: "target output exceeded limit"}
	}
	if ctx.Err() != nil {
		return machine.Executed{Err: "target timed out"}
	}
	if waitErr != nil && len(line.b) == 0 && stderr.Len() == 0 {
		return machine.Executed{Err: waitErr.Error()}
	}
	emission := line.b
	if len(emission) == 0 {
		// What the target said on the way out is the observation.
		if lines := bytes.Split(bytes.TrimSpace(stderr.Bytes()), []byte("\n")); len(lines[len(lines)-1]) > 0 {
			return machine.Executed{Err: "target exited without emitting: " + string(lines[len(lines)-1])}
		}
	}
	log.Logf("execute  target emitted %d bytes (sha256 %s)", len(emission), record.Short(record.Sum(emission)))
	return machine.Executed{Obs: emission}
}

// verifiedCopy checks a surface's program against its registered hash and
// publishes exactly the bytes it checked at exe, the path its profile lets
// run (43 D5). The registered path can change between a check and an exec;
// exe is under the floor, which nothing but the runtime can write. A copy
// already there is checked again, not trusted.
func verifiedCopy(s boundary.Surface, exe string) error {
	if have, err := os.ReadFile(exe); err == nil && record.Sum(have) == s.SHA256 {
		return nil
	}
	bin, err := os.ReadFile(s.Path)
	if err != nil {
		return err
	}
	if record.Sum(bin) != s.SHA256 {
		return errors.New("surface binary does not match its registered hash")
	}
	if filepath.Base(exe) != s.SHA256 {
		return errors.New("verified copy is not named by its hash")
	}
	dir := filepath.Dir(exe)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".copy-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // a no-op once renamed
	_, err = tmp.Write(bin)
	if err == nil {
		err = tmp.Chmod(0o500)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	// Published whole or not at all: a reader never sees part of a copy.
	return os.Rename(tmp.Name(), exe)
}

// spawn performs a Spawn: start skyra again with the bytes on stdin, send it
// the id of the Boundary they were emitted under (the core committed it), and
// wait for its stamp.
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
	if err := handshake.Send(down, handshake.Under{Boundary: x.Under}); err != nil {
		return nil, nil, nil, err
	}
	var st handshake.StampMsg
	if _, err := handshake.Recv(down, &st); err != nil {
		_ = cmd.Wait()
		return down, cmd, machine.Executed{Err: err.Error()}, nil
	}
	return down, cmd, machine.Executed{Obs: []byte(st.Stamp)}, nil
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
