// The first invocation parks via browser. The resumed invocation checks for
// the exact frame directories the ingress host held, by device/inode, then
// returns its verdict through the normal runtime. No path access is needed.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

type identity struct {
	Device int32  `json:"device"`
	Inode  uint64 `json:"inode"`
}

func main() {
	var c struct {
		Frame string `json:"frame"`
		Steps []struct {
			Value string `json:"value"`
		} `json:"steps"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&c); err != nil {
		panic(err)
	}
	if len(c.Steps) == 0 {
		fmt.Println(`{"to":"browser","payload":"wait"}`)
		return
	}
	var frames []identity
	if err := json.Unmarshal([]byte(c.Steps[len(c.Steps)-1].Value), &frames); err != nil || len(frames) == 0 {
		panic("resume probe received no frame identities")
	}
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		panic(err)
	}
	for fd := 3; uint64(fd) < limit.Cur; fd++ {
		var st syscall.Stat_t
		if syscall.Fstat(fd, &st) != nil {
			continue
		}
		for _, frame := range frames {
			if st.Dev == frame.Device && st.Ino == frame.Inode {
				verdict(c.Frame, "leaked-frame")
				return
			}
		}
	}
	verdict(c.Frame, "clean")
}

// verdict returns to the slot this continuation's frame discharges.
func verdict(frame, v string) {
	b, _ := json.Marshal(map[string]string{"to": frame, "payload": v})
	fmt.Println(string(b))
}
