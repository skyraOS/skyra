// Package drive is the shell for proposal 61's machine (model.Machine): it
// runs the leaves the machine pushes, asks the model when it startles, and
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
	"strconv"
	"syscall"
	"time"

	"skyra/typesys-prototype/model"
)

// A Model writes a guard when the machine startles, reading the frame's $n.
type Model interface {
	Guard(ask model.Ask) (string, error)
}

// A World runs a leaf and gives back what its process gave: stdout, exit
// status and stderr. An error is a process that could not be run.
type World interface {
	Run(r model.RunLeaf) (model.Answered, error)
}

// Exec runs a leaf's argv as a process, never through a shell, so a
// value passed as $n is one argument whatever it holds. A process killed by
// a signal exits 128 plus the signal, as a shell reports it.
type Exec struct {
	Dir     string
	Timeout time.Duration
}

func (e Exec) Run(r model.RunLeaf) (model.Answered, error) {
	if len(r.Argv) == 0 {
		return model.Answered{}, errors.New("a leaf with no binary")
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
	return model.Answered{Bytes: out.Bytes(), Exit: code, Stderr: errOut.Bytes()}, err
}

// Ids are the runtime's: SHA-256 over the canonical JSON of what each
// location is made of.
type Ids struct{}

func (Ids) Child(p model.InstLoc, n int, s model.Scope) model.InstLoc {
	return model.InstLoc(hash("child", p, n, s.Text()))
}

func (Ids) First(d model.DefLoc, g model.Grant) model.InstLoc {
	return model.InstLoc(hash("first", d, g.Defs))
}

func (Ids) Slot(i model.InstLoc, n int) model.InstLoc {
	return model.InstLoc(string(i) + "#" + strconv.Itoa(n))
}

func (Ids) Program(p model.Program) model.DefLoc { return model.DefLoc(hash("program", p)) }

// Version is the language this shell runs: 58's shapes and cuts, CEL, the
// template grammar the prompts teach, and return's value.
const Version model.DefLoc = "skyra/61"

// ModelDef is the one model every delegation consults: the shell's Model.
const ModelDef model.DefLoc = "model"

// Program is the program a shell grants its tools under: act reaches them,
// delegate the model; the classes are tools it declares interchangeable.
func Program(c model.Constants, tools []string, classes ...model.Class) model.Program {
	defs := make([]model.DefLoc, len(tools))
	for i, t := range tools {
		defs[i] = model.DefLoc(t)
	}
	p := model.NewProgram(Version, c, defs, []model.DefLoc{ModelDef})
	p.Classes = classes
	return p
}

// NewMachine is a run of the tools under the constants and classes, over
// memory.
func NewMachine(c model.Constants, tools []string, mem model.Memory, classes ...model.Class) (model.Machine, error) {
	return model.NewMachine(Ids{}, Program(c, tools, classes...), mem)
}

func hash(parts ...any) string {
	b, _ := json.Marshal(parts)
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Run drives a machine from its goal to its end: what its return gave
// (model.ReturnValue), or the error observation of an exhausted bound, the
// transition bound or one transition's refinement bound, which the machine
// itself decides.
func Run(m model.Machine, goal []byte, mdl Model, world World) (model.Machine, any, error) {
	m, eff, err := m.Step(model.Answered{Bytes: goal})
	for err == nil {
		switch e := eff.(type) {
		case model.Ask:
			var guard string
			if guard, err = mdl.Guard(e); err != nil {
				return m, nil, err
			}
			m, eff, err = m.Step(model.Wrote{Guard: guard})
		case model.RunLeaf:
			var out model.Answered
			if out, err = world.Run(e); err != nil {
				return m, nil, err
			}
			m, eff, err = m.Step(out)
		case model.Finished:
			return m, e.Value, nil
		case model.Exhausted:
			return m, e.Value, nil
		default:
			return m, nil, errors.New("an effect the shell does not handle")
		}
	}
	return m, nil, err
}
