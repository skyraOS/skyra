// Package park is how an instance hands a frame to an actor host (next/16),
// and how the host hands it back. Parking passes custody: the instance's
// state, its payload and its whole frame stack (as fds) go to the host, and
// the instance exits only once the host has acknowledged them.
//
// On the wire: a 4-byte big-endian length, then that much JSON; the frame
// stack rides on the same write as SCM_RIGHTS. The host answers with the same
// framing: {"err": ""} for custody taken.
package park

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Frame is a parked invocation. The host holds a reference to the work and
// its descriptors, never its authority: resume derives the instance from the
// tree (next/20, finding 1; next/24).
type Frame struct {
	Root    string          `json:"root"`    // the .skyra it belongs to
	Stamp   string          `json:"stamp"`   // the parked invoke; its frame is the slot the event returns to
	Frames  int             `json:"frames"`  // how many of Files are the frame stack; resume checks it against the derived Boundary
	Payload json.RawMessage `json:"payload"` // what the invocation asked of the host
	Files   []*os.File      `json:"-"`       // the frame stack and resources: custody
}

// Send parks a frame with the host at sock and waits for its answer.
const maxMessage = 16 << 20

func Send(sock string, f Frame) error {
	c, err := net.Dial("unix", sock)
	if err != nil {
		return err
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(30 * time.Second))
	uc := c.(*net.UnixConn)
	body, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if len(body) > maxMessage {
		return errors.New("park message too large")
	}
	msg := binary.BigEndian.AppendUint32(nil, uint32(len(body)))
	msg = append(msg, body...)
	fds := make([]int, len(f.Files))
	for i, file := range f.Files {
		fds[i] = int(file.Fd())
	}
	// The fds ride on the header. The body follows with an ordinary write,
	// which loops until it is all sent: a single sendmsg on a stream socket
	// may send only part of a large message.
	if _, _, err := uc.WriteMsgUnix(msg[:4], syscall.UnixRights(fds...), nil); err != nil {
		return err
	}
	if _, err := uc.Write(msg[4:]); err != nil {
		return err
	}
	var ack struct {
		Err string `json:"err"`
	}
	if err := readFramed(uc, &ack); err != nil {
		return fmt.Errorf("no answer from the host: %w", err)
	}
	if ack.Err != "" {
		return errors.New(ack.Err)
	}
	return nil
}

// Receive reads one parked frame from a connection, with its fds.
func Receive(c *net.UnixConn) (result Frame, failure error) {
	c.SetDeadline(time.Now().Add(30 * time.Second))
	head := make([]byte, 4)
	oob := make([]byte, syscall.CmsgSpace(4*64))
	n, oobn, flags, _, err := c.ReadMsgUnix(head, oob)
	if err != nil {
		return Frame{}, err
	}
	var files []*os.File
	defer func() {
		if failure != nil {
			for _, f := range files {
				f.Close()
			}
		}
	}()
	if oobn > 0 {
		msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
		if err != nil {
			return Frame{}, err
		}
		for _, msg := range msgs {
			fds, err := syscall.ParseUnixRights(&msg)
			if err != nil {
				return Frame{}, err
			}
			for _, fd := range fds {
				syscall.CloseOnExec(fd)
				files = append(files, os.NewFile(uintptr(fd), "frame"))
			}
		}
	}
	if flags&syscall.MSG_CTRUNC != 0 {
		return Frame{}, errors.New("truncated descriptors")
	}
	if n < 4 {
		if _, err := io.ReadFull(c, head[n:]); err != nil {
			return Frame{}, err
		}
	}
	size := binary.BigEndian.Uint32(head)
	if size > maxMessage {
		return Frame{}, errors.New("park message too large")
	}
	body := make([]byte, size)
	if _, err := io.ReadFull(c, body); err != nil {
		return Frame{}, err
	}
	var f Frame
	if err := json.Unmarshal(body, &f); err != nil {
		return Frame{}, err
	}
	f.Files = files
	depth := f.Frames
	if depth < 1 || len(f.Files) != depth {
		return Frame{}, fmt.Errorf("parked with %d descriptors, for %d frames", len(f.Files), depth)
	}
	return f, nil
}

// Answer acknowledges a parked frame: custody taken, or refused with a reason.
func Answer(c *net.UnixConn, refusal string) error {
	body, _ := json.Marshal(map[string]string{"err": refusal})
	_, err := c.Write(append(binary.BigEndian.AppendUint32(nil, uint32(len(body))), body...))
	return err
}

// Resume hands a parked frame back to the runtime: it starts `skyra resume`
// with the frame stack as fds 3.., and the parked stamp and the event on
// stdin. The runtime derives the instance from the tree; the event is the
// target's output, handed off to a successor that checks it like any other.
func Resume(skyra string, f Frame, event []byte, stderr io.Writer) error {
	defer f.Close()
	in, _ := json.Marshal(Resumed{Stamp: f.Stamp, Frames: f.Frames, Out: string(event)})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, skyra, "resume", f.Root)
	cmd.Stdin = bytes.NewReader(in)
	cmd.Stderr = stderr
	cmd.ExtraFiles = f.Files
	if err := cmd.Start(); err != nil {
		return err
	}
	// Drive returns only after its successor has validated custody. A process
	// that starts but rejects the resume must be reported to the caller.
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("resume custody was not confirmed: %w", err)
	}
	return nil
}

// Resumed is what `skyra resume` reads on stdin: which work, what its
// descriptors are, and the event. Nothing in it is authority.
type Resumed struct {
	Stamp  string `json:"stamp"`
	Frames int    `json:"frames"`
	Out    string `json:"out"`
}

// Close releases the host's hold on the frame stack.
func (f Frame) Close() {
	for _, file := range f.Files {
		file.Close()
	}
}

func readFramed(c io.Reader, v any) error {
	head := make([]byte, 4)
	if _, err := io.ReadFull(c, head); err != nil {
		return err
	}
	size := binary.BigEndian.Uint32(head)
	if size > maxMessage {
		return errors.New("park message too large")
	}
	body := make([]byte, size)
	if _, err := io.ReadFull(c, body); err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}
