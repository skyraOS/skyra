package keychain

import (
	"os/exec"
	"os/user"
	"strings"
)

func Get(service string) string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	out, err := exec.Command("security", "find-generic-password", "-a", u.Username, "-s", service, "-w").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
