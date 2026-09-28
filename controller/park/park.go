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
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"

	"skyra/controller/custody"
	"skyra/controller/machine"
)

// Frame is a parked invocation: everything needed to resume it.
type Frame struct {
	Root     string           `json:"root"`     // the .skyra it belongs to
	Instance machine.Instance `json:"instance"` // the instance, in Execute, waiting for its target's output
	Payload  json.RawMessage  `json:"payload"`  // what the invocation asked of the host
	Files    []*os.File       `json:"-"`        // the frame stack: custody
}

// Send parks a frame with the host at sock and waits for its answer.
func Send(sock string, f Frame) error {
	c, err := net.Dial("unix", sock)
	if err != nil {
		return err
	}
	defer c.Close()
	uc := c.(*net.UnixConn)
	body, _ := json.Marshal(f)
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
func Receive(c *net.UnixConn) (Frame, error) {
	head := make([]byte, 4)
	oob := make([]byte, syscall.CmsgSpace(4*64))
	n, oobn, _, _, err := c.ReadMsgUnix(head, oob)
	if err != nil {
		return Frame{}, err
	}
	if n < 4 {
		if _, err := io.ReadFull(c, head[n:]); err != nil {
			return Frame{}, err
		}
	}
	body := make([]byte, binary.BigEndian.Uint32(head))
	if _, err := io.ReadFull(c, body); err != nil {
		return Frame{}, err
	}
	var f Frame
	if err := json.Unmarshal(body, &f); err != nil {
		return Frame{}, err
	}
	if oobn > 0 {
		msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
		if err != nil || len(msgs) != 1 {
			return Frame{}, errors.New("bad control message")
		}
		fds, err := syscall.ParseUnixRights(&msgs[0])
		if err != nil {
			return Frame{}, err
		}
		for _, fd := range fds {
			syscall.CloseOnExec(fd)
			f.Files = append(f.Files, os.NewFile(uintptr(fd), "frame"))
		}
	}
	depth := f.Instance.Self.Depth()
	if len(f.Files) != depth+len(f.Instance.Resources) {
		f.Close()
		return Frame{}, fmt.Errorf("parked with %d descriptors, its Horizon needs %d frames and %d resources", len(f.Files), depth, len(f.Instance.Resources))
	}
	if err := custody.Validate(f.Instance.Resources, f.Files[depth:]); err != nil {
		f.Close()
		return Frame{}, err
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
// with the frame stack as fds 3.., and the instance and the event on stdin.
// The event is the target's output; the resumed instance hands it off to a
// successor, which checks it like any other output.
func Resume(skyra string, f Frame, event []byte, stderr io.Writer) error {
	in, _ := json.Marshal(Resumed{Instance: f.Instance, Out: string(event)})
	cmd := exec.Command(skyra, "resume", f.Root)
	cmd.Stdin = bytes.NewReader(in)
	cmd.Stderr = stderr
	cmd.ExtraFiles = f.Files
	if err := cmd.Start(); err != nil {
		return err
	}
	f.Close() // the resumed instance holds custody now
	go cmd.Wait()
	return nil
}

// Resumed is what `skyra resume` reads on stdin.
type Resumed struct {
	Instance machine.Instance `json:"instance"`
	Out      string           `json:"out"`
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
	body := make([]byte, binary.BigEndian.Uint32(head))
	if _, err := io.ReadFull(c, body); err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}
