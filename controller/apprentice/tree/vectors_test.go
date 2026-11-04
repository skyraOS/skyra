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
	Known  []string `json:"known"`
	Guards []struct {
		Base  string          `json:"base"`
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
	for _, k := range []Kind{KindAct, KindDelegate} {
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
	s := Store{}
	for _, k := range j.Known {
		s.known = append(s.known, ReturnBase{k})
	}
	for _, g := range j.Guards {
		var guard Guard
		if string(g.Guard) != `"default"` {
			guard = WrittenGuard(dec[struct {
				Written string `json:"written"`
			}](t, g.Guard).Written)
		}
		s = s.set(ReturnBase{g.Base}, guard)
	}
	return s
}

// --- encoding the Go's answers as Lean writes them ---

// Lean's toy hashes, written out the same way: Lean's toString of a list is
// "[a, b]".
var toyHashes = hashes{
	input: func(c Contract, passed []ReturnBase) string {
		ps := make([]string, len(passed))
		for i, p := range passed {
			ps[i] = p.id
		}
		return "input(" + string(c.binary) + "|" + string(c.shape) + ",[" + strings.Join(ps, ", ") + "])"
	},
	ret: func(i InputBase, o Schema) string { return "return(" + i.id + "," + string(o) + ")" },
}

var (
	shapeOut = Schema(`{"items":{"type":"number"},"type":"array"}`)
	shapeIn  = Schema(`{"properties":{"path":{"type":"string"}},"type":"object"}`)
)

// domain is Lean's: three return base types, two shapes back from one input
// and one from another, made the same way.
var domain = func() []ReturnBase {
	empty := Schema(`{"type":"object"}`)
	iA := createInvoke(toyHashes, contract(Binary("cdhash:"+strings.Repeat("a", 40)), empty), nil)
	iB := createInvoke(toyHashes, contract(Binary("sha256:"+strings.Repeat("1", 64)), empty), nil)
	return []ReturnBase{must(createReturn(toyHashes, iA, shapeOut)),
		must(createReturn(toyHashes, iA, shapeIn)), must(createReturn(toyHashes, iB, shapeOut))}
}()

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
		"perform_effect": eContract(a.performEffect), "create_input": eContract(a.createInput),
		"delegate": eContract(a.delegate), "goal": eContract(a.goal),
		"create_delegation": eContract(a.createDelegation), "create_act": eContract(a.createAct),
		"package_reach": eContract(a.reach)}
}

func eStore(s Store) any {
	guards := []any{}
	for _, b := range domain {
		if g, ok := s.Guard(b); ok {
			var jg any = "default"
			if cel, w := g.Written(); w {
				jg = map[string]any{"written": cel}
			}
			guards = append(guards, map[string]any{"base": b.id, "guard": jg})
		}
	}
	return map[string]any{"known": eList(s.known, func(b ReturnBase) any { return b.id }), "guards": guards}
}

func toy(b ReturnBase, g Guard) RefinementID {
	if cel, w := g.Written(); w {
		return RefinementID("refine(" + b.id + ",written:" + cel + ")")
	}
	return RefinementID("refine(" + b.id + ",default)")
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
	returns = map[ReturnResult]string{ReturnContinue: "continue", ReturnFallback: "fallback"}
	invokes = map[InvokeResult]string{InvokeProcedure: "procedure", InvokeModel: "model"}
	dirs    = map[string]Dir{"invoke": DirInvoke, "ret": DirRet}
	seenStr = map[Seen]string{SeenSeen: "seen", SeenUnseen: "unseen"}
)

// The apprentice every case uses: apprenticeOf.json's first case.
func vectorApprentice(t *testing.T) Apprentice {
	t.Helper()
	c := loadVectors(t, "apprenticeOf")[0]
	cs := contractsOf(t, c.In["contracts"])
	a, ok := ApprenticeOf(dec[Binary](t, c.In["binary"]), cs[0], cs[1], cs[2], cs[3], cs[4], cs[5], cs[6], cs[7], cs[8], cs[9], cs[10])
	if !ok {
		t.Fatal("the vectors' apprentice is refused")
	}
	return a
}

type replay func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any

// Input and return base types are rebuilt as Lean wrote them, made in the
// package: every one Lean writes was made by createInvoke or createReturn.
func inputsOf(t *testing.T, raw json.RawMessage) []InputBase {
	t.Helper()
	var out []InputBase
	for _, id := range dec[[]string](t, raw) {
		out = append(out, InputBase{id})
	}
	return out
}

func returnsOf(t *testing.T, raw json.RawMessage) []ReturnBase {
	t.Helper()
	var out []ReturnBase
	for _, id := range dec[[]string](t, raw) {
		out = append(out, ReturnBase{id})
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
		a, ok := ApprenticeOf(dec[Binary](t, in["binary"]), cs[0], cs[1], cs[2], cs[3], cs[4], cs[5], cs[6], cs[7], cs[8], cs[9], cs[10])
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
		return MayReturn(obsOf(t, in["created"]), obsOf(t, in["returned"]))
	},
	"classify": func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any {
		k, ok := Classify(a, dec[jContract](t, in["self"]).get(t), contractsOf(t, in["reach"]))
		return eOpt(k, ok, func(k Kind) any { return k.String() })
	},
	"rootKind": func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any {
		k, ok := RootKind(a, dec[jContract](t, in["self"]).get(t), contractsOf(t, in["reach"]), kindOf(t, in["kind"]))
		return eOpt(k, ok, func(k Kind) any { return k.String() })
	},
	"parts": func(t *testing.T, a Apprentice, in map[string]json.RawMessage) any {
		return eList(Parts(a, kindOf(t, in["kind"])), eContract)
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
	"createInvoke": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return createInvoke(toyHashes, dec[jContract](t, in["contract"]).get(t), returnsOf(t, in["passed"])).id
	},
	"createReturn": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		r, ok := createReturn(toyHashes, InputBase{dec[string](t, in["input"])}, dec[Schema](t, in["shape"]))
		return eOpt(r, ok, func(r ReturnBase) any { return r.id })
	},
	"do": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return seenStr[Do(inputsOf(t, in["history"]), InputBase{dec[string](t, in["input"])})]
	},
	"see": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		n, s := See(storeOf(t, in["store"]), ReturnBase{dec[string](t, in["base"])})
		return map[string]any{"seen": seenStr[n], "store": eStore(s)}
	},
	"rewrite": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		s, ok := Rewrite(storeOf(t, in["store"]), ReturnBase{dec[string](t, in["base"])}, dec[string](t, in["cel"]))
		return eOpt(s, ok, eStore)
	},
	"baseOf": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		r, ok := baseOf(toyHashes, dec[jType](t, in["type"]).get(t), returnsOf(t, in["passed"]))
		return eOpt(r, ok, func(r ReturnBase) any { return r.id })
	},
	"refinement": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		r, ok := Refinement(toy, storeOf(t, in["store"]), ReturnBase{dec[string](t, in["base"])})
		return eOpt(r, ok, func(r RefinementID) any { return r })
	},
	"ret": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return returns[ReturnOf(pick(t, in["seen"], seens), dec[bool](t, in["guardHolds"]))]
	},
	"invoke": func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		return invokes[InvokeOf(pick(t, in["seen"], seens), dec[bool](t, in["guardHolds"]))]
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
	allowed := map[string]bool{"bytes": true, "crypto/sha256": true, "encoding/hex": true,
		"encoding/json": true, "slices": true, "strings": true}
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
