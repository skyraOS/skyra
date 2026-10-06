// Package custody describes non-frame descriptors carried with a lineage.
// Descriptors never enter records; this manifest travels beside them at each
// handoff and is checked before a target may receive one.
package custody

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"skyra/controller/core/horizon"
)

type Resource = horizon.Resource

func Validate(resources []Resource, files []*os.File) error {
	if len(resources) != len(files) {
		return fmt.Errorf("%d resources have %d descriptors", len(resources), len(files))
	}
	for i, r := range resources {
		if r.Kind != "connection" || r.Owner == "" || r.Incarnation == "" {
			return fmt.Errorf("invalid resource %d: %+v", i, r)
		}
		st, err := files[i].Stat()
		if err != nil {
			return err
		}
		if st.Mode()&os.ModeSocket == 0 {
			return fmt.Errorf("resource %d is not a socket", i)
		}
	}
	return nil
}

func Find(resources []Resource, files []*os.File, owner, incarnation string) (*os.File, error) {
	if err := Validate(resources, files); err != nil {
		return nil, err
	}
	for i, r := range resources {
		if r.Kind == "connection" && r.Owner == owner && r.Incarnation == incarnation {
			return files[i], nil
		}
	}
	return nil, fmt.Errorf("no connection for owner %s and incarnation %s", owner, incarnation)
}

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
