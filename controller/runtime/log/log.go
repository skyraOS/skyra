// Package log owns the process's log lines on stderr.
package log

import (
	"fmt"
	"os"
)

// Who labels this process's lines: "boot", or the instance's short stamp.
var Who = "boot"

func Logf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "%-14s %s\n", "["+Who+"]", fmt.Sprintf(format, a...))
}
