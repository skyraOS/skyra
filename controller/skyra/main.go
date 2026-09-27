// Command skyra is the stateless controller. It owns only argv: which mode,
// and the operator's flags.
//
//	skyra boot -root DIR -registry FILE [-to CAPABILITY] [-payload JSON | -text TEXT]
//	skyra instance DIR   (internal: started only by a parent holding fd 3)
//	skyra replay DIR     re-derive a finished lineage from its records
//	skyra resume DIR     (internal: started only by an actor host holding a parked frame)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"skyra/controller/boot"
	"skyra/controller/instance"
	"skyra/controller/machine"
	"skyra/controller/park"
	"skyra/controller/record"
	"skyra/controller/replay"
	"skyra/controller/trace"
)

func main() {
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
			_, err = boot.Run(real, *reg, *to, json.RawMessage(*payload), *detach)
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
	case "resume":
		// Started only by an actor host, holding a parked frame stack as fds 3..
		trace.Who = fmt.Sprintf("resume pid %d", os.Getpid())
		if len(os.Args) != 3 {
			err = fmt.Errorf("usage: skyra resume DIR")
		} else {
			err = resume(os.Args[2])
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
	if err := os.MkdirAll(p, 0o755); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// resume continues a parked instance: the host hands back its frame stack as
// fds 3.., and on stdin the instance and the outside event, as its target's
// output.
func resume(root string) error {
	var r park.Resumed
	if err := json.NewDecoder(os.Stdin).Decode(&r); err != nil {
		return err
	}
	stack := make([]*os.File, r.Instance.Self.Depth())
	for i := range stack {
		stack[i] = os.NewFile(uintptr(3+i), "frame")
	}
	trace.Who = "resume " + record.Short(r.Instance.Stamp)
	trace.Logf("resume   the host returns custody: an outside event arrived (%d bytes)", len(r.Out))
	_, err := instance.Drive(root, nil, stack, r.Instance, machine.Emitted{Out: []byte(r.Out)})
	return err
}
