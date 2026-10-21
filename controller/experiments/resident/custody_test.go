//go:build darwin

package resident

// This is an isolated transport experiment, not a new production execution
// mode. Helpers below stand in for successive trusted runtime instances.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
	"skyra/controller/runtime/handshake"
)

const helperEnv = "SKYRA_RESIDENT_EXPERIMENT"

type resource struct {
	Kind        string `json:"kind"`
	Owner       string `json:"owner"`
	Incarnation string `json:"incarnation,omitempty"`
}

type bundle struct {
	Resources []resource `json:"resources"`
	PID       int        `json:"pid"`
	Profile   string     `json:"profile"`
	Files     []*os.File `json:"-"`
}

func (b *bundle) close() {
	for _, f := range b.Files {
		f.Close()
	}
	b.Files = nil
}

func (b bundle) validate() error {
	if len(b.Resources) != len(b.Files) {
		return errors.New("resource manifest does not match descriptors")
	}
	for i, r := range b.Resources {
		st, err := b.Files[i].Stat()
		if err != nil {
			return err
		}
		switch r.Kind {
		case "frame":
			if !st.IsDir() {
				return errors.New("frame is not a directory")
			}
		case "connection":
			if st.Mode()&os.ModeSocket == 0 || r.Incarnation == "" {
				return errors.New("connection is not an identified socket")
			}
		default:
			return errors.New("unknown resource kind")
		}
		if r.Owner == "" {
			return errors.New("resource has no owner")
		}
	}
	return nil
}

// grant selects a connection for an authorized operation. Carrying custody
// through an unrelated tool does not grant that tool the descriptor.
func (b bundle) grant(owner, incarnation string) (*os.File, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}
	for i, r := range b.Resources {
		if r.Kind == "connection" && r.Owner == owner && r.Incarnation == incarnation {
			return b.Files[i], nil
		}
	}
	return nil, errors.New("connection grant does not match owner and incarnation")
}

type request struct {
	Op          string   `json:"op"`
	Owner       string   `json:"owner"`
	Incarnation string   `json:"incarnation"`
	Deliver     []string `json:"deliver"`
	ReadLines   int      `json:"read_lines"`
	WaitMS      int      `json:"wait_ms"`
	Prefix      string   `json:"prefix,omitempty"`
}

type observation struct {
	Lines      []string `json:"lines"`
	Partial    string   `json:"partial"`
	Delivery   string   `json:"delivery"`
	Read       string   `json:"read"`
	ObservedAt int64    `json:"observed_at"`
	PeerEOF    bool     `json:"peer_eof"`
	Error      string   `json:"error"`
}

func streamPair() (*os.File, *os.File, error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, nil, err
	}
	for _, fd := range fds {
		syscall.CloseOnExec(fd)
	}
	return os.NewFile(uintptr(fds[0]), "connection"), os.NewFile(uintptr(fds[1]), "provider"), nil
}

func helperCommand(mode, floor string, sandbox bool) (*exec.Cmd, string, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, "", err
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return nil, "", err
	}
	bin, err := os.ReadFile(self)
	if err != nil {
		return nil, "", err
	}
	// The profile lets only the runtime's verified copy run (43 D5).
	h := horizon.Horizon{Surface: &horizon.Surface{Path: self, SHA256: record.Sum(bin)}}
	exe := horizon.Verified(h, floor)
	if have, err := os.ReadFile(exe); err != nil || record.Sum(have) != h.Surface.SHA256 {
		if err := os.MkdirAll(filepath.Dir(exe), 0o700); err != nil {
			return nil, "", err
		}
		tmp := fmt.Sprintf("%s.%d", exe, os.Getpid())
		if err := os.WriteFile(tmp, bin, 0o500); err != nil {
			return nil, "", err
		}
		if err := os.Rename(tmp, exe); err != nil {
			return nil, "", err
		}
	}
	profile := horizon.Render(h, floor)
	var cmd *exec.Cmd
	if sandbox {
		cmd = exec.Command("/usr/bin/sandbox-exec", "-p", profile, exe, "-test.run=^TestResidentHelper$")
	} else {
		cmd = exec.Command(self, "-test.run=^TestResidentHelper$")
	}
	cmd.Env = []string{"PATH=/usr/bin:/bin", helperEnv + "=" + mode}
	cmd.Stderr = os.Stderr
	return cmd, record.Sum([]byte(profile)), nil
}

