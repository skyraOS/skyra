package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func snapshot(t *testing.T, a *app, revision int, entries ...entry) {
	t.Helper()
	raw, err := json.Marshal(message{Type: "snapshot", UIIncarnation: "i", Revision: revision,
		Conversations: []conversation{{ID: "c", Phase: "pending", Entries: entries}}})
	if err != nil {
		t.Fatal(err)
	}
	a.receive(raw)
}

func TestCallsLaunchOnlyWhenTheyAppear(t *testing.T) {
	a := &app{revision: -1, selected: "c", width: 120, height: 36}
	old := entry{Role: "call", Call: &call{To: "read", State: "returned", Stamp: "old"}}
	snapshot(t, a, 1, entry{Role: "user", Text: "hi"}, old)
	if len(a.calls.lights) != 0 {
		t.Fatal("a call already there when the conversation came on screen launched a light")
	}

	running := entry{Role: "call", Call: &call{To: "read", State: "running", Stamp: "new"}}
	snapshot(t, a, 2, entry{Role: "user", Text: "hi"}, old, running)
	if len(a.calls.lights) != 1 || a.calls.lights[0].home >= 0 {
		t.Fatalf("a new running call: %+v", a.calls.lights)
	}
	if !a.calls.busy(float64(a.wallFrame)) {
		t.Fatal("a light in flight does not quiet the sky")
	}

	a.wallFrame = 5
	returned := entry{Role: "call", Call: &call{To: "read", State: "returned", Stamp: "new"}}
	snapshot(t, a, 3, entry{Role: "user", Text: "hi"}, old, returned)
	if a.calls.lights[0].home != 5 {
		t.Fatalf("the returned call's light is not coming home: %+v", a.calls.lights[0])
	}
	if a.calls.busy(5 + callOut + callRide + callHome + 1) {
		t.Fatal("the light never landed")
	}
}

func TestRefusedCallLightsTheEdgeAndMeteorsWait(t *testing.T) {
	var c callLights
	c.launch("x", "fault", 0)
	arc := false
	c.draw(120, 36, callOutEdge+3, func(sx, sy int, fg string, _ uint8) {
		if p := parseRGB(fg); p[0] > p[2]+40 {
			arc = true // amber
		}
	})
	if !arc {
		t.Fatal("no arc at the edge after a refused call arrived")
	}
	// A meteor is crossing at frame 15; with a light out, it waits.
	busy := drawSky(120, 36, 15, nil, &c)
	quiet := drawSky(120, 36, 15, nil, nil)
	count := func(rows [][]wallCell) (n int) {
		for _, row := range rows {
			for _, cell := range row {
				if cell.priority == 3 {
					n++
				}
			}
		}
		return n
	}
	if count(quiet) == 0 || count(busy) != 0 {
		t.Fatalf("meteor cells: quiet %d, busy %d", count(quiet), count(busy))
	}
}

func TestLedgerFitsItsLine(t *testing.T) {
	ms := int64(412)
	for _, c := range []call{
		{To: "read", Arg: "README.md", State: "returned", Result: "returned 13.6 kB", Stamp: "4b7d45a7e", MS: &ms},
		{To: "native.commandExecution", State: "fault", Fault: "boundary", Stamp: "2ae3f9ec"},
		{To: "read", Arg: strings.Repeat("very/long/path/", 20), State: "running", Stamp: "abc"},
	} {
		for _, width := range []int{40, 80, 140} {
			n := 0
			var text strings.Builder
			for _, s := range ledger(c, width) {
				n += len([]rune(s.text))
				text.WriteString(s.text)
			}
			if n > width {
				t.Errorf("width %d: %d runes: %q", width, n, text.String())
			}
		}
	}
	var text strings.Builder
	for _, s := range ledger(call{To: "native.commandExecution", State: "fault", Fault: "boundary", Stamp: "2ae3f9ec"}, 80) {
		text.WriteString(s.text)
	}
	if !strings.Contains(text.String(), "outside the Boundary") {
		t.Fatalf("refused call reads %q", text.String())
	}
}
