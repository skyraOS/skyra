package main

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// The interface's colors: desktop/Sources/Skyra/Sky/Theme.swift and the
// frontend's tailwind.config.ts use the same values.
const (
	surface  = "8;10;14"
	userBG   = "15;20;29"
	operator = "96;196;166"
	accent   = "126;196;204"
	ink      = "214;219;229"
	dim      = "120;128;146"
	faint    = "60;66;80"
	red      = "226;112;112"
	amber    = "222;168;104" // a refused call: the Horizon doing its job
	callTo   = "169;201;206"
)

// line is one transcript row: text on a panel with a rail, or, for a tool
// call, a run of segments straight on the sky.
type line struct {
	text, bg, fg, rail string
	segs               []segment
}

// segment is one run of text in one color.
type segment struct{ fg, text string }

// How much of the sky each panel covers, as the desktop app draws them: the
// sky shows through, dimmed, so the whole screen is one background. Your
// messages sit on a panel of their own; the agent's on a veil of the sky.
const (
	userAlpha  = 0.8
	agentAlpha = 0.55
)

// pixel is one composed terminal cell.
type pixel struct {
	r      rune
	fg, bg [3]uint8
}

// canvas is the screen being composed: the sky first, then translucent
// panels and text on top.
type canvas struct {
	w, h       int
	titleLight float64 // how lit the title is by a satellite or meteor near it
	cells      [][]pixel
	fg         [][]rgb
	bg         [][]rgb
	marks      [][]rune
}

func parseRGB(s string) rgb {
	var c rgb
	fmt.Sscanf(s, "%f;%f;%f", &c[0], &c[1], &c[2])
	return c
}

func newCanvas(w, h int, frame uint64, sim *skySim, calls *callLights) *canvas {
	c := &canvas{w: w, h: h}
	sky := drawSky(w, h, frame, sim, calls)
	for y := 0; y < h; y++ {
		fgRow, bgRow, markRow := make([]rgb, w), make([]rgb, w), make([]rune, w)
		for x, cell := range sky[y] {
			bgRow[x] = rgb{surfaceRGB[0] + cell.lift[0], surfaceRGB[1] + cell.lift[1], surfaceRGB[2] + cell.lift[2]}
			markRow[x] = ' '
			if cell.mark != 0 {
				markRow[x], fgRow[x] = cell.mark, parseRGB(cell.fg)
			}
		}
		c.fg, c.bg, c.marks = append(c.fg, fgRow), append(c.bg, bgRow), append(c.marks, markRow)
	}
	// A satellite or meteor passing the title lights it. Rows are about
	// twice as tall as columns are wide, so rows count double.
	for y := 0; y < min(h, 4); y++ {
		for x := 0; x < min(w, 16); x++ {
			if p := sky[y][x].priority; p == 3 || p == 4 {
				dx, dy := float64(x-4), float64(2*y)
				c.titleLight = math.Max(c.titleLight, math.Exp(-(dx*dx+dy*dy)/40))
			}
		}
	}
	return c
}

// The title's light, from receded (a conversation is on screen) through open
// (an empty session) to lit (something in the sky passing it).
var (
	titleWordReceded = rgb{28, 34, 44}
	titleWordOpen    = rgb{74, 84, 100}
	titleWordLit     = rgb{132, 150, 166}
	titleMarkReceded = rgb{52, 84, 92}
	titleMarkOpen    = rgb{96, 154, 162}
	titleMarkLit     = rgb{150, 208, 214}
)

// titleColors are the mark's and word's colors for how open the title is
// (0 to 1) and how lit.
func titleColors(open, lit float64) (mark, word string) {
	mark = mix(lerp(titleMarkReceded, titleMarkOpen, open), titleMarkLit, lit)
	word = mix(lerp(titleWordReceded, titleWordOpen, open), titleWordLit, lit)
	return mark, word
}

// panel lays a translucent color over a rectangle; marks beneath it stay,
// sunk into the panel by the same amount.
func (c *canvas) panel(x0, y0, w, h int, color string, alpha float64) {
	p := parseRGB(color)
	for y := max(0, y0); y < min(c.h, y0+h); y++ {
		for x := max(0, x0); x < min(c.w, x0+w); x++ {
			c.bg[y][x] = lerp(c.bg[y][x], p, alpha)
			c.fg[y][x] = lerp(c.fg[y][x], c.bg[y][x], alpha)
		}
	}
}

