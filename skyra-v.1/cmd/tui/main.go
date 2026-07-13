package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"sort"
	"strings"

	"skyra-v1/src/debug"
	"skyra-v1/src/inference"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dnav2 "skyra-v1/src/skyra/skyra-dna-v2"
	userdna "skyra-v1/src/skyra/user-dna"
	"skyra-v1/src/skyra/world"
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	subtle    = lipgloss.Color("8")
	accent    = lipgloss.Color("12")
	brand     = lipgloss.Color("5")
	warn      = lipgloss.Color("11")
	success   = lipgloss.Color("10")
	moveColor = lipgloss.Color("14") // cyan for move mode

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(brand)

	btnStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle).
			Foreground(subtle)

	btnActiveStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Foreground(accent).
			Bold(true)

	outputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(0, 1)

	pickerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(warn).
			Padding(0, 1)

	itemStyle    = lipgloss.NewStyle().Foreground(subtle)
	itemSelStyle = lipgloss.NewStyle().Foreground(warn).Bold(true)
	labelStyle   = lipgloss.NewStyle().Foreground(warn)
	dimStyle     = lipgloss.NewStyle().Foreground(subtle)
	senderUser   = lipgloss.NewStyle().Bold(true).Foreground(accent)
	senderSkyra  = lipgloss.NewStyle().Bold(true).Foreground(brand)
	hintStyle    = lipgloss.NewStyle().Foreground(subtle).Italic(true)
)

// ── state machine ─────────────────────────────────────────────────────────────

type uiState int

const (
	stateInput        uiState = iota // primitive selector + text input
	stateEntityPicker                // engage entity selection overlay
	stateWaiting                     // impulse fired, waiting for motor
	stateMove                        // arrow-key navigation mode
	statePraxis                      // read-only causal substrate inspection
)

// ── messages ──────────────────────────────────────────────────────────────────

type resultMsg struct {
	framing string
	other   string
	self    string
	motor   string
	err     string
}

type inboundMsg struct{ impulse skyra.Impulse }

type tickMsg struct{}

// ── model ─────────────────────────────────────────────────────────────────────

type model struct {
	// runtime
	realm     *world.World
	user      *userdna.UserCell
	being     *dnav2Being
	kernel    *relation.Relation
	inboundCh chan skyra.Impulse

	// ui state
	state     uiState
	primitive string // "perceive" or "engage"
	targetID  string // selected entity for engage

	// entity picker
	entities []string // sorted entity IDs
	selIdx   int

	// positions
	michaelCoord skyra.Coord

	// map camera — independent of agent position
	cameraOffset skyra.Coord // pan offset from michaelCoord (0,0 = centered on michael)
	viewCells    int         // world cells shown across the viewport; 0 = auto (fill screen)
	dragging     bool
	dragStartPx  skyra.Coord // terminal cell where drag began
	dragStartCam skyra.Coord // cameraOffset at drag start

	// widgets
	input  textarea.Model
	output viewport.Model
	ready  bool
	width  int
	height int

	// output log entries (rendered lines)
	lines []string
}

func newModel(realm *world.World, user *userdna.UserCell, being *dnav2Being, kernel *relation.Relation, inboundCh chan skyra.Impulse) model {
	ti := textarea.New()
	ti.Placeholder = "write something..."
	ti.Focus()
	ti.SetHeight(1)
	ti.ShowLineNumbers = false
	ti.KeyMap.InsertNewline.SetKeys("alt+enter")
	ti.CharLimit = 0

	return model{
		realm:        realm,
		user:         user,
		being:        being,
		kernel:       kernel,
		inboundCh:    inboundCh,
		primitive:    "perceive",
		input:        ti,
		michaelCoord: skyra.Coord{X: world.MazeEntranceX, Y: world.MazeOriginY - 2},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, waitInbound(m.inboundCh))
}

func waitInbound(ch chan skyra.Impulse) tea.Cmd {
	return func() tea.Msg {
		return inboundMsg{impulse: <-ch}
	}
}

