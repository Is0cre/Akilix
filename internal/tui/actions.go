package tui

import (
	"fmt"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Action byte

const (
	AddScope         Action = 'A'
	AddExclusion     Action = 'X'
	NetworkDiscovery Action = 'N'
	PortDiscovery    Action = 'P'
	ViewDiscoveries  Action = 'D'
	OpenLiveLog      Action = 'L'
	LeaveWorkbook    Action = 'Q'
)

type Choice struct {
	Hotkey      Action
	Label       string
	Description string
}

var workbookChoices = [...]Choice{
	{AddScope, "Add scope target", "authorize a target"},
	{AddExclusion, "Add exclusion", "deny a target"},
	{NetworkDiscovery, "Network discovery", "Nmap host discovery"},
	{PortDiscovery, "Port discovery", "Naabu port discovery"},
	{ViewDiscoveries, "View discoveries", "inspect observed hosts and ports"},
	{OpenLiveLog, "Open live log", "follow workbook activity"},
	{LeaveWorkbook, "Leave workbook", "return to shell"},
}

type ActionMsg struct{ Action Action }

type ActionsModel struct {
	choices  [len(workbookChoices)]Choice
	cursor   int
	selected Action
	prefix   string
}

func NewActionsModel(prefix string) *ActionsModel {
	return &ActionsModel{choices: workbookChoices, prefix: prefix}
}

func (m *ActionsModel) Init() tea.Cmd { return nil }

func (m *ActionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ActionMsg:
		m.selected = msg.Action
		return m, tea.Quit
	case tea.KeyPressMsg:
		code := msg.Key().Code
		switch code {
		case tea.KeyUp, 'k', 'K':
			m.cursor = (m.cursor - 1 + len(m.choices)) % len(m.choices)
			return m, nil
		case tea.KeyDown, 'j', 'J':
			m.cursor = (m.cursor + 1) % len(m.choices)
			return m, nil
		case tea.KeyEnter:
			return m, selectAction(m.choices[m.cursor].Hotkey)
		}
		switch Action(unicode.ToUpper(code)) {
		case AddScope:
			return m, onAddScope()
		case AddExclusion:
			return m, onAddExclusion()
		case NetworkDiscovery:
			return m, onNetworkDiscovery()
		case PortDiscovery:
			return m, onPortDiscovery()
		case ViewDiscoveries:
			return m, onViewDiscoveries()
		case OpenLiveLog:
			return m, onOpenLiveLog()
		case LeaveWorkbook:
			return m, onLeaveWorkbook()
		}
	}
	return m, nil
}

func (m *ActionsModel) View() tea.View {
	accent := lipgloss.Color("#87ff00")
	muted := lipgloss.Color("#7f8b8f")
	line := lipgloss.NewStyle().Foreground(lipgloss.Color("#657a3e"))
	active := lipgloss.NewStyle().Foreground(accent).Bold(true)
	inactive := lipgloss.NewStyle().Foreground(lipgloss.Color("#e8e6dd"))
	description := lipgloss.NewStyle().Foreground(muted)
	badge := lipgloss.NewStyle().Foreground(accent).Bold(true)
	focusedBadge := badge.Reverse(true)

	var b strings.Builder
	b.Grow(len(m.prefix) + 768)
	b.WriteString(m.prefix)
	b.WriteString(line.Render(" Actions " + strings.Repeat("─", 70)))
	b.WriteByte('\n')
	for index, choice := range m.choices {
		key := badge.Render(fmt.Sprintf("[%c]", choice.Hotkey))
		label := inactive.Render(fmt.Sprintf("%-21s", choice.Label))
		if index == m.cursor {
			key = focusedBadge.Render(fmt.Sprintf("[%c]", choice.Hotkey))
			label = active.Render(fmt.Sprintf("%-21s", choice.Label))
		}
		fmt.Fprintf(&b, "  %s  %s %s\n", key, label, description.Render(choice.Description))
	}
	b.WriteString(line.Render(" " + strings.Repeat("─", 78)))
	b.WriteString("\n Select action > ")
	b.WriteString(description.Render("Navigate with j/k, or press the matching hotkey"))

	return tea.NewView(b.String())
}

func (m *ActionsModel) Cursor() int             { return m.cursor }
func (m *ActionsModel) Selected() Action        { return m.selected }
func (m *ActionsModel) Choices() []Choice       { return m.choices[:] }
func (m *ActionsModel) SetPrefix(prefix string) { m.prefix = prefix }

func selectAction(action Action) tea.Cmd {
	return func() tea.Msg { return ActionMsg{Action: action} }
}

func onAddScope() tea.Cmd         { return selectAction(AddScope) }
func onAddExclusion() tea.Cmd     { return selectAction(AddExclusion) }
func onNetworkDiscovery() tea.Cmd { return selectAction(NetworkDiscovery) }
func onPortDiscovery() tea.Cmd    { return selectAction(PortDiscovery) }
func onViewDiscoveries() tea.Cmd  { return selectAction(ViewDiscoveries) }
func onOpenLiveLog() tea.Cmd      { return selectAction(OpenLiveLog) }
func onLeaveWorkbook() tea.Cmd    { return selectAction(LeaveWorkbook) }
