package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Is0cre/Akilix/internal/journal"
)

func updateWorkbench(t *testing.T, model *WorkbenchModel, message tea.Msg) (*WorkbenchModel, tea.Cmd) {
	t.Helper()
	next, command := model.Update(message)
	updated, ok := next.(*WorkbenchModel)
	if !ok {
		t.Fatalf("model type=%T", next)
	}
	return updated, command
}

func TestScopeModalValidatesAndReturnsFocusAfterDurableAdd(t *testing.T) {
	var added string
	model := NewWorkbenchModel("OLD DASHBOARD\n", "case", func(value string) error { added = value; return nil }, func() (string, error) { return "Includes 1  🎯 SCOPE ACTIVE\n", nil }, nil)
	model, _ = updateWorkbench(t, model, ActionMsg{Action: AddScope})
	if model.Mode() != "scope" {
		t.Fatalf("mode=%s", model.Mode())
	}
	for _, r := range "192.168.2.9/24" {
		model, _ = updateWorkbench(t, model, key(string(r), r))
	}
	model, command := updateWorkbench(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if command == nil {
		t.Fatal("valid scope returned no command")
	}
	model, _ = updateWorkbench(t, model, command())
	if added != "192.168.2.0/24" || model.Mode() != "actions" || !strings.Contains(model.View().Content, "Includes 1") {
		t.Fatalf("added=%q mode=%s view=%s", added, model.Mode(), model.View().Content)
	}
}

func TestScopeModalInvalidInputClearsAndRetriesWithoutLeaving(t *testing.T) {
	model := NewWorkbenchModel("", "case", func(string) error { t.Fatal("invalid input reached backend"); return nil }, func() (string, error) { return "", nil }, nil)
	model, _ = updateWorkbench(t, model, ActionMsg{Action: AddScope})
	for _, r := range "not-a-network" {
		model, _ = updateWorkbench(t, model, key(string(r), r))
	}
	model, _ = updateWorkbench(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if model.Mode() != "scope" || !strings.Contains(model.View().Content, "INVALID CIDR/IP FORMAT") || strings.Contains(model.View().Content, "not-a-network") {
		t.Fatalf("mode=%s view=%s", model.Mode(), model.View().Content)
	}
	model, _ = updateWorkbench(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if model.Mode() != "actions" {
		t.Fatalf("escape mode=%s", model.Mode())
	}
}

func TestJournalViewportCapsAtFiveHundredAndEscapeRestoresActions(t *testing.T) {
	model := NewWorkbenchModel("DASHBOARD\n", "case", nil, nil, nil)
	model.mode = modeJournal
	lines := make([]string, maxJournalLines+20)
	for i := range lines {
		event, err := journal.NewEvent("PORT_FOUND", "RECON", map[string]any{"value": fmt.Sprintf("192.0.2.1:%d", i)}, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := jsonMarshal(event)
		if err != nil {
			t.Fatal(err)
		}
		lines[i] = encoded
	}
	model, _ = updateWorkbench(t, model, journalLinesMsg{lines: lines})
	if model.JournalLineCount() != maxJournalLines {
		t.Fatalf("lines=%d", model.JournalLineCount())
	}
	model, _ = updateWorkbench(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if model.Mode() != "actions" {
		t.Fatalf("escape mode=%s", model.Mode())
	}
}

func TestJournalFormatterShowsSocketAndSemanticModule(t *testing.T) {
	event, err := journal.NewEvent("PORT_FOUND", "RECON", map[string]any{"endpoint": "192.0.2.1:53"}, time.Date(2026, 9, 1, 1, 53, 12, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := jsonMarshal(event)
	if err != nil {
		t.Fatal(err)
	}
	line := formatJournalLine(encoded)
	if !strings.Contains(line, "[01:53:12]") || !strings.Contains(line, "🟢 RECON") || !strings.Contains(line, "192.0.2.1:53") {
		t.Fatalf("line=%q", line)
	}
}

func jsonMarshal(value any) (string, error) {
	data, err := json.Marshal(value)
	return string(data), err
}
