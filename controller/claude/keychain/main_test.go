package main

import (
	"bytes"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const service = "Claude Code-credentials"

// fakeSecurity stands in for /usr/bin/security: it records its arguments and
// stdin, and answers a find with a fixed secret.
func fakeSecurity(t *testing.T) (log string) {
	t.Helper()
	dir := t.TempDir()
	log = filepath.Join(dir, "calls")
	script := filepath.Join(dir, "security")
	body := "#!/bin/sh\necho \"$*\" >> " + log + "\nif [ \"$1\" = -i ]; then cat >> " + log + "; fi\n" +
		"if [ \"$1\" = find-generic-password ]; then echo secret-token; fi\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	old := security
	security = script
	t.Cleanup(func() { security = old })
	return log
}

func calls(t *testing.T, log string) string {
	b, err := os.ReadFile(log)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(b)
}

// direct is a shim wired straight to a broker, without a socket.
func direct(t *testing.T) func(request) (response, error) {
	b, err := newBroker(service)
	if err != nil {
		t.Fatal(err)
	}
	return func(r request) (response, error) { return b.answer(r), nil }
}

func run1(t *testing.T, call func(request) (response, error), stdin string, args ...string) (int, string, string) {
	var out, errOut bytes.Buffer
	code := shim(args, strings.NewReader(stdin), &out, &errOut, call)
	return code, out.String(), errOut.String()
}

func TestShimReadsThePinnedItem(t *testing.T) {
	log := fakeSecurity(t)
	code, out, _ := run1(t, direct(t), "", "find-generic-password", "-a", "anyone", "-w", "-s", service)
	if code != 0 || out != "secret-token\n" {
		t.Fatalf("find: %d %q", code, out)
	}
	u, _ := user.Current()
	if want := "find-generic-password -a " + u.Username + " -w -s " + service; !strings.Contains(calls(t, log), want) {
		t.Fatalf("broker ran %q, want %q", calls(t, log), want)
	}
}

// Any other item is reported missing, and the real tool is never run for it.
func TestShimReachesNoOtherItem(t *testing.T) {
	log := fakeSecurity(t)
	for _, svc := range []string{"Claude Code", "Chrome Safe Storage", service + "-x"} {
		code, out, _ := run1(t, direct(t), "", "find-generic-password", "-a", "me", "-w", "-s", svc)
		if code != notFound || out != "" {
			t.Fatalf("%s: %d %q", svc, code, out)
		}
	}
	code, _, _ := run1(t, direct(t), `add-generic-password -U -a "me" -s "Other" -X "6869"`+"\n", "-i")
	if code == 0 {
		t.Fatal("updated another item")
	}
	if c := calls(t, log); c != "" {
		t.Fatalf("real tool ran: %q", c)
	}
}

// A token refresh writes back through `security -i`, as the CLI sends it.
func TestShimWritesARefreshBack(t *testing.T) {
	log := fakeSecurity(t)
	code, _, _ := run1(t, direct(t), `add-generic-password -U -a "me" -s "`+service+`" -X "7b7d"`+"\n", "-i")
	if code != 0 {
		t.Fatalf("add: %d", code)
	}
	u, _ := user.Current()
	want := `add-generic-password -U -a "` + u.Username + `" -s "` + service + `" -X "7b7d"`
	if c := calls(t, log); !strings.HasPrefix(c, "-i\n") || !strings.Contains(c, want) {
		t.Fatalf("broker ran %q", c)
	}
	code, _, _ = run1(t, direct(t), "", "add-generic-password", "-U", "-a", "me", "-s", service, "-X", "7b7d")
	if code != 0 {
		t.Fatalf("argv add: %d", code)
	}
}

func TestShimRefusesEverythingElse(t *testing.T) {
	log := fakeSecurity(t)
	for _, args := range [][]string{
		{"delete-generic-password", "-a", "me", "-s", service},
		{"find-internet-password", "-s", "github.com"},
		{"dump-keychain"},
		{"find-generic-password", "-a", "me", "-w", "-s", service, "-g"},
		{"add-generic-password", "-U", "-a", "me", "-s", service, "-w", "plaintext"},
		{"add-generic-password", "-U", "-a", "me", "-s", service, "-X", "not hex"},
		{},
	} {
		if code, _, _ := run1(t, direct(t), "", args...); code == 0 {
			t.Fatalf("%v: allowed", args)
		}
	}
	if code, _, _ := run1(t, direct(t), "delete-generic-password -a me -s \""+service+"\"\n", "-i"); code == 0 {
		t.Fatal("interactive delete allowed")
	}
	if c := calls(t, log); c != "" {
		t.Fatalf("real tool ran: %q", c)
	}
}

// The broker serves over its socket until the watch descriptor closes.
func TestBrokerServesUntilWatchCloses(t *testing.T) {
	fakeSecurity(t)
	dir, err := os.MkdirTemp("/tmp", "kc")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	socket := filepath.Join(dir, "s")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- serve(socket, service, int(r.Fd())) }()
	var code int
	var out string
	for i := 0; i < 100; i++ {
		if code, out, _ = run1(t, dial(socket), "", "find-generic-password", "-a", "me", "-w", "-s", service); code == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if code != 0 || out != "secret-token\n" {
		t.Fatalf("over the socket: %d %q", code, out)
	}
	w.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("broker outlived its watch")
	}
	if _, err := os.Stat(socket); !os.IsNotExist(err) {
		t.Fatal("socket left behind")
	}
}

func TestSplit(t *testing.T) {
	words, ok := split(`add-generic-password -U -a "my name" -s "Claude Code-credentials" -X "7b"`)
	want := []string{"add-generic-password", "-U", "-a", "my name", "-s", "Claude Code-credentials", "-X", "7b"}
	if !ok || strings.Join(words, "|") != strings.Join(want, "|") {
		t.Fatalf("%q", words)
	}
	if _, ok := split(`add "unterminated`); ok {
		t.Fatal("unterminated quote accepted")
	}
}
