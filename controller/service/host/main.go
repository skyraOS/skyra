// Command service-host keeps long-lived programs running outside the kernel
// (proposal 49). The launcher starts it with the programs the operator
// registered; the adapter (service/io) reaches it on its socket.
//
//	service-host -programs FILE -sock DIR/service.sock -state DIR
//
// The socket's directory must be this user's private directory, named
// skyra-service-*, which is how an adapter tells it from other sockets it is
// granted. Each session's idle backoff starts at -idle-start and stops the
// program once its interval would pass -idle-max with no operation.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	claudelaunch "skyra/controller/claude/launch"
	codexlaunch "skyra/controller/codex/launch"
	"skyra/controller/service"
)

// The model CLIs a program can name as its provider (proposal 47).
func init() {
	service.Register("codex", codexlaunch.Provider{})
	service.Register("claude", claudelaunch.Provider{})
}

func main() {
	programs := flag.String("programs", "", "the programs this host may run")
	sock := flag.String("sock", "", "the socket to serve: DIR/service.sock")
	state := flag.String("state", "", "the host's own directory")
	start := flag.Duration("idle-start", 30*time.Second, "a session's first idle interval")
	max := flag.Duration("idle-max", 16*time.Minute, "the idle interval past which a session stops")
	flag.Parse()
	if *programs == "" || *sock == "" || *state == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*programs, *sock, *state, *start, *max); err != nil {
		fmt.Fprintln(os.Stderr, "service-host:", err)
		os.Exit(1)
	}
}

func run(programs, sock, state string, start, max time.Duration) error {
	if filepath.Base(sock) != "service.sock" || !strings.HasPrefix(filepath.Base(filepath.Dir(sock)), "skyra-service-") {
		return fmt.Errorf("%s is not DIR/service.sock under a skyra-service-* directory", sock)
	}
	if err := service.Private(filepath.Dir(sock)); err != nil {
		return err
	}
	p, err := service.Load(programs)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(state, 0o700); err != nil {
		return err
	}
	os.Remove(sock)
	l, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	h := &service.Host{Programs: p, State: state, Start: start, Max: max, Log: os.Stderr}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		l.Close()
	}()
	err = h.Serve(l)
	h.Stop()
	if _, closed := err.(*net.OpError); closed {
		return nil
	}
	return err
}
