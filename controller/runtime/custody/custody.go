// Package custody keeps what a runtime process holds from leaking into the
// programs it starts. Frames travel as descriptors, explicitly; nothing else
// does (proposal 49 moved long-lived processes out of the kernel, behind adapters).
package custody

import (
	"os"
	"strconv"
	"syscall"
)

// SealInherited marks every descriptor above stderr close-on-exec. A process
// may be started holding descriptors its launcher leaked (a terminal, launchd,
// a CI runner); none of them may reach a child except by an explicit handoff:
// ExtraFiles or SCM_RIGHTS. Each runtime process calls it before anything else.
func SealInherited() {
	if entries, err := os.ReadDir("/dev/fd"); err == nil {
		for _, e := range entries {
			if fd, err := strconv.Atoi(e.Name()); err == nil && fd > 2 {
				syscall.CloseOnExec(fd)
			}
		}
		return
	}
	var lim syscall.Rlimit
	max := uint64(1 << 16)
	if syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim) == nil && lim.Cur < max {
		max = lim.Cur
	}
	for fd := 3; uint64(fd) < max; fd++ {
		syscall.CloseOnExec(fd)
	}
}
