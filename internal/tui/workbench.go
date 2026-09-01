package tui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/scope"
	"github.com/Is0cre/Akilix/internal/workbookview"
)

const maxJournalLines = 500
const maxDiscoveryLines = 5000

type workbenchMode byte

const (
	modeActions workbenchMode = iota
	modeScope
	modeJournal
	modeDiscoveries
	modeScanInput
	modeScanRunning
	modeWarning
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
type scanFinishedMsg struct {
	image string
	err   error
}
type scanTickMsg struct{}
type discoveriesLoadedMsg struct {
	items []workbookview.Discovery
	err   error
}
type ScanRunner func(context.Context, Action, string) (string, error)
type DiscoveryLoader func() ([]workbookview.Discovery, error)

type WorkbenchModel struct {
	actions          *ActionsModel
	mode             workbenchMode
	input            []rune
	inputError       string
	addScope         func(string) error
	refresh          func() (string, error)
	tail             *journal.Tail
	viewport         viewport.Model
	journalLines     []string
	journalError     string
	workbookName     string
	selected         Action
	scanAction       Action
	scanRunner       ScanRunner
	discoveryLoader  DiscoveryLoader
	discoveryCount   int
	discoveryItems   []workbookview.Discovery
	discoveryFilter  []rune
	discoveryEditing bool
	warningImage     string
	warningText      string
	scanFrame        int
	scanCancel       context.CancelFunc
	scanCancelling   bool
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
	case modeDiscoveries:
		return m.updateDiscoveries(msg)
	case modeScanInput:
		return m.updateScanInput(msg)
	case modeScanRunning:
		if key, ok := msg.(tea.KeyPressMsg); ok && key.Key().Code == tea.KeyEscape {
			if m.scanCancel != nil {
				m.scanCancelling = true
				m.scanCancel()
			}
			return m, nil
		}
		if _, ok := msg.(scanTickMsg); ok {
			m.scanFrame = (m.scanFrame + 1) % 8
			return m, scanTick()
		}
		if finished, ok := msg.(scanFinishedMsg); ok {
			if m.scanCancel != nil {
				m.scanCancel()
			}
			m.scanCancel = nil
			m.input = m.input[:0]
			if errors.Is(finished.err, context.Canceled) {
				m.mode, m.scanCancelling = modeActions, false
				if m.refresh != nil {
					if prefix, err := m.refresh(); err == nil {
						m.actions.SetPrefix(prefix)
					}
				}
				return m, nil
			}
			if finished.err != nil {
				m.mode, m.warningImage, m.warningText, m.scanCancelling = modeWarning, finished.image, classifyScanError(finished.err), false
				return m, nil
			}
			m.mode = modeActions
			if m.refresh != nil {
				if prefix, err := m.refresh(); err == nil {
					m.actions.SetPrefix(prefix)
				}
			}
		}
		return m, nil
	case modeWarning:
		if _, ok := msg.(tea.KeyPressMsg); ok {
			m.mode, m.input, m.inputError, m.warningImage, m.warningText = modeActions, m.input[:0], "", "", ""
		}
		return m, nil
	}
	if selection, ok := msg.(ActionMsg); ok {
		switch selection.Action {
		case AddScope:
			m.mode, m.input, m.inputError = modeScope, m.input[:0], ""
			return m, nil
		case OpenLiveLog:
			m.mode, m.journalError = modeJournal, ""
			return m, tea.Batch(pollJournal(m.tail), journalTick())
		case ViewDiscoveries:
			m.mode, m.journalError = modeDiscoveries, ""
			m.viewport.SetContent("")
			loader := m.discoveryLoader
			return m, func() tea.Msg {
				if loader == nil {
					return discoveriesLoadedMsg{err: fmt.Errorf("discovery inventory unavailable")}
				}
				items, err := loader()
				return discoveriesLoadedMsg{items: items, err: err}
			}
		case NetworkDiscovery, PortDiscovery:
			if m.scanRunner != nil {
				m.mode, m.scanAction, m.input, m.inputError = modeScanInput, selection.Action, m.input[:0], ""
				return m, nil
			}
		default:
			m.selected = selection.Action
			return m, tea.Quit
		}
	}
	_, cmd := m.actions.Update(msg)
	return m, cmd
}

