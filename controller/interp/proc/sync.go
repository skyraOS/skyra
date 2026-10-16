package proc

import (
	"os"
	"syscall"
)

// flush orders the store's writes (proposal 40): F_BARRIERFSYNC, so what was
// flushed before persists before anything after it. The interpreter runs as
// a target, and what it does reaches the outside only through the
// controller, whose next release point forces the drive's cache (tree.Force)
// and with it these writes. SKYRA_SYNC=full makes every flush full, as
// before 40.
func flush(f *os.File) error {
	if os.Getenv("SKYRA_SYNC") == "full" {
		return f.Sync()
	}
	c, err := f.SyscallConn()
	if err != nil {
		return err
	}
	unordered := false
	if err := c.Control(func(fd uintptr) {
		_, _, e := syscall.Syscall(syscall.SYS_FCNTL, fd, fBarrierFsync, 0)
		unordered = e != 0
	}); err != nil {
		return err
	}
	if unordered {
		return f.Sync() // no barriers here: a plain fsync would not order, so flush fully
	}
	return nil
}

const fBarrierFsync = 85 // F_BARRIERFSYNC, <sys/fcntl.h>; not in package syscall
