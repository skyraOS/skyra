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
	"math"
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
	Call *call  `json:"call"` // role "call": a tool call in the agent's turn
}

// call is a tool call, as the webhost reads it from the lineage.
type call struct {
	To     string `json:"to"`
	Arg    string `json:"arg"`
	State  string `json:"state"` // running | returned | fault
	Result string `json:"result"`
	Fault  string `json:"fault"`
	Cause  string `json:"cause"`
	Stamp  string `json:"stamp"`
	MS     *int64 `json:"ms"`
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
	kind   string
	sock   *socket
	data   []byte
	key    string
	err    error
	button int // mouse: the SGR button code
	col    int // mouse: 1-based cell
	row    int
}

type app struct {
	sock          *socket
	connected     bool
	incarnation   string
	revision      int
	conversations []conversation
	selected      string
	input         []rune
	caret         int // position in input, in runes
	opening       bool
	openBefore    string
	pendingText   string
	scroll        int // transcript lines above the bottom
	wallFrame     uint64
	screen        [][]pixel // what the terminal shows now
	sim           *skySim   // the satellites, which the pointer can pull
	mouseOff      bool      // Ctrl+T turns pointer tracking off and back on
	pointer       *point    // the mouse, in braille dots; nil when unknown
	titleOpen     float64   // 1 on an empty session, easing to 0 in a conversation
	calls         callLights
	onScreen      string            // the conversation whose calls the sky follows
	seenCalls     map[string]string // each call's state when last seen, by conversation and stamp
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
	fmt.Print("\x1b[?1049h\x1b[?25l\x1b[2J" + mouseOn)
	defer fmt.Print(mouseOffSeq + "\x1b[0m\x1b[?25h\x1b[?1049l")

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
			a.stepSky()
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
			case "mouse":
				if !a.mouse(e) {
					continue // pointer motion shows on the next tick
				}
			case "focus-out":
				a.pointer = nil
				continue
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

// Any-motion mouse reports in SGR form, and focus reports so the pointer is
// forgotten when the window loses focus.
const (
	mouseOn     = "\x1b[?1003h\x1b[?1006h\x1b[?1004h"
	mouseOffSeq = "\x1b[?1003l\x1b[?1006l\x1b[?1004l"
)

// readMouse reads the rest of an SGR mouse report, "<b;col;row" then M or m.
func readMouse(r *bufio.Reader) (event, bool) {
	var buf []byte
	for len(buf) < 32 {
		c, err := r.ReadByte()
		if err != nil {
			return event{}, false
		}
		if c == 'M' || c == 'm' {
			parts := strings.Split(string(buf), ";")
			if len(parts) != 3 {
				return event{}, false
			}
			b, e1 := strconv.Atoi(parts[0])
			col, e2 := strconv.Atoi(parts[1])
			row, e3 := strconv.Atoi(parts[2])
			if e1 != nil || e2 != nil || e3 != nil {
				return event{}, false
			}
			return event{kind: "mouse", button: b, col: col, row: row}, true
		}
		buf = append(buf, c)
	}
	return event{}, false
}

// mouse handles a mouse report: the wheel scrolls the transcript, and any
// report places the pointer. It reports whether the screen needs a redraw.
func (a *app) mouse(e event) bool {
	if a.mouseOff {
		return false
	}
	a.pointer = &point{float64((e.col-1)*2 + 1), float64((e.row-1)*4 + 2)}
	switch e.button &^ 32 { // 32 marks motion
	case 64:
		a.scroll += 3
		return true
	case 65:
		a.scroll = max(0, a.scroll-3)
		return true
	}
	return false
}

// titleFade is how far the title eases each frame: two seconds end to end.
const titleFade = 0.05

// stepSky moves the satellites one frame, toward the pointer when it is near,
// and eases the title toward open or receded.
func (a *app) stepSky() {
	target := 0.0
	if c := a.current(); c == nil || len(c.Entries) == 0 {
		target = 1
	}
	if a.titleOpen < target {
		a.titleOpen = math.Min(target, a.titleOpen+titleFade)
	} else {
		a.titleOpen = math.Max(target, a.titleOpen-titleFade)
	}

	if a.sim == nil || a.sim.width != a.width || a.sim.height != a.height {
		a.sim = newSkySim(a.width, a.height, a.wallFrame)
		return
	}
	a.sim.step(a.wallFrame, a.pointer)
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
		e, err := readEvent(r)
		if err != nil {
			events <- event{kind: "input-error", err: err}
			return
		}
		if e.kind != "" {
			events <- e
		}
	}
}

