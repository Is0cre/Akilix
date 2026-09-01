package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/scope"
)

const maxJournalLines = 500

type workbenchMode byte

const (
	modeActions workbenchMode = iota
	modeScope
	modeJournal
)

type scopeAddedMsg struct {
	value string
	err   error
}
type journalLinesMsg struct {
	lines []string
	err   error
}
type journalTickMsg struct{}

type WorkbenchModel struct {
	actions      *ActionsModel
	mode         workbenchMode
	input        []rune
	inputError   string
	addScope     func(string) error
	refresh      func() (string, error)
	tail         *journal.Tail
	viewport     viewport.Model
	journalLines []string
	journalError string
	workbookName string
	selected     Action
}

func NewWorkbenchModel(prefix, workbookName string, addScope func(string) error, refresh func() (string, error), tail *journal.Tail) *WorkbenchModel {
	view := viewport.New(viewport.WithWidth(78), viewport.WithHeight(12))
	view.SoftWrap = false
	return &WorkbenchModel{actions: NewActionsModel(prefix), addScope: addScope, refresh: refresh, tail: tail, viewport: view, workbookName: workbookName, journalLines: []string{}}
}

func (m *WorkbenchModel) Init() tea.Cmd { return nil }

func (m *WorkbenchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeScope:
		return m.updateScope(msg)
	case modeJournal:
		return m.updateJournal(msg)
	}
	if selection, ok := msg.(ActionMsg); ok {
		switch selection.Action {
		case AddScope:
			m.mode, m.input, m.inputError = modeScope, m.input[:0], ""
			return m, nil
		case OpenLiveLog:
			m.mode, m.journalError = modeJournal, ""
			return m, tea.Batch(pollJournal(m.tail), journalTick())
		default:
			m.selected = selection.Action
			return m, tea.Quit
		}
	}
	_, cmd := m.actions.Update(msg)
	return m, cmd
}

func (m *WorkbenchModel) updateScope(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case scopeAddedMsg:
		if message.err != nil {
			m.input, m.inputError = m.input[:0], message.err.Error()
			return m, nil
		}
		prefix, err := m.refresh()
		if err != nil {
			m.input, m.inputError = m.input[:0], err.Error()
			return m, nil
		}
		m.actions.SetPrefix(prefix)
		m.mode, m.input, m.inputError = modeActions, m.input[:0], ""
		return m, nil
	case tea.KeyPressMsg:
		key := message.Key()
		switch key.Code {
		case tea.KeyEscape:
			m.mode, m.input, m.inputError = modeActions, m.input[:0], ""
			return m, nil
		case tea.KeyBackspace:
			if len(m.input) != 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil
		case tea.KeyEnter:
			normalized, err := scope.NormalizeIPTarget(string(m.input))
			if err != nil {
				m.input, m.inputError = m.input[:0], "INVALID CIDR/IP FORMAT. Retrying..."
				return m, nil
			}
			return m, func() tea.Msg { return scopeAddedMsg{value: normalized, err: m.addScope(normalized)} }
		}
		if key.Text != "" && utf8.ValidString(key.Text) && len(m.input)+utf8.RuneCountInString(key.Text) <= 128 {
			m.input = append(m.input, []rune(key.Text)...)
			m.inputError = ""
		}
	}
	return m, nil
}

func (m *WorkbenchModel) updateJournal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case journalTickMsg:
		return m, tea.Batch(pollJournal(m.tail), journalTick())
	case journalLinesMsg:
		if message.err != nil {
			m.journalError = message.err.Error()
			return m, nil
		}
		for _, line := range message.lines {
			m.journalLines = append(m.journalLines, formatJournalLine(line))
		}
		if len(m.journalLines) > maxJournalLines {
			m.journalLines = append([]string(nil), m.journalLines[len(m.journalLines)-maxJournalLines:]...)
		}
		m.viewport.SetContentLines(m.journalLines)
		m.viewport.GotoBottom()
		return m, nil
	case tea.KeyPressMsg:
		if message.Key().Code == tea.KeyEscape || message.String() == "q" {
			m.mode, m.journalError = modeActions, ""
			return m, nil
		}
		updated, cmd := m.viewport.Update(message)
		m.viewport = updated
		return m, cmd
	}
	return m, nil
}

