// Command root selects a mode from argv and hands it to package demos. The
// runtime itself is split by responsibility: boot brings a system up and down,
// kernel runs the main loop, state records what exists, execution owns a
// context's authority and frames, extension grows it, ports carries Moves, and
// trace records every transition.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"skyra/history/dna19/root/boot"
	"skyra/history/dna19/root/demos"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	run := func() error {
		if len(os.Args) >= 6 && os.Args[1] == "codex-demo" {
			return demos.Codex(ctx, boot.CodexConfig{Binary: os.Args[2], Descriptor: os.Args[3], Provider: os.Args[4], Echo: os.Args[5], ProviderArgs: os.Args[6:]}, os.Stdout)
		}
		if len(os.Args) >= 4 && os.Args[1] == "user-demo" {
			addr := ""
			if len(os.Args) >= 5 {
				addr = os.Args[4]
			}
			return demos.User(boot.UserConfig{Binary: os.Args[2], Echo: os.Args[3], Addr: addr}, os.Stdout)
		}
		if len(os.Args) >= 8 && os.Args[1] == "user-codex-demo" {
			return demos.UserCodex(boot.UserCodexConfig{
				UserBinary: os.Args[2], CodexBinary: os.Args[3], Descriptor: os.Args[4],
				Provider: os.Args[5], Echo: os.Args[6], Addr: os.Args[7], ProviderArgs: os.Args[8:],
			}, os.Stdout)
		}
		return demos.Run(ctx, os.Args[1:], os.Stdout)
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
