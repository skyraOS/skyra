// Package boot is the operator's process: it writes genesis, then runs as an
// instance whose target (the operator) has just emitted the seed. The
// operator's frame is the first dependent: boot waits on it, and the instance
// that resolves it is what wakes boot.
package boot

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/instance"
	"skyra/controller/runtime/registry"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/trace"
	"skyra/controller/runtime/tree"
)

// Outcome is how the operator's seed ended: the record that discharged its
// slot, and the head the lineage is replayed from.
type Outcome struct {
	Head      string
	Discharge record.Record
}

// Run boots: genesis from the operator's registry, then the operator emits
// the seed (addressed to a capability in the root reach), then boot waits for
// its frame to resolve, prints the lineage, and replays it. A scratch
// directory, when given, joins the room's authority: everything under the
// root may read and write it, and run nothing from it.
func Run(root, registryFile, scratch, to string, payload json.RawMessage, detach bool) (Outcome, error) {
	if err := tree.Init(root); err != nil {
		return Outcome{}, err
	}
	reach, authority, err := registry.Load(registryFile)
	if err != nil {
		return Outcome{}, err
	}
	if scratch != "" {
		if authority.Scratch, err = scratchDir(root, scratch); err != nil {
			return Outcome{}, err
		}
	}
	reach[machine.System] = machine.Syscalls() // the Horizon's own contracts: its syscalls (next/32)
	for _, s := range reach {                  // each actor's own state, the one part of the floor its sandbox may touch
		if s.State != "" {
			if err := os.MkdirAll(filepath.Join(root, s.State), 0o700); err != nil {
				return Outcome{}, err
			}
			if s.BundlePath != "" {
				bundle, err := os.ReadFile(s.BundlePath)
				if err != nil {
					return Outcome{}, err
				}
				if record.Sum(bundle) != s.BundleSHA256 {
					return Outcome{}, errors.New("registered procedure bundle changed")
				}
				binary, err := os.ReadFile(s.Path)
				if err != nil {
					return Outcome{}, err
				}
				if record.Sum(binary) != s.SHA256 {
					return Outcome{}, errors.New("registered installer binary changed")
				}
				cmd := exec.Command(s.Path, "install-bundle", filepath.Join(root, s.State), s.BundlePath)
				if out, err := cmd.CombinedOutput(); err != nil {
					return Outcome{}, fmt.Errorf("provision %s: %w: %s", s.State, err, out)
				}
			}
		}
	}

	// Genesis: the room. The root Horizon's reach and authority come from the
	// operator, and everything under it shares them until a narrowing. It is
	// the one origin with no predecessor, so it is authored, and it carries a
	// nonce: two boots of the same registry are two histories, never one.
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return Outcome{}, err
	}
	roomBytes, _ := json.Marshal(map[string]any{"reach": reach, "authority": authority})
	g0, err := tree.Commit(root, record.Record{Version: 2, Input: record.Sum(roomBytes), State: record.Invoke, Reach: reach, Authority: &authority,
		Author: "operator", Nonce: hex.EncodeToString(nonce)}) // frame 0: the operator's, in invoke
	if err != nil {
		return Outcome{}, err
	}
	node0, err := tree.OpenFrame(root, g0)
	if err != nil {
		return Outcome{}, err
	}
	h0 := horizon.Horizon{Stamp: g0, Reach: reach, Authority: authority, Frame: g0}
	trace.Logf("genesis  committed %s  root reach %v", record.Short(g0), h0.Names())

	// Boot is an instance whose target is the operator: it is already live
	// under the root Horizon, and its target has just emitted the seed. The
	// seed is operands like any other output; intake derives its state.
	seed, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	s0 := machine.Instance{Floor: root, State: record.Invoke, Phase: machine.Execute, Stamp: g0, Self: h0}
	if _, err := instance.Drive(root, nil, []*os.File{node0}, s0, machine.Emitted{Out: seed}); err != nil {
		return Outcome{}, err
	}

	if detach {
		fmt.Printf("seeded %s; state: %s\n", record.Short(g0), root)
		return Outcome{}, nil
	}

	// The operator's frame is discharged in one of two ways: the seed was
	// refused (its fault discharges frame 0's slot), or it ran and its reply
	// resolved frame 0's round.
	head, err := wait(root, g0)
	if err != nil {
		return Outcome{}, err
	}
	fmt.Println()
	fmt.Println("lineage (newest first):")
	for s := head; s != ""; {
		r, err := tree.Load(root, s)
		if err != nil {
			return Outcome{}, err
		}
		fmt.Printf("  %s  %s\n", record.Short(s), record.Describe(r))
		s = r.Predecessor
	}
	n, err := replay.Replay(root, head)
	if err != nil {
		return Outcome{}, fmt.Errorf("replay: %w", err)
	}
	fmt.Printf("replay: %d records re-derived from genesis by step, byte-identical\n", n)

	out := Outcome{Head: head}
	h, _ := tree.Load(root, head)
	out.Discharge = h
	if h.State == record.Resolve {
		if out.Discharge, err = tree.Load(root, h.Deps[0]); err != nil {
			return Outcome{}, err
		}
	}
	switch out.Discharge.State {
	case record.Return:
		var answer struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(out.Discharge.Payload, &answer) == nil && answer.Text != "" {
			fmt.Printf("answer: %s\n", answer.Text)
		} else {
			fmt.Printf("answer: %s\n", out.Discharge.Payload)
		}
	case record.Fault:
		fmt.Printf("answer: fault (%s): %s\n", out.Discharge.Fault.Kind, out.Discharge.Fault.Cause)
	}
	return out, nil
}

