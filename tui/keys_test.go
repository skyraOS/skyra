package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestEscapeSequencesAreReadWhole(t *testing.T) {
	for _, tc := range []struct{ in, kind, key string }{
		{"\x1b[A", "key", "up"},
		{"\x1b[B", "key", "down"},
		{"\x1b[C", "key", "right"},
		{"\x1b[D", "key", "left"},
		{"\x1bOA", "key", "up"}, // application cursor mode
		{"\x1bOD", "key", "left"},
		{"\x1b[1;5C", "key", "word-right"}, // Ctrl+Right
		{"\x1b[1;3D", "key", "word-left"},  // Option+Left
		{"\x1bb", "key", "word-left"},      // Option+Left, macOS Terminal
		{"\x1bf", "key", "word-right"},
		{"\x1b[H", "key", "home"},
		{"\x1b[F", "key", "end"},
		{"\x1b[1~", "key", "home"},
		{"\x1b[4~", "key", "end"},
		{"\x1b[3~", "key", "delete"},
		{"\x1b[5~", "key", "pageup"},
		{"\x1b[6~", "key", "pagedown"},
		{"\x1b[O", "focus-out", ""},
		{"\x1b[I", "", ""},    // focus in
		{"\x1b[200~", "", ""}, // paste start
		{"\x1b[15~", "", ""},  // F5
		{"\x1b[1;2P", "", ""}, // Shift+F1
		{"\x1b[<35;4;5M", "mouse", ""},
	} {
		// Each sequence is followed by "x": it must arrive as the next key,
		// proving nothing of the sequence was left behind.
		r := bufio.NewReader(strings.NewReader(tc.in + "x"))
		e, err := readEvent(r)
		if err != nil || e.kind != tc.kind || e.key != tc.key {
			t.Errorf("%q: got %q %q %v, want %q %q", tc.in, e.kind, e.key, err, tc.kind, tc.key)
			continue
		}
		if next, _ := readEvent(r); next.key != "x" {
			t.Errorf("%q left %q behind", tc.in, next.key)
		}
	}
}

func typeText(a *app, s string) {
	for _, r := range s {
		a.key(string(r))
	}
}

func TestEditingHappensAtTheCaret(t *testing.T) {
	a := app{width: 80, height: 24}
	typeText(&a, "helo world")
	for i := 0; i < 7; i++ { // to between "hel" and "o"
		a.key("left")
	}
	typeText(&a, "l")
	if got := string(a.input); got != "hello world" {
		t.Fatalf("insert at caret: %q", got)
	}
	a.key("home")
	typeText(&a, ">")
	a.key("end")
	a.key("\x7f")
	if got := string(a.input); got != ">hello worl" {
		t.Fatalf("home, end and backspace: %q", got)
	}
	a.key("home")
	a.key("delete")
	if got := string(a.input); got != "hello worl" || a.caret != 0 {
		t.Fatalf("delete: %q caret %d", got, a.caret)
	}
	a.key("word-right")
	if a.caret != 5 {
		t.Fatalf("word right landed at %d", a.caret)
	}
	a.key("end")
	a.key("word-left")
	if a.caret != 6 {
		t.Fatalf("word left landed at %d", a.caret)
	}
	a.key("\x01") // Ctrl+A
	a.key("left")
	if a.caret != 0 {
		t.Fatal("left moved past the start")
	}
	a.key("\x05") // Ctrl+E
	a.key("right")
	if a.caret != len(a.input) {
		t.Fatal("right moved past the end")
	}
}

func TestNavigationKeysNeverType(t *testing.T) {
	a := app{width: 80, height: 24}
	typeText(&a, "draft")
	for _, k := range []string{"up", "down", "left", "right", "word-left", "word-right", "home", "end", "pageup", "pagedown"} {
		a.key(k)
	}
	if got := string(a.input); got != "draft" {
		t.Fatalf("navigation changed the draft to %q", got)
	}
}

func TestUpAndDownMoveThroughALongDraftThenScroll(t *testing.T) {
	a := app{width: 80, height: 24} // 72 characters to a composer line
	typeText(&a, strings.Repeat("a", 100))
	a.key("up")
	if a.caret != 100-a.textWidth() || a.scroll != 0 {
		t.Fatalf("up in a two-line draft: caret %d scroll %d", a.caret, a.scroll)
	}
	a.key("up")
	if a.scroll != 1 {
		t.Fatal("up on the first line did not scroll the transcript")
	}
	a.key("down")
	if a.caret != 100 || a.scroll != 1 {
		t.Fatalf("down from the first line: caret %d scroll %d", a.caret, a.scroll)
	}
	a.key("down")
	if a.scroll != 0 {
		t.Fatal("down on the last line did not scroll the transcript back")
	}
}

func TestTheScreenCursorFollowsTheCaret(t *testing.T) {
	a := app{width: 80, height: 24}
	typeText(&a, "hello")
	a.compose()
	end := a.cursorCol
	a.key("left")
	a.key("left")
	a.compose()
	if a.cursorCol != end-2 {
		t.Fatalf("cursor at column %d, want %d", a.cursorCol, end-2)
	}
}

func TestScrollStopsAtTheTop(t *testing.T) {
	a := app{width: 80, height: 24, selected: "c", conversations: []conversation{{ID: "c", Phase: "ready",
		Entries: []entry{{Role: "user", Text: "hi"}, {Role: "agent", Text: "hello"}}}}}
	for i := 0; i < 10; i++ {
		a.key("pageup")
	}
	a.compose()
	if a.scroll != 0 { // the whole transcript fits
		t.Fatalf("scrolled %d lines past the top", a.scroll)
	}
}

// Every key the footer lists must do something.
func TestFooterKeysWork(t *testing.T) {
	convo := func() app {
		return app{width: 80, height: 24, selected: "c", conversations: []conversation{{ID: "c", Phase: "ready",
			Entries: []entry{{Role: "user", Text: strings.Repeat("line\n", 40)}}}}}
	}

	a := convo()
	a.key("\x0e") // ^n new
	if a.selected != "" {
		t.Error("^n did not start a new session")
	}

	a = convo()
	a.key("pageup")
	if a.scroll == 0 {
		t.Error("pgup did not scroll")
	}
	a.key("pagedown")
	if a.scroll != 0 {
		t.Error("pgdn did not scroll back")
	}

	a = convo()
	typeText(&a, "draft")
	a.key("\x03") // ^c clear
	if len(a.input) != 0 || len(a.current().Entries) != 1 {
		t.Error("^c did not clear the draft alone")
	}

	a = app{width: 80, height: 24}
	a.key("\x04") // ^d close, with nothing to close
	if a.notice == "" {
		t.Error("^d with no session said nothing")
	}
	a = convo()
	a.conversations[0].Phase = "pending"
	a.key("\x04")
	if a.notice == "" {
		t.Error("^d during a turn said nothing")
	}

	a = convo()
	a.key("\x14") // ^t mouse
	if !a.mouseOff {
		t.Error("^t did not turn the mouse off")
	}
	a.key("\x14")
	if a.mouseOff {
		t.Error("^t did not turn the mouse back on")
	}
	// enter is covered by TestFirstMessageOpensFreshSessionAndSends, and
	// ^c ^c by TestCtrlCClearsTheComposerAndTwiceQuits.
}