// readEvent reads one key, mouse report or focus change. Escape sequences
// are always read to their end, so an unknown one is dropped whole rather
// than leaking characters into the draft. An empty kind means nothing to do.
func readEvent(r *bufio.Reader) (event, error) {
	b, err := r.ReadByte()
	if err != nil {
		return event{}, err
	}
	switch {
	case b == 0x1b:
		return readEscape(r), nil
	case b >= utf8.RuneSelf:
		buf := []byte{b}
		for !utf8.FullRune(buf) && len(buf) < 4 {
			c, err := r.ReadByte()
			if err != nil {
				break
			}
			buf = append(buf, c)
		}
		return event{kind: "key", key: string(buf)}, nil
	}
	return event{kind: "key", key: string(b)}, nil
}

func readEscape(r *bufio.Reader) event {
	second, err := r.ReadByte()
	if err != nil {
		return event{}
	}
	key := func(k string) event {
		if k == "" {
			return event{} // a key this client does not use
		}
		return event{kind: "key", key: k}
	}
	switch second {
	case 'b':
		return key("word-left") // Option+Left in macOS Terminal
	case 'f':
		return key("word-right")
	case 'O': // SS3: arrows in application mode
		third, err := r.ReadByte()
		if err != nil {
			return event{}
		}
		return key(map[byte]string{'A': "up", 'B': "down", 'C': "right", 'D': "left", 'H': "home", 'F': "end"}[third])
	case '[':
	default:
		return event{}
	}
	first, err := r.ReadByte()
	if err != nil {
		return event{}
	}
	if first == '<' {
		if e, ok := readMouse(r); ok {
			return e
		}
		return event{}
	}
	// A CSI sequence: parameter bytes, then one final byte.
	var params []byte
	final := first
	for final >= 0x30 && final <= 0x3f && len(params) < 16 {
		params = append(params, final)
		if final, err = r.ReadByte(); err != nil {
			return event{}
		}
	}
	p := string(params)
	modified := strings.Contains(p, ";") // Ctrl, Option or Shift held
	switch final {
	case 'A':
		return key("up")
	case 'B':
		return key("down")
	case 'C':
		if modified {
			return key("word-right")
		}
		return key("right")
	case 'D':
		if modified {
			return key("word-left")
		}
		return key("left")
	case 'H':
		return key("home")
	case 'F':
		return key("end")
	case 'O':
		return event{kind: "focus-out"}
	case '~':
		return key(map[string]string{"1": "home", "7": "home", "4": "end", "8": "end", "3": "delete", "5": "pageup", "6": "pagedown"}[strings.Split(p, ";")[0]])
	}
	return event{} // focus in, and anything else
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
			a.setInput([]rune(a.pendingText))
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
		a.followCalls()
	case "rejected":
		a.notice = m.Code + ": " + m.Message
		if a.opening {
			a.setInput([]rune(a.pendingText))
			a.pendingText, a.opening = "", false
		}
	case "accepted":
	default:
		a.notice = "unknown server message"
	}
}

// followCalls launches a light for each call that appeared in the
// conversation on screen, and brings home each one that stopped running. A
// conversation's calls that were there when it came on screen launch nothing.
func (a *app) followCalls() {
	if a.seenCalls == nil {
		a.seenCalls = map[string]string{}
	}
	id := a.selected
	quiet := id != a.onScreen
	a.onScreen = id
	c := a.current()
	if c == nil {
		return
	}
	f := float64(a.wallFrame)
	for _, e := range c.Entries {
		if e.Call == nil {
			continue
		}
		key := id + "/" + e.Call.Stamp
		before, seen := a.seenCalls[key]
		a.seenCalls[key] = e.Call.State
		switch {
		case quiet:
		case !seen:
			a.calls.launch(key, e.Call.State, f)
		case before == "running" && e.Call.State != "running":
			a.calls.settle(key, f)
		}
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
			a.pendingText = text
			a.setInput(nil)
		}
		return
	}
	before := ""
	if len(a.conversations) > 0 {
		before = a.conversations[0].ID
	}
	if a.send(map[string]string{"type": "open-conversation", "requestId": requestID()}) {
		a.opening, a.openBefore, a.pendingText = true, before, text
		a.setInput(nil)
		a.notice = "opening a conversation…"
	}
}

