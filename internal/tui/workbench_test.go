package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/workbookview"
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

func TestJournalFormatterLabelsEngineeringAndEvidenceSources(t *testing.T) {
	for module, marker := range map[string]string{"ENGINEERING": "🛠 ENG", "EVIDENCE": "🔒 EVID"} {
		event, err := journal.NewEvent("INVOCATION_COMPLETED", module, map[string]any{"tool": "vim"}, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := jsonMarshal(event)
		if line := formatJournalLine(encoded); !strings.Contains(line, marker) {
			t.Fatalf("module=%s line=%q", module, line)
		}
	}
}

func TestDiscoveriesActionLoadsScrollablePassiveProjection(t *testing.T) {
	model := NewWorkbenchModel("DASHBOARD\n", "case", nil, nil, nil)
	model.SetDiscoveryLoader(func() ([]workbookview.Discovery, error) {
		return []workbookview.Discovery{{Kind: "host", Value: "192.0.2.10", Hostname: "router.test", LastSeen: "2026-09-01T01:00:00.000Z", Occurrences: 2, LastProvenanceID: "J-0123456789abcdef0123"}}, nil
	})
	model, command := updateWorkbench(t, model, ActionMsg{Action: ViewDiscoveries})
	if model.Mode() != "discoveries" || command == nil {
		t.Fatalf("mode=%s command=%v", model.Mode(), command)
	}
	model, _ = updateWorkbench(t, model, command())
	view := model.View().Content
	if !strings.Contains(view, "192.0.2.10") || !strings.Contains(view, "router.test") || !strings.Contains(view, "seen=2") {
		t.Fatalf("view=%s", view)
	}
	model, _ = updateWorkbench(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if model.Mode() != "actions" {
		t.Fatalf("mode=%s", model.Mode())
	}
}

func TestDiscoveriesFilterNarrowsAndClearsWithoutReloading(t *testing.T) {
	loads := 0
	model := NewWorkbenchModel("", "case", nil, nil, nil)
	model.SetDiscoveryLoader(func() ([]workbookview.Discovery, error) {
		loads++
		return []workbookview.Discovery{{Kind: "host", Value: "192.0.2.10", Hostname: "router.test"}, {Kind: "port", Value: "192.0.2.10:443"}}, nil
	})
	model, command := updateWorkbench(t, model, ActionMsg{Action: ViewDiscoveries})
	model, _ = updateWorkbench(t, model, command())
	model, _ = updateWorkbench(t, model, key("/", '/'))
	for _, r := range "router" {
		model, _ = updateWorkbench(t, model, key(string(r), r))
	}
	model, _ = updateWorkbench(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if model.discoveryCount != 1 || strings.Contains(model.viewport.View(), ":443") || loads != 1 {
		t.Fatalf("count=%d loads=%d view=%s", model.discoveryCount, loads, model.viewport.View())
	}
	model, _ = updateWorkbench(t, model, key("c", 'c'))
	if model.discoveryCount != 2 || loads != 1 {
		t.Fatalf("count=%d loads=%d", model.discoveryCount, loads)
	}
}

func TestOfflineMissingImageShowsWarningAndAnyKeyRecovers(t *testing.T) {
	model := NewWorkbenchModel("DASHBOARD\n", "case", nil, nil, nil)
	model.SetScanRunner(func(action Action, target string) (string, error) {
		if action != NetworkDiscovery || target != "192.168.2.0/24" {
			t.Fatalf("action=%c target=%q", action, target)
		}
		return "localhost/local-nmap", fmt.Errorf("podman: image not known")
	})
	model, _ = updateWorkbench(t, model, ActionMsg{Action: NetworkDiscovery})
	for _, r := range "192.168.2.9/24" {
		model, _ = updateWorkbench(t, model, key(string(r), r))
	}
	model, command := updateWorkbench(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if model.Mode() != "scan-running" || command == nil {
		t.Fatalf("mode=%s command=%v", model.Mode(), command)
	}
	batch, ok := command().(tea.BatchMsg)
	if !ok {
		t.Fatalf("scan command type=%T", command())
	}
	var finished tea.Msg
	for _, command := range batch {
		if message := command(); message != nil {
			if _, ok := message.(scanFinishedMsg); ok {
				finished = message
				break
			}
		}
	}
	if finished == nil {
		t.Fatal("batch contained no scan completion")
	}
	model, _ = updateWorkbench(t, model, finished)
	view := model.View().Content
	if model.Mode() != "warning" || !strings.Contains(view, "localhost/local-nmap") || !strings.Contains(view, "Technological Sovereignty") {
		t.Fatalf("mode=%s view=%s", model.Mode(), view)
	}
	model, _ = updateWorkbench(t, model, key("x", 'x'))
	if model.Mode() != "actions" || strings.Contains(model.View().Content, "OCI Error") {
		t.Fatalf("mode=%s view=%s", model.Mode(), model.View().Content)
	}
}

func TestScanInputEscapeClearsState(t *testing.T) {
	model := NewWorkbenchModel("", "case", nil, nil, nil)
	model.SetScanRunner(func(Action, string) (string, error) { t.Fatal("runner called"); return "", nil })
	model, _ = updateWorkbench(t, model, ActionMsg{Action: PortDiscovery})
	model, _ = updateWorkbench(t, model, key("1", '1'))
	model, _ = updateWorkbench(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if model.Mode() != "actions" || len(model.input) != 0 {
		t.Fatalf("mode=%s input=%q", model.Mode(), string(model.input))
	}
}

func TestScanErrorClassificationRecognizesPodmanOfflineFailures(t *testing.T) {
	for _, message := range []string{"image not known", "Error pulling image", "no such image", "image not found"} {
		if got := classifyScanError(fmt.Errorf("podman: %s", message)); got != "missing-image" {
			t.Fatalf("message=%q got=%q", message, got)
		}
	}
	if got := classifyScanError(fmt.Errorf("permission denied")); got != "permission denied" {
		t.Fatalf("got=%q", got)
	}
}

func jsonMarshal(value any) (string, error) {
	data, err := json.Marshal(value)
	return string(data), err
}
