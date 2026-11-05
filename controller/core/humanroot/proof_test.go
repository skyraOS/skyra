package humanroot

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDecisionsAgreeWithLean(t *testing.T) {
	raw, err := os.ReadFile("testdata/lean.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		Sets    [][]string          `json:"sets"`
		Entries [][]Capability      `json:"entries"`
		Targets []Capability        `json:"targets"`
		Author  [][]json.RawMessage `json:"mayAuthor"`
		Pass    [][]json.RawMessage `json:"mayPass"`
		Grant   [][]json.RawMessage `json:"mayGrant"`
		Invoke  [][]json.RawMessage `json:"mayInvoke"`
	}
	if err = json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	index := func(row []json.RawMessage, i int) int {
		var n int
		if err := json.Unmarshal(row[i], &n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	answer := func(row []json.RawMessage) bool {
		var b bool
		if err := json.Unmarshal(row[len(row)-1], &b); err != nil {
			t.Fatal(err)
		}
		return b
	}
	ids := func(i int) []ID {
		var out []ID
		for _, s := range vectors.Sets[i] {
			out = append(out, ID(s))
		}
		return out
	}
	caps := func(i int) []Capability {
		var out []Capability
		for _, s := range vectors.Sets[i] {
			out = append(out, Capability(s))
		}
		return out
	}
	for _, r := range vectors.Author {
		if MayAuthor(Ingress(index(r, 0))) != answer(r) {
			t.Fatal("Lean mayAuthor disagreement")
		}
	}
	for _, r := range vectors.Pass {
		if MayPass(ids(index(r, 0)), ids(index(r, 1))) != answer(r) {
			t.Fatal("Lean mayPass disagreement")
		}
	}
	for _, r := range vectors.Grant {
		if MayGrant(caps(index(r, 0)), caps(index(r, 1))) != answer(r) {
			t.Fatal("Lean mayGrant disagreement")
		}
	}
	for _, r := range vectors.Invoke {
		got := MayInvoke(vectors.Entries[index(r, 0)], vectors.Targets[index(r, 1)], ids(index(r, 2)), ids(index(r, 3)), caps(index(r, 4)), caps(index(r, 5)))
		if got != answer(r) {
			t.Fatalf("Lean mayInvoke disagreement: %s", r)
		}
	}
	if len(vectors.Invoke) != 15552 || len(vectors.Author) != 4 || len(vectors.Pass) != 36 || len(vectors.Grant) != 36 {
		t.Fatal("Lean coverage changed unexpectedly")
	}
}

func TestRootCoreRemainsPureAndIndependentOfWorkerTree(t *testing.T) {
	allowed := map[string]bool{"bytes": true, "crypto/sha256": true, "encoding/hex": true, "encoding/json": true, "errors": true, "slices": true}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		ast, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range ast.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if !allowed[path] {
				t.Errorf("pure human root must not import %q", path)
			}
		}
	}
}
