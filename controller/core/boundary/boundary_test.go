package boundary

import (
	"skyra/controller/core/contract"
	"strings"
	"testing"
)

func TestAuthorityAllowsByPathComponent(t *testing.T) {
	ceiling := Authority{Read: []PathGrant{{Path: "/private/work", Scope: "subpath"}}, Mach: []string{"service.allowed"}}
	if !ceiling.Allows(Authority{Read: []PathGrant{{Path: "/private/work/file", Scope: "literal"}}, Mach: []string{"service.allowed"}}) {
		t.Fatal("a child file and named service must fit the ceiling")
	}
	for _, requested := range []Authority{
		{Read: []PathGrant{{Path: "/private/workspace/file", Scope: "literal"}}},
		{Read: []PathGrant{{Path: "/private", Scope: "subpath"}}},
		{Mach: []string{"service.other"}},
		{Read: []PathGrant{{Path: "/private/work", Scope: "bogus"}}},
	} {
		if ceiling.Allows(requested) {
			t.Fatalf("authority escaped its ceiling: %+v", requested)
		}
	}
}

func TestRenderOnlyDeclaredResources(t *testing.T) {
	h := Boundary{Surface: &Surface{Path: "/private/bin/tool"}, Authority: Authority{
		Read:  []PathGrant{{Path: "/private/work/input", Scope: "literal"}},
		Write: []PathGrant{{Path: "/private/work/output", Scope: "literal"}},
		Mach:  []string{"service.allowed"},
	}}
	profile := Render(h, "/private/work/.skyra")
	for _, line := range []string{
		`(allow file-read* (literal "/private/work/input"))`,
		`(allow file-write* (literal "/private/work/output"))`,
		`(allow mach-lookup (global-name "service.allowed"))`,
		`(deny file-read* file-write* (subpath "/private/work/.skyra"))`,
	} {
		if !strings.Contains(profile, line) {
			t.Fatalf("missing %s in profile:\n%s", line, profile)
		}
	}
	for _, line := range []string{"(allow file-read*)", "(allow mach-lookup)", `(allow network-outbound (remote ip "*:*"))`} {
		if strings.Contains(profile, line) {
			t.Fatalf("blanket grant %s in profile:\n%s", line, profile)
		}
	}
}

// Everything under one Boundary gets the same grants; only the program being
// run changes, and it is the only thing exec is allowed for (next/24,
// "Exclusions").
func TestOneProfilePerBoundaryExecOnlyTheRunningProgram(t *testing.T) {
	grants := Authority{Read: []PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	ha, hb := strings.Repeat("a", 64), strings.Repeat("b", 64)
	a := Render(Boundary{Surface: &Surface{Path: "/private/bin/a", SHA256: ha}, Authority: grants}, "/private/work/.skyra")
	b := Render(Boundary{Surface: &Surface{Path: "/private/bin/b", SHA256: hb}, Authority: grants}, "/private/work/.skyra")
	if strings.ReplaceAll(a, ha, "X") != strings.ReplaceAll(b, hb, "X") {
		t.Fatalf("two programs under one Boundary must get the same profile:\n%s\n%s", a, b)
	}
	// What runs is the runtime's verified copy, never the registered path (43 D5).
	if n := strings.Count(a, "(allow process-exec"); n != 1 || !strings.Contains(a, `(allow process-exec (literal "/private/work/.skyra/surfaces/`+ha+`"))`) || strings.Contains(a, "/private/bin/a") {
		t.Fatalf("exec must be allowed for the running program's verified copy only:\n%s", a)
	}
	// A surface without a well-formed hash is allowed to run nothing.
	for _, sum := range []string{"", "../../x", strings.Repeat("A", 64)} {
		if p := Render(Boundary{Surface: &Surface{Path: "/private/bin/a", SHA256: sum}, Authority: grants}, "/private/work/.skyra"); strings.Contains(p, "process-exec") {
			t.Fatalf("hash %q: exec allowed:\n%s", sum, p)
		}
	}
}

// A contract restated in another order is the contract held (42).
func TestHoldsAContractRestatedInAnotherOrder(t *testing.T) {
	bound := &contract.Schema{Type: "object", Required: []string{"a", "b"}, Fields: map[string]*contract.Schema{"a": {Type: "string"}, "b": {Type: "string"}}}
	restated := &contract.Schema{Type: "object", Required: []string{"b", "a"}, Fields: bound.Fields}
	other := &contract.Schema{Type: "object", Required: []string{"a"}, Fields: bound.Fields}
	held := CList{"echo": {bound}}
	h := Boundary{Reach: map[string]Surface{"echo": {Input: bound, Contracts: []*contract.Schema{other}}}, Held: &held}
	if !h.Holds("echo", restated) || h.Holds("echo", other) {
		t.Fatal("a restated contract is the held one; a different one is not")
	}
}
