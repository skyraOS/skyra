// Latency sits between the actor host and an app-server (usually codex/fake)
// and holds back each model output, a tool call or an agent message, for -ms
// before forwarding it: simulated inference, with no change to the controller.
// Every other line passes through at once. See ../../../OVERHEAD-VS-INFERENCE.md.
//
//	latency -ms 1000 ./bin/fake-codex
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	ms := flag.Int("ms", 0, "milliseconds to hold back each model output")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: latency -ms N APP-SERVER [ARGS...]")
		os.Exit(2)
	}
	cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
	cmd.Stdin, cmd.Stderr = os.Stdin, os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "latency:", err)
		os.Exit(1)
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "latency:", err)
		os.Exit(1)
	}
	delay := time.Duration(*ms) * time.Millisecond
	r := bufio.NewReaderSize(out, 1<<20)
	for {
		line, err := r.ReadString('\n')
		if line != "" {
			if strings.Contains(line, `"method":"item/tool/call"`) || strings.Contains(line, `"method":"item/completed"`) {
				time.Sleep(delay)
			}
			if _, werr := io.WriteString(os.Stdout, line); werr != nil {
				break
			}
		}
		if err != nil {
			break
		}
	}
	if cmd.Wait() != nil {
		os.Exit(1)
	}
}
