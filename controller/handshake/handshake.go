// Package handshake owns the socket between a parent and the successor it
// spawned. Only those two hold it; the target never does.
//
//	parent → child   Horizon   the Horizon the bytes were emitted under
//	child  → parent  StampMsg  hash of the child's committed record
//	parent → child   Horizon   the child's own Horizon, plus its frame stack: the release
//
// The frame stack is the call stack as capabilities: one directory fd per
// frame, from the frame this Horizon's reply discharges down to the operator's.
package handshake

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"syscall"
)

type StampMsg struct {
	Stamp string `json:"stamp"`
}

// Pair creates a datagram socketpair. The parent keeps the conn; the file
// becomes the child's fd 3.
func Pair() (*net.UnixConn, *os.File, error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return nil, nil, err
	}
	// Only explicit ExtraFiles/SCM_RIGHTS grants may survive an exec. Raw
	// socketpair descriptors otherwise leak beside the intended fd 3 copy.
	for _, fd := range fds {
		syscall.CloseOnExec(fd)
	}
	// macOS caps a unix datagram at net.local.dgram.maxdgram (2 KiB) unless the
	// socket's buffers are raised. A Horizon carries contracts and the
	// continuations waiting on it, so give it room.
	for _, fd := range fds {
		for _, opt := range []int{syscall.SO_SNDBUF, syscall.SO_RCVBUF} {
			if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, opt, maxMessage); err != nil {
				return nil, nil, err
			}
		}
	}
	c, err := Conn(os.NewFile(uintptr(fds[0]), "handshake"))
	if err != nil {
		return nil, nil, err
	}
	return c, os.NewFile(uintptr(fds[1]), "handshake-child"), nil
}

// Inherited wraps n fds granted through ExtraFiles, from fd first on. An
// inherited fd has close-on-exec cleared, so without this every target the
// process runs would inherit it too: a frame directory fd would let a
// sandboxed target list a directory its profile denies.
func Inherited(first, n int) []*os.File {
	files := make([]*os.File, n)
	for i := range files {
		syscall.CloseOnExec(first + i)
		files[i] = os.NewFile(uintptr(first+i), "frame")
	}
	return files
}

// Conn turns an inherited socket fd into a conn.
func Conn(f *os.File) (*net.UnixConn, error) {
	defer f.Close()
	c, err := net.FileConn(f)
	if err != nil {
		return nil, err
	}
	uc, ok := c.(*net.UnixConn)
	if !ok {
		return nil, errors.New("handshake is not a unix socket")
	}
	return uc, nil
}

const (
	maxMessage = 1 << 20
	maxFDs     = 64
)

// Send writes one message, with any fds.
func Send(c *net.UnixConn, v any, fds ...*os.File) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var oob []byte
	if len(fds) > 0 {
		raw := make([]int, len(fds))
		for i, f := range fds {
			raw[i] = int(f.Fd())
		}
		oob = syscall.UnixRights(raw...)
	}
	_, _, err = c.WriteMsgUnix(b, oob, nil)
	return err
}

// Recv reads one message, and the fds that came with it, in order.
func Recv(c *net.UnixConn, v any) ([]*os.File, error) {
	buf := make([]byte, maxMessage)
	oob := make([]byte, syscall.CmsgSpace(4*maxFDs))
	n, oobn, _, _, err := c.ReadMsgUnix(buf, oob)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("peer closed the handshake")
	}
	if err := json.Unmarshal(buf[:n], v); err != nil {
		return nil, err
	}
	if oobn == 0 {
		return nil, nil
	}
	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil || len(msgs) != 1 {
		return nil, errors.New("bad control message")
	}
	raw, err := syscall.ParseUnixRights(&msgs[0])
	if err != nil {
		return nil, err
	}
	files := make([]*os.File, len(raw))
	for i, fd := range raw {
		syscall.CloseOnExec(fd)
		files[i] = os.NewFile(uintptr(fd), "frame")
	}
	return files, nil
}
