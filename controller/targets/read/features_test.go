package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBatchAroundNumberingAndOutlines(t *testing.T) {
	root := t.TempDir()
	write(t, root, "a.go", "package a\n\ntype Item struct{}\n\nfunc Run() {\n}\n")
	write(t, root, "b.md", "# Start\ntext\n## Next\nbody\n")
	around := input{Path: "a.go", Numbered: true}
	around.Around = &struct {
		Line    int `json:"line"`
		Context int `json:"context"`
	}{5, 1}
	r := observe(root, "", around)
	if *r.Text != "4: \n5: func Run() {\n6: }\n" {
		t.Fatal(r)
	}
	r = observe(root, "", input{Path: "a.go", Outline: true})
	if len(r.Outline) != 2 || r.Outline[1].Name != "Run" || r.Outline[1].StartLine != 5 || r.Outline[1].EndLine != 6 || r.Text != nil {
		t.Fatal(r)
	}
	r = observe(root, "", input{Path: "b.md", Outline: true})
	if len(r.Outline) != 2 || r.Outline[0].EndLine != 2 {
		t.Fatal(r)
	}
	r = observeMany(root, "", input{Paths: []json.RawMessage{json.RawMessage(`"b.md"`), json.RawMessage(`{"path":"a.go","start_line":3,"end_line":3}`), json.RawMessage(`"missing"`)}})
	if len(r.Results) != 3 || *r.Results[1].Text != "type Item struct{}\n" || r.Results[2].Error != "not_found" {
		t.Fatal(r)
	}
	if r := observe(root, "", input{Path: "a.go", StartLine: num(4), EndLine: num(2)}); !strings.HasPrefix(r.Error, "bad_range") {
		t.Fatal(r)
	}
}
func TestBatchBudgetDoesNotMultiplyReadLimit(t *testing.T) {
	root := t.TempDir()
	write(t, root, "large", strings.Repeat(strings.Repeat("a", 99)+"\n", 2000))
	r := observeMany(root, "", input{Paths: []json.RawMessage{json.RawMessage(`"large"`), json.RawMessage(`"large"`)}})
	total := 0
	for _, f := range r.Results {
		if f.Text != nil {
			total += len(*f.Text)
		}
	}
	encoded, _ := json.Marshal(r)
	if len(encoded) > maxText {
		t.Fatal(len(encoded))
	}
	if total > maxText {
		t.Fatal(total)
	}
}
