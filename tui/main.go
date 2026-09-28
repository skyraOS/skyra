// Command tui is a terminal client of the webhost's conversation protocol.
// It holds no runtime authority and never reads or writes a lineage directly.
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
)

type entry struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type conversation struct {
	ID      string  `json:"conversation"`
	Phase   string  `json:"phase"`
	Entries []entry `json:"entries"`
}

type message struct {
	Type          string         `json:"type"`
	UIIncarnation string         `json:"uiIncarnation"`
	Revision      int            `json:"revision"`
	Conversations []conversation `json:"conversations"`
	Code          string         `json:"code"`
	Message       string         `json:"message"`
}

type event struct {
	kind string
	sock *socket
	data []byte
	key  string
	err  error
}

type app struct {
	sock          *socket
	connected     bool
	incarnation   string
	revision      int
	conversations []conversation
	selected      string
	input         []rune
	opening       bool
	openBefore    string
	pendingText   string
	scroll        int // transcript lines above the bottom
	wallFrame     uint64
	screen        [][]pixel // what the terminal shows now
	cursorRow     int
	cursorCol     int
	notice        string
	quitArmed     bool // Ctrl+C on an empty composer; a second press quits
	width, height int
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "webhost address")
	flag.Parse()
	if err := run(*addr); err != nil {
		fmt.Fprintln(os.Stderr, "skyra tui:", err)
		os.Exit(1)
	}
}

func run(addr string) error {
	if !isTerminal() {
		return errors.New("run this command in a terminal")
	}
	old, err := stty("-g")
	if err != nil {
		return err
	}
	if _, err := stty("raw", "-echo"); err != nil {
		return err
	}
	defer func() { _, _ = stty(strings.TrimSpace(old)) }()
	fmt.Print("\x1b[?1049h\x1b[?25l\x1b[2J")
	defer fmt.Print("\x1b[0m\x1b[?25h\x1b[?1049l")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan event, 64)
	go socketLoop(ctx, addr, events)
	go keyLoop(events)
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGWINCH, syscall.SIGTERM)
	defer signal.Stop(signals)
	a := app{revision: -1}
	a.resize()
	a.draw()
	wallTicker := time.NewTicker(100 * time.Millisecond)
	defer wallTicker.Stop()
	for {
		select {
		case <-wallTicker.C:
			a.wallFrame++
			a.tick()
		case e := <-events:
			switch e.kind {
			case "connected":
				a.sock, a.connected, a.notice = e.sock, true, ""
			case "disconnected":
				a.sock, a.connected = nil, false
				if e.err != nil {
					a.notice = "connection lost; retrying"
				}
			case "message":
				a.receive(e.data)
			case "key":
				if a.key(e.key) {
					return nil
				}
			case "input-error":
				return e.err
			}
			a.draw()
		case sig := <-signals:
			if sig == syscall.SIGTERM {
				return nil
			}
			a.resize()
			a.draw()
		}
	}
}

func isTerminal() bool {
	var stat syscall.Stat_t
	if err := syscall.Fstat(int(os.Stdin.Fd()), &stat); err != nil {
		return false
	}
	return stat.Mode&syscall.S_IFMT == syscall.S_IFCHR
}

func stty(args ...string) (string, error) {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = os.Stdin
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("stty %s: %w: %s", strings.Join(args, " "), err, b)
	}
	return string(b), nil
}

func (a *app) resize() {
	b, err := stty("size")
	if err != nil {
		a.width, a.height = 80, 24
		return
	}
	parts := strings.Fields(b)
	if len(parts) != 2 {
		a.width, a.height = 80, 24
		return
	}
	a.height, _ = strconv.Atoi(parts[0])
	a.width, _ = strconv.Atoi(parts[1])
	if a.width < 20 {
		a.width = 20
	}
	if a.height < 8 {
		a.height = 8
	}
}

func socketLoop(ctx context.Context, addr string, events chan<- event) {
	for ctx.Err() == nil {
		s, err := connect(addr)
		if err == nil {
			select {
			case events <- event{kind: "connected", sock: s}:
			case <-ctx.Done():
				s.close()
				return
			}
			for ctx.Err() == nil {
				b, e := s.read()
				if e != nil {
					err = e
					break
				}
				select {
				case events <- event{kind: "message", data: b}:
				case <-ctx.Done():
					s.close()
					return
				}
			}
			s.close()
		}
		select {
		case events <- event{kind: "disconnected", err: err}:
		case <-ctx.Done():
			return
		}
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			return
		}
	}
}

func keyLoop(events chan<- event) {
	r := bufio.NewReader(os.Stdin)
	for {
		b, err := r.ReadByte()
		if err != nil {
			events <- event{kind: "input-error", err: err}
			return
		}
		key := string(b)
		if b == 0x1b {
			second, err := r.ReadByte()
			if err != nil {
				continue
			}
			if second != '[' {
				continue
			}
			third, err := r.ReadByte()
			if err != nil {
				continue
			}
			switch third {
			case 'A':
				key = "up"
			case 'B':
				key = "down"
			case '5', '6':
				_, _ = r.ReadByte() // trailing ~
				if third == '5' {
					key = "pageup"
				} else {
					key = "pagedown"
				}
			default:
				continue
			}
		} else if b >= utf8.RuneSelf {
			buf := []byte{b}
			for !utf8.FullRune(buf) && len(buf) < 4 {
				c, e := r.ReadByte()
				if e != nil {
					break
				}
				buf = append(buf, c)
			}
			key = string(buf)
		}
		events <- event{kind: "key", key: key}
	}
}