// launch runs in a short-lived trusted process. It exits after handing off
// the endpoint; the resident provider is not supervised by a permanent host.
func launch(floor string) (bundle, error) {
	local, peer, err := streamPair()
	if err != nil {
		return bundle{}, err
	}
	defer peer.Close()
	frame, err := os.Open(floor)
	if err != nil {
		local.Close()
		return bundle{}, err
	}
	cmd, profile, err := helperCommand("provider", floor, true)
	if err != nil {
		local.Close()
		frame.Close()
		return bundle{}, err
	}
	cmd.Stdin, cmd.Stdout = peer, peer
	if err := cmd.Start(); err != nil {
		local.Close()
		frame.Close()
		return bundle{}, err
	}
	pid := cmd.Process.Pid
	incarnation := fmt.Sprintf("%d-%d", pid, time.Now().UnixNano())
	b := bundle{PID: pid, Profile: profile, Files: []*os.File{frame, local}, Resources: []resource{
		{Kind: "frame", Owner: "conversation"},
		{Kind: "connection", Owner: "conversation", Incarnation: incarnation},
	}}
	// Read a readiness line without reading beyond it. The actual exchange
	// workers start with the provider's first protocol response still unread.
	c, err := net.FileConn(local)
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		b.close()
		return bundle{}, err
	}
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := exactLine(c, "")
	c.Close()
	if err != nil || line != "ready\n" {
		cmd.Process.Kill()
		cmd.Wait()
		b.close()
		return bundle{}, fmt.Errorf("sandboxed provider readiness: %q, %v", line, err)
	}
	cmd.Process.Release()
	return b, nil
}

// exactLine deliberately does not use bufio: bytes following the newline
// must remain in the socket for the next ephemeral worker. Partial reads are
// returned explicitly, including when a deadline interrupts a line.
func exactLine(r io.Reader, prefix string) (string, error) {
	line := []byte(prefix)
	var one [1]byte
	for len(line) < 1<<20 {
		n, err := r.Read(one[:])
		if n == 1 {
			line = append(line, one[0])
			if one[0] == '\n' {
				return string(line), nil
			}
		}
		if err != nil {
			return string(line), err
		}
	}
	return string(line), errors.New("line exceeds experiment limit")
}

func exchange(file *os.File, q request) (o observation) {
	o.Delivery, o.Read = "not_sent", "observed"
	o.Lines = []string{}
	defer func() { o.ObservedAt = time.Now().UnixMilli() }()
	c, err := net.FileConn(file)
	if err != nil {
		o.Error = err.Error()
		return
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(time.Duration(q.WaitMS) * time.Millisecond))
	if q.Op == "stop" {
		// Shutdown affects all copies of this socket endpoint. Closing this
		// worker's fd alone cannot signal EOF while a custodian retains one.
		if err = syscall.Shutdown(int(file.Fd()), syscall.SHUT_WR); err != nil {
			o.Error = err.Error()
			return
		}
	} else {
		for _, line := range q.Deliver {
			o.Delivery = "unknown"
			if _, err = io.Copy(c, bytes.NewBufferString(line+"\n")); err != nil {
				o.Error = err.Error()
				return
			}
		}
		if len(q.Deliver) > 0 {
			o.Delivery = "written" // transport write, not provider acknowledgment
		}
	}
	if q.Op == "die-after-write" {
		os.Exit(71)
	}
	prefix := q.Prefix
	for i := 0; i < q.ReadLines || q.Op == "stop"; i++ {
		line, readErr := exactLine(c, prefix)
		prefix = ""
		if q.Op == "die-after-read" {
			os.Exit(72)
		}
		if readErr != nil {
			o.Partial = line
			o.PeerEOF = errors.Is(readErr, io.EOF)
			if !o.PeerEOF {
				o.Error = readErr.Error()
			}
			return
		}
		o.Lines = append(o.Lines, line)
	}
	return
}

// operate launches a fresh sandboxed child with only its explicit connection
// grant. The custodian retains the original endpoint during the operation.
func operate(b bundle, q request, floor string) observation {
	f, err := b.grant(q.Owner, q.Incarnation)
	if err != nil {
		return observation{Delivery: "not_sent", Read: "not_started", ObservedAt: time.Now().UnixMilli(), Error: err.Error()}
	}
	cmd, _, err := helperCommand("operation", floor, true)
	if err != nil {
		return observation{Delivery: "not_sent", Read: "not_started", Error: err.Error()}
	}
	in, _ := json.Marshal(q)
	cmd.Stdin = bytes.NewReader(in)
	cmd.ExtraFiles = []*os.File{f}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		return observation{Delivery: "not_sent", Read: "not_started", ObservedAt: time.Now().UnixMilli(), Error: err.Error()}
	}
	if err := cmd.Wait(); err != nil {
		return observation{Delivery: "unknown", Read: "unknown", ObservedAt: time.Now().UnixMilli(), Error: err.Error()}
	}
	var o observation
	if err := json.Unmarshal(out.Bytes(), &o); err != nil {
		return observation{Delivery: "unknown", Read: "unknown", ObservedAt: time.Now().UnixMilli(), Error: err.Error()}
	}
	return o
}

func receive(c *net.UnixConn) (bundle, error) {
	var b bundle
	files, err := handshake.Recv(c, &b)
	b.Files = files
	if err == nil {
		err = b.validate()
	}
	if err != nil {
		b.close()
	}
	return b, err
}

func transfer(c *net.UnixConn, b *bundle) error {
	if err := handshake.Send(c, *b, b.Files...); err != nil {
		return err
	}
	var ack string
	if _, err := handshake.Recv(c, &ack); err != nil {
		return err
	}
	if ack != "accepted" {
		return errors.New("custody not accepted")
	}
	b.close()
	return nil
}
