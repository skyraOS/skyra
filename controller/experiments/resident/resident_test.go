//go:build darwin

package resident

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"skyra/controller/handshake"
)

// TestResidentHelper is re-executed as real, separate runtime, provider, and
// operation processes. Only the provider/operation/tool helpers are sandboxed.
func TestResidentHelper(t *testing.T) {
	mode := os.Getenv(helperEnv)
	if mode == "" {
		return
	}
	if err := helper(mode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0) // do not write the testing package's PASS onto provider stdout
}

func helper(mode string) error {
	switch mode {
	case "provider":
		fmt.Println("ready")
		for {
			line, err := exactLine(os.Stdin, "")
			if err == io.EOF {
				fmt.Println("bye")
				return nil
			}
			if err != nil {
				return err
			}
			switch strings.TrimSuffix(line, "\n") {
			case "burst":
				// All three lines may enter the socket in one write. Only the
				// first is read before the tool and parking handoffs.
				fmt.Print("tool:echo\nqueued:one\nqueued:two\n")
			case "partial":
				fmt.Print("part")
			case "finish":
				fmt.Print("ial\n")
			default:
				fmt.Print("echo:" + line)
			}
		}
	case "operation":
		var q request
		if err := json.NewDecoder(os.Stdin).Decode(&q); err != nil {
			return err
		}
		f := os.NewFile(3, "granted-connection")
		defer f.Close()
		return json.NewEncoder(os.Stdout).Encode(exchange(f, q))
	case "tool":
		// No resource from the custodian may leak through exec. A directory
		// fd would bypass the intended scope even without a connection fd.
		for fd := 3; fd < 256; fd++ {
			var st syscall.Stat_t
			if syscall.Fstat(fd, &st) == nil && (st.Mode&syscall.S_IFMT == syscall.S_IFSOCK || st.Mode&syscall.S_IFMT == syscall.S_IFDIR) {
				return fmt.Errorf("tool inherited resource fd %d", fd)
			}
		}
		floor, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		if _, err := os.ReadDir(string(floor)); err == nil {
			return fmt.Errorf("tool could read protected floor")
		}
		// The child cannot launch another sandbox to give itself authority.
		if err := exec.Command("/usr/bin/sandbox-exec", "-p", "(version 1)(allow default)", "/bin/echo", "escaped").Run(); err == nil {
			return fmt.Errorf("tool launched a nested sandbox")
		}
		fmt.Print("tool-ok")
		return nil
	}

	c, err := handshake.Conn(os.NewFile(3, "custody-control"))
	if err != nil {
		return err
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(10 * time.Second))
	var floor string
	if _, err := handshake.Recv(c, &floor); err != nil {
		return err
	}
	if mode == "launch" {
		b, err := launch(floor)
		if err != nil {
			return err
		}
		defer b.close()
		return transfer(c, &b)
	}
	if mode != "holder" {
		return fmt.Errorf("unknown experiment helper %q", mode)
	}
	b, err := receive(c)
	if err != nil {
		return err
	}
	defer b.close()
	if err := handshake.Send(c, "accepted"); err != nil {
		return err
	}
	var q request
	if _, err := handshake.Recv(c, &q); err != nil {
		return err
	}
	var o observation
	switch q.Op {
	case "park":
		// Stand-in for the ingress custodian: no provider reads, no timer
		// policy, and no provider semantics. Outside input releases it.
		if err := handshake.Send(c, "parked"); err != nil {
			return err
		}
		var event string
		if _, err := handshake.Recv(c, &event); err != nil || event != "resume" {
			return fmt.Errorf("park resume: %q, %v", event, err)
		}
		o = observation{Delivery: "not_sent", Read: "not_started", ObservedAt: time.Now().UnixMilli()}
	case "tool":
		cmd, _, err := helperCommand("tool", floor, true)
		if err != nil {
			return err
		}
		cmd.Stdin = strings.NewReader(floor)
		out, err := cmd.Output() // no ExtraFiles: custody is not a target grant
		if err != nil || string(out) != "tool-ok" {
			return fmt.Errorf("isolated tool: %q, %v", out, err)
		}
		o = observation{Lines: []string{string(out)}, Delivery: "not_sent", Read: "not_started", ObservedAt: time.Now().UnixMilli()}
	default:
		o = operate(b, q, floor)
	}
	if err := handshake.Send(c, o); err != nil {
		return err
	}
	return transfer(c, &b)
}

