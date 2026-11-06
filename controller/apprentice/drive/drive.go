// Package drive is the shell for proposal 58's machine (tree.Machine): it
// runs the invokes the machine names, asks the model when it startles, and
// feeds back what comes of each. It stands beside the old shell until the
// cutover.
package drive

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os/exec"
	"syscall"
	"time"

	"skyra/apprentice/tree"
)

// A Model writes a guard when the machine startles, reading the frame's $n.
type Model interface {
	Guard(ask tree.AskModel) (string, error)
}

// A World runs an invoke and gives back what its process gave: stdout,
// exit status and stderr. An error is a process that could not be run.
type World interface {
	Run(r tree.RunInvoke) (tree.Answered, error)
}

// Exec runs an invoke's argv as a process, never through a shell, so a
// value passed as $n is one argument whatever it holds. A process killed by
// a signal exits 128 plus the signal, as a shell reports it.
type Exec struct {
	Dir     string
	Timeout time.Duration
}

func (e Exec) Run(r tree.RunInvoke) (tree.Answered, error) {
	if len(r.Argv) == 0 {
		return tree.Answered{}, errors.New("an invoke with no binary")
	}
	ctx := context.Background()
	if e.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.Timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, r.Argv[0], r.Argv[1:]...)
	cmd.Dir = e.Dir
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	code := 0
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		err, code = nil, exit.ExitCode() // it ran: its failure is an observation
		if ws, ok := exit.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			code = 128 + int(ws.Signal())
		}
	}
	return tree.Answered{Bytes: out.Bytes(), Exit: code, Stderr: errOut.Bytes()}, err
}

// Hashes are the runtime's: SHA-256 over the canonical JSON of what a base
// is made of.
var Hashes = tree.WalkHashes{
	Input: func(contract string, passed []string) string {
		return hash("input", contract, passed)
	},
	Return: func(input, shape string) string { return hash("return", input, shape) },
}

func hash(parts ...any) string {
	b, _ := json.Marshal(parts)
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Run drives a machine from its goal to its end: the value `return $n`
// gave, or the error observation of an exhausted bound, the transition bound or one
// transition's refinement bound, which the machine itself decides.
func Run(m tree.Machine, goal []byte, model Model, world World) (tree.Machine, any, error) {
	m, eff, err := m.Step(tree.Answered{Bytes: goal})
	for err == nil {
		switch e := eff.(type) {
		case tree.AskModel:
			var guard string
			if guard, err = model.Guard(e); err != nil {
				return m, nil, err
			}
			m, eff, err = m.Step(tree.Wrote{Guard: guard})
		case tree.RunInvoke:
			var out tree.Answered
			if out, err = world.Run(e); err != nil {
				return m, nil, err
			}
			m, eff, err = m.Step(out)
		case tree.Finished:
			return m, e.Value, nil
		case tree.Exhausted:
			return m, e.Value, nil
		default:
			return m, nil, errors.New("an effect the shell does not handle")
		}
	}
	return m, nil, err
}