// wait blocks until the operator's frame is discharged, woken by the
// filesystem (kqueue) whenever an entry is added to the frame's directory:
// either its result, or the fan-in claim of its round.
func wait(root, g0 string) (string, error) {
	dir, err := os.Open(filepath.Join(root, "frames", g0))
	if err != nil {
		return "", err
	}
	defer dir.Close()
	kq, err := syscall.Kqueue()
	if err != nil {
		return "", err
	}
	defer syscall.Close(kq)
	watch := syscall.Kevent_t{Ident: uint64(dir.Fd()), Filter: syscall.EVFILT_VNODE,
		Flags: syscall.EV_ADD | syscall.EV_CLEAR, Fflags: syscall.NOTE_WRITE}
	if _, err := syscall.Kevent(kq, []syscall.Kevent_t{watch}, nil, nil); err != nil {
		return "", err
	}
	deadline := time.Now().Add(5 * time.Minute)
	for {
		if s, err := tree.Resolved(root, g0, g0); err == nil {
			return s, nil
		}
		if s, err := tree.Result(root, g0); err == nil {
			return s, nil
		}
		left := time.Until(deadline)
		if left <= 0 {
			return "", errors.New("the operator's frame was not discharged")
		}
		ts := syscall.NsecToTimespec(left.Nanoseconds())
		events := make([]syscall.Kevent_t, 1)
		if _, err := syscall.Kevent(kq, nil, events, &ts); err != nil && err != syscall.EINTR {
			return "", err
		}
	}
}

// Head finds a finished lineage's head in a state directory: the operator's
// frame, discharged either by its result or by the resolve that won its claim.
func Head(root string) (string, error) {
	frames, err := os.ReadDir(filepath.Join(root, "frames"))
	if err != nil {
		return "", err
	}
	for _, f := range frames {
		r, err := tree.Load(root, f.Name())
		if err != nil || r.Predecessor != "" {
			continue // not genesis
		}
		if s, err := tree.Resolved(root, f.Name(), f.Name()); err == nil {
			return s, nil
		}
		if s, err := tree.Result(root, f.Name()); err == nil {
			return s, nil
		}
		return "", errors.New("the lineage has not finished: the operator's frame is still waiting")
	}
	return "", errors.New("no genesis frame")
}

// scratchDir creates a scratch directory and returns its real path. It may
// not hold the state directory or sit inside it: the floor is never granted.
func scratchDir(root, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	for _, pair := range [][2]string{{real, root}, {root, real}} {
		if rel, err := filepath.Rel(pair[0], pair[1]); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("scratch %s overlaps the state directory %s", real, root)
		}
	}
	return real, nil
}
