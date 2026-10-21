package instance

import (
	"os"
	"path/filepath"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
)

// What runs is the bytes that were checked (43 D5): once the copy is
// published, changing the registered path changes nothing that runs, and a
// copy that no longer matches its name is replaced, not trusted.
func TestVerifiedCopyIsTheCheckedBytes(t *testing.T) {
	dir := t.TempDir()
	path, floor := filepath.Join(dir, "tool"), filepath.Join(dir, ".skyra")
	good := []byte("#!/bin/sh\necho checked\n")
	if err := os.WriteFile(path, good, 0o700); err != nil {
		t.Fatal(err)
	}
	s := horizon.Surface{Path: path, SHA256: record.Sum(good)}
	exe := horizon.Verified(horizon.Horizon{Surface: &s}, floor)
	if err := verifiedCopy(s, exe); err != nil {
		t.Fatal(err)
	}
	// The swap the check used to race with.
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho swapped\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := verifiedCopy(s, exe); err != nil {
		t.Fatalf("a published copy runs without the registered path: %v", err)
	}
	if got, _ := os.ReadFile(exe); string(got) != string(good) {
		t.Fatalf("the copy changed with the registered path: %q", got)
	}
	if info, err := os.Stat(exe); err != nil || info.Mode().Perm() != 0o500 {
		t.Fatalf("copy mode: %v %v", info.Mode(), err)
	}
	// A copy that does not match its name is not trusted: the registered
	// path is checked again, and it no longer matches either.
	os.Chmod(exe, 0o700)
	os.WriteFile(exe, []byte("tampered"), 0o700)
	if err := verifiedCopy(s, exe); err == nil {
		t.Fatal("a tampered copy and a swapped binary were accepted")
	}
	os.WriteFile(path, good, 0o700)
	if err := verifiedCopy(s, exe); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(exe); string(got) != string(good) {
		t.Fatalf("a tampered copy was not replaced: %q", got)
	}
	// A copy is only ever published under its own hash.
	if err := verifiedCopy(s, filepath.Join(floor, "surfaces", "other")); err == nil {
		t.Fatal("a copy was published under another name")
	}
	if e := run(machine.Run{Surface: s}).(machine.Executed); e.Err == "" {
		t.Fatal("a Run with no verified copy ran")
	}
}
