package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"skyra-v1/src/debug"
	"skyra-v1/src/inference"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dna "skyra-v1/src/skyra/skyra-dna"
	userdna "skyra-v1/src/skyra/user-dna"
	"skyra-v1/src/skyra/world"
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	subtle  = lipgloss.Color("8")
	accent  = lipgloss.Color("12")
	brand   = lipgloss.Color("5")
	warn    = lipgloss.Color("11")
	success = lipgloss.Color("10")

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

	itemStyle     = lipgloss.NewStyle().Foreground(subtle)
	itemSelStyle  = lipgloss.NewStyle().Foreground(warn).Bold(true)
	labelStyle    = lipgloss.NewStyle().Foreground(warn)
	dimStyle      = lipgloss.NewStyle().Foreground(subtle)
	senderUser    = lipgloss.NewStyle().Bold(true).Foreground(accent)
	senderSkyra   = lipgloss.NewStyle().Bold(true).Foreground(brand)
	hintStyle     = lipgloss.NewStyle().Foreground(subtle).Italic(true)
)

// ── state machine ─────────────────────────────────────────────────────────────

type uiState int

const (
	stateInput        uiState = iota // primitive selector + text input
	stateEntityPicker                // engage entity selection overlay
	stateWaiting                     // impulse fired, waiting for motor
)

// ── messages ──────────────────────────────────────────────────────────────────

type resultMsg struct {
	framing string
	other   string
	self    string
	motor   string
	err     string
}

type tickMsg struct{}

// ── model ─────────────────────────────────────────────────────────────────────

type model struct {
	// runtime
	realm  *world.World
	user   *userdna.UserCell
	being  *dna.MembraneCell
	kernel *relation.Relation

	// ui state
	state     uiState
	primitive string // "perceive" or "engage"
	targetID  string // selected entity for engage

	// entity picker
	entities []string // sorted entity IDs
	selIdx   int

	// widgets
	input  textarea.Model
	output viewport.Model
	ready  bool
	width  int
	height int

	// output log entries (rendered lines)
	lines []string
}

func newModel(realm *world.World, user *userdna.UserCell, being *dna.MembraneCell, kernel *relation.Relation) model {
	ti := textarea.New()
	ti.Placeholder = "write something..."
	ti.Focus()
	ti.SetHeight(1)
	ti.ShowLineNumbers = false
	ti.KeyMap.InsertNewline.SetKeys("alt+enter")
	ti.CharLimit = 0

	return model{
		realm:     realm,
		user:      user,
		being:     being,
		kernel:    kernel,
		primitive: "perceive",
		input:     ti,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
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
		if msg.err != "" {
			m.pushLine(dimStyle.Render("error: " + msg.err))
		} else {
			m.pushLine(labelStyle.Render("framing: ") + msg.framing)
			m.pushLine(dimStyle.Render("  other: ") + msg.other)
			m.pushLine(dimStyle.Render("   self: ") + msg.self)
			m.pushLine(dimStyle.Render("  motor: ") + msg.motor)
		}
		m.pushLine("")
		m.refreshOutput()
		return m, nil

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

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	}

	return m, nil
}

func (m model) fire(text string) (tea.Model, tea.Cmd) {
	m.state = stateWaiting

	src := &skyra.ImpulseSource{
		SpaceID: m.user.ID(),
		Coord:   skyra.Coord{X: 500, Y: 500},
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
		imp = skyra.Impulse{
			Primitive: world.PrimitivePerceive,
			ResultMap: world.MapPerception,
			Payload:   text,
			Amplitude: 1.0,
			Source:    src,
		}
		m.pushLine(senderUser.Render("michael") + dimStyle.Render(" → perceive"))
	}
	m.pushLine(dimStyle.Render("  \"" + text + "\""))
	m.input.Reset()
	m.refreshOutput()

	before := m.being.Motor().ActionCount()
	realm := m.realm
	user := m.user
	being := m.being

	return m, func() tea.Msg {
		user.Send(imp)
		realm.Observe(being, imp)

		deadline := time.Now().Add(3 * time.Minute)
		for time.Now().Before(deadline) {
			if being.Motor().ActionCount() > before {
				mid := being.Sensory().MidLevel()
				selfPred, otherPred := predictions(mid)
				return resultMsg{
					framing: mid.Gist(),
					other:   otherPred,
					self:    selfPred,
					motor:   being.Motor().LastAction(),
				}
			}
			time.Sleep(150 * time.Millisecond)
		}
		return resultMsg{err: "timed out waiting for motor"}
	}
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if !m.ready {
		return "loading..."
	}

	var sections []string

	// title
	sections = append(sections, titleStyle.Render(" skyra v.1"))

	// primitive selector
	sections = append(sections, m.renderPrimitiveBar())

	// entity picker (inline, only when in picker state)
	if m.state == stateEntityPicker {
		sections = append(sections, m.renderEntityPicker())
	}

	// output log
	sections = append(sections,
		outputStyle.Width(m.outputWidth()).Height(m.outputHeight()).Render(m.output.View()),
	)

	// target hint for engage
	if m.primitive == "engage" && m.targetID != "" {
		sections = append(sections, hintStyle.Render("  target: "+m.targetID))
	}

	// input
	waitLabel := ""
	if m.state == stateWaiting {
		waitLabel = dimStyle.Render(" waiting...")
	}
	inputContent := m.input.View() + waitLabel
	sections = append(sections,
		inputStyle.Width(m.inputWidth()).Render(inputContent),
	)

	// hints
	hints := "tab: perceive/engage  enter: send  ↑↓: navigate picker  esc: cancel  ctrl+c: quit"
	sections = append(sections, hintStyle.Render(hints))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
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

func sortedEntityIDs(entities map[string]skyra.Space) []string {
	ids := make([]string, 0, len(entities))
	for id := range entities {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func predictions(mid *dna.MidLevelCell) (self, other string) {
	sl := mid.Slice()
	sl.Lock()
	defer sl.Unlock()

	bestTS := -1
	var best skyra.Space
	for _, gr := range mid.Groupings() {
		mc, ok := gr.(*dna.MemoryCell)
		if !ok {
			continue
		}
		if g, ok := mc.Content().(dna.Gist); ok && g.Timestamp > bestTS {
			bestTS, best = g.Timestamp, gr
		}
	}
	mc, ok := best.(*dna.MemoryCell)
	if !ok {
		return "", ""
	}
	for _, child := range mc.Children() {
		if p, ok := child.Content().(dna.Prediction); ok {
			switch p.Role {
			case "self":
				self = p.Text
			case "other":
				other = p.Text
			}
		}
	}
	return
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
	being := dna.NewMembraneCell("be helpful", dna.BuildContext{
		BeingID:  "skyra",
		NewSlice: func(id string) skyra.Slice { return world.NewSlice(id, "skyra") },
		Kernel:   kernel,
		Infer:    inference.CallCodex,
	})
	user := userdna.NewUserCell("michael", func(id string) skyra.Slice {
		return world.NewSlice(id, "michael")
	})
	realm.Place(user, 500, 500)
	realm.Place(being, 501, 500)
	user.Seed(being)

	// Wire motor → world: motor's concrete action is delivered to the world so the
	// target entity (e.g. michael) actually receives skyra's response.
	being.Motor().SetEmit(func(imp skyra.Impulse) {
		realm.Observe(being, imp)
	})

	m := newModel(realm, user, being, kernel)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui error:", err)
		os.Exit(1)
	}
}