type running struct {
	conn *net.UnixConn
	cmd  *exec.Cmd
}

func startHelper(t *testing.T, mode, floor string) running {
	t.Helper()
	c, child, err := handshake.Pair()
	if err != nil {
		t.Fatal(err)
	}
	cmd, _, err := helperCommand(mode, floor, false)
	if err != nil {
		t.Fatal(err)
	}
	cmd.ExtraFiles = []*os.File{child}
	if err := cmd.Start(); err != nil {
		child.Close()
		c.Close()
		t.Fatal(err)
	}
	child.Close()
	c.SetDeadline(time.Now().Add(10 * time.Second))
	t.Cleanup(func() {
		c.Close()
		if cmd.ProcessState == nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	})
	if err := handshake.Send(c, floor); err != nil {
		t.Fatal(err)
	}
	return running{c, cmd}
}

func acceptBundle(t *testing.T, r running) bundle {
	t.Helper()
	b, err := receive(r.conn)
	if err != nil {
		t.Fatal(err)
	}
	if err := handshake.Send(r.conn, "accepted"); err != nil {
		b.close()
		t.Fatal(err)
	}
	if err := r.cmd.Wait(); err != nil {
		b.close()
		t.Fatalf("custodian failed: %v", err)
	}
	r.conn.Close()
	return b
}

func startResident(t *testing.T) (*bundle, string) {
	t.Helper()
	floor, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	b := acceptBundle(t, startHelper(t, "launch", floor))
	if b.PID == 0 || b.Profile == "" {
		t.Fatal("start did not identify its resident and rendered profile")
	}
	t.Cleanup(func() {
		// On failure as well as success: EOF on all endpoint copies, never
		// kill by a PID which may since have been reused.
		for i, r := range b.Resources {
			if r.Kind == "connection" && i < len(b.Files) {
				syscall.Shutdown(int(b.Files[i].Fd()), syscall.SHUT_WR)
			}
		}
		b.close()
	})
	return &b, floor
}

func requestFor(b *bundle, op string, lines int, deliver ...string) request {
	return request{Op: op, Owner: "conversation", Incarnation: b.Resources[1].Incarnation,
		Deliver: deliver, ReadLines: lines, WaitMS: 2000}
}

func cycle(t *testing.T, b *bundle, floor string, q request) observation {
	t.Helper()
	r := startHelper(t, "holder", floor)
	if err := transfer(r.conn, b); err != nil {
		t.Fatal(err)
	}
	// The old custodian now holds zero resources; the helper is the only
	// holder, including while parked and while its operation child exits.
	if len(b.Files) != 0 {
		t.Fatal("sender kept resource copies after transfer")
	}
	if err := handshake.Send(r.conn, q); err != nil {
		t.Fatal(err)
	}
	if q.Op == "park" {
		var state string
		if _, err := handshake.Recv(r.conn, &state); err != nil || state != "parked" {
			t.Fatalf("park: %q, %v", state, err)
		}
		if err := handshake.Send(r.conn, "resume"); err != nil {
			t.Fatal(err)
		}
	}
	var o observation
	if _, err := handshake.Recv(r.conn, &o); err != nil {
		t.Fatal(err)
	}
	*b = acceptBundle(t, r)
	if o.ObservedAt == 0 {
		t.Fatal("operation returned no time observation")
	}
	return o
}

func wantLines(t *testing.T, o observation, lines ...string) {
	t.Helper()
	if o.Error != "" || !reflect.DeepEqual(o.Lines, lines) {
		t.Fatalf("observation = %+v, want lines %q", o, lines)
	}
}

