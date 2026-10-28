// Package service keeps long-lived programs running outside the kernel
// (proposal 49). The kernel only runs programs that exit; a program that must
// outlive a step, a model CLI above all, is a session of the service host. The
// host keeps each session's stdio between steps and answers one operation at
// a time on its socket. The adapter (service/io) is the only client: an
// ordinary actor that performs one operation and exits.
//
// A session is reached only by its handle, which open returns. The procedure
// that opened it keeps the handle in its recorded state and passes it on
// every operation; neither the kernel nor the host tracks who owns what.
package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Request is one operation on the service host, as the adapter sends it.
type Request struct {
	Frame       string   `json:"frame,omitempty"`   // the adapter's frame, for the host's log
	Program     string   `json:"program,omitempty"` // open: a program the host has registered
	Incarnation string   `json:"incarnation"`       // every other operation: the session's handle
	Op          string   `json:"op"`                // open | exchange | inspect | stop
	Deliver     []string `json:"deliver"`
	WaitMS      int      `json:"wait_ms"`
	Prefix      string   `json:"prefix"`
}

// Observation is what an operation observed: the adapter's record of it.
type Observation struct {
	Lines       []string `json:"lines"`
	Partial     string   `json:"partial"`
	Delivery    string   `json:"delivery"`
	Read        string   `json:"read"`
	ObservedAt  int64    `json:"observed_at"`
	PeerEOF     bool     `json:"peer_eof"`
	Error       string   `json:"error"`
	Incarnation string   `json:"incarnation,omitempty"` // open: the new session's handle
	Pid         int      `json:"pid,omitempty"`         // open
	Profile     string   `json:"profile,omitempty"`     // open: the hash of the profile the session runs under
}

const (
	// A tool result can carry a 2 MiB screenshot as base64 plus text, and a
	// program may echo it back in a notification.
	maxLine     = 4 << 20 // one incoming protocol line, including its newline
	maxDelivery = 4 << 20
	maxBatch    = 256  // lines one exchange returns
	maxStop     = 1024 // lines a stop drains
	// Quiet is how an operation that waited out its deadline is reported;
	// the adapter treats it as silence.
	Quiet = "i/o timeout: no output within the wait"
)

// Ask sends one request to the service host at socket and returns its
// observation. A host that cannot be reached is a session that is gone:
// nothing was delivered, and the peer has closed.
func Ask(socket string, q Request) Observation {
	gone := Observation{Lines: []string{}, Delivery: "not_sent", Read: "not_started", PeerEOF: true, ObservedAt: time.Now().UnixMilli()}
	c, err := net.DialTimeout("unix", socket, 5*time.Second)
	if err != nil {
		return gone
	}
	defer c.Close()
	// An open prepares the program before it answers: allow for that.
	c.SetDeadline(time.Now().Add(time.Duration(q.WaitMS)*time.Millisecond + 60*time.Second))
	b, _ := json.Marshal(q)
	if _, err := c.Write(append(b, '\n')); err != nil {
		return gone
	}
	line, err := bufio.NewReader(io.LimitReader(c, 512<<20)).ReadBytes('\n')
	var o Observation
	if err != nil || json.Unmarshal(line, &o) != nil {
		// The request went out; whether it reached the session is unknown.
		return Observation{Lines: []string{}, Delivery: "unknown", Read: "observed", Error: "service host did not answer", ObservedAt: time.Now().UnixMilli()}
	}
	if o.Lines == nil {
		o.Lines = []string{}
	}
	return o
}

// Dir is this user's private socket directory. A socket path must stay under
// 104 bytes on macOS, so it is short, and it is checked to be this user's
// own, private directory, since /private/tmp is shared.
func Dir() (string, error) {
	dir := fmt.Sprintf("/private/tmp/skyra-%d", os.Getuid())
	if err := os.Mkdir(dir, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return "", err
	}
	if err := Private(dir); err != nil {
		return "", err
	}
	return filepath.Clean(dir), nil
}

// Private checks that dir is a directory this user owns and no one else can
// enter.
func Private(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !info.IsDir() || !ok || int(st.Uid) != os.Getuid() || info.Mode().Perm() != 0o700 {
		return fmt.Errorf("%s is not this user's private directory", dir)
	}
	return nil
}