// ── update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.output = viewport.New(m.outputWidth(), m.outputHeight())
			m.ready = true
		} else {
			m.output.Width = m.outputWidth()
			m.output.Height = m.outputHeight()
		}
		m.input.SetWidth(m.inputWidth())
		m.refreshOutput()
		return m, nil

	case resultMsg:
		m.state = stateInput
		m.refreshOutput()
		return m, nil

	case inboundMsg:
		imp := msg.impulse
		src := "skyra"
		if imp.Source != nil && imp.Source.SpaceID != "" {
			src = imp.Source.SpaceID
		}
		msgWidth := m.outputWidth() - 4
		switch imp.Primitive {
		case world.PrimitivePerceive:
			m.pushLine(senderSkyra.Render(src) + dimStyle.Render(" perceived you"))
			m.pushLine("")
		case world.PrimitiveEngage:
			text, _ := imp.Payload.(string)
			m.pushLine(senderSkyra.Render(src) + dimStyle.Render(" → engage → michael"))
			for _, line := range strings.Split(wordWrap(text, msgWidth), "\n") {
				m.pushLine(dimStyle.Render("  " + line))
			}
			m.pushLine("")
			// Exit agent teleports michael back outside the entrance — update local coord.
			if src == world.ExitAgentID {
				m.michaelCoord = skyra.Coord{X: world.MazeEntranceX, Y: world.MazeOriginY - 2}
			}
		case world.PrimitiveBump:
			result, _ := imp.Payload.(world.BumpResult)
			other := result.OtherID
			if other == "" {
				other = src
			}
			m.pushLine(senderSkyra.Render(src) + dimStyle.Render(" bumped into "+other))
			m.pushLine("")
		}
		m.refreshOutput()
		return m, waitInbound(m.inboundCh)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// pass keys to textarea when in input state
	if m.state == stateInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	const zoomStep = 8
	const zoomMin = 20
	const zoomMax = 300

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		// zoom in — show fewer world cells
		cells := m.viewCells
		if cells == 0 {
			cells = m.width - 4
		}
		cells -= zoomStep
		if cells < zoomMin {
			cells = zoomMin
		}
		m.viewCells = cells
		return m, nil

	case tea.MouseButtonWheelDown:
		// zoom out — show more world cells
		cells := m.viewCells
		if cells == 0 {
			cells = m.width - 4
		}
		cells += zoomStep
		if cells > zoomMax {
			cells = zoomMax
		}
		m.viewCells = cells
		return m, nil
	}

	// Left button drag — pan the camera
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			m.dragging = true
			m.dragStartPx = skyra.Coord{X: msg.X, Y: msg.Y}
			m.dragStartCam = m.cameraOffset
		}
	case tea.MouseActionRelease:
		m.dragging = false
	case tea.MouseActionMotion:
		if m.dragging {
			m.cameraOffset = skyra.Coord{
				X: m.dragStartCam.X + (m.dragStartPx.X - msg.X),
				Y: m.dragStartCam.Y + (m.dragStartPx.Y - msg.Y),
			}
		}
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+p" {
		if m.state == statePraxis {
			m.state = stateInput
			m.input.Focus()
			return m, textarea.Blink
		}
		m.state = statePraxis
		m.input.Blur()
		return m, nil
	}
	switch m.state {

	case stateInput:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyTab:
			if m.primitive == "perceive" {
				m.primitive = "engage"
			} else {
				m.primitive = "perceive"
			}
			return m, nil

		case tea.KeyEsc:
			if strings.TrimSpace(m.input.Value()) == "" {
				m.state = stateMove
				return m, nil
			}

		case tea.KeyEnter:
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			if m.primitive == "engage" {
				// open entity picker before firing
				m.entities = sortedEntityIDs(m.user.KnownEntities())
				m.selIdx = 0
				m.state = stateEntityPicker
				return m, nil
			}
			// perceive — fire immediately
			return m.fire(text)
		}

		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case stateMove:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			m.state = stateInput
			m.cameraOffset = skyra.Coord{}
			m.input.Focus()
			return m, textarea.Blink

		case tea.KeyUp:
			return m.move(0, -1)
		case tea.KeyDown:
			return m.move(0, 1)
		case tea.KeyLeft:
			return m.move(-1, 0)
		case tea.KeyRight:
			return m.move(1, 0)

		default:
			// any other key exits move mode
			m.state = stateInput
			m.input.Focus()
			return m, textarea.Blink
		}

	case stateEntityPicker:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			m.state = stateInput
			return m, nil

		case tea.KeyUp:
			if m.selIdx > 0 {
				m.selIdx--
			}
			return m, nil

		case tea.KeyDown:
			if m.selIdx < len(m.entities)-1 {
				m.selIdx++
			}
			return m, nil

		case tea.KeyEnter:
			if len(m.entities) == 0 {
				m.state = stateInput
				return m, nil
			}
			m.targetID = m.entities[m.selIdx]
			m.state = stateInput
			// now fire with the selected target
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				// just confirmed entity, wait for user to type
				return m, nil
			}
			return m.fire(text)
		}

	case statePraxis:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			m.state = stateInput
			m.input.Focus()
			return m, textarea.Blink
		}
	}

	return m, nil
}

