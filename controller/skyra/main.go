// Command skyra is the stateless controller. It owns only argv: which mode,
// and the operator's flags.
//
//	skyra boot -root DIR -registry FILE [-scratch DIR] [-to CAPABILITY] [-payload JSON | -text TEXT]
//	skyra instance DIR   (internal: started only by a parent holding fd 3)
//	skyra replay DIR     re-derive a finished lineage from its records
//	skyra metric DIR     the research metric, from the records (next/32)
//	skyra resume DIR     (internal: started only by an actor host holding a parked frame)
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/boot"
	"skyra/controller/runtime/custody"
	"skyra/controller/runtime/handshake"
	"skyra/controller/runtime/instance"
	"skyra/controller/runtime/metric"
	"skyra/controller/runtime/park"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/trace"
	"skyra/controller/runtime/tree"

	"skyra/controller/runtime/hold"
)

func main() {
	custody.SealInherited() // nothing the launcher leaked reaches a target
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: skyra boot -root DIR -registry FILE [-to CAPABILITY] [-payload JSON] | skyra instance DIR")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "boot":
		fs := flag.NewFlagSet("boot", flag.ExitOnError)
		root := fs.String("root", "", "state directory (.skyra)")
		reg := fs.String("registry", "", "the operator's registry of surfaces and their contracts")
		scratch := fs.String("scratch", "", "a directory everything in the room may read and write, and run nothing from")
		to := fs.String("to", "echo", "the capability the seed addresses")
		payload := fs.String("payload", `{"text":"hello"}`, "the seed's payload (JSON)")
		text := fs.String("text", "", "shorthand for -payload '{\"text\": TEXT}'")
		detach := fs.Bool("detach", false, "seed and return without waiting: for a conversation, which ends when the user closes it")
		fs.Parse(os.Args[2:])
		if *text != "" {
			b, _ := json.Marshal(map[string]string{"text": *text})
			*payload = string(b)
		}
		if *root == "" || *reg == "" {
			fs.Usage()
			os.Exit(2)
		}
		var real string
		if real, err = realDir(*root); err == nil {
			_, err = boot.Run(real, *reg, *scratch, *to, json.RawMessage(*payload), *detach)
		}
	case "replay":
		if len(os.Args) != 3 {
			err = fmt.Errorf("usage: skyra replay DIR")
			break
		}
		var head string
		if head, err = boot.Head(os.Args[2]); err == nil {
			var n int
			if n, err = replay.Replay(os.Args[2], head); err == nil {
				fmt.Printf("replay: %d records re-derived from genesis by step, byte-identical\n", n)
			}
		}
	case "metric":
		if len(os.Args) != 3 {
			err = fmt.Errorf("usage: skyra metric DIR")
			break
		}
		var m metric.Metric
		if m, err = metric.Read(os.Args[2]); err == nil {
			fmt.Println(m)
		}
	case "resume":
		// Started only by an actor host, holding a parked frame stack as fds 3..
		trace.Who = fmt.Sprintf("resume pid %d", os.Getpid())
		if len(os.Args) != 3 {
			err = fmt.Errorf("usage: skyra resume DIR")
		} else {
			err = resume(os.Args[2])
		}
	case "hold":
		// Started only by a resident's recorded start, holding the resident's
		// stream at fd 3: it serves the edge's socket (proposal 38).
		trace.Who = fmt.Sprintf("hold pid %d", os.Getpid())
		if len(os.Args) != 4 {
			err = fmt.Errorf("usage: skyra hold SOCKET INCARNATION")
		} else {
			err = holdResident(os.Args[2], os.Args[3])
		}
	case "instance":
		trace.Who = fmt.Sprintf("inst pid %d", os.Getpid())
		if len(os.Args) != 3 {
			err = fmt.Errorf("usage: skyra instance DIR")
		} else {
			err = instance.Run(os.Args[2])
		}
	default:
		err = fmt.Errorf("unknown mode %q", os.Args[1])
	}
	if err != nil {
		trace.Logf("error    %v", err)
		os.Exit(1)
	}
}

// realDir resolves symlinks so the sandbox denies the real path
// (/tmp is /private/tmp on macOS).
func realDir(p string) (string, error) {
	if err := os.MkdirAll(p, 0o700); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// resume continues a parked instance: the host hands back its frame stack as
// fds 3.., and on stdin the parked stamp and the outside event, as its
// target's output. The host holds a reference and descriptors, never
// authority: the instance, its Horizon and the floor are derived here, from
// the tree (next/20, finding 1).
func resume(root string) error {
	var r park.Resumed
	if err := json.NewDecoder(os.Stdin).Decode(&r); err != nil {
		return err
	}
	stack := handshake.Inherited(3, r.Frames)
	s, err := parked(root, r)
	if err != nil {
		for _, f := range stack {
			f.Close()
		}
		return fmt.Errorf("resume %s: %w", record.Short(r.Stamp), err)
	}
	trace.Who = "resume " + record.Short(r.Stamp)
	trace.Logf("resume   the host returns custody: an outside event arrived (%d bytes)", len(r.Out))
	_, err = instance.Drive(root, nil, stack, s, machine.Emitted{Out: []byte(r.Out)})
	return err
}

// parked derives the instance a host is resuming. It must be an invoke of an
// async surface that was released and never handed on: not another step, and
// not one already resumed.
func parked(root string, r park.Resumed) (machine.Instance, error) {
	rec, err := tree.Load(root, r.Stamp)
	if err != nil {
		return machine.Instance{}, err
	}
	self, err := replay.HorizonAt(root, r.Stamp)
	if err != nil {
		return machine.Instance{}, err
	}
	switch {
	case rec.State != record.Invoke || self.Surface == nil || self.Surface.Host == "":
		return machine.Instance{}, errors.New("not a parked invocation")
	case self.Depth() != r.Frames:
		return machine.Instance{}, fmt.Errorf("carries %d frames, its Horizon has %d", r.Frames, self.Depth())
	}
	if _, err := tree.Result(root, self.Frame); err == nil {
		return machine.Instance{}, errors.New("already settled")
	}
	handed, err := tree.HasSuccessor(root, r.Stamp)
	if err != nil {
		return machine.Instance{}, err
	}
	if handed {
		return machine.Instance{}, errors.New("already resumed")
	}
	return machine.Instance{Floor: root, State: record.Invoke, Phase: machine.Execute, R: rec, Stamp: r.Stamp, Self: self, Sockets: hold.Sockets()}, nil
}

// holdResident keeps the resident's stream at fd 3 behind the edge's socket.
// The idle backoff starts at 30 s and stops the resident once its interval
// would pass 16 min with no request: about half an hour of idleness, the old
// rule. SKYRA_HOLD_START and SKYRA_HOLD_MAX (Go durations) override them.
func holdResident(socket, incarnation string) error {
	f := os.NewFile(3, "resident-stream")
	c, err := net.FileConn(f)
	f.Close()
	if err != nil {
		return err
	}
	stream, ok := c.(*net.UnixConn)
	if !ok {
		return errors.New("the resident's stream is not a unix socket")
	}
	start, max := 30*time.Second, 16*time.Minute
	if d, err := time.ParseDuration(os.Getenv("SKYRA_HOLD_START")); err == nil {
		start = d
	}
	if d, err := time.ParseDuration(os.Getenv("SKYRA_HOLD_MAX")); err == nil {
		max = d
	}
	return hold.Serve(stream, hold.Config{Socket: socket, Incarnation: incarnation, Start: start, Max: max, Log: os.Stderr})
}
