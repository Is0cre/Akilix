package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func key(text string, code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: code})
}

func updateActions(t *testing.T, model *ActionsModel, msg tea.Msg) (*ActionsModel, tea.Cmd) {
	t.Helper()
	next, cmd := model.Update(msg)
	updated, ok := next.(*ActionsModel)
	if !ok {
		t.Fatalf("updated model has type %T", next)
	}
	return updated, cmd
}

func TestActionsNavigationWraps(t *testing.T) {
	model := NewActionsModel("")
	model, _ = updateActions(t, model, key("k", 'k'))
	if model.Cursor() != len(model.Choices())-1 {
		t.Fatalf("up wrap cursor=%d", model.Cursor())
	}
	model, _ = updateActions(t, model, key("j", 'j'))
	if model.Cursor() != 0 {
		t.Fatalf("down wrap cursor=%d", model.Cursor())
	}
	model, _ = updateActions(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	if model.Cursor() != 1 {
		t.Fatalf("down cursor=%d", model.Cursor())
	}
	model, _ = updateActions(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if model.Cursor() != 0 {
		t.Fatalf("up cursor=%d", model.Cursor())
	}
}

func TestActionsEnterAndCaseInsensitiveHotkeys(t *testing.T) {
	for _, test := range []struct {
		name string
		msg  tea.KeyPressMsg
		want Action
	}{
		{"enter", tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}), AddScope},
		{"scope lowercase", key("a", 'a'), AddScope},
		{"exclusion uppercase", key("X", 'x'), AddExclusion},
		{"network lowercase", key("n", 'n'), NetworkDiscovery},
		{"port uppercase", key("P", 'p'), PortDiscovery},
		{"discoveries lowercase", key("d", 'd'), ViewDiscoveries},
		{"hardware lowercase", key("h", 'h'), HardwareInventory},
		{"log lowercase", key("l", 'l'), OpenLiveLog},
		{"leave uppercase", key("Q", 'q'), LeaveWorkbook},
	} {
		t.Run(test.name, func(t *testing.T) {
			model, cmd := updateActions(t, NewActionsModel(""), test.msg)
			if cmd == nil {
				t.Fatal("selection returned no command")
			}
			message, ok := cmd().(ActionMsg)
			if !ok || message.Action != test.want {
				t.Fatalf("message=%#v, want %q", message, test.want)
			}
			model, quit := updateActions(t, model, message)
			if model.Selected() != test.want || quit == nil {
				t.Fatalf("selected=%q quit=%v", model.Selected(), quit != nil)
			}
		})
	}
}

func TestActionsViewContainsAlignedSemanticRows(t *testing.T) {
	view := NewActionsModel("WORKBOOK\n").View().Content
	for _, want := range []string{
		"Actions ", "[A]", "Add scope target", "authorize a target",
		"[D]", "View discoveries", "[H]", "Hardware inventory", "[Q]", "Leave workbook", "Navigate with j/k",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q: %q", want, view)
		}
	}
}

func BenchmarkActionsNavigation(b *testing.B) {
	model := NewActionsModel("")
	message := key("j", 'j')
	b.ReportAllocs()
	for range b.N {
		next, _ := model.Update(message)
		model = next.(*ActionsModel)
	}
}
