package dnav3

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// The world. A sandboxed filesystem, and that is a deliberate choice.
//
// v2 put a learning substrate in a 20x20 maze, where the only thing reality could ever say was
// "blocked." A filesystem says: this file exists. It contains that. The build exited 1. The test
// failed. Free, instant, unambiguous, and UNGAMEABLE — which is precisely what the grades need and
// precisely what the maze could not give.
//
// The world REPORTS. It does not GRADE. It has never seen the being's prediction and it must not:
// its independence is the entire reason its evidence is worth anything.

// Realm is everything the being requires of a world: execute one act and report, and answer a
// targeted perceive. Two methods, because the being touches its world in exactly two places —
// motor (express) and the settler. Anything that can answer an Act with an Outcome can host the
// being; the sandboxed filesystem below is one Realm, not the definition of one.
type Realm interface {
	Do(Act) Outcome
	Perceive(Triple) Outcome
}

// Outcome is what a world worker returns. Statuses are kept distinct on purpose — "blocked" and
// "denied" and "unavailable" are three different lessons, and collapsing them into "it didn't work"
// throws away most of the signal. "superseded" means an independently observed world event made
// the rest of the current expression obsolete; its Effects carry the observation that did so.
type Outcome struct {
	Status  string // succeeded | blocked | denied | unavailable | indeterminate | superseded
	Effects []Triple
	Err     string
}

// World is a sandbox rooted at Root. Nothing outside Root is readable or writable — the being can
// be wrong in here without being dangerous.
type World struct {
	Root string
}

func NewWorld(root string) *World { return &World{Root: root} }

// Do executes one act and REPORTS WHAT HAPPENED. It never inspects an expect and never returns a
// judgment. Motor calls this; motor does not think.
func (w *World) Do(a Act) Outcome {
	switch a.Primitive {
	case "write":
		return w.write(a)
	case "exec":
		return w.exec(a)
	case "build":
		return w.build(a)
	case "perceive":
		return w.Perceive(Triple{Subject: a.Target, Predicate: attrOf(a.Payload)})
	}
	return Outcome{Status: "unavailable", Err: fmt.Sprintf("no primitive %q here", a.Primitive)}
}

// build has a HIDDEN PREREQUISITE: it succeeds only if BOTH the source and the config exist. This is
// the thing a cold plan forgets and precedent remembers — the forcing function for the retrieval
// curve. The world does not lie about it: a failed build reports WHAT was missing, so the being can
// learn the prerequisite from a real observation rather than being told.
func (w *World) build(a Act) Outcome {
	// A THREE-PART RECIPE the being is never told. The build reports the FIRST missing input, so the
	// being discovers the recipe one prerequisite at a time — and precedent lets it front-load all
	// three next time.
	status, missing := "ok", ""
	for _, need := range []string{"app.src", "app.conf", "app.lock"} {
		path, _ := w.resolve(need)
		if _, err := os.Stat(path); err != nil {
			status, missing = "failed", need
			break
		}
	}
	return Outcome{Status: "succeeded", Effects: []Triple{
		{Subject: "build:/", Predicate: "status", Object: status},
		{Subject: "build:/", Predicate: "missing", Object: missing},
	}}
}

// Perceive reads world state at (subject, predicate) and reports it.
//
// THIS IS WHAT MAKES ABSENCE DECIDABLE. A targeted perceive comes back authoritative for exactly
// the triple that was bet on, so within it, absence IS falsity — and the general closed-world
// problem never has to be solved. The bet manufactures its own evidence.
func (w *World) Perceive(t Triple) Outcome {
	path, err := w.resolve(t.Subject)
	if err != nil {
		return Outcome{Status: "denied", Err: err.Error()}
	}
	switch t.Predicate {
	case "exists":
		_, statErr := os.Stat(path)
		return Outcome{Status: "succeeded", Effects: []Triple{
			{Subject: t.Subject, Predicate: "exists", Object: boolStr(statErr == nil)},
		}}
	case "status":
		// perceiving a build RE-RUNS it: build status is a live fact, not a stored one.
		if strings.HasPrefix(t.Subject, "build:") {
			return w.build(Act{Target: t.Subject})
		}
	}
	// A PARAMETERIZED PREDICATE does not fit subject-predicate-object: in
	// `file:/x contains "func main"`, the object is the ARGUMENT, not the value. The first live run
	// proved this the hard way — Perceive returned `contains: true`, the grader compared it against
	// `"func main"`, and the bet was structurally guaranteed to be false forever (P=0.12).
	//
	// So the argument folds INTO the predicate, making each parameterization its own addressable
	// world state:
	//
	//	file:/x  contains:"func main"  →  true | false
	if arg, ok := strings.CutPrefix(t.Predicate, "contains:"); ok {
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			// It does not exist, so it contains nothing. That is a fact, and it is the fact the
			// bet needed.
			return Outcome{Status: "succeeded", Effects: []Triple{
				{Subject: t.Subject, Predicate: t.Predicate, Object: "false"},
			}}
		}
		return Outcome{Status: "succeeded", Effects: []Triple{
			{Subject: t.Subject, Predicate: t.Predicate, Object: boolStr(strings.Contains(string(b), arg))},
		}}
	}
	return Outcome{Status: "unavailable", Err: fmt.Sprintf("cannot perceive %q", t.Predicate)}
}

func (w *World) write(a Act) Outcome {
	path, err := w.resolve(a.Target)
	if err != nil {
		return Outcome{Status: "denied", Err: err.Error()}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Outcome{Status: "blocked", Err: err.Error()}
	}
	if err := os.WriteFile(path, []byte(a.Payload), 0o644); err != nil {
		return Outcome{Status: "blocked", Err: err.Error()}
	}
	// The world reports what it committed. Note it does NOT report "exists: true" — it reports
	// that it wrote. Whether the file exists is something the being must go and PERCEIVE, which is
	// LAW 1 doing its job: expression cannot write world state.
	return Outcome{Status: "succeeded", Effects: []Triple{
		{Subject: a.Target, Predicate: "written", Object: "true"},
	}}
}

func (w *World) exec(a Act) Outcome {
	dir, err := w.resolve(strings.TrimPrefix(a.Target, "dir:"))
	if err != nil {
		return Outcome{Status: "denied", Err: err.Error()}
	}
	cmd := exec.Command("sh", "-c", a.Payload)
	cmd.Dir = dir
	out, runErr := cmd.CombinedOutput()
	status := "ok"
	if runErr != nil {
		status = "failed"
	}
	return Outcome{Status: "succeeded", Effects: []Triple{
		{Subject: a.Target, Predicate: "status", Object: status},
		{Subject: a.Target, Predicate: "output", Object: strings.TrimSpace(string(out))},
	}}
}

// resolve maps a world-addressable ID onto a real path inside the sandbox, and refuses to leave it.
//
// World-given identity (file:, dir:) is what makes coreference free — the world adjudicates it, so
// no model is needed to decide whether two names refer to the same thing.
func (w *World) resolve(id string) (string, error) {
	rel := id
	for _, ns := range []string{"file:", "dir:", "build:", "test:"} {
		rel = strings.TrimPrefix(rel, ns)
	}
	rel = strings.TrimPrefix(rel, "/")
	path := filepath.Join(w.Root, rel)
	clean, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	root, err := filepath.Abs(w.Root)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(clean, root) {
		return "", fmt.Errorf("%q escapes the sandbox", id)
	}
	return clean, nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// attrOf lets a perceive act name which predicate it wants; defaults to existence.
func attrOf(payload string) string {
	if p := strings.TrimSpace(payload); p != "" {
		return p
	}
	return "exists"
}
