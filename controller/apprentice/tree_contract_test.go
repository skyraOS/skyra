package main

import (
	"encoding/json"
	"testing"

	"skyra/apprentice/proc"
	"skyra/apprentice/shell"
)

// The tree contract the kernel checks a tree's input against is the one the
// shell reads: the profile a root sets is in both.
func TestTheDeclaredTreeContractIsTheShells(t *testing.T) {
	var d struct{ Contracts []*proc.Schema }
	if err := json.Unmarshal(declared, &d); err != nil {
		t.Fatal(err)
	}
	for _, c := range d.Contracts {
		if c.Title != "tree" {
			continue
		}
		got, _ := json.Marshal(c)
		want, _ := json.Marshal(shell.Contract())
		if string(got) != string(want) {
			t.Fatalf("contract.json's tree contract differs from the shell's:\n%s\n%s", got, want)
		}
		return
	}
	t.Fatal("contract.json declares no tree contract")
}