// setInput replaces the draft and puts the caret at its end.
func (a *app) setInput(r []rune) {
	a.input, a.caret = r, len(r)
}

// textWidth is how many characters fit on one line of the composer.
func (a *app) textWidth() int {
	return max(1, a.width-8)
}

// wordLeft and wordRight are where Ctrl or Option with an arrow lands: the
// start of the previous word, or just past the end of the next.
func (a *app) wordLeft() int {
	i := a.caret
	for i > 0 && unicode.IsSpace(a.input[i-1]) {
		i--
	}
	for i > 0 && !unicode.IsSpace(a.input[i-1]) {
		i--
	}
	return i
}

func (a *app) wordRight() int {
	i := a.caret
	for i < len(a.input) && unicode.IsSpace(a.input[i]) {
		i++
	}
	for i < len(a.input) && !unicode.IsSpace(a.input[i]) {
		i++
	}
	return i
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
			a.setInput(nil)
		} else if a.quitArmed {
			return true
		} else {
			a.quitArmed, a.notice = true, quitNotice
		}
	case "\x14": // Ctrl+T: pointer tracking off and on
		a.mouseOff = !a.mouseOff
		if a.mouseOff {
			a.pointer = nil
			fmt.Print(mouseOffSeq)
			a.notice = "mouse off: the terminal selects text again"
		} else {
			fmt.Print(mouseOn)
			a.notice = "mouse on: satellites follow the pointer"
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
	case "left":
		a.caret = max(0, a.caret-1)
	case "right":
		a.caret = min(len(a.input), a.caret+1)
	case "word-left":
		a.caret = a.wordLeft()
	case "word-right":
		a.caret = a.wordRight()
	case "home", "\x01": // Home, Ctrl+A
		a.caret = 0
	case "end", "\x05": // End, Ctrl+E
		a.caret = len(a.input)
	case "up": // up a line of a long draft, else scroll the transcript
		if a.caret >= a.textWidth() {
			a.caret -= a.textWidth()
		} else {
			a.scroll++
		}
	case "down":
		if a.caret/a.textWidth() < len(a.input)/a.textWidth() {
			a.caret = min(len(a.input), a.caret+a.textWidth())
		} else {
			a.scroll = max(0, a.scroll-1)
		}
	case "\x04": // Ctrl+D: close
		c := a.current()
		switch {
		case c == nil:
			a.notice = "no session to close"
		case c.Phase == "ready" || c.Phase == "failed":
			a.send(map[string]string{"type": "close-interaction", "uiIncarnation": a.incarnation, "conversation": c.ID, "requestId": requestID()})
		case c.Phase == "closed":
			a.notice = "this session is already closed"
		default:
			a.notice = "agent working; close the session when it finishes"
		}
	case "\x15":
		a.setInput(nil) // Ctrl+U
	case "\x7f", "\b":
		if a.caret > 0 {
			a.input = append(a.input[:a.caret-1], a.input[a.caret:]...)
			a.caret--
		}
	case "delete":
		if a.caret < len(a.input) {
			a.input = append(a.input[:a.caret], a.input[a.caret+1:]...)
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
				a.setInput(nil)
				a.scroll = 0
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
		r, size := utf8.DecodeRuneInString(k)
		if r != utf8.RuneError && size == len(k) && unicode.IsPrint(r) {
			a.caret = min(a.caret, len(a.input))
			a.input = append(a.input[:a.caret], append([]rune{r}, a.input[a.caret:]...)...)
			a.caret++
		}
	}
	return false
}
