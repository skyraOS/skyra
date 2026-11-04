// The vectors: every case Core.lean writes, replayed against the Go. A file
// with no replay, or a replay with no file, fails, so a change to either
// side alone fails CI (55, "Plan", step 3).
package tree

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const vectorDir = "../invariants/vectors"

type vcase struct {
	In  map[string]json.RawMessage `json:"in"`
	Out json.RawMessage            `json:"out"`
}

func loadVectors(t *testing.T, name string) []vcase {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(vectorDir, name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Function string  `json:"function"`
		Cases    []vcase `json:"cases"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if f.Function != name || len(f.Cases) == 0 {
		t.Fatalf("%s: function %q, %d cases", name, f.Function, len(f.Cases))
	}
	return f.Cases
}

// --- decoding Lean's inputs ---

func dec[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("decoding %s: %v", raw, err)
	}
	return v
}

type jContract struct {
	Binary Binary `json:"binary"`
	Shape  Schema `json:"shape"`
}

type jType struct {
	Contract jContract `json:"contract"`
	Output   Schema    `json:"output"`
}

type jObs struct {
	Type    jType  `json:"type"`
	Members []jObs `json:"members"`
}

type jStore struct {
	Known  []BaseID `json:"known"`
	Guards []struct {
		Base  BaseID          `json:"base"`
		Guard json.RawMessage `json:"guard"`
	} `json:"guards"`
}

// Contracts and types are rebuilt through their constructors: every one
// Lean writes passed them.
func (j jContract) get(t *testing.T) Contract {
	t.Helper()
	c, ok := Identify(j.Binary, j.Shape)
	if !ok {
		t.Fatalf("a vector's contract is refused: %+v", j)
	}
	return c
}

func (j jType) get(t *testing.T) ObservationType {
	t.Helper()
	o, ok := TypeOf(j.Contract.get(t), j.Output)
	if !ok {
		t.Fatalf("a vector's type is refused: %+v", j)
	}
	return o
}

// Observations are rebuilt as Lean wrote them, made in the package: every
// one Lean writes was made by observe.
func (j jObs) get(t *testing.T) Observation {
	t.Helper()
	o := Observation{typ: j.Type.get(t)}
	for _, m := range j.Members {
		o.members = append(o.members, m.get(t))
	}
	return o
}

func contractsOf(t *testing.T, raw json.RawMessage) []Contract {
	t.Helper()
	var out []Contract
	for _, j := range dec[[]jContract](t, raw) {
		out = append(out, j.get(t))
	}
	return out
}

func obsOf(t *testing.T, raw json.RawMessage) []Observation {
	t.Helper()
	var out []Observation
	for _, j := range dec[[]jObs](t, raw) {
		out = append(out, j.get(t))
	}
	return out
}

func kindOf(t *testing.T, raw json.RawMessage) Kind {
	t.Helper()
	for _, k := range []Kind{KindEffect, KindInvoke, KindDelegate} {
		if k.String() == dec[string](t, raw) {
			return k
		}
	}
	t.Fatalf("no kind %s", raw)
	return 0
}

func storeOf(t *testing.T, raw json.RawMessage) Store {
	t.Helper()
	j := dec[jStore](t, raw)
	s := Store{known: j.Known}
	for _, g := range j.Guards {
		var guard Guard
		if string(g.Guard) != `"default"` {
			guard = WrittenGuard(dec[struct {
				Written string `json:"written"`
			}](t, g.Guard).Written)
		}
		s = s.set(g.Base, guard)
	}
	return s
}

// --- encoding the Go's answers as Lean writes them ---

var domain = []BaseID{"base:x", "base:y", "base:z"}

func eContract(c Contract) any { return map[string]any{"binary": c.binary, "shape": c.shape} }

func eType(o ObservationType) any {
	return map[string]any{"contract": eContract(o.contract), "output": o.output}
}

func eObs(o Observation) any {
	return map[string]any{"type": eType(o.typ), "members": eList(o.members, eObs)}
}

func eList[T any](xs []T, f func(T) any) any {
	out := []any{}
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func eOpt[T any](v T, ok bool, f func(T) any) any {
	if !ok {
		return nil
	}
	return f(v)
}

func eApprentice(a Apprentice) any {
	return map[string]any{"binary": a.binary, "observe": eContract(a.observe), "root": eContract(a.root),
		"exit": eContract(a.exit), "return": eContract(a.ret),
		"effect_observation": eContract(a.effectObservation), "invoke": eContract(a.invoke),
		"delegate": eContract(a.delegate), "goal": eContract(a.goal)}
}

func eStore(s Store) any {
	guards := []any{}
	for _, b := range domain {
		if g, ok := s.Guard(b); ok {
			var jg any = "default"
			if cel, w := g.Written(); w {
				jg = map[string]any{"written": cel}
			}
			guards = append(guards, map[string]any{"base": b, "guard": jg})
		}
	}
	return map[string]any{"known": eList(s.known, func(b BaseID) any { return b }), "guards": guards}
}

func toy(b BaseID, g Guard) RefinementID {
	if cel, w := g.Written(); w {
		return RefinementID("refine(" + string(b) + ",written:" + cel + ")")
	}
	return RefinementID("refine(" + string(b) + ",default)")
}

func pick[T comparable](t *testing.T, raw json.RawMessage, names map[string]T) T {
	t.Helper()
	v, ok := names[dec[string](t, raw)]
	if !ok {
		t.Fatalf("no value %s", raw)
	}
	return v
}

var (
	seens   = map[string]Seen{"seen": SeenSeen, "unseen": SeenUnseen}
	inputs  = map[string]InputBaseNovelty{"seen": InputSeen, "unseen": InputUnseen}
	nexts   = map[Next]string{NextContinue: "continue", NextFallback: "fallback"}
	enters  = map[Enter]string{EnterProcedure: "procedure", EnterModel: "model"}
	dirs    = map[string]Dir{"invoke": DirInvoke, "ret": DirRet}
	seenStr = map[Seen]string{SeenSeen: "seen", SeenUnseen: "unseen"}
	inStr   = map[InputBaseNovelty]string{InputSeen: "seen", InputUnseen: "unseen"}
)

// The apprentice every case uses: apprenticeOf.json's first case.
func vectorApprentice(t *testing.T) Apprentice {
	t.Helper()
	c := loadVectors(t, "apprenticeOf")[0]
	cs := contractsOf(t, c.In["contracts"])
	a, ok := ApprenticeOf(dec[Binary](t, c.In["binary"]), cs[0], cs[1], cs[2], cs[3], cs[4], cs[5], cs[6], cs[7])
	if !ok {
		t.Fatal("the vectors' apprentice is refused")
	}
	return a
}

type replay func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any

func inputBases(t *testing.T, raw json.RawMessage) []InputBase {
	t.Helper()
	var out []InputBase
	for _, j := range dec[[]struct {
		Binary       Binary  `json:"binary"`
		Observations []jType `json:"observations"`
	}](t, raw) {
		i := InputBase{Binary: j.Binary}
		for _, o := range j.Observations {
			i.Observations = append(i.Observations, o.get(t))
		}
		out = append(out, i)
	}
	return out
}

var replays = map[string]replay{
	"identify": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		c, ok := Identify(dec[Binary](t, in["binary"]), dec[Schema](t, in["shape"]))
		return eOpt(c, ok, eContract)
	},
	"observationTypeOf": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		o, ok := TypeOf(dec[jContract](t, in["contract"]).get(t), dec[Schema](t, in["output"]))
		return eOpt(o, ok, eType)
	},
	"apprenticeOf": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		cs := contractsOf(t, in["contracts"])
		a, ok := ApprenticeOf(dec[Binary](t, in["binary"]), cs[0], cs[1], cs[2], cs[3], cs[4], cs[5], cs[6], cs[7])
		return eOpt(a, ok, eApprentice)
	},
	"admits": func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any {
		return Admits(a, kindOf(t, in["kind"]), dec[jType](t, in["type"]).get(t), obsOf(t, in["members"]))
	},
	"observe": func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any {
		o, ok := Observe(a, kindOf(t, in["kind"]), dec[jType](t, in["type"]).get(t), obsOf(t, in["members"]))
		return eOpt(o, ok, eObs)
	},
	"rootObservation": func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any {
		o, ok := RootObservation(a, dec[jType](t, in["type"]).get(t))
		return eOpt(o, ok, eObs)
	},
	"mayPass": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return MayPass(obsOf(t, in["sender"]), obsOf(t, in["passed"]))
	},
	"receive": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return eList(Receive(pick(t, in["dir"], dirs), obsOf(t, in["receiver"]), obsOf(t, in["passed"])), eObs)
	},
	"lineage": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return eList(Lineage(dec[jObs](t, in["observation"]).get(t)), eObs)
	},
	"nodeMayReturn": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return NodeMayReturn(obsOf(t, in["received"]), obsOf(t, in["returned"]))
	},
	"mayReturn": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return MayReturn(kindOf(t, in["kind"]), obsOf(t, in["created"]), obsOf(t, in["returned"]))
	},
	"classify": func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any {
		k, ok := Classify(a, dec[jContract](t, in["self"]).get(t), contractsOf(t, in["reach"]))
		return eOpt(k, ok, func(k Kind) any { return k.String() })
	},
	"loaded": func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any {
		return eList(Loaded(a, kindOf(t, in["kind"])), eContract)
	},
	"passable": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return Passable(contractsOf(t, in["held"]), contractsOf(t, in["passed"]))
	},
	"offered": func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any {
		o := Offered(a, kindOf(t, in["kind"]), contractsOf(t, in["passed"]))
		return map[string]any{"invoke": eList(o.Invoke, eContract), "passOnly": eList(o.PassOnly, eContract),
			"apprentice": eList(o.Apprentice, eContract)}
	},
	"inputBaseNovelty": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		i := inputBases(t, json.RawMessage("["+string(in["input"])+"]"))[0]
		return inStr[InputNovelty(inputBases(t, in["history"]), i)]
	},
	"see": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		n, s := See(storeOf(t, in["store"]), dec[BaseID](t, in["base"]))
		return map[string]any{"seen": seenStr[n], "store": eStore(s)}
	},
	"rewrite": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		s, ok := Rewrite(storeOf(t, in["store"]), dec[BaseID](t, in["base"]), dec[string](t, in["cel"]))
		return eOpt(s, ok, eStore)
	},
	"refinement": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		r, ok := Refinement(toy, storeOf(t, in["store"]), dec[BaseID](t, in["base"]))
		return eOpt(r, ok, func(r RefinementID) any { return r })
	},
	"next": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return nexts[NextOf(pick(t, in["seen"], seens), dec[bool](t, in["guardHolds"]))]
	},
	"enter": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return enters[EnterOf(pick(t, in["novelty"], inputs), dec[bool](t, in["guardHolds"]))]
	},
}

// normal is a value as JSON decodes it, so Lean's answer and the Go's
// compare by content.
func normal(t *testing.T, v any) any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestVectors(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(vectorDir, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no vectors in %s: %v", vectorDir, err)
	}
	var names []string
	for _, f := range files {
		names = append(names, strings.TrimSuffix(filepath.Base(f), ".json"))
	}
	for name := range replays {
		if !slices.Contains(names, name) {
			t.Errorf("a replay for %s, and no vectors", name)
		}
	}
	a := vectorApprentice(t)
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			r, ok := replays[name]
			if !ok {
				t.Fatalf("vectors for %s, and no replay", name)
			}
			for i, c := range loadVectors(t, name) {
				got, want := normal(t, r(t, a, c.In)), normal(t, c.Out)
				if !reflect.DeepEqual(got, want) {
					in, _ := json.Marshal(c.In)
					t.Fatalf("case %d: %s\n got  %v\n want %v", i, in, got, want)
				}
			}
		})
	}
}

// The core touches no OS, tree or network: its imports are an allowlist.
func TestCoreIsPure(t *testing.T) {
	allowed := map[string]bool{"slices": true, "strings": true}
	files, _ := filepath.Glob("*.go")
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		pf, err := parser.ParseFile(token.NewFileSet(), f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range pf.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if !allowed[p] {
				t.Errorf("%s imports %q: package tree must not touch the OS, the tree or the network", f, p)
			}
		}
	}
}