func (m model) move(dx, dy int) (tea.Model, tea.Cmd) {
	dest := skyra.Coord{X: m.michaelCoord.X + dx, Y: m.michaelCoord.Y + dy}
	if m.realm.IsWall(dest.X, dest.Y) {
		return m, nil
	}
	m.michaelCoord = dest

	imp := skyra.Impulse{
		Primitive: world.PrimitiveMove,
		TargetID:  fmt.Sprintf("%d,%d", dest.X, dest.Y),
		Amplitude: 1.0,
		Source:    &skyra.ImpulseSource{SpaceID: m.user.ID(), Coord: m.michaelCoord},
	}

	realm := m.realm
	user := m.user
	return m, func() tea.Msg {
		realm.Observe(user, imp)
		return nil
	}
}

func (m model) fire(text string) (tea.Model, tea.Cmd) {
	m.state = stateWaiting

	src := &skyra.ImpulseSource{
		SpaceID: m.user.ID(),
		Coord:   m.michaelCoord,
	}

	var imp skyra.Impulse
	if m.primitive == "engage" {
		imp = skyra.Impulse{
			Primitive: world.PrimitiveEngage,
			TargetID:  m.targetID,
			Payload:   text,
			Amplitude: 1.0,
			Source:    src,
		}
		m.pushLine(senderUser.Render("michael") + dimStyle.Render(" → engage → "+m.targetID))
	} else {
		// perceive carries no utterance — it's spatial sensing, not communication
		imp = skyra.Impulse{
			Primitive: world.PrimitivePerceive,
			ResultMap: world.MapPerception,
			Amplitude: 1.0,
			Source:    src,
		}
		m.pushLine(senderUser.Render("michael") + dimStyle.Render(" → perceive"))
	}
	if m.primitive == "engage" {
		for _, line := range strings.Split(wordWrap(text, m.outputWidth()-4), "\n") {
			m.pushLine(dimStyle.Render("  " + line))
		}
	}
	m.input.Reset()
	m.refreshOutput()

	realm := m.realm
	user := m.user

	return m, func() tea.Msg {
		user.Send(imp)
		realm.Observe(user, imp)
		return resultMsg{}
	}
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if !m.ready {
		return "loading..."
	}

	var sections []string

	// title + status bar
	sections = append(sections, m.renderTitleBar())

	if m.state == statePraxis {
		report := "Praxis inspection unavailable"
		if m.being != nil && m.being.dna != nil {
			report = dnav2.RenderPraxis(m.being.dna.InspectPraxis(""))
		}
		sections = append(sections, outputStyle.Width(m.outputWidth()).Height(m.outputHeight()+4).Render(report))
		sections = append(sections, hintStyle.Render("  ctrl+p/esc: back to chat  ctrl+c: quit"))
	} else if m.state == stateMove {
		// move mode: show map panel + directional hint
		sections = append(sections, m.renderMapPanel())
		sections = append(sections, hintStyle.Render("  ↑↓←→: move  esc: back to chat  ctrl+c: quit"))
	} else {
		// chat mode
		sections = append(sections, m.renderPrimitiveBar())

		if m.state == stateEntityPicker {
			sections = append(sections, m.renderEntityPicker())
		}

		sections = append(sections,
			outputStyle.Width(m.outputWidth()).Height(m.outputHeight()).Render(m.output.View()),
		)

		if m.primitive == "engage" && m.targetID != "" {
			sections = append(sections, hintStyle.Render("  target: "+m.targetID))
		}

		sections = append(sections,
			inputStyle.Width(m.inputWidth()).Render(m.input.View()),
		)

		sections = append(sections, hintStyle.Render("  tab: perceive/engage  enter: send  esc: move mode  ctrl+p: praxis  ctrl+c: quit"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m model) renderTitleBar() string {
	title := titleStyle.Render(" skyra v.1")

	// michael's coord
	mCoord := fmt.Sprintf("michael@%d,%d", m.michaelCoord.X, m.michaelCoord.Y)

	// skyra's coord — read from the world directly so it updates on any move,
	// not only when skyra herself perceives.
	sCoord := "skyra@?"
	if m.being != nil {
		if coord, ok := m.realm.CoordOf(m.being); ok {
			sCoord = fmt.Sprintf("skyra@%d,%d", coord.X, coord.Y)
		}
	}

	var modeStr string
	if m.state == stateMove {
		modeStr = lipgloss.NewStyle().Bold(true).Foreground(moveColor).Render("[move]")
	} else {
		modeStr = dimStyle.Render("[chat]")
	}

	right := dimStyle.Render(mCoord+"  "+sCoord) + "  " + modeStr
	gap := m.width - lipgloss.Width(title) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	return title + strings.Repeat(" ", gap) + right
}

func (m model) renderPrimitiveBar() string {
	perceiveBtn := btnStyle.Render("perceive")
	engageBtn := btnStyle.Render("engage")

	if m.primitive == "perceive" {
		perceiveBtn = btnActiveStyle.Render("perceive")
	} else {
		engageBtn = btnActiveStyle.Render("engage")
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, " ", perceiveBtn, "  ", engageBtn)
}

func (m model) renderEntityPicker() string {
	if len(m.entities) == 0 {
		return pickerStyle.Render(dimStyle.Render("  no known entities"))
	}

	var rows []string
	rows = append(rows, labelStyle.Render("  select target:"))
	for i, id := range m.entities {
		if i == m.selIdx {
			rows = append(rows, itemSelStyle.Render("  › "+id))
		} else {
			rows = append(rows, itemStyle.Render("    "+id))
		}
	}
	return pickerStyle.Width(m.outputWidth()).Render(strings.Join(rows, "\n"))
}

// renderMapPanel draws a 2D view of the maze centered on Michael's position.
// Symbols: █ wall  · path  @ michael  S skyra  (space) unknown
func (m model) renderMapPanel() string {
	contentW := m.width - 4
	contentH := m.height - 4
	if contentW < 10 {
		contentW = 10
	}
	if contentH < 3 {
		contentH = 3
	}

	// viewCells controls zoom: how many world cells span the terminal width.
	// 0 = auto (fill screen). Smaller = zoomed in, larger = zoomed out.
	worldW := contentW
	if m.viewCells > 0 {
		worldW = m.viewCells
	}
	// Maintain aspect ratio for world height.
	worldH := worldW * contentH / contentW
	if worldH < 1 {
		worldH = 1
	}

	// Camera center = michael's position + independent pan offset.
	centerX := m.michaelCoord.X + m.cameraOffset.X
	centerY := m.michaelCoord.Y + m.cameraOffset.Y
	startX := centerX - worldW/2
	startY := centerY - worldH/2

	skyraCoord, hasSkyra := m.realm.CoordOf(m.being)

	var sb strings.Builder
	for row := 0; row < contentH; row++ {
		// Map terminal row → world y (scale worldH → contentH).
		wy := startY + row*worldH/contentH
		for col := 0; col < contentW; col++ {
			// Map terminal col → world x (scale worldW → contentW).
			wx := startX + col*worldW/contentW
			switch {
			case wx == m.michaelCoord.X && wy == m.michaelCoord.Y:
				sb.WriteString(lipgloss.NewStyle().Foreground(accent).Bold(true).Render("@"))
			case hasSkyra && wx == skyraCoord.X && wy == skyraCoord.Y:
				sb.WriteString(lipgloss.NewStyle().Foreground(brand).Bold(true).Render("S"))
			case wx == world.MazeExitX && wy == world.MazeExitY:
				sb.WriteString(lipgloss.NewStyle().Foreground(success).Bold(true).Render("X"))
			case wx == world.MazeEntranceX && wy == world.MazeOriginY:
				sb.WriteString(lipgloss.NewStyle().Foreground(warn).Bold(true).Render("E"))
			case m.realm.IsWall(wx, wy):
				sb.WriteString(lipgloss.NewStyle().Foreground(subtle).Render("█"))
			case m.realm.IsPath(wx, wy):
				sb.WriteString(dimStyle.Render("·"))
			default:
				sb.WriteString(" ")
			}
		}
		if row < contentH-1 {
			sb.WriteByte('\n')
		}
	}

	return outputStyle.Width(contentW).Render(sb.String())
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (m *model) pushLine(line string) {
	m.lines = append(m.lines, line)
}

func (m *model) refreshOutput() {
	if !m.ready {
		return
	}
	m.output.SetContent(strings.Join(m.lines, "\n"))
	m.output.GotoBottom()
}

func (m model) outputWidth() int {
	w := m.width - 4
	if w < 20 {
		w = 20
	}
	return w
}

func (m model) outputHeight() int {
	if m.state == stateMove {
		// title(1) + hints(1) + padding(2)
		h := m.height - 4
		if h < 3 {
			h = 3
		}
		return h
	}
	// title(1) + primitive bar(3) + input(3) + hints(1) + padding(4)
	h := m.height - 12
	if m.state == stateEntityPicker {
		h -= (len(m.entities) + 3)
	}
	if h < 3 {
		h = 3
	}
	return h
}

func (m model) inputWidth() int {
	w := m.width - 4
	if w < 20 {
		w = 20
	}
	return w
}

// wordWrap breaks text at word boundaries so it fits within width characters.
func wordWrap(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}
	var result strings.Builder
	lineLen := 0
	for i, word := range strings.Fields(text) {
		if i > 0 && lineLen+1+len(word) > width {
			result.WriteByte('\n')
			lineLen = 0
		} else if i > 0 {
			result.WriteByte(' ')
			lineLen++
		}
		result.WriteString(word)
		lineLen += len(word)
	}
	return result.String()
}

func sortedEntityIDs(entities map[string]skyra.Space) []string {
	ids := make([]string, 0, len(entities))
	for id := range entities {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	if err := debug.Init("logs"); err != nil {
		fmt.Fprintln(os.Stderr, "log init:", err)
		os.Exit(1)
	}
	defer debug.Close()

	kernel := relation.NewRelation()
	realm := world.New(kernel)

	// The being is the slice-based-dispatch (dnav2) pipeline: one DNA nucleus + the
	// orchestrator running on its OWN kernel. dnav2Being is the thin world bridge —
	// its Observe fires orch, and DNA.Emit (wired below) sends motor output back out.
	d := dnav2.NewDNA("skyra", world.NewSlice("skyra:nucleus", "skyra"))
	dnav2.WireGenes(d)
	il := dnav2.NewInferLogger("skyra", inference.CallCodex) // times + logs every inference call
	d.SetInfer(il.Infer)

	board := dnav2.NewBoard()
	beingKernel := relation.NewRelation()
	nsSlices := map[string]skyra.Slice{}
	for _, ns := range []string{"sensory", "other", "self", "midlevel", "motor"} {
		nsSlices[ns] = world.NewSlice("skyra:"+ns, "skyra")
	}
	orch := dnav2.NewOrchestrator(d, board, beingKernel, world.NewSlice("skyra:membrane", "skyra"), nsSlices)
	being := newDnav2Being("skyra", world.NewSlice("skyra:being", "skyra"), orch, d)

	user := userdna.NewUserCell("michael", func(id string) skyra.Slice {
		return world.NewSlice(id, "michael")
	})
	realm.GenerateMaze(42)
	entrance := world.NewEntranceAgent(realm)
	exit := world.NewExitAgent(realm)
	realm.Place(entrance, world.MazeEntranceX, world.MazeEntranceY)
	realm.Place(exit, world.MazeExitX, world.MazeExitY)
	realm.Place(user, world.MazeEntranceX, world.MazeOriginY-2)
	realm.Place(being, world.MazeEntranceX+2, world.MazeOriginY-2)
	user.Seed(entrance)
	user.Seed(exit)
	user.Seed(being)

	// Wire motor → world: dnav2's motor:collapse emits skyra's concrete action through
	// DNA.Emit; deliver it to the world so the target (michael) actually receives it.
	d.SetEmit(func(imp skyra.Impulse) {
		realm.Observe(being, imp)
	})

	// Wire inbound notifications → TUI channel so michael's received impulses
	// (skyra's engage replies, perceive pings) appear in the UI.
	// Filter: skip return waves and self-delivered impulses (the world echoes
	// michael's own engage back to him when he's in the delivery set).
	inboundCh := make(chan skyra.Impulse, 16)
	userID := user.ID()
	user.OnInbound = func(imp skyra.Impulse) {
		if imp.Source != nil && imp.Source.SpaceID == userID {
			return // own echo — skip
		}
		select {
		case inboundCh <- imp:
		default:
		}
	}

	m := newModel(realm, user, being, kernel, inboundCh)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui error:", err)
		os.Exit(1)
	}
}
