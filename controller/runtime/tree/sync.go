package tree

import (
	"os"
	"syscall"
	"time"

	"skyra/controller/runtime/trace"
)

// The durability policy (proposal 40, built in 41 V5).
//
// Every flush in the tree orders: F_BARRIERFSYNC, so what was flushed before
// it persists before anything issued after it. That keeps the tree
// consistent after a power loss (no record without its input, no slot
// without its round, no name with partial content) without emptying the
// drive's cache each time.
//
// Force empties the cache: one F_FULLFSYNC, which makes everything every
// earlier barrier sent persistent. The shell forces before an act that
// leaves the failure domain (a target runs, a host takes
// custody) and boot forces before it tells the operator anything. So an
// effect is never released before its intent is durable; what can be lost
// in a power loss is only the tail of controller-only steps since the last
// forced flush, as a unit.
//
// SKYRA_SYNC=full makes every flush a full flush again, as before 40.
// SKYRA_SYNC_TRACE=1 logs each flush, its kind, and how long it took.
var (
	syncFull  = os.Getenv("SKYRA_SYNC") == "full"
	syncTrace = os.Getenv("SKYRA_SYNC_TRACE") != ""
)

const fBarrierFsync = 85 // F_BARRIERFSYNC, <sys/fcntl.h>; not in package syscall

// flush orders everything written to f before what is written after it.
func flush(f *os.File, what string) error {
	t := time.Now()
	kind, err := barrier(f)
	if syncTrace {
		trace.Logf("fsync    %s %s %dus", kind, what, time.Since(t).Microseconds())
	}
	return err
}

func barrier(f *os.File) (string, error) {
	if syncFull {
		return "full", f.Sync()
	}
	c, err := f.SyscallConn()
	if err != nil {
		return "barrier", err
	}
	unordered := false
	if err := c.Control(func(fd uintptr) {
		_, _, e := syscall.Syscall(syscall.SYS_FCNTL, fd, fBarrierFsync, 0)
		unordered = e != 0
	}); err != nil {
		return "barrier", err
	}
	if unordered {
		// A file system without barriers: a plain fsync would neither order
		// nor persist on macOS, so this flush is a full one (40, C).
		return "full", f.Sync()
	}
	return "barrier", nil
}

// Force makes everything written to the tree's device so far persistent:
// one F_FULLFSYNC, through the tree's root (any descriptor on the device
// flushes the drive's whole cache). Call it before an effect leaves the
// failure domain.
func Force(root string) error {
	t := time.Now()
	d, err := os.Open(root)
	if err != nil {
		return err
	}
	defer d.Close()
	err = d.Sync()
	if syncTrace {
		trace.Logf("fsync    force %dus", time.Since(t).Microseconds())
	}
	return err
}
