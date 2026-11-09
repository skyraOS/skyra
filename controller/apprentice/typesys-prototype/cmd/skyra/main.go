// Command skyra runs a goal as a tree over real tools, startling the
// claude CLI where memory has nothing, and keeps what it learns in a store:
// the log on disk, which memory is rebuilt from at each start. A second
// run of a goal it has learned asks nothing.
//
//	skyra -tools ls,cat "show the contents of the only file in this directory:
//	/tmp/x"
//
// The goal is the arguments, joined by spaces, or stdin when there are
// none. What the tree returns is printed, its exit status the command's;
// a bound exhausted prints its error and exits 2.
//
//	skyra -show
//
// lists what the store has learned, each key by an id. Keys are shared by
// every program that meets them, so this is the whole store.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"skyra/typesys-prototype/drive"
	"skyra/typesys-prototype/model"
	"skyra/typesys-prototype/tree"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, drive.NewClaudeModel(), drive.Exec{}))
}

// classes is the -class flag: each use one class, its members by comma.
type classes [][]model.DefLoc

func (c *classes) String() string { return fmt.Sprint(*c) }

func (c *classes) Set(s string) error {
	members := defs(s)
	if len(members) < 2 {
		return errors.New("a class has two members or more")
	}
	*c = append(*c, members)
	return nil
}

func defs(s string) []model.DefLoc {
	var ds []model.DefLoc
	for _, d := range strings.Split(s, ",") {
		if d = strings.TrimSpace(d); d != "" {
			ds = append(ds, model.DefLoc(d))
		}
	}
	return ds
}

// verbose is the model, each ask and the guard it gave written to w.
type verbose struct {
	drive.Model
	w io.Writer
}

func (v verbose) Guard(a model.Ask) (string, error) {
	g, err := v.Model.Guard(a)
	fmt.Fprintf(v.w, "ask at depth %d\n%s=> %s\n\n", a.Depth, drive.Render(a), g)
	return g, err
}

// run is the command, its model and world given: what main runs, and what
// a test runs without the CLI.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer, mdl drive.Model, world drive.World) int {
	fs := flag.NewFlagSet("skyra", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tools := fs.String("tools", "", "the tools the tree may run, by comma (required)")
	readOnly := fs.String("readonly", "", "the tools declared read-only, by comma")
	var cls classes
	fs.Var(&cls, "class", "read-only tools held interchangeable, by comma; repeat for each class")
	act := fs.Int("act", model.DefaultConstants.Act, "k: an agent holding more tool types than this delegates")
	store := fs.String("store", defaultStore(), "the log on disk")
	talk := fs.Bool("v", false, "write each ask and the guard written to stderr")
	show := fs.Bool("show", false, "list what the store has learned")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	program := func() model.Program {
		c := model.DefaultConstants
		c.Act = *act
		var names []string
		for _, d := range defs(*tools) {
			names = append(names, string(d))
		}
		p := drive.Program(c, names)
		p.ReadOnly, p.Classes = defs(*readOnly), cls
		return p
	}
	s := drive.Store{Path: *store}
	if *show {
		mem, err := s.Load()
		if err != nil {
			fmt.Fprintln(stderr, "skyra:", err)
			return 2
		}
		ls, err := drive.Learn(mem)
		if err == nil {
			err = drive.Show(stdout, ls)
		}
		if err != nil {
			fmt.Fprintln(stderr, "skyra:", err)
			return 2
		}
		return 0
	}
	if *tools == "" {
		fmt.Fprintln(stderr, "skyra: -tools is required")
		return 2
	}
	goal := strings.Join(fs.Args(), " ")
	if goal == "" {
		b, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintln(stderr, "skyra:", err)
			return 2
		}
		goal = strings.TrimSuffix(string(b), "\n")
	}
	if goal == "" {
		fmt.Fprintln(stderr, "skyra: no goal: give it as arguments or on stdin")
		return 2
	}

	p := program()
	if *talk {
		mdl = verbose{Model: mdl, w: stderr}
	}

	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		fmt.Fprintln(stderr, "skyra:", err)
		return 2
	}
	start, err := s.Load()
	if err != nil {
		fmt.Fprintln(stderr, "skyra:", err)
		return 2
	}
	m, err := model.NewMachine(drive.Ids{}, p, start)
	if err != nil {
		fmt.Fprintln(stderr, "skyra:", err)
		return 2
	}
	m, v, runErr := drive.Run(m, []byte(goal), mdl, world)
	// What was learned before an error is kept: each event was whole.
	if err := s.Record(start, m); err != nil {
		fmt.Fprintln(stderr, "skyra: the store:", err)
		return 2
	}
	if runErr != nil {
		fmt.Fprintln(stderr, "skyra:", runErr)
		return 2
	}
	return report(v, stdout, stderr)
}

// report prints what the tree gave: a return's stdout, and its stderr when
// it names one, exiting with its status; or a bound's error, exiting 2.
func report(v any, stdout, stderr io.Writer) int {
	r, _ := v.(map[string]any)
	if e, ok := r["error"]; ok {
		fmt.Fprintln(stderr, "skyra:", drive.Text(e))
		return 2
	}
	out := drive.Text(r["stdout"])
	fmt.Fprint(stdout, out)
	if !strings.HasSuffix(out, "\n") && out != "" {
		fmt.Fprintln(stdout)
	}
	if why, ok := r["stderr"]; ok {
		fmt.Fprintln(stderr, drive.Text(why))
	}
	exit, _ := r["exit"].(tree.ExitStatus)
	return int(exit)
}

// defaultStore is ~/.skyra/log.jsonl, or one in the working directory
// without a home.
func defaultStore() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "skyra.log.jsonl"
	}
	return filepath.Join(home, ".skyra", "log.jsonl")
}