func (a *app) receive(raw []byte) {
	var m message
	if err := json.Unmarshal(raw, &m); err != nil {
		a.notice = "invalid server message"
		return
	}
	switch m.Type {
	case "snapshot":
		if m.UIIncarnation == a.incarnation && m.Revision < a.revision {
			return
		}
		if a.incarnation != "" && m.UIIncarnation != a.incarnation && a.opening {
			a.input = []rune(a.pendingText)
			a.pendingText, a.opening = "", false
			a.notice = "host restarted; draft kept, press Enter to send"
		}
		a.incarnation, a.revision, a.conversations = m.UIIncarnation, m.Revision, m.Conversations
		if a.opening && len(a.conversations) > 0 && a.conversations[0].ID != a.openBefore {
			newest := a.conversations[0]
			a.selected, a.scroll = newest.ID, 0
			if newest.Phase == "ready" {
				if a.pendingText != "" {
					if !a.send(map[string]string{"type": "submit", "uiIncarnation": a.incarnation, "conversation": newest.ID, "requestId": requestID(), "text": a.pendingText}) {
						return
					}
				}
				a.pendingText, a.opening = "", false
				a.notice = ""
			}
			return
		}
		if a.selected != "" && a.current() == nil {
			a.selected, a.scroll = "", 0
		}
	case "rejected":
		a.notice = m.Code + ": " + m.Message
		if a.opening {
			a.input = []rune(a.pendingText)
			a.pendingText, a.opening = "", false
		}
	case "accepted":
	default:
		a.notice = "unknown server message"
	}
}

func (a *app) current() *conversation {
	for i := range a.conversations {
		if a.conversations[i].ID == a.selected {
			return &a.conversations[i]
		}
	}
	return nil
}

func requestID() string {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return fmt.Sprintf("tui-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(id[:])
}

func (a *app) send(v any) bool {
	if !a.connected || a.sock == nil {
		a.notice = "disconnected"
		return false
	}
	if err := a.sock.send(v); err != nil {
		a.notice = "send failed: " + err.Error()
		return false
	}
	return true
}

func (a *app) startNew(text string) {
	if !a.connected || a.incarnation == "" {
		a.notice = "connecting; keep typing, then press Enter"
		return
	}
	if a.opening {
		if a.pendingText == "" && text != "" {
			a.pendingText, a.input = text, nil
		}
		return
	}
	before := ""
	if len(a.conversations) > 0 {
		before = a.conversations[0].ID
	}
	if a.send(map[string]string{"type": "open-conversation", "requestId": requestID()}) {
		a.opening, a.openBefore, a.pendingText = true, before, text
		a.input = nil
		a.notice = "opening a conversation…"
	}
}

const quitNotice = "press ctrl+c again to quit"

func (a *app) key(k string) bool {
	if k != "\x03" && a.quitArmed {
		a.quitArmed = false
		if a.notice == quitNotice {
			a.notice = ""
		}
	}
	switch k {
	case "\x03": // Ctrl+C: clear the composer; on an empty one, twice quits
		if len(a.input) > 0 {
			a.input = nil
		} else if a.quitArmed {
			return true
		} else {
			a.quitArmed, a.notice = true, quitNotice
		}
	case "\x0e": // Ctrl+N
		if a.opening {
			a.notice = "opening a session; wait for it to finish"
		} else {
			a.selected, a.scroll, a.notice = "", 0, ""
		}
	case "pageup":
		a.scroll += max(3, a.height/2)
	case "pagedown":
		a.scroll = max(0, a.scroll-max(3, a.height/2))
	case "\x04": // Ctrl+D: close
		if c := a.current(); c != nil && (c.Phase == "ready" || c.Phase == "failed") {
			a.send(map[string]string{"type": "close-interaction", "uiIncarnation": a.incarnation, "conversation": c.ID, "requestId": requestID()})
		}
	case "\x15":
		a.input = nil // Ctrl+U
	case "\x7f", "\b":
		if len(a.input) > 0 {
			a.input = a.input[:len(a.input)-1]
		}
	case "\r", "\n":
		c := a.current()
		text := strings.TrimSpace(string(a.input))
		if text == "" {
			break
		}
		if !a.connected || a.incarnation == "" {
			a.notice = "connecting; your draft is kept"
			break
		}
		if a.opening {
			a.startNew(text)
		} else if c != nil && c.Phase == "ready" && a.connected && a.incarnation != "" {
			if a.send(map[string]string{"type": "submit", "uiIncarnation": a.incarnation, "conversation": c.ID, "requestId": requestID(), "text": text}) {
				a.input, a.scroll = nil, 0
			}
		} else if c == nil || c.Phase == "closed" || c.Phase == "failed" {
			a.startNew(text)
		} else {
			a.notice = "agent working; your draft is kept"
		}
	default:
		if len(a.input) >= 10000 {
			a.notice = "message limit reached"
			break
		}
		r, _ := utf8.DecodeRuneInString(k)
		if r != utf8.RuneError && unicode.IsPrint(r) {
			a.input = append(a.input, r)
		}
	}
	return false
}