func pollJournal(tail *journal.Tail) tea.Cmd {
	return func() tea.Msg {
		if tail == nil {
			return journalLinesMsg{err: fmt.Errorf("journal unavailable")}
		}
		lines, err := tail.Poll()
		return journalLinesMsg{lines: lines, err: err}
	}
}

func journalTick() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg { return journalTickMsg{} })
}

func (m *WorkbenchModel) View() tea.View {
	switch m.mode {
	case modeScope:
		return tea.NewView(m.actions.prefix + renderScopeInput(string(m.input), m.inputError))
	case modeJournal:
		return tea.NewView(m.actions.prefix + renderJournal(m.workbookName, m.viewport.View(), m.journalError))
	default:
		return m.actions.View()
	}
}

func renderScopeInput(value, problem string) string {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#87ff00"))
	amber := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f00")).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#7f8b8f"))
	line := accent.Render(" Scope Management " + strings.Repeat("─", 59))
	field := accent.Render("[ Enter Allowed CIDR / IP Target ]: ") + value + muted.Render(strings.Repeat("_", max(1, 38-len([]rune(value)))))
	view := line + "\n" + field + "\n" + accent.Render(strings.Repeat("─", 78))
	if problem != "" {
		view += "\n" + amber.Render("[!] "+problem)
	}
	return view + "\n" + muted.Render("(Press ESC to abort, Enter to append and durably journal target scope)")
}

func renderJournal(name, content, problem string) string {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#87ff00"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#7f8b8f"))
	view := accent.Render(" Workbook Journal Stream: "+name+" "+strings.Repeat("─", max(1, 51-len([]rune(name))))) + "\n" + content
	if problem != "" {
		view += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f00")).Render("[!] "+problem)
	}
	return view + "\n" + muted.Render("(Navigate via j/k/PgUp/PgDn | ESC or q returns to Workbook Operations)")
}

func formatJournalLine(line string) string {
	var event journal.Event
	if json.Unmarshal([]byte(line), &event) != nil {
		return "[invalid journal record]"
	}
	timestamp := event.Timestamp
	if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", timestamp); err == nil {
		timestamp = parsed.Format("15:04:05")
	}
	value := event.Event
	for _, key := range []string{"value", "endpoint", "device", "tool"} {
		if item, ok := event.Payload[key].(string); ok && item != "" {
			value += ": " + item
			break
		}
	}
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#585858"))
	module := lipgloss.NewStyle().Foreground(lipgloss.Color("#8ab4f8"))
	message := lipgloss.NewStyle().Foreground(lipgloss.Color("#e8e6dd"))
	if event.Module == "SCOPE" || strings.Contains(event.Event, "FOUND") || strings.Contains(event.Event, "COMPLETED") {
		module = lipgloss.NewStyle().Foreground(lipgloss.Color("#87ff00")).Bold(true)
		message = lipgloss.NewStyle().Foreground(lipgloss.Color("#87ff00"))
	}
	if strings.Contains(event.Event, "FAILED") || strings.Contains(event.Event, "VIOLATION") || strings.Contains(event.Event, "OUT_OF_SCOPE") {
		module = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f00")).Bold(true)
		message = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f00"))
	}
	moduleName := event.Module
	switch event.Module {
	case "CORE":
		moduleName = "⚙ CORE"
	case "HARDWARE":
		moduleName = "🔌 HW"
	case "SCOPE":
		moduleName = "🎯 SCOPE"
	case "OCI":
		moduleName = "📦 OCI"
	case "RECON":
		moduleName = "⚡ RECON"
		if event.Event == "PORT_FOUND" {
			moduleName = "🟢 RECON"
		}
	}
	return muted.Render("["+timestamp+"]") + " " + module.Render(fmt.Sprintf("[%-8s]", moduleName)) + " " + message.Render(value) + "  " + muted.Render(event.ProvenanceID)
}

func (m *WorkbenchModel) Selected() Action { return m.selected }
func (m *WorkbenchModel) Mode() string {
	switch m.mode {
	case modeScope:
		return "scope"
	case modeJournal:
		return "journal"
	default:
		return "actions"
	}
}
func (m *WorkbenchModel) JournalLineCount() int { return len(m.journalLines) }
