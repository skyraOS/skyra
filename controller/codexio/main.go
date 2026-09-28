// Command codexio performs one recorded operation on a resident stream. The
// runtime grants only the matching connection at fd 3; the worker owns no
// process launcher, host socket or side store.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

type request struct {
	Incarnation string   `json:"incarnation"`
	Op          string   `json:"op"`
	Deliver     []string `json:"deliver"`
	WaitMS      int      `json:"wait_ms"`
	Prefix      string   `json:"prefix"`
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

func main() {
	var q request
	if err := json.NewDecoder(os.Stdin).Decode(&q); err != nil {
		fail(err)
	}
	if q.Incarnation == "" || (q.Op != "exchange" && q.Op != "stop" && q.Op != "inspect") || q.WaitMS < 1 || q.WaitMS > 30000 {
		fail(errors.New("invalid operation"))
	}
	file := os.NewFile(3, "resident-connection")
	if file == nil {
		fail(errors.New("no connection granted"))
	}
	defer file.Close()
	o := operate(file, q)
	b, _ := json.Marshal(map[string]any{"to": "reply", "payload": o})
	fmt.Println(string(b))
}

func fail(err error) { fmt.Fprintln(os.Stderr, "codexio:", err); os.Exit(1) }

func operate(file *os.File, q request) (o observation) {
	o.Lines = []string{}
	o.Delivery, o.Read = "not_sent", "not_started"
	defer func() { o.ObservedAt = time.Now().UnixMilli() }()
	conn, err := net.FileConn(file)
	if err != nil {
		o.Error = err.Error()
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Duration(q.WaitMS) * time.Millisecond))
	if q.Op == "stop" {
		if err = syscall.Shutdown(int(file.Fd()), syscall.SHUT_WR); err != nil {
			o.Error = err.Error()
			return
		}
		o.Delivery = "shutdown"
	} else {
		for _, line := range q.Deliver {
			if !json.Valid([]byte(line)) || len(line) > 1<<20 {
				o.Error = "delivery is not one JSON value"
				return
			}
			if _, err = io.WriteString(conn, line+"\n"); err != nil {
				o.Delivery = "unknown"
				o.Error = err.Error()
				return
			}
		}
		if len(q.Deliver) > 0 {
			o.Delivery = "written"
		}
	}
	if q.Op == "inspect" && len(q.Deliver) == 0 {
		o.Error = "inspect needs a protocol request"
		return
	}
	o.Read = "observed"
	// One exact line per invocation. No bufio reader may consume bytes that
	// belong to the next operation child.
	partial := []byte(q.Prefix)
	var one [1]byte
	for len(partial) < 1<<20 {
		n, readErr := conn.Read(one[:])
		if n == 1 {
			partial = append(partial, one[0])
			if one[0] == '\n' {
				o.Lines = append(o.Lines, string(partial))
				return
			}
		}
		if readErr != nil {
			o.Partial = string(partial)
			o.PeerEOF = errors.Is(readErr, io.EOF)
			if !o.PeerEOF {
				o.Error = readErr.Error()
			}
			return
		}
	}
	o.Partial, o.Error = string(partial), "line exceeds 1 MiB"
	return
}