func (m *WorkbenchModel) updateDiscoveries(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case discoveriesLoadedMsg:
		if message.err != nil {
			m.journalError = message.err.Error()
			return m, nil
		}
		m.discoveryItems = append(m.discoveryItems[:0], message.items...)
		m.applyDiscoveryFilter()
		return m, nil
	case tea.KeyPressMsg:
		if m.discoveryEditing {
			switch message.Key().Code {
			case tea.KeyEscape:
				m.discoveryEditing = false
				return m, nil
			case tea.KeyEnter:
				m.discoveryEditing = false
				m.applyDiscoveryFilter()
				return m, nil
			case tea.KeyBackspace:
				if len(m.discoveryFilter) > 0 {
					m.discoveryFilter = m.discoveryFilter[:len(m.discoveryFilter)-1]
				}
				m.applyDiscoveryFilter()
				return m, nil
			}
			if message.Key().Text != "" && len(m.discoveryFilter)+utf8.RuneCountInString(message.Key().Text) <= 128 {
				m.discoveryFilter = append(m.discoveryFilter, []rune(message.Key().Text)...)
				m.applyDiscoveryFilter()
			}
			return m, nil
		}
		if message.String() == "/" {
			m.discoveryEditing = true
			return m, nil
		}
		if message.String() == "c" {
			m.discoveryFilter = m.discoveryFilter[:0]
			m.applyDiscoveryFilter()
			return m, nil
		}
		if message.Key().Code == tea.KeyEscape || message.String() == "q" {
			m.mode, m.journalError, m.discoveryCount = modeActions, "", 0
			m.discoveryItems, m.discoveryFilter, m.discoveryEditing = m.discoveryItems[:0], m.discoveryFilter[:0], false
			return m, nil
		}
		updated, cmd := m.viewport.Update(message)
		m.viewport = updated
		return m, cmd
	}
	return m, nil
}

func (m *WorkbenchModel) applyDiscoveryFilter() {
	query := strings.ToLower(strings.TrimSpace(string(m.discoveryFilter)))
	items := make([]workbookview.Discovery, 0, len(m.discoveryItems))
	for _, item := range m.discoveryItems {
		searchable := strings.ToLower(item.Kind + " " + item.Value + " " + item.Hostname + " " + item.LastInvocationID + " " + item.LastProvenanceID)
		if query == "" || strings.Contains(searchable, query) {
			items = append(items, item)
		}
	}
	m.discoveryCount = len(items)
	m.journalError = ""
	if len(items) > maxDiscoveryLines {
		items = items[:maxDiscoveryLines]
		m.journalError = fmt.Sprintf("showing first %d of %d matching observations", len(items), m.discoveryCount)
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, formatDiscovery(item))
	}
	if len(lines) == 0 {
		lines = append(lines, "No host or port observations recorded yet.")
	}
	m.viewport.SetContentLines(lines)
	m.viewport.GotoTop()
}

func (m *WorkbenchModel) updateScanInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch key.Key().Code {
	case tea.KeyEscape:
		m.mode, m.input, m.inputError = modeActions, m.input[:0], ""
		return m, nil
	case tea.KeyBackspace:
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
	case tea.KeyEnter:
		target, err := scope.NormalizeIPTarget(string(m.input))
		if err != nil {
			m.input, m.inputError = m.input[:0], "INVALID CIDR/IP FORMAT. Retrying..."
			return m, nil
		}
		action, runner := m.scanAction, m.scanRunner
		image := "localhost/local-nmap"
		if action == PortDiscovery {
			image = "localhost/local-naabu"
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		m.mode, m.inputError, m.scanCancel, m.scanCancelling = modeScanRunning, "", cancel, false
		return m, tea.Batch(func() tea.Msg {
			_, err := runner(ctx, action, target)
			if ctx.Err() != nil {
				err = ctx.Err()
			}
			return scanFinishedMsg{image: image, err: err}
		}, scanTick())
	}
	if key.Key().Text != "" && len(m.input)+utf8.RuneCountInString(key.Key().Text) <= 128 {
		m.input = append(m.input, []rune(key.Key().Text)...)
		m.inputError = ""
	}
	return m, nil
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
	case modeDiscoveries:
		return tea.NewView(m.actions.prefix + renderDiscoveries(m.workbookName, m.discoveryCount, string(m.discoveryFilter), m.discoveryEditing, m.viewport.View(), m.journalError))
	case modeScanInput:
		return tea.NewView(m.actions.prefix + renderScanInput(m.scanAction, string(m.input), m.inputError))
	case modeScanRunning:
		trail := []string{"1", "0", "1", "1", "0", "0", "1", "0"}
		state := "Running scoped discovery"
		if m.scanCancelling {
			state = "Cancelling scoped discovery"
		}
		return tea.NewView(m.actions.prefix + lipgloss.NewStyle().Foreground(lipgloss.Color("#87ff00")).Render(state+"  🦎 "+strings.Join(trail[:m.scanFrame+1], " ")) + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#7f8b8f")).Render("(ESC cancels and records the interrupted invocation)"))
	case modeWarning:
		return tea.NewView(m.actions.prefix + renderExecutionWarning(m.warningImage, m.warningText))
	default:
		return m.actions.View()
	}
}

func scanTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg { return scanTickMsg{} })
}

func renderScanInput(action Action, value, problem string) string {
	name := "Network discovery"
	if action == PortDiscovery {
		name = "Port discovery"
	}
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#87ff00"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#7f8b8f"))
	amber := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f00")).Bold(true)
	view := accent.Render(" "+name+" "+strings.Repeat("─", max(1, 76-len(name)))) + "\n" + accent.Render("[ Enter scoped CIDR / IP Target ]: ") + value + muted.Render(strings.Repeat("_", max(1, 40-len([]rune(value)))))
	if problem != "" {
		view += "\n" + amber.Render("[!] "+problem)
	}
	hint := "Enter explicitly starts managed native Nmap"
	return view + "\n" + muted.Render("(ESC aborts | "+hint+")")
}

func classifyScanError(err error) string {
	value := strings.ToLower(err.Error())
	if strings.Contains(value, "image not known") || strings.Contains(value, "error pulling image") || strings.Contains(value, "no such image") || strings.Contains(value, "image not found") {
		return "missing-image"
	}
	return err.Error()
}

func renderExecutionWarning(image, problem string) string {
	amber := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f00")).Bold(true)
	text := lipgloss.NewStyle().Foreground(lipgloss.Color("#e8e6dd"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#7f8b8f"))
	message := "[Execution Error]: " + problem
	if problem == "missing-image" {
		message = "[📦 OCI Error]: Target tool image '" + image + "' is missing from local storage."
		return amber.Render("⚠ SYSTEM EXECUTION ERROR "+strings.Repeat("─", 52)) + "\n" + text.Render(message) + "\n" + amber.Render("[💡 Invariant]: Technological Sovereignty requires offline image availability.") + "\n\n" + text.Render("Please explicitly build or import the profile image before retrying.") + "\n" + amber.Render(strings.Repeat("─", 78)) + "\n" + muted.Render("(Press any key to clear this warning and return to operations menu)")
	}
	return amber.Render("⚠ SYSTEM EXECUTION ERROR "+strings.Repeat("─", 52)) + "\n" + text.Render(message) + "\n" + amber.Render(strings.Repeat("─", 78)) + "\n" + muted.Render("(Press any key to clear this warning and return to operations menu)")
}

func renderDiscoveries(name string, count int, filter string, editing bool, content, problem string) string {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#87ff00"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#7f8b8f"))
	view := accent.Render(fmt.Sprintf(" Workbook Discoveries: %s · %d records ", name, count)+strings.Repeat("─", max(1, 45-len([]rune(name))))) + "\n" + content
	if filter != "" || editing {
		cursor := ""
		if editing {
			cursor = "_"
		}
		view += "\n" + accent.Render("Filter: ") + filter + cursor
	}
	if problem != "" {
		view += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f00")).Render("[!] "+problem)
	}
	return view + "\n" + muted.Render("(j/k/PgUp/PgDn navigate | / filter | c clear | ESC or q return)")
}

func formatDiscovery(item workbookview.Discovery) string {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#87ff00")).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#585858"))
	kind := strings.ToUpper(item.Kind)
	value := item.Value
	if item.Hostname != "" {
		value += "  " + muted.Render(item.Hostname)
	}
	return accent.Render(fmt.Sprintf("[%-4s]", kind)) + " " + value + muted.Render(fmt.Sprintf("  seen=%d  last=%s  %s", item.Occurrences, item.LastSeen, item.LastProvenanceID))
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
	for _, key := range []string{"value", "endpoint", "address", "device", "tool"} {
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
		if event.Event == "PORT_FOUND" || event.Event == "HOST_DISCOVERED" {
			moduleName = "🟢 RECON"
		}
	case "ENGINEERING":
		moduleName = "🛠 ENG"
	case "EVIDENCE":
		moduleName = "🔒 EVID"
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
	case modeDiscoveries:
		return "discoveries"
	case modeScanInput:
		return "scan-input"
	case modeScanRunning:
		return "scan-running"
	case modeWarning:
		return "warning"
	default:
		return "actions"
	}
}
func (m *WorkbenchModel) JournalLineCount() int                     { return len(m.journalLines) }
func (m *WorkbenchModel) SetScanRunner(runner ScanRunner)           { m.scanRunner = runner }
func (m *WorkbenchModel) SetDiscoveryLoader(loader DiscoveryLoader) { m.discoveryLoader = loader }