// glow lifts the background around a rectangle toward color, strongest at
// its edge and falling off softly with distance. Rows are about twice as tall
// as columns are wide, so rows count double. The rectangle itself is left to
// the panel drawn over it; marks beneath the glow stay as they are.
func (c *canvas) glow(x0, y0, w, h int, color string, strength float64) {
	const reach, spread = 7, 12.0 // columns of falloff (rows: half), and its width
	p := parseRGB(color)
	for y := max(0, y0-reach/2); y < min(c.h, y0+h+reach/2); y++ {
		for x := max(0, x0-reach); x < min(c.w, x0+w+reach); x++ {
			dx := float64(max(x0-x, 0, x-(x0+w-1)))
			dy := 2 * float64(max(y0-y, 0, y-(y0+h-1)))
			d2 := dx*dx + dy*dy
			if d2 == 0 {
				continue
			}
			c.bg[y][x] = lerp(c.bg[y][x], p, strength*math.Exp(-d2/spread))
		}
	}
}

// text writes s from x, clipped to width. Spaces are left transparent, so
// the sky (or the panel under it) shows between words.
func (c *canvas) text(x, y, width int, s, fg string) int {
	if y < 0 || y >= c.h {
		return 0
	}
	col := parseRGB(fg)
	n := 0
	for _, r := range strings.ReplaceAll(safe(s), "\n", " ") {
		if n >= width || x+n >= c.w {
			break
		}
		if r != ' ' {
			c.marks[y][x+n], c.fg[y][x+n] = r, col
		}
		n++
	}
	return n
}

func (c *canvas) segments(x, y, width int, segs ...segment) {
	for _, s := range segs {
		n := c.text(x, y, width, s.text, s.fg)
		x, width = x+n, width-n
	}
}

func quantize(c rgb) [3]uint8 {
	var q [3]uint8
	for i, v := range c {
		q[i] = uint8(math.Round(math.Max(0, math.Min(255, v))))
	}
	return q
}

func (c *canvas) pixels() [][]pixel {
	out := make([][]pixel, c.h)
	for y := range out {
		out[y] = make([]pixel, c.w)
		for x := range out[y] {
			out[y][x] = pixel{c.marks[y][x], quantize(c.fg[y][x]), quantize(c.bg[y][x])}
		}
	}
	return out
}

func style(p pixel) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm\x1b[38;2;%d;%d;%dm", p.bg[0], p.bg[1], p.bg[2], p.fg[0], p.fg[1], p.fg[2])
}

// paint writes every cell of after that differs from before (all of them
// when before is nil), then returns the cursor to the composer.
func paint(before, after [][]pixel, cursorRow, cursorCol int) string {
	var out strings.Builder
	lastStyle := ""
	for y, row := range after {
		next := -1 // column the terminal cursor is at, if on this row
		for x, p := range row {
			if before != nil && y < len(before) && x < len(before[y]) && before[y][x] == p {
				continue
			}
			if out.Len() == 0 {
				out.WriteString("\x1b[?25l")
			}
			if x != next {
				fmt.Fprintf(&out, "\x1b[%d;%dH", y+1, x+1)
			}
			if st := style(p); st != lastStyle {
				out.WriteString(st)
				lastStyle = st
			}
			out.WriteRune(p.r)
			next = x + 1
		}
	}
	if out.Len() == 0 {
		return ""
	}
	fmt.Fprintf(&out, "\x1b[0m\x1b[%d;%dH\x1b[?25h", cursorRow, cursorCol)
	return out.String()
}

// draw repaints the whole screen.
func (a *app) draw() {
	frame := a.compose()
	if frame == nil {
		a.screen = nil
		c := newCanvas(a.width, a.height, a.wallFrame, nil, nil)
		c.segments(0, 0, a.width, segment{faint, " ✦ "}, segment{dim, "skyra · enlarge terminal"})
		fmt.Print("\x1b[2J" + paint(nil, c.pixels(), 1, 1))
		return
	}
	fmt.Print(paint(nil, frame, a.cursorRow, a.cursorCol))
	a.screen = frame
}

// tick advances the sky, writing only the cells that changed.
func (a *app) tick() {
	if a.screen == nil {
		return
	}
	frame := a.compose()
	if frame == nil {
		return
	}
	fmt.Print(paint(a.screen, frame, a.cursorRow, a.cursorCol))
	a.screen = frame
}

