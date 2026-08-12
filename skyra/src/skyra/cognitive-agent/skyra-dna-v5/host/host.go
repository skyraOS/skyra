// Package host is the capability-class layer: the kernel calls through a class
// interface, a driver fills it in, and the kernel never knows what is behind it.
//
//	Skyra kernel
//	    ├── execution     spawn. the one needed first.
//	    ├── storage       read/write. cheap ops that must not become spawns.
//	    ├── display  ─┐
//	    └── audio    ─┴─ the PROJECTION substrate (NOT NAMED YET)
//	          ↑
//	      OS drivers            linux · macos · windows
//
// Classes rather than one method, because an ops struct with optional members can
// grow without breaking existing implementations and a one-method interface
// cannot grow at all. That matters more here than usual: the primitive list is
// the one surface SELECTION CANNOT CORRECT.
//
// An absent implementation is NOT a declaration — it is a build-time fact, and
// calling into a class a host does not implement fails loudly.
//
// The test for what belongs here is whether anything INTERPRETS. A driver call
// reaches MECHANISM with fixed, reviewed semantics; an entity reaches an
// INTERPRETER whose behaviour is authored elsewhere. So a filesystem read is a
// driver call and an external API is an ENTITY, never a class. Network is not a
// class either — it is TRANSPORT, beneath everything.
package host

import (
	"fmt"
	"os"
	"path/filepath"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// Host is reached through capability classes. Constitutive rather than declared:
// something with no execution class is not a host, it is a device we cannot use.
type Host interface {
	ID() ids.HostID
	Execution() ExecutionOps // nil if not implemented — a build-time fact
	Storage() StorageOps
}

// ExecutionOps is the trust boundary itself, not merely a call across it.
//
//	THE CALL      TRUSTED.     fixed semantics, kernel code.
//	THE PAYLOAD   UNTRUSTED.   the process does whatever it wants.
//
// A trusted primitive whose entire purpose is releasing something untrusted —
// which is why grants on it cannot be checked at the call site and must be
// REALIZED by the platform layer instead, or the grant describes nothing.
type ExecutionOps interface {
	Spawn(argv []string, dir string) (Handle, error)
}

// StorageOps exists because spawn alone is expressively sufficient and
// cost-catastrophic: "read a file" becoming "spawn a reader" is a process per
// read. Linux has read() as a syscall for exactly this reason.
type StorageOps interface {
	Read(path string) ([]byte, error)
	Write(path string, body []byte) error
}

// Handle is what spawn hands back. PORTS ARE RETURNED, NEVER SPECIFIED: how to
// reach the thing is inside it, and the caller never learns what a port is.
//
// AN OPAQUE REFERENCE AND THE HOST IT IS VALID ON. Nothing else. It carried
// Stdout, Stderr and Exit, which is a SHELL'S SHAPE leaking into a class
// interface — a guess nobody was forced to make, on the one surface selection
// cannot correct later, in a method that has never run. A process's output is
// not part of what a handle IS; it is part of what some future execution driver
// returns beside one.
type Handle struct {
	Host ids.HostID
	Ref  string
}

// ---------------------------------------------------------------------------
// the local host — `here` is the axiom
// ---------------------------------------------------------------------------

// Local is the host the runtime is executing on. It is GIVEN, not discovered —
// the same shape as the standing root: the thing you cannot derive because you
// are standing on it.
type Local struct {
	Root string // everything this host can touch lives under here
}

func (l *Local) ID() ids.HostID          { return "here" }
func (l *Local) Execution() ExecutionOps { return &localExec{l} }
func (l *Local) Storage() StorageOps     { return &localStore{l} }

type localExec struct{ l *Local }

func (e *localExec) Spawn(argv []string, dir string) (Handle, error) {
	return Handle{}, fmt.Errorf("execution class not implemented on %s", e.l.ID())
}

type localStore struct{ l *Local }

func (s *localStore) Read(path string) ([]byte, error) {
	p, err := s.resolve(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(p)
}

func (s *localStore) Write(path string, body []byte) error {
	p, err := s.resolve(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, body, 0o644)
}

// resolve is the platform layer refusing to leave its root.
//
// This is what a grant would have to be realized BY. kernel.md is explicit: if
// the kernel checks a grant and then the platform layer does not confine, the
// grant described nothing and the record reads as bounded. The confinement has to
// be real down here or it is decoration up there.
func (s *localStore) resolve(path string) (string, error) {
	clean := filepath.Clean(filepath.Join(s.l.Root, path))
	root := filepath.Clean(s.l.Root)
	if clean != root && !filepath.HasPrefix(clean, root+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes host root", path)
	}
	return clean, nil
}
