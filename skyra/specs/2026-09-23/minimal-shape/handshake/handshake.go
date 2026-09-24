// Package handshake owns the socket between a parent and the successor it
// spawned. Only those two hold it; the target never does.
//
//	parent → child   Horizon   the Horizon the bytes were emitted under
//	child  → parent  StampMsg  hash of the child's committed record
//	parent → child   Horizon   the child's own Horizon (+ the node fd): the release
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
	c, err := Conn(os.NewFile(uintptr(fds[0]), "handshake"))
	if err != nil {
		return nil, nil, err
	}
	return c, os.NewFile(uintptr(fds[1]), "handshake-child"), nil
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

// Send writes one message, optionally carrying one fd.
func Send(c *net.UnixConn, v any, fd *os.File) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var oob []byte
	if fd != nil {
		oob = syscall.UnixRights(int(fd.Fd()))
	}
	_, _, err = c.WriteMsgUnix(b, oob, nil)
	return err
}

// Recv reads one message, and the fd that came with it, if any.
func Recv(c *net.UnixConn, v any) (*os.File, error) {
	buf := make([]byte, 2048)
	oob := make([]byte, syscall.CmsgSpace(4))
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
	fds, err := syscall.ParseUnixRights(&msgs[0])
	if err != nil || len(fds) != 1 {
		return nil, errors.New("expected exactly one fd")
	}
	return os.NewFile(uintptr(fds[0]), "node"), nil
}