// compose builds the screen: the sky over the whole terminal, then the
// header, transcript, composer and footer on top of it.
func (a *app) compose() [][]pixel {
	w, h := a.width, a.height
	if h < 10 || w < 35 {
		return nil
	}
	c := newCanvas(w, h, a.wallFrame, a.sim, &a.calls)

	status, statusColor := "offline", red
	if a.connected {
		status, statusColor = "online", operator
	}
	phase := "new session"
	if cur := a.current(); cur != nil {
		phase = cur.Phase
	}
	// The title floats on the sky in near-sky tones: readable on an empty
	// session, receding to little more than its mark once a conversation is
	// on screen.
	markColor, wordColor := titleColors(a.titleOpen, c.titleLight)
	c.segments(0, 0, w, segment{markColor, " ✦ "}, segment{wordColor, "skyra"})
	right := []segment{{faint, phase + "   "}, {statusColor, "•"}, {dim, " " + status + " "}}
	rightWidth := 0
	for _, s := range right {
		rightWidth += len([]rune(s.text))
	}
	if w-8-rightWidth > 0 {
		c.segments(w-rightWidth, 0, rightWidth, right...)
	}

	composerHeight := 3
	if h < 12 {
		composerHeight = 1
	}
	// One row stays open between the transcript and the composer, for the
	// composer's glow.
	bodyHeight := h - composerHeight - 6
	readingWidth := w - 4
	inset := 2
	for i, r := range a.transcript(readingWidth, bodyHeight) {
		y := 2 + i
		switch {
		case len(r.segs) > 0:
			c.segments(inset+3, y, readingWidth-3, r.segs...)
		case r.rail != "":
			alpha := agentAlpha
			if r.bg == userBG {
				alpha = userAlpha
			}
			c.panel(inset, y, readingWidth, 1, r.bg, alpha)
			c.text(inset+1, y, 1, "▎", r.rail)
			c.text(inset+3, y, readingWidth-3, r.text, r.fg)
		case r.text != "":
			c.text(inset, y, readingWidth, r.text, r.fg)
		}
	}

	placeholder := "Message the agent"
	ready := a.connected && !a.opening
	if !a.connected {
		placeholder = "Connecting… type a draft"
	} else if a.opening {
		placeholder = "Opening a session…"
	} else if cur := a.current(); cur == nil || cur.Phase == "closed" || cur.Phase == "failed" {
		placeholder = "Type a message to begin"
	} else if cur.Phase == "pending" {
		placeholder, ready = "Agent working… type a draft", false
	}
	prompt := faint
	if ready {
		prompt = accent
	}
	textWidth := a.textWidth()
	contentRows := composerHeight - 2
	contentOffset := 1
	if composerHeight == 1 {
		contentRows, contentOffset = 1, 0
	}
	// The composer shows the lines around the caret.
	a.caret = max(0, min(a.caret, len(a.input)))
	lineCount := len(a.input)/textWidth + 1
	caretLine := a.caret / textWidth
	firstLine := max(0, caretLine-contentRows+1)
	top := 3 + bodyHeight
	// As in the desktop app, the composer is a dark field with no outline,
	// set apart by a soft white light: brighter when it is ready for a
	// message, fainter while the agent works or it is offline. The light is
	// around the field, not on it.
	glowStrength := 0.05
	if ready {
		glowStrength = 0.14
	}
	c.glow(inset, top, readingWidth, composerHeight, ink, glowStrength)
	c.panel(inset, top, readingWidth, composerHeight, surface, 1)
	for r := 0; r < contentRows; r++ {
		y, lineIndex := top+contentOffset+r, firstLine+r
		if r == 0 {
			c.text(inset+1, y, 1, "›", prompt)
		}
		if len(a.input) == 0 && lineIndex == 0 {
			c.text(inset+3, y, textWidth, placeholder, faint)
		} else if lineIndex < lineCount {
			start := lineIndex * textWidth
			c.text(inset+3, y, textWidth, string(a.input[start:min(start+textWidth, len(a.input))]), ink)
		}
	}

	var footer []segment
	if a.notice != "" {
		footer = []segment{{dim, "  " + a.notice}}
	} else {
		for i, k := range [][2]string{{"enter", "send"}, {"^n", "new"}, {"pgup/pgdn", "scroll"}, {"^c", "clear"}, {"^d", "close"}, {"^t", "mouse"}, {"^c ^c", "quit"}} {
			sep := "   "
			if i == 0 {
				sep = "  "
			}
			footer = append(footer, segment{dim, sep + k[0]}, segment{faint, " " + k[1]})
		}
	}
	c.segments(0, top+composerHeight+1, w, footer...)

	a.cursorRow = top + contentOffset + caretLine - firstLine + 1
	a.cursorCol = inset + 4 + a.caret%textWidth
	return c.pixels()
}

