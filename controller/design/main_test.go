package main

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestDraftCannotLeaveDesign(t *testing.T) {
	scratch := t.TempDir()
	outside := t.TempDir()
	for _, path := range []string{"../escape.md", "/tmp/escape.md", "."} {
		if _, err := draft(scratch, request{Path: path, Text: "x"}); err == nil {
			t.Fatalf("accepted %q", path)
		}
	}
	if _, err := draft(scratch, request{Path: "review/spec.md", Text: "# Design"}); err != nil {
		t.Fatal(err)
	}
	os.Symlink(outside, filepath.Join(scratch, "design", "escape"))
	if _, err := draft(scratch, request{Path: "escape/no.md", Text: "x"}); err == nil {
		t.Fatal("symlink escaped")
	}
	os.WriteFile(filepath.Join(outside, "keep.md"), []byte("keep"), 0600)
	os.Symlink(filepath.Join(outside, "keep.md"), filepath.Join(scratch, "design", "alias.md"))
	if _, err := draft(scratch, request{Path: "alias.md", Text: "x"}); err == nil {
		t.Fatal("alias overwritten")
	}
	data, _ := os.ReadFile(filepath.Join(outside, "keep.md"))
	if string(data) != "keep" {
		t.Fatal("outside file modified")
	}
}
func TestDraftRejectsSymlinkRoot(t *testing.T) {
	scratch := t.TempDir()
	os.Symlink(t.TempDir(), filepath.Join(scratch, "design"))
	if _, err := draft(scratch, request{Path: "draft.md", Text: "x"}); err == nil {
		t.Fatal("symlink root accepted")
	}
}

func TestMissingAndStaleGrantsHaveTheirOwnErrors(t *testing.T) {
	t.Setenv("SKYRA_CONNECTIONS", "[]")
	_, err := run(request{Operation: "context"})
	got := errorResult(err, "context")["error"].(*failure)
	if got.Code != "no_bridge_grant" || got.Stage != "bridge" {
		t.Fatalf("%+v", got)
	}
	t.Setenv("SKYRA_CONNECTIONS", `["/private/tmp/skyra-penpot-nonexistent/penpot.sock"]`)
	_, err = run(request{Operation: "context"})
	got = errorResult(err, "context")["error"].(*failure)
	if got.Code != "bridge_unavailable" {
		t.Fatalf("%+v", got)
	}
}

func TestInvalidBridgeResponseMarksAnEditUncertain(t *testing.T) {
	dir, err := os.MkdirTemp("", "skyra-penpot-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "penpot.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	paths, _ := json.Marshal([]string{path})
	t.Setenv("SKYRA_CONNECTIONS", string(paths))
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		var r request
		json.NewDecoder(conn).Decode(&r)
		conn.Write([]byte(`{"unexpected":"response"}`))
	}()
	_, err = run(request{Operation: "execute", Code: "return 1;"})
	got := errorResult(err, "execute")["error"].(*failure)
	if got.Code != "invalid_bridge_response" || got.EditOutcome != "unknown" {
		t.Fatalf("%+v", got)
	}
}
