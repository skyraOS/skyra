package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"skyra-v1/src/reality"
)

var (
	sidebarWidth = 30

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

	sidebarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1, 1)

	chatStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12"))

	senderUser = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

	senderBeing = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("5"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))
)

type ChatEntry struct {
	From    string
	Content string
	IsUser  bool
}

type ResponseMsg struct {
	From    string
	Content string
}

type Model struct {
	world    *reality.Relation
	width    int
	height   int
	input    textarea.Model
	chat     viewport.Model
	messages []ChatEntry
	ready    bool
	inputCh  chan string
	outputCh chan ResponseMsg
}

func New(world *reality.Relation, inputCh chan string, outputCh chan ResponseMsg) Model {
	ti := textarea.New()
	ti.Placeholder = "speak..."
	ti.Focus()
	ti.SetHeight(1)
	ti.ShowLineNumbers = false
	ti.KeyMap.InsertNewline.SetKeys("alt+enter")
	ti.CharLimit = 0

	return Model{
		world:    world,
		input:    ti,
		messages: []ChatEntry{},
		inputCh:  inputCh,
		outputCh: outputCh,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		listenResponse(m.outputCh),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				return m, nil
			}
			m.messages = append(m.messages, ChatEntry{
				From: "michael", Content: val, IsUser: true,
			})
			m.input.Reset()
			m.updateChat()
			ch := m.inputCh
			return m, func() tea.Msg {
				ch <- val
				return nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			chatW := m.width - sidebarWidth - 6
			chatH := m.chatHeight()
			m.chat = viewport.New(chatW, chatH)
			m.ready = true
		}
		m.input.SetWidth(m.width - 4)
		m.resizeChat()
		m.updateChat()
		return m, nil

	case ResponseMsg:
		m.messages = append(m.messages, ChatEntry{
			From: msg.From, Content: msg.Content,
		})
		m.updateChat()
		return m, listenResponse(m.outputCh)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) chatHeight() int {
	h := m.height - 8
	if h < 1 {
		h = 1
	}
	return h
}

func (m *Model) resizeChat() {
	if !m.ready {
		return
	}
	m.chat.Width = m.width - sidebarWidth - 6
	m.chat.Height = m.chatHeight()
}

func (m *Model) updateChat() {
	if !m.ready {
		return
	}
	var sb strings.Builder
	for _, entry := range m.messages {
		var name string
		if entry.IsUser {
			name = senderUser.Render(entry.From)
		} else {
			name = senderBeing.Render(entry.From)
		}
		lines := wrapText(entry.Content, m.chat.Width-4)
		sb.WriteString(name + "\n")
		for _, line := range lines {
			sb.WriteString("  " + line + "\n")
		}
		sb.WriteString("\n")
	}
	m.chat.SetContent(sb.String())
	m.chat.GotoBottom()
}

func (m Model) View() string {
	if !m.ready {
		return "loading..."
	}

	sidebar := m.renderSidebar()
	sidebarRendered := sidebarStyle.
		Width(sidebarWidth).
		Height(m.height - 4).
		Render(sidebar)

	chatRendered := chatStyle.
		Width(m.width - sidebarWidth - 6).
		Height(m.height - 4).
		Render(m.chat.View())

	top := lipgloss.JoinHorizontal(lipgloss.Top, sidebarRendered, chatRendered)

	inputRendered := inputStyle.
		Width(m.width - 4).
		Render(m.input.View())

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(" skyra v.1"),
		top,
		inputRendered,
	)
}

func (m Model) renderSidebar() string {
	var sb strings.Builder

	// world state
	sb.WriteString(titleStyle.Render("WORLD") + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", sidebarWidth-4)) + "\n")

	if m.world != nil {
		m.world.RLock()
		count := len(m.world.World)
		m.world.RUnlock()
		sb.WriteString(dimStyle.Render(fmt.Sprintf(" Entities: %d", count)) + "\n")

		m.world.RLock()
		for name, entity := range m.world.World {
			w := entity.Activation(m.world)
			sb.WriteString(fmt.Sprintf(" %s %s",
				labelStyle.Render(name),
				dimStyle.Render(fmt.Sprintf("(%.2f)", w)),
			) + "\n")
		}
		m.world.RUnlock()
	}

	// relation state
	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("RELATION") + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", sidebarWidth-4)) + "\n")

	if m.world != nil {
		m.world.RLock()
		sb.WriteString(dimStyle.Render(fmt.Sprintf(" Origin: %s", m.world.Origin)) + "\n")
		sb.WriteString(dimStyle.Render(fmt.Sprintf(" Impulse: %s", truncate(m.world.Impulse, 20))) + "\n")
		sliceCount := len(m.world.Slices)
		m.world.RUnlock()
		sb.WriteString(dimStyle.Render(fmt.Sprintf(" Slices: %d", sliceCount)) + "\n")
	}

	// slices
	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("SLICES") + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", sidebarWidth-4)) + "\n")

	if m.world != nil {
		m.world.RLock()
		for name, slice := range m.world.Slices {
			slice.Lock()
			thoughts := len(slice.TrainOfThought)
			signal := slice.Signal
			load := slice.Load
			goal := truncate(slice.Goal, 18)
			slice.Unlock()
			sb.WriteString(fmt.Sprintf(" %s\n", labelStyle.Render(name)))
			sb.WriteString(dimStyle.Render(fmt.Sprintf("   sig:%.2f load:%d thoughts:%d", signal, load, thoughts)) + "\n")
			if goal != "" {
				sb.WriteString(dimStyle.Render(fmt.Sprintf("   goal: %s", goal)) + "\n")
			}
		}
		m.world.RUnlock()
	}

	return sb.String()
}

func listenResponse(ch chan ResponseMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func wrapText(text string, width int) []string {
	if width <= 0 {
		width = 80
	}
	var lines []string
	for _, raw := range strings.Split(text, "\n") {
		if len(raw) <= width {
			lines = append(lines, raw)
			continue
		}
		words := strings.Fields(raw)
		var current string
		for _, word := range words {
			if current == "" {
				current = word
			} else if len(current)+1+len(word) <= width {
				current += " " + word
			} else {
				lines = append(lines, current)
				current = word
			}
		}
		if current != "" {
			lines = append(lines, current)
		}
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}
