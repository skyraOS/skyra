package horizon

import (
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
	h := Horizon{Surface: &Surface{Path: "/private/bin/tool", Authority: Authority{
		Read:  []PathGrant{{Path: "/private/work/input", Scope: "literal"}},
		Write: []PathGrant{{Path: "/private/work/output", Scope: "literal"}},
		Mach:  []string{"service.allowed"},
	}}}
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