func (a *app) transcript(width, height int) []line {
	rows := make([]line, height)
	for i := range rows {
		rows[i] = line{"", surface, ink, "", nil}
	}
	c := a.current()
	if c == nil {
		return rows
	}
	var all []line
	if len(c.Entries) == 0 {
		all = append(all, line{"  opening the session…", surface, faint, "", nil})
	}
	for i, e := range c.Entries {
		if e.Call != nil {
			if i == 0 || c.Entries[i-1].Call == nil {
				all = append(all, line{"", surface, ink, "", nil})
			}
			all = append(all, line{bg: surface, segs: ledger(*e.Call, width-3)})
			continue
		}
		// The desktop tells you from the agent by side and panel; a terminal
		// line has no side, so a quiet rail and label say it.
		label, bg, rail := "agent", surface, faint
		if e.Role == "user" {
			label, bg, rail = "you", userBG, dim
		}
		fg := ink
		if strings.HasPrefix(e.Text, "⚠") {
			fg, rail = red, red
		}
		all = append(all, line{"", surface, ink, "", nil}, line{label, bg, dim, rail, nil})
		for _, part := range wrap(safe(e.Text), width-5) {
			all = append(all, line{part, bg, fg, rail, nil})
		}
	}
	if c.Phase == "pending" {
		// Three dots breathing, like the stars, as the desktop app shows it.
		var dots []segment
		for i := range 3 {
			breath := (1 + math.Sin(float64(a.wallFrame)*0.3-float64(i)*0.8)) / 2
			dots = append(dots, segment{mix(surfaceRGB, parseRGB(accent), 0.25+0.5*breath), "• "})
		}
		all = append(all, line{"", surface, ink, "", nil}, line{bg: surface, segs: dots})
	}
	// Scrolling stops at the first line, so scrolling back down answers at once.
	a.scroll = max(0, min(a.scroll, len(all)-height))
	start := max(0, len(all)-height-a.scroll)
	if start+height > len(all) {
		start = max(0, len(all)-height)
	}
	for i := 0; i < height && start+i < len(all); i++ {
		rows[i] = all[start+i]
	}
	return rows
}

// ledger is a tool call on one line: what was called, with what, and what
// came of it, the outcome and stamp flush right.
func ledger(c call, width int) []segment {
	toColor, outColor := callTo, dim
	var out string
	switch {
	case c.State == "running":
		out = "running"
	case c.State == "returned":
		out = c.Result
		if c.MS != nil {
			out += fmt.Sprintf(" · %.1f s", float64(*c.MS)/1000)
		}
	case c.Fault == "horizon":
		toColor, outColor, out = amber, amber, "outside the Horizon"
	case c.Fault == "contract":
		toColor, outColor, out = amber, amber, "didn't fit "+c.To+"'s input"
	default:
		toColor, outColor, out = red, red, c.Cause
	}
	stamp := c.Stamp
	if len(stamp) > 8 {
		stamp = stamp[:8]
	}
	to := []rune(c.To)
	right := []rune(out + "  " + stamp)
	arg := []rune(c.Arg)
	room := width - 2 - len(to) - len(right) - 2
	if room < 0 {
		// Too narrow for everything: the outcome matters more than the
		// stamp, and both more than the whole tool name.
		right = []rune(out)
		if len(to) > width-2-len(right)-2 {
			to = to[:max(0, width-2-len(right)-2)]
		}
		room = width - 2 - len(to) - len(right) - 2
	}
	if len(arg) > max(0, room) {
		arg = arg[:max(0, room)]
	}
	pad := max(1, width-2-len(to)-1-len(arg)-len(right))
	return []segment{
		{faint, "↳ "}, {toColor, string(to)}, {dim, " " + string(arg) + strings.Repeat(" ", pad)},
		{outColor, out}, {faint, string(right[len([]rune(out)):])},
	}
}

func color(bg, fg, text string) string {
	return "\x1b[48;2;" + bg + "m\x1b[38;2;" + fg + "m" + text + "\x1b[0m"
}

func safe(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\n' {
			b.WriteRune(r)
		} else if r == '\t' {
			b.WriteRune(' ')
		} else if unicode.IsPrint(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(strings.ReplaceAll(safe(s), "\n", " "))
	if len(r) > width {
		r = r[:width]
	}
	return string(r) + strings.Repeat(" ", width-len(r))
}

func wrap(s string, width int) []string {
	if width < 1 {
		width = 1
	}
	var out []string
	for _, paragraph := range strings.Split(s, "\n") {
		r := []rune(paragraph)
		if len(r) == 0 {
			out = append(out, "")
			continue
		}
		for len(r) > 0 {
			n := min(width, len(r))
			out = append(out, string(r[:n]))
			r = r[n:]
		}
	}
	return out
}
