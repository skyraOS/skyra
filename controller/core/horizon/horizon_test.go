package horizon

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
	h := Horizon{Surface: &Surface{Path: "/private/bin/tool"}, Authority: Authority{
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

// Everything under one Horizon gets the same grants; only the program being
// run changes, and it is the only thing exec is allowed for (next/24,
// "Exclusions").
func TestOneProfilePerHorizonExecOnlyTheRunningProgram(t *testing.T) {
	grants := Authority{Read: []PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	a := Render(Horizon{Surface: &Surface{Path: "/private/bin/a"}, Authority: grants}, "/private/work/.skyra")
	b := Render(Horizon{Surface: &Surface{Path: "/private/bin/b"}, Authority: grants}, "/private/work/.skyra")
	if strings.ReplaceAll(a, "/private/bin/a", "X") != strings.ReplaceAll(b, "/private/bin/b", "X") {
		t.Fatalf("two programs under one Horizon must get the same profile:\n%s\n%s", a, b)
	}
	if n := strings.Count(a, "(allow process-exec"); n != 1 || !strings.Contains(a, `(allow process-exec (literal "/private/bin/a"))`) {
		t.Fatalf("exec must be allowed for the running program only:\n%s", a)
	}
}

// A work tool's contract restated in another order is the tool it was bound
// as (42).
func TestPermitsAToolRestatedInAnotherOrder(t *testing.T) {
	bound := &contract.Schema{Type: "object", Required: []string{"a", "b"}, Fields: map[string]*contract.Schema{"a": {Type: "string"}, "b": {Type: "string"}}}
	restated := &contract.Schema{Type: "object", Required: []string{"b", "a"}, Fields: bound.Fields}
	other := &contract.Schema{Type: "object", Required: []string{"a"}, Fields: bound.Fields}
	tools := map[string]ToolBinding{"t": {Target: "echo", Contracts: []*contract.Schema{bound}}}
	h := Horizon{WorkTools: &tools}
	if !h.PermitsTool("echo", restated) || h.PermitsTool("echo", other) {
		t.Fatal("a restated contract is the bound tool; a different one is not")
	}
}