func TestResidentCustody(t *testing.T) {
	j := newJournal(t)
	j.invoke(request{Op: "start"})
	b, floor := startResident(t)
	identity, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	j.returned(observation{Lines: []string{"ready\n", string(identity)}, Delivery: "not_sent", Read: "observed", ObservedAt: time.Now().UnixMilli()})
	run := func(q request) observation {
		j.invoke(q) // commit the operands before the operation can run
		o := cycle(t, b, floor, q)
		j.returned(o) // commit the observation before the next request
		return o
	}
	pid, incarnation := b.PID, b.Resources[1].Incarnation
	o := run(requestFor(b, "exchange", 1, "burst"))
	wantLines(t, o, "tool:echo\n")
	o = run(requestFor(b, "tool", 0))
	wantLines(t, o, "tool-ok")
	run(requestFor(b, "park", 0))
	o = run(requestFor(b, "exchange", 2))
	wantLines(t, o, "queued:one\n", "queued:two\n")
	o = run(requestFor(b, "exchange", 1, "tool-result:hello"))
	wantLines(t, o, "echo:tool-result:hello\n")
	if b.PID != pid || b.Resources[1].Incarnation != incarnation {
		t.Fatal("handoff replaced resident identity")
	}
	o = run(requestFor(b, "stop", 0))
	wantLines(t, o, "bye\n")
	if !o.PeerEOF {
		t.Fatal("stop did not observe peer EOF with custodian copy still open")
	}
	// Shutdown must also revoke the retained descriptor's ability to write.
	c, err := net.FileConn(b.Files[1])
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := c.Write([]byte("after-stop\n")); err == nil {
		t.Fatal("retained endpoint still writable after stop")
	}
	j.verify()
	t.Log("start → exchange → isolated tool → park/resume → exchange → stop; launch and every custodian exited")
}

func TestStaleAndForeignGrantsDoNotTouchConnection(t *testing.T) {
	b, floor := startResident(t)
	for _, foreign := range []bool{false, true} {
		q := requestFor(b, "stop", 0)
		if foreign {
			q.Owner = "other-conversation"
		} else {
			q.Incarnation = "stale"
		}
		o := cycle(t, b, floor, q)
		if o.Delivery != "not_sent" || o.Error == "" {
			t.Fatalf("invalid grant was accepted: %+v", o)
		}
	}
	wantLines(t, cycle(t, b, floor, requestFor(b, "exchange", 1, "still-alive")), "echo:still-alive\n")
}

func TestPartialLineSurvivesWorkerExit(t *testing.T) {
	b, floor := startResident(t)
	q := requestFor(b, "exchange", 1, "partial")
	q.WaitMS = 100
	o := cycle(t, b, floor, q)
	if o.Partial != "part" || o.Error == "" || len(o.Lines) != 0 {
		t.Fatalf("partial timeout not preserved: %+v", o)
	}
	q = requestFor(b, "exchange", 1, "finish")
	q.Prefix = o.Partial
	wantLines(t, cycle(t, b, floor, q), "partial\n")
}

func TestWorkerDeathIsUncertainAndNeverRetried(t *testing.T) {
	for _, op := range []string{"die-after-write", "die-after-read"} {
		t.Run(op, func(t *testing.T) {
			b, floor := startResident(t)
			o := cycle(t, b, floor, requestFor(b, op, 1, "once"))
			if o.Delivery != "unknown" || o.Read != "unknown" || o.Error == "" {
				t.Fatalf("worker failure was not uncertain: %+v", o)
			}
			if op == "die-after-write" {
				wantLines(t, cycle(t, b, floor, requestFor(b, "exchange", 1)), "echo:once\n")
			}
			// A read-before-crash is unrecoverable here; the uncertainty must
			// be surfaced, not hidden by resubmitting the consumed request.
			wantLines(t, cycle(t, b, floor, requestFor(b, "exchange", 1, "next")), "echo:next\n")
		})
	}
}

func TestDroppingLastCustodySendsEOF(t *testing.T) {
	b, _ := startResident(t)
	pid := b.PID
	b.close() // no shutdown: ordinary last-close must make this provider exit
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); err == syscall.ESRCH {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("provider %d remained after its last connection holder closed", pid)
}
